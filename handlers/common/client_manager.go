// 管理员操作
package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"gorm.io/gorm"
)

// 检查是否是管理员
func CheckIsAdmin(ctx context.Context, clientID, userID string) bool {
	if CheckIsOwner(ctx, clientID, userID) {
		return true
	}
	var status int
	if err := session.DB(ctx).Table("client_users").Where("client_id = ? AND user_id = ?", clientID, userID).Select("status").Scan(&status).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tools.Println(err)
		}
		return false
	}
	return status == models.ClientUserStatusAdmin
}

func checkIsSuperManager(userID string) bool {
	for _, v := range config.Config.SuperManager {
		if v == userID {
			return true
		}
	}
	return false
}

func CheckIsOwner(ctx context.Context, clientID, userID string) bool {
	c, err := GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		tools.Println(userID, clientID)
		return false
	}
	return c.OwnerID == userID
}

func UpdateClientConversationStatus(ctx context.Context, u *models.ClientUser, status string) error {
	if !CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	return nil
}

func GetClientConversationStatus(ctx context.Context, clientID string) string {
	status, err := session.Redis(ctx).SyncGet(ctx, fmt.Sprintf("client-conversation-%s", clientID)).Result()
	if err != nil || status == "" {
		SetClientConversationStatusByIDAndStatus(ctx, clientID, models.ClientConversationStatusNormal)
		return models.ClientConversationStatusNormal
	}
	return status
}

func SetClientConversationStatusByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, fmt.Sprintf("client-conversation-%s", clientID), status, -1)
}

const (
	ClientNewMemberNoticeOn  = "1"
	ClientNewMemberNoticeOff = "0"

	ClientProxyStatusOn  = "1"
	ClientProxyStatusOff = "0"
)

func GetClientNewMemberNotice(ctx context.Context, clientID string) string {
	status, err := session.Redis(ctx).SyncGet(ctx, fmt.Sprintf("client-new-member-%s", clientID)).Result()
	if err != nil || status == "" {
		SetClientNewMemberNoticeByIDAndStatus(ctx, clientID, ClientNewMemberNoticeOn)
		return ClientNewMemberNoticeOn
	}
	return status
}

func GetClientProxy(ctx context.Context, clientID string) string {
	status, err := session.Redis(ctx).SyncGet(ctx, fmt.Sprintf("client-proxy-%s", clientID)).Result()
	if err != nil || status == "" {
		SetClientProxyStatusByIDAndStatus(ctx, clientID, ClientProxyStatusOff)
		return ClientProxyStatusOff
	}
	return status
}

func SetClientNewMemberNoticeByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, fmt.Sprintf("client-new-member-%s", clientID), status, -1)
}

func SetClientProxyStatusByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, fmt.Sprintf("client-proxy-%s", clientID), status, -1)
}
