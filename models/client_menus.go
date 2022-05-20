package models

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
)

const client_menus_DDL = `
CREATE TABLE IF NOT EXISTS client_menus (
	client_id varchar(36) NOT NULL,
	icon varchar(255) NOT NULL,
	name_zh varchar(255) NOT NULL,
	name_en varchar(255) NOT NULL,
	name_ja varchar(255) NOT NULL,
	url varchar(255) NOT NULL,
	idx int NOT NULL DEFAULT 0,
	created_at timestamp NOT NULL DEFAULT NOW()
);
`

type ClientMenu struct {
	ClientID  string    `json:"client_id"`
	Icon      string    `json:"icon"`
	NameZh    string    `json:"name_zh"`
	NameEn    string    `json:"name_en"`
	NameJa    string    `json:"name_ja"`
	URL       string    `json:"url"`
	Idx       int       `json:"idx"`
	CreatedAt time.Time `json:"created_at"`
}

func getClientMenus(ctx context.Context, clientID string) ([]*ClientMenu, error) {
	cms := make([]*ClientMenu, 0)
	if err := session.Database(ctx).ConnQuery(ctx, `SELECT * FROM client_menus WHERE client_id = $1 ORDER BY idx DESC`, func(rows pgx.Rows) error {
		for rows.Next() {
			var cm ClientMenu
			if err := rows.Scan(&cm.ClientID, &cm.Icon, &cm.NameZh, &cm.NameEn, &cm.NameJa, &cm.URL, &cm.Idx, &cm.CreatedAt); err != nil {
				return err
			}
			cms = append(cms, &cm)
		}
		return nil
	}, clientID); err != nil {
		return nil, err
	}
	return cms, nil
}
