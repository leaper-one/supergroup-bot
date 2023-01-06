package common

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
	created_at timestamp NOT NULL DEFAULT NOW(),
	PRIMARY KEY (mining_id, user_id)
);
`

type LiquidityMiningUser struct {
	MiningID  string    `json:"mining_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func CreateLiquidityMiningUser(ctx context.Context, m *LiquidityMiningUser) error {
	query := durable.InsertQueryOrUpdate("liquidity_mining_users", "mining_id, user_id", "")
	_, err := session.DB(ctx).Exec(ctx, query, m.MiningID, m.UserID)
	return err
}

func GetLiquidityMiningUsersByID(ctx context.Context, clientID, miningID string) ([]*ClientUser, error) {
	m := make([]*ClientUser, 0)
	err := session.DB(ctx).ConnQuery(ctx, `
SELECT user_id, client_id, access_token, authorization_id, scope, private_key, ed25519 FROM client_users WHERE user_id IN (
	SELECT user_id FROM liquidity_mining_users WHERE mining_id=$1
) AND client_id=$2;
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var u ClientUser
			if err := rows.Scan(&u.UserID, &u.ClientID, &u.AccessToken, &u.AuthorizationID, &u.Scope, &u.PrivateKey, &u.Ed25519); err != nil {
				return err
			}
			m = append(m, &u)
		}
		return nil
	}, miningID, clientID)
	return m, err
}
