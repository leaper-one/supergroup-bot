package common

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
	"unicode"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

// 检查管理员的消息 是否 quote 了 留言消息，如果是的话，就在这个函数里处理 return true

// 检测是否是刚刚入群5分钟内
func checkIsJustJoinGroup(u *models.ClientUser) bool {
	return u.CreatedAt.Add(time.Minute * 5).After(time.Now())
}

// 检测是否达到贴纸消息的限制

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
	c, err := GetClientByIDOrHost(_ctx, clientID)
	if err != nil {
		tools.Println(err)
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
	if forbiddenMsgCategory[msg.Category] {
		// 发送禁止消息
		go SendForbidMsg(u.ClientID, u.UserID, msg.Category)
		return true
	}

	if msg.Category == mixin.MessageCategoryPlainContact ||
		msg.Category == "ENCRYPTED_CONTACT" {
		data := tools.Base64Decode(msg.Data)
		var c mixin.ContactMessage
		if err := json.Unmarshal(data, &c); err != nil {
			return true
		}
		contactUser, err := SearchUser(_ctx, u.ClientID, c.UserID)
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
		status == models.ClientUserStatusAdmin ||
		status == models.ClientUserStatusGuest {
		return true
	}
	return CheckHasClientMemberAuth(ctx, clientID, category, status)
}
