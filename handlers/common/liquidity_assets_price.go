package common

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const liquidity_assets_price_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_assets_price (
	asset_id   VARCHAR(36) NOT NULL PRIMARY KEY,
	price_usd  VARCHAR NOT NULL DEFAULT '0',
	updated_at timestamp NOT NULL DEFAULT NOW()
);
`

type LiquidityAssetsPrice struct {
	AssetID   string    `json:"asset_id"`
	PriceUSD  string    `json:"price_usd"`
	UpdatedAt time.Time `json:"updated_at"`
}

func updateLiquidityAssetsPrice(ctx context.Context, assetID string, priceUSD decimal.Decimal) error {
	if priceUSD.IsZero() {
		return nil
	}
	_, err := session.DB(ctx).Exec(ctx, `
INSERT INTO liquidity_assets_price(asset_id, price_usd)
VALUES ($1, $2) 
ON CONFLICT (asset_id) 
DO UPDATE SET price_usd=$2, updated_at=now()
`, priceUSD, assetID)
	return err
}

func getLiquidityAssetPrice(ctx context.Context, assetID string) (decimal.Decimal, error) {
	var priceUSD decimal.Decimal
	err := session.DB(ctx).QueryRow(ctx, `
SELECT price_usd FROM liquidity_assets_price WHERE asset_id=$1
`, assetID).Scan(&priceUSD)
	if err != nil {
		if err == pgx.ErrNoRows {
			return decimal.Zero, nil
		}
		return decimal.Zero, err
	}
	return priceUSD, nil
}
