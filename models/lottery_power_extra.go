package models

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const power_extra_DDL = `
CREATE TABLE IF NOT EXISTS power_extra (
	client_id 	VARCHAR(36) NOT NULL PRIMARY KEY,
	description VARCHAR DEFAULT '',
	multiplier 	VARCHAR DEFAULT '2',
	start_at 		DATE DEFAULT '1970-01-01',
	end_at 			DATE DEFAULT '1970-01-01',
	created_at 	TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

type PowerExtra struct {
	ClientID    string          `json:"client_id"`
	Multiplier  decimal.Decimal `json:"multiplier"`
	Description string          `json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

func getDoubleClaimClientList(ctx context.Context) []*Client {
	clientList := make([]*Client, 0)
	if err := session.Database(ctx).ConnQuery(ctx, `
select c.name, c.description, c.icon_url, c.identity_number, c.client_id, c.created_at, 
pe.description
FROM power_extra pe
LEFT JOIN client c ON pe.client_id = c.client_id
WHERE pe.start_at <= now() AND pe.end_at >= now()
	`, func(rows pgx.Rows) error {
		for rows.Next() {
			var c Client
			err := rows.Scan(&c.Name, &c.Description, &c.IconURL, &c.IdentityNumber, &c.ClientID, &c.CreatedAt, &c.Welcome)
			if err != nil {
				return err
			}
			clientList = append(clientList, &c)
		}
		return nil
	}); err != nil {
		session.Logger(ctx).Println(err)
	}
	return clientList
}

func checkIsDoubleClaimClient(ctx context.Context, clientID string) bool {
	list := getDoubleClaimClientList(ctx)
	for _, v := range list {
		if v.ClientID == clientID {
			return true
		}
	}
	return false
}
