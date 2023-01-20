package common

import (
	"context"
	"strings"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

func checkCategoryIsValid(key string) bool {
	return map[string]bool{
		mixin.MessageCategoryPlainText:    true,
		mixin.MessageCategoryPlainSticker: true,
		mixin.MessageCategoryPlainImage:   true,
		mixin.MessageCategoryPlainVideo:   true,
		mixin.MessageCategoryPlainPost:    true,
		mixin.MessageCategoryPlainData:    true,
		mixin.MessageCategoryPlainLive:    true,
		mixin.MessageCategoryPlainContact: true,
		mixin.MessageCategoryAppCard:      true,
		"PLAIN_TRANSCRIPT":                true,
		"lucky_coin":                      true,
		"url":                             true,
	}[key]
}

func CheckHasClientMemberAuth(ctx context.Context, clientID, category string, userStatus int) bool {
	if strings.HasPrefix(category, "ENCRYPTED_") {
		category = strings.Replace(category, "ENCRYPTED_", "PLAIN_", 1)
	}
	if category == mixin.MessageCategoryPlainText {
		return true
	}
	if !checkCategoryIsValid(category) {
		tools.Println(category)
		return false
	}
	if userStatus > 5 {
		userStatus = 5
	}
	var hasAuth bool

	if err := session.DB(ctx).
		Table("client_member_auth").
		Where("client_id = ? AND user_status = ?", clientID, userStatus).
		Select(category).
		Scan(&hasAuth).Error; err != nil {
		tools.Println(err)
		return false
	}

	return hasAuth
}
