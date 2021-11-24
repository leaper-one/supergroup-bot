package models

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
)

const liquidity_mining_users_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_mining_users (
	mining_id VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	created_at timestamp NOT NULL DEFAULT NOW()
);
`

type LiquidityMiningUser struct {
	MiningID  string    `json:"mining_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateLiquidityMiningUser(ctx context.Context, m *LiquidityMiningUser) error {
	query := durable.InsertQuery("liquidity_mining_users", "mining_id, user_id")
	_, err := session.Database(ctx).Exec(ctx, query, m.MiningID, m.UserID)
	return err
}

func GetLiquidityMiningUsersByID(ctx context.Context, id string) ([]*User, error) {
	m := make([]*User, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT user_id, access_token FROM users WHERE user_id IN (
	SELECT user_id FROM liquidity_mining_users WHERE mining_id=$1
);
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.UserID, &u.AccessToken); err != nil {
				return err
			}
			m = append(m, &u)
		}
		return nil
	}, id)
	return m, err
}
