package models

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const liquidity_mining_record_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_mining_record (
	mining_id VARCHAR(36) NOT NULL,
	record_id VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	amount varchar NOT NULL DEFAULT '0',
	profit varchar NOT NULL DEFAULT '0',
	created_at timestamp NOT NULL DEFAULT NOW()
);
`

type LiquidityMiningRecord struct {
	MiningID  string          `json:"mining_id,omitempty"`
	RecordID  string          `json:"record_id,omitempty"`
	UserID    string          `json:"user_id,omitempty"`
	AssetID   string          `json:"asset_id,omitempty"`
	Amount    decimal.Decimal `json:"amount,omitempty"`
	Profit    decimal.Decimal `json:"profit,omitempty"`
	CreatedAt time.Time       `json:"created_at,omitempty"`

	Status int    `json:"status,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	Date   string `json:"date,omitempty"`
}

func CreateLiquidityMiningRecord(ctx context.Context, tx pgx.Tx, r *LiquidityMiningRecord) error {
	query := durable.InsertQuery("liquidity_mining_record", "mining_id, record_id, user_id, asset_id, amount, profit")
	_, err := tx.Exec(ctx, query, r.MiningID, r.RecordID, r.UserID, r.AssetID, r.Amount, r.Profit)
	return err
}

func GetLiquidtityMiningRecordByMiningIDAndUserID(ctx context.Context, u *ClientUser, mintID, page, status string) ([]*LiquidityMiningRecord, error) {
	lmrs := make([]*LiquidityMiningRecord, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT a.symbol, lmr.record_id, lmr.amount, lmr.profit, to_char(lmr.created_at, 'YYYY-MM-DD') AS date
FROM liquidity_mining_record lmr
LEFT JOIN assets a ON a.asset_id = lmr.asset_id
WHERE lmr.mining_id = $1 
AND lmr.user_id = $2
ORDER BY lmr.created_at DESC 
	`, func(rows pgx.Rows) error {
		for rows.Next() {
			var lmr LiquidityMiningRecord
			if err := rows.Scan(&lmr.Symbol, &lmr.RecordID, &lmr.Amount, &lmr.Profit, &lmr.Date); err != nil {
				return err
			}
			status, err := getLiquidityMiningTxStatusByRecordID(ctx, lmr.RecordID)
			if err != nil {
				return err
			}
			lmr.Status = status
			lmrs = append(lmrs, &lmr)
		}
		return nil
	}, mintID, u.UserID)
	return lmrs, err
}
