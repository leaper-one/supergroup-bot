package models

import (
	"context"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"time"
)

const client_block_user_DDL = `
CREATE TABLE IF NOT EXISTS client_block_user (
  client_id           VARCHAR(36) NOT NULL,
  user_id             VARCHAR(36) NOT NULL,
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  PRIMARY KEY (client_id,user_id)
);
`

const block_user_DDL = `
CREATE TABLE IF NOT EXISTS block_user (
  user_id             VARCHAR(36) NOT NULL PRIMARY KEY,
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type ClientBlockUser struct {
	ClientID  string
	UserID    string
	CreatedAt time.Time
}

type BlockUser struct {
	UserID    string
	CreatedAt time.Time
}

var cacheBlockClientUserIDMap = make(map[string]map[string]bool)

func checkIsBlockUser(ctx context.Context, clientID, userID string) bool {
	if cacheBlockClientUserIDMap[clientID] == nil {
		blockUsers := make(map[string]bool)
		if err := session.Database(ctx).ConnQuery(ctx, `SELECT user_id FROM block_user`, func(rows pgx.Rows) error {
			for rows.Next() {
				var u string
				if err := rows.Scan(&u); err != nil {
					return err
				}
				blockUsers[u] = true
			}
			return nil
		}); err != nil {
			return false
		}
		if err := session.Database(ctx).ConnQuery(ctx, `SELECT user_id FROM client_block_user WHERE client_id=$1`, func(rows pgx.Rows) error {
			for rows.Next() {
				var u string
				if err := rows.Scan(&u); err != nil {
					return err
				}
				blockUsers[u] = true
			}
			return nil
		}, clientID); err != nil {
			return false
		}
		cacheBlockClientUserIDMap[clientID] = blockUsers
	}
	return cacheBlockClientUserIDMap[clientID][userID]
}
