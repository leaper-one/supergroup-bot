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
func checkIsManager(ctx context.Context, clientID, userID string) bool {
	var status int
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT status FROM client_users WHERE client_id=$1 AND user_id=$2
`, clientID, userID).Scan(&status); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			session.Logger(ctx).Println(err)
		}
		return false
	}
	return status == ClientUserStatusManager
}

func checkClientIsMute(ctx context.Context, clientID string) bool {
	res := session.Redis(ctx).QGet(ctx, durable.GetRedisClientMuteStatus(clientID))
	return res == "1"
}

func setClientMuteByIDAndStatus(ctx context.Context, clientID string, isOpen bool) error {
	if isOpen {
		return session.Redis(ctx).QSet(ctx, durable.GetRedisClientMuteStatus(clientID), "1")
	} else {
		return session.Redis(ctx).Del(ctx, durable.GetRedisClientMuteStatus(clientID)).Err()
	}
}
