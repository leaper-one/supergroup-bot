package models

import (
	"context"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"time"
)

const client_asset_lp_check_DDL = `
-- 机器人 lp token 换算表
CREATE TABLE IF NOT EXISTS client_asset_lp_check (
  client_id          VARCHAR(36),
  asset_id           VARCHAR(36),
  updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  PRIMARY KEY(client_id, asset_id)
);

`

type ClientAssetLpCheck struct {
	ClientID  string
	AssetID   string
	UpdatedAt time.Time
	CreatedAt time.Time

	PriceUsd decimal.Decimal
}

func UpdateClientAssetLPCheck(ctx context.Context, clientID, assetID string) error {
	query := durable.InsertQueryOrUpdate("client_asset_lp_check", "client_id,asset_id", "updated_at")
	_, err := session.Database(ctx).Exec(ctx, query, clientID, assetID, time.Now())
	return err
}

var cacheClientLpCheckList = make(map[string]map[string]decimal.Decimal)

func GetClientAssetLPCheckMapByID(ctx context.Context, clientID string) (map[string]decimal.Decimal, error) {
	if cacheClientLpCheckList[clientID] == nil {
		cacheClientLpCheckList[clientID] = make(map[string]decimal.Decimal)
		err := session.Database(ctx).ConnQuery(ctx, `
SELECT calc.client_id,calc.asset_id,a.price_usd
FROM client_asset_lp_check AS calc
LEFT JOIN assets AS a ON calc.asset_id=a.asset_id
WHERE calc.client_id=$1
`, func(rows pgx.Rows) error {
			for rows.Next() {
				var ca ClientAssetLpCheck
				if err := rows.Scan(&ca.ClientID, &ca.AssetID, &ca.PriceUsd); err != nil {
					return err
				}
				//cal = append(cal, &ca)
				cacheClientLpCheckList[clientID][ca.AssetID] = ca.PriceUsd
			}
			return nil
		}, clientID)
		if err != nil {
			return nil, err
		}
		//cacheClientLpCheckList[clientID] = cal
	}
	return cacheClientLpCheckList[clientID], nil
}
