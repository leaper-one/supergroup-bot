// 管理员操作
package models

import (
	"context"
	"errors"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
)

// 检查是否是管理员
func checkIsManager(ctx context.Context, clientID, userID string) bool {
	var status int
	if err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT status FROM client_users WHERE client_id=$1 AND user_id=$2
`, func(row pgx.Row) error {
		return row.Scan(&status)
	}, clientID, userID); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			session.Logger(ctx).Println(err)
		}
		return false
	}
	return status == ClientUserStatusManager
}
