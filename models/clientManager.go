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

func getClientConversationStatus(ctx context.Context, clientID string) string {
	return session.Redis(ctx).QGet(ctx, durable.GetRedisClientConversationStatus(clientID))
}

const (
	ClientConversationStatusNormal    = "0"
	ClientConversationStatusMute      = "1"
	ClientConversationStatusAudioLive = "2"
)

func setClientConversationStatusByIDAndStatus(ctx context.Context, clientID string, status string) error {
	return session.Redis(ctx).QSet(ctx, durable.GetRedisClientConversationStatus(clientID), status)
}
