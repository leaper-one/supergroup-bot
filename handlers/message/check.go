package message

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
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
	"mvdan.cc/xurls"
)

func checkIsMutedUser(user *models.ClientUser) bool {
	now := time.Now()
	if user.MutedAt.After(now) {
		duration := decimal.NewFromFloat(user.MutedAt.Sub(now).Hours())
		hour := duration.IntPart()
		minute := duration.Sub(decimal.NewFromInt(hour)).Mul(decimal.NewFromInt(60)).IntPart()
		go sendMutedMsg(user.ClientID, user.UserID, user.MutedTime, int(hour), int(minute))
		return true
	}
	return false
}

// 检查管理员的消息 是否 quote 了 留言消息，如果是的话，就在这个函数里处理 return true
func checkIsQuoteLeaveMessage(ctx context.Context, u *models.ClientUser, msg *mixin.MessageView) (bool, error) {
	if msg.QuoteMessageID == "" {
		return false, nil
	}
	dm, err := common.GetDistributeMsgByMsgIDFromRedis(ctx, msg.QuoteMessageID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if dm.Status != models.DistributeMessageStatusLeaveMessage {
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
		if err := common.MuteClientUser(ctx, u.ClientID, dm.RepresentativeID, muteTime); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	if data == "/block" {
		if err := common.BlockClientUser(ctx, u.ClientID, u.UserID, dm.RepresentativeID, false); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	// 2. 转发给其他管理员和该用户
	go handleLeaveMsg(u.ClientID, u.UserID, dm.OriginMessageID, msg)
	return true, nil
}

// 检测是否是刚刚入群5分钟内
func checkIsJustJoinGroup(u *models.ClientUser) bool {
	return u.CreatedAt.Add(time.Minute * 5).After(time.Now())
}

// 检测是否含有链接
func checkHasURLMsg(ctx context.Context, clientID string, msg *mixin.MessageView) bool {
	if msg.Category == mixin.MessageCategoryPlainImage ||
		msg.Category == "ENCRYPTED_IMAGE" {
		client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
		if err != nil {
			return false
		}
		if url, err := tools.MessageQRFilter(ctx, client.Client, msg); err == nil {
			if url != "" && !CheckUrlIsWhiteURL(ctx, clientID, url) {
				return true
			}
		} else {
			tools.Println(err)
		}
	} else if msg.Category == mixin.MessageCategoryPlainText ||
		msg.Category == "ENCRYPTED_TEXT" {
		data := string(tools.Base64Decode(msg.Data))
		// if checkHasBotID(data) {
		// 	return true
		// }
		urls := xurls.Relaxed.FindAllString(data, -1)
		for _, url := range urls {
			if !CheckUrlIsWhiteURL(ctx, clientID, url) {
				return true
			}
		}
	}
	return false
}

func CheckUrlIsWhiteURL(ctx context.Context, clientID, targetURL string) bool {
	ws, err := GetClientWhiteURLByClientID(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return false
	}
	if strings.HasPrefix(targetURL, "http") {
		targetURLObj, err := url.Parse(targetURL)
		if err != nil {
			return false
		}
		for _, w := range ws {
			if targetURLObj.Host == w {
				return true
			}
		}
	} else {
		for _, w := range ws {
			if strings.HasPrefix(targetURL, w) {
				return true
			}
		}
	}
	return false
}
func GetClientWhiteURLByClientID(ctx context.Context, clientID string) ([]string, error) {
	var result []string
	err := session.DB(ctx).Table("client_white_url").Where("client_id = ?", clientID).Pluck("white_url", &result).Error
	return result, err
}

// 检测是否达到贴纸消息的限制
func checkStickerLimit(ctx context.Context, clientID string, msg *mixin.MessageView) bool {
	count := 0
	if err := session.DB(ctx).QueryRow(ctx, `
SELECT count(1) FROM messages 
WHERE client_id=$1 AND user_id=$2 AND category=ANY($3)
AND now()-created_at<interval '5 seconds'
`, clientID, msg.UserID, []string{mixin.MessageCategoryPlainSticker, "ENCRYPTED_STICKER"}).Scan(&count); err != nil {
		tools.Println(err)
		return false
	}
	if count == 2 {
		go SendStickerLimitMsg(clientID, msg.UserID)
	}
	return count >= 5
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
func checkMsgLanguage(msg *mixin.MessageView, clientID string) bool {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != "ENCRYPTED_TEXT" {
		return false
	}
	lang := config.Config.Lang
	if lang == "zh" {
		return false
	}
	c, err := GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return false
	}
	if c.Lang == "zh" {
		return false
	}
	data := string(emojiRx.ReplaceAllString(string(tools.Base64Decode(msg.Data)), ``))
	if len(data) == 0 {
		return false
	}
	return languageRateCheck(data, lang)
}

func languageRateCheck(data, lang string) bool {
	var t *unicode.RangeTable
	switch lang {
	case "en":
		t = nil
	case "zh":
		t = new(unicode.RangeTable)
	}
	langPer := tools.LanguageCount(data, t)
	return langPer.LessThan(config.LangCheckPer)
}

var forbiddenMsgCategory = map[string]bool{
	mixin.MessageCategoryPlainAudio:     true,
	"ENCRYPTED_AUDIO":                   true,
	mixin.MessageCategoryPlainLocation:  true,
	"ENCRYPTED_LOCATION":                true,
	mixin.MessageCategoryAppButtonGroup: true,
}

// 单独检测 禁止发的消息类型 这三种消息不能发。
func checkMsgIsForbid(u *models.ClientUser, msg *mixin.MessageView) bool {
	ctx := models.Ctx
	if forbiddenMsgCategory[msg.Category] {
		// 发送禁止消息
		go common.SendForbidMsg(u.ClientID, u.UserID, msg.Category)
		return true
	}

	if msg.Category == mixin.MessageCategoryPlainContact ||
		msg.Category == "ENCRYPTED_CONTACT" {
		data := tools.Base64Decode(msg.Data)
		var c mixin.ContactMessage
		if err := json.Unmarshal(data, &c); err != nil {
			return true
		}
		contactUser, err := common.SearchUser(ctx, u.ClientID, c.UserID)
		if err != nil {
			return true
		}
		id, _ := strconv.Atoi(contactUser.IdentityNumber)
		if id < 7000000000 {
			// 联系人卡片消息
			go sendForbidMsg(u.ClientID, u.UserID, msg.Category)
			return true
		}
	}

	return false
}

// 检查消息频率
func checkMessageCountLimit(ctx context.Context, clientID, userID string, status int) bool {
	count := 0
	if err := session.DB(ctx).QueryRow(ctx, `
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
	return CheckHasClientMemberAuth(ctx, clientID, category, status)
}

var cacheSendJoinMsg = tools.NewMutex()

// 检测用户是否5分钟内发过消息
func checkIsSendJoinMsg(userID string) bool {
	t := cacheSendJoinMsg.Read(userID)
	if t == nil {
		cacheSendJoinMsg.Write(userID, time.Now())
		return false
	}
	if t.(time.Time).Add(time.Minute * 5).Before(time.Now()) {
		cacheSendJoinMsg.Write(userID, time.Now())
		return false
	}
	return true
}

// 检查 conversation 是否是会话
func checkIsContact(ctx context.Context, clientID, conversationID string) bool {
	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return false
	}
	c, err := client.ReadConversation(ctx, conversationID)
	if err != nil {
		tools.Println(err)
		return false
	}
	return c.Category == mixin.ConversationCategoryContact
}

// 检测是否能够发送红包
func checkCanNotSendLuckyCoin(ctx context.Context, clientID, data, status string) bool {
	var m mixin.AppCardMessage
	err := json.Unmarshal(tools.Base64Decode(data), &m)
	if err != nil {
		tools.Println(err)
		return true
	}
	u, err := url.Parse(m.Action)
	if err != nil {
		tools.Println(err)
		return true
	}
	query, _ := url.ParseQuery(u.RawQuery)
	if len(query["uid"]) == 0 {
		return true
	}
	uid := query["uid"][0]
	if common.CheckIsBlockUser(ctx, clientID, uid) {
		return true
	}
	user, err := common.GetClientUserByClientIDAndUserID(ctx, clientID, uid)
	if err != nil || user.UserID == "" {
		tools.Println(err, user)
		return true
	}
	if !common.CheckHasClientMemberAuth(ctx, clientID, "lucky_coin", user.Status) {
		return true
	}
	if (status == models.ClientConversationStatusMute ||
		status == models.ClientConversationStatusAudioLive) &&
		!common.CheckIsAdmin(ctx, clientID, uid) {
		return true
	}

	return false
}

// 检查 是否是 帮转/禁言/拉黑 的消息
func checkIsButtonOperation(ctx context.Context, clientID string, msg *mixin.MessageView) (bool, error) {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != "ENCRYPTED_TEXT" {
		return false, nil
	}
	data := string(tools.Base64Decode(msg.Data))
	if !strings.HasPrefix(data, "---operation") {
		return false, nil
	}
	// 确定是操作的内容了
	operationAction := strings.Split(data, ",")
	if len(operationAction) != 3 {
		return true, nil
	}
	originMsg, err := getMsgByClientIDAndMessageID(ctx, clientID, operationAction[2])
	if err != nil {
		return true, err
	}
	switch operationAction[1] {
	// 1. 帮转发
	case "forward":
		if err := common.CreateAndDistributeMessage(ctx, clientID, &mixin.MessageView{
			ConversationID: originMsg.ConversationID,
			UserID:         originMsg.UserID,
			MessageID:      originMsg.MessageID,
			Category:       originMsg.Category,
			Data:           originMsg.Data,
			CreatedAt:      msg.CreatedAt,
		}); err != nil {
			return true, err
		}
	// 2. 禁言
	case "mute":
		if err := common.MuteClientUser(ctx, clientID, originMsg.UserID, "12"); err != nil {
			tools.Println(err)
		}
	// 3. 拉黑
	case "block":
		if err := common.BlockClientUser(ctx, clientID, msg.UserID, originMsg.UserID, false); err != nil {
			tools.Println(err)
		}
	}

	return true, nil
}

func checkIsOperationMsg(ctx context.Context, u *models.ClientUser, msg *mixin.MessageView) (bool, error) {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != "ENCRYPTED_TEXT" {
		return false, nil
	}
	data := string(tools.Base64Decode(msg.Data))
	if data == "/mute open" || data == "/mute close" {
		muteStatus := data == "/mute open"
		common.MuteClientOperation(muteStatus, u.ClientID)
		return true, nil
	}
	if isOperation, err := handleUnmuteAndUnblockMsg(ctx, data, u); err != nil {
		tools.Println(err)
	} else if isOperation {
		return true, nil
	}

	return handleRecallOrMuteOrBlockOrInfoMsg(ctx, data, u.ClientID, msg)
}
func handleUnmuteAndUnblockMsg(ctx context.Context, data string, u *models.ClientUser) (bool, error) {
	operation := strings.Split(data, " ")
	if len(operation) < 2 || len(operation[1]) <= 4 {
		return false, nil
	}
	if strings.HasPrefix(data, "/unmute") {
		_u, err := common.SearchUser(ctx, u.ClientID, operation[1])
		if err != nil {
			tools.Println(err)
			return true, nil
		}
		if err := common.MuteClientUser(ctx, u.ClientID, _u.UserID, "0"); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	if strings.HasPrefix(data, "/unblock") {
		_u, err := common.SearchUser(ctx, u.ClientID, operation[1])
		if err != nil {
			tools.Println(err)
			return true, nil
		}
		if err := common.BlockClientUser(ctx, u.ClientID, u.UserID, _u.UserID, true); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	if strings.HasPrefix(data, "/blockall") {
		if checkIsSuperManager(u.UserID) {
			_u, err := common.SearchUser(ctx, u.ClientID, operation[1])
			if err != nil {
				tools.Println(err)
				return true, nil
			}
			memo := ""
			if len(operation) == 3 {
				memo = operation[2]
			}
			if err := common.AddBlockUser(ctx, u.UserID, u.ClientID, _u.UserID, memo); err != nil {
				tools.Println(err)
			}
			common.SendClientUserTextMsg(u.ClientID, u.UserID, "success", "")
		}
		return true, nil
	}
	return false, nil
}

func checkIsSuperManager(userID string) bool {
	for _, v := range config.Config.SuperManager {
		if v == userID {
			return true
		}
	}
	return false
}
