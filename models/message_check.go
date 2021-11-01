package models

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"mvdan.cc/xurls"
)

// 检查管理员的消息 是否 quote 了 留言消息，如果是的话，就在这个函数里处理 return true
func checkIsQuoteLeaveMessage(ctx context.Context, clientUser *ClientUser, msg *mixin.MessageView) (bool, error) {
	if msg.QuoteMessageID == "" {
		return false, nil
	}
	dm, err := getDistributeMessageByClientIDAndMessageID(ctx, clientUser.ClientID, msg.QuoteMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if dm.Status != DistributeMessageStatusLeaveMessage {
		return false, nil
	}
	// 确定是 quote 的留言信息了
	// 1. 看是不是 mute 和 block
	data := string(tools.Base64Decode(msg.Data))
	if strings.HasPrefix(data, "/mute") {
		muteTime := "12"
		tmp := strings.Split(data, " ")
		if len(tmp) > 1 {
			t, err := strconv.Atoi(tmp[1])
			if err == nil && t >= 0 {
				muteTime = tmp[1]
			}
		}
		if err := muteClientUser(ctx, clientUser.ClientID, dm.RepresentativeID, muteTime); err != nil {
			session.Logger(ctx).Println(err)
		}
		return true, nil
	}

	if data == "/block" {
		if err := blockClientUser(ctx, clientUser.ClientID, dm.RepresentativeID, false); err != nil {
			session.Logger(ctx).Println(err)
		}
		return true, nil
	}

	// 2. 转发给其他管理员和该用户
	go handleLeaveMsg(clientUser.ClientID, clientUser.UserID, dm.OriginMessageID, msg)
	return true, nil
}

var cacheSendJoinMsg = make(map[string]time.Time)

// 检测用户是否5分钟内发过消息
func checkIsSendJoinMsg(userID string) bool {
	if cacheSendJoinMsg[userID].IsZero() {
		cacheSendJoinMsg[userID] = time.Now()
		return false
	}
	if cacheSendJoinMsg[userID].Add(time.Minute * 5).Before(time.Now()) {
		cacheSendJoinMsg[userID] = time.Now()
		return false
	}
	return true
}

// 检测是否是刚刚入群5分钟内
func checkIsJustJoinGroup(u *ClientUser) bool {
	return u.CreatedAt.Add(time.Minute * 5).After(time.Now())
}

// 检测是否含有链接
func checkHasURLMsg(ctx context.Context, clientID string, msg *mixin.MessageView) bool {
	hasURL := false
	if msg.Category == mixin.MessageCategoryPlainImage {
		if url, err := tools.MessageQRFilter(ctx, GetMixinClientByID(ctx, clientID).Client, msg); err == nil {
			if url != "" && !CheckUrlIsWhiteURL(ctx, clientID, url) {
				hasURL = true
			}
		} else {
			session.Logger(ctx).Println(err)
		}
	} else if msg.Category == mixin.MessageCategoryPlainText {
		data := tools.Base64Decode(msg.Data)
		urls := xurls.Relaxed.FindAllString(string(data), -1)
		for _, url := range urls {
			if !CheckUrlIsWhiteURL(ctx, clientID, url) {
				hasURL = true
				break
			}
		}
	}
	return hasURL
}

// 检测是否达到贴纸消息的限制
func checkStickerLimit(ctx context.Context, clientID string, msg *mixin.MessageView) bool {
	count := 0
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT count(1) FROM messages 
WHERE client_id=$1 AND user_id=$2 AND category=$3
AND now()-created_at<interval '5 seconds'
`, clientID, msg.UserID, mixin.MessageCategoryPlainSticker).Scan(&count); err != nil {
		session.Logger(ctx).Println(err)
		return false
	}
	if count == 2 {
		go SendStickerLimitMsg(clientID, msg.UserID)
	}
	return count >= 5
}

// 检查 conversation 是否是会话
func checkIsContact(ctx context.Context, clientID, conversationID string) bool {
	c, err := GetMixinClientByID(ctx, clientID).ReadConversation(ctx, conversationID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return false
	}
	if c.Category == mixin.ConversationCategoryContact {
		return true
	}
	return false
}

// 检测是否能够发送红包
func checkCanNotSendLuckyCoin(ctx context.Context, clientID, data, status string) bool {
	var m mixin.AppCardMessage
	err := json.Unmarshal(tools.Base64Decode(data), &m)
	if err != nil {
		session.Logger(ctx).Println(err)
		return true
	}
	u, err := url.Parse(m.Action)
	if err != nil {
		session.Logger(ctx).Println(err)
		return true
	}
	query, _ := url.ParseQuery(u.RawQuery)
	if len(query["uid"]) == 0 {
		return true
	}
	uid := query["uid"][0]
	if checkIsBlockUser(ctx, clientID, uid) {
		return true
	}
	user, err := GetClientUserByClientIDAndUserID(ctx, clientID, uid)
	if err != nil || user == nil || user.UserID == "" {
		session.Logger(ctx).Println(err, user)
		return true
	}
	if !checkHasClientMemberAuth(ctx, clientID, "lucky_coin", user.Status) {
		return true
	}
	if (status == ClientConversationStatusMute ||
		status == ClientConversationStatusAudioLive) &&
		!checkIsAdmin(ctx, clientID, uid) {
		return true
	}

	return false
}

var ignoreMsgList = []string{"Hi", "你好"}

// 检测是否是忽略的消息
func checkIsIgnoreLeaveMsg(msg *mixin.MessageView) bool {
	data := string(tools.Base64Decode(msg.Data))
	for _, s := range ignoreMsgList {
		if data == s {
			return true
		}
	}
	return false
}

// 语言检测
func checkMsgLanguage(msg *mixin.MessageView) bool {
	if msg.Category != mixin.MessageCategoryPlainText {
		return false
	}
	data := string(emojiRx.ReplaceAllString(string(tools.Base64Decode(msg.Data)), ``))
	if len(data) == 0 {
		return false
	}
	lang := config.Config.Lang
	return languaueRateCheck(data, lang)
}

func languaueRateCheck(data, lang string) bool {
	if lang == "zh" {
		return false
	}
	t := new(unicode.RangeTable)
	if lang == "en" {
		t = nil
	}
	c, tc := tools.LanguageCount(data, t)
	return (decimal.NewFromInt(int64(c)).Div(decimal.NewFromInt(int64(tc)))).
		LessThan(decimal.NewFromInt(2).Div(decimal.NewFromInt(3)))
}

var forbiddenMsgCategory = map[string]bool{
	mixin.MessageCategoryPlainAudio:     true,
	mixin.MessageCategoryPlainLocation:  true,
	mixin.MessageCategoryAppButtonGroup: true,
}

// 单独检测 禁止发的消息类型 这三种消息不能发。
func checkMsgIsForbid(u *ClientUser, msg *mixin.MessageView) bool {
	if forbiddenMsgCategory[msg.Category] {
		// 发送禁止消息
		go SendForbidMsg(u.ClientID, u.UserID, msg.Category)
		return true
	}

	if msg.Category == mixin.MessageCategoryPlainContact {
		data := tools.Base64Decode(msg.Data)
		var c mixin.ContactMessage
		if err := json.Unmarshal(data, &c); err != nil {
			return true
		}
		contactUser, err := SearchUser(_ctx, c.UserID)
		if err != nil {
			return true
		}
		id, _ := strconv.Atoi(contactUser.IdentityNumber)
		if id < 7000000000 {
			// 联系人卡片消息
			go SendForbidMsg(u.ClientID, u.UserID, msg.Category)
			return true
		}
	}

	return false
}

// 检查消息频率
func checkMessageCountLimit(ctx context.Context, clientID, userID string, status int) bool {
	count := 0
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT count(1) FROM messages 
WHERE client_id=$1 
AND user_id=$2 
AND now()-created_at<interval '1 minutes'
`, clientID, userID).Scan(&count); err != nil {
		return false
	}
	limit := statusLimitMap[status]
	return count < limit
}

// 检查用户是否可以发送目标的消息类型
func checkCategory(ctx context.Context, clientID, category string, status int) bool {
	if category == mixin.MessageCategoryMessageRecall ||
		status == ClientUserStatusAdmin ||
		status == ClientUserStatusGuest {
		return true
	}
	return checkHasClientMemberAuth(ctx, clientID, category, status)
}
