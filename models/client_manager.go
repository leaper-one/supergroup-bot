// 管理员操作
package models

import (
	"context"
	"errors"

	"github.com/MixinNetwork/supergroup/durable"
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

func checkIsOwner(ctx context.Context, clientID, userID string) bool {
	var ownerID string
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT owner_id FROM client WHERE client_id=$1
`, clientID).Scan(&ownerID); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			session.Logger(ctx).Println(err)
		}
		return false
	}
	return ownerID == userID
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
	status := session.Redis(ctx).QGet(ctx, durable.GetRedisConversationStatus(clientID))
	if status == "" {
		setClientConversationStatusByIDAndStatus(ctx, clientID, ClientConversationStatusNormal)
		return ClientConversationStatusNormal
	}
	return status
}

func setClientConversationStatusByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, durable.GetRedisConversationStatus(clientID), status)
}

const (
	ClientNewMemberNoticeOn  = "1"
	ClientNewMemberNoticeOff = "0"
)

func UpdateClientNewMemberNotice(ctx context.Context, clientID string, status string) error {
	return setClientNewMemberNoticeByIDAndStatus(ctx, clientID, status)
}
func getClientNewMemberNotice(ctx context.Context, clientID string) string {
	status := session.Redis(ctx).QGet(ctx, durable.GetRedisNewMemberNotice(clientID))
	if status == "" {
		setClientNewMemberNoticeByIDAndStatus(ctx, clientID, ClientNewMemberNoticeOn)
		return ClientNewMemberNoticeOn
	}
	return status
}

func setClientNewMemberNoticeByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, durable.GetRedisNewMemberNotice(clientID), status)
}

type ClientAdvanceSetting struct {
	ConversationStatus string `json:"conversation_status"`
	NewMemberNotice    string `json:"new_member_notice"`
}

func GetClientAdvanceSetting(ctx context.Context, u *ClientUser) (*ClientAdvanceSetting, error) {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	var sr ClientAdvanceSetting
	sr.ConversationStatus = getClientConversationStatus(ctx, u.ClientID)
	sr.NewMemberNotice = getClientNewMemberNotice(ctx, u.ClientID)
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
		return UpdateClientNewMemberNotice(ctx, u.ClientID, sr.NewMemberNotice)
	}
	return session.BadRequestError(ctx)
}
