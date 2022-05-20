// 管理员操作
package models

import (
	"context"
	"errors"
	"fmt"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
)

// 检查是否是管理员
func checkIsAdmin(ctx context.Context, clientID, userID string) bool {
	if checkIsOwner(ctx, clientID, userID) {
		return true
	}
	var status int
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT status FROM client_users WHERE client_id=$1 AND user_id=$2
`, clientID, userID).Scan(&status); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			session.Logger(ctx).Println(err)
		}
		return false
	}
	return status == ClientUserStatusAdmin
}

func checkIsSuperManager(userID string) bool {
	for _, v := range config.Config.SuperManager {
		if v == userID {
			return true
		}
	}
	return false
}

func checkIsOwner(ctx context.Context, clientID, userID string) bool {
	c, err := GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		session.Logger(ctx).Println(userID, clientID)
		return false
	}
	return c.OwnerID == userID
}

const (
	ClientConversationStatusNormal    = "0"
	ClientConversationStatusMute      = "1"
	ClientConversationStatusAudioLive = "2"
)

func UpdateClientConversationStatus(ctx context.Context, u *ClientUser, status string) error {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	muteClientOperation(status != ClientConversationStatusNormal, u.ClientID)
	return nil
}

func getClientConversationStatus(ctx context.Context, clientID string) string {
	status, err := session.Redis(ctx).SyncGet(ctx, fmt.Sprintf("client-conversation-%s", clientID)).Result()
	if err != nil || status == "" {
		setClientConversationStatusByIDAndStatus(ctx, clientID, ClientConversationStatusNormal)
		return ClientConversationStatusNormal
	}
	return status
}

func setClientConversationStatusByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, fmt.Sprintf("client-conversation-%s", clientID), status, -1)
}

const (
	ClientNewMemberNoticeOn  = "1"
	ClientNewMemberNoticeOff = "0"

	ClientProxyStatusOn  = "1"
	ClientProxyStatusOff = "0"
)

func getClientNewMemberNotice(ctx context.Context, clientID string) string {
	status, err := session.Redis(ctx).SyncGet(ctx, fmt.Sprintf("client-new-member-%s", clientID)).Result()
	if err != nil || status == "" {
		setClientNewMemberNoticeByIDAndStatus(ctx, clientID, ClientNewMemberNoticeOn)
		return ClientNewMemberNoticeOn
	}
	return status
}
func GetClientProxy(ctx context.Context, clientID string) string {
	status, err := session.Redis(ctx).SyncGet(ctx, fmt.Sprintf("client-proxy-%s", clientID)).Result()
	if err != nil || status == "" {
		setClientProxyStatusByIDAndStatus(ctx, clientID, ClientProxyStatusOff)
		return ClientProxyStatusOff
	}
	return status
}

func setClientNewMemberNoticeByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, fmt.Sprintf("client-new-member-%s", clientID), status, -1)
}

func setClientProxyStatusByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, fmt.Sprintf("client-proxy-%s", clientID), status, -1)
}

type ClientAdvanceSetting struct {
	ConversationStatus string `json:"conversation_status"`
	NewMemberNotice    string `json:"new_member_notice"`
	ProxyStatus        string `json:"proxy_status"`
}

func GetClientAdvanceSetting(ctx context.Context, u *ClientUser) (*ClientAdvanceSetting, error) {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	var sr ClientAdvanceSetting
	sr.ConversationStatus = getClientConversationStatus(ctx, u.ClientID)
	sr.NewMemberNotice = getClientNewMemberNotice(ctx, u.ClientID)
	sr.ProxyStatus = GetClientProxy(ctx, u.ClientID)
	return &sr, nil
}

func UpdateClientAdvanceSetting(ctx context.Context, u *ClientUser, sr ClientAdvanceSetting) error {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if sr.ConversationStatus == "0" || sr.ConversationStatus == "1" {
		return UpdateClientConversationStatus(ctx, u, sr.ConversationStatus)
	}
	if sr.NewMemberNotice == "0" || sr.NewMemberNotice == "1" {
		return setClientNewMemberNoticeByIDAndStatus(ctx, u.ClientID, sr.NewMemberNotice)
	}
	if sr.ProxyStatus == "0" || sr.ProxyStatus == "1" {
		return setClientProxyStatusByIDAndStatus(ctx, u.ClientID, sr.ProxyStatus)
	}
	return session.BadRequestError(ctx)
}
