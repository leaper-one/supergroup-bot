package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const liquidity_mining_record_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_mining_record (
  trace_id VARCHAR(36) NOT NULL PRIMARY KEY,
	mining_id VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	amount varchar NOT NULL DEFAULT '0',
	status smallint NOT NULL DEFAULT 0,
	created_at timestamp NOT NULL DEFAULT NOW()
);
`

type LiquidityMiningRecord struct {
	MiningID  string          `json:"mining_id"`
	UserID    string          `json:"user_id"`
	AssetID   string          `json:"asset_id"`
	Amount    decimal.Decimal `json:"amount"`
	Status    int             `json:"status"`
	TraceID   string          `json:"trace_id"`
	CreatedAt time.Time       `json:"created_at"`
}

const (
	LiquidityMiningRecordStatusPending = 1 // 等待点击领取
	LiquidityMiningRecordStatusSuccess = 2 // 转账成功
	LiquidityMiningRecordStatusFailed  = 3 // 资产检查失败
)

func CreateLiquidityMiningRecord(ctx context.Context, m *LiquidityMiningRecord) error {
	query := durable.InsertQuery("liquidity_mining_record", "mining_id, user_id, asset_id, trace_id, amount, status")
	_, err := session.Database(ctx).Exec(ctx, query, m.MiningID, m.UserID, m.AssetID, m.TraceID, m.Amount, m.Status)
	return err
}

func CreateLiquidityMiningRecordWithTx(ctx context.Context, tx pgx.Tx, m *LiquidityMiningRecord) error {
	query := durable.InsertQuery("liquidity_mining_record", "mining_id, user_id, asset_id, trace_id, amount, status")
	_, err := tx.Exec(ctx, query, m.MiningID, m.UserID, m.AssetID, m.TraceID, m.Amount, m.Status)
	return err
}

func updateLiquidityMiningRecord(ctx context.Context, traceID string, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `
UPDATE liquidity_mining_record SET status=$1 WHERE trace_id=$2
	`, status, traceID)
	return err
}

func getLiquidityMiningRecordByTraceID(ctx context.Context, id string) (*LiquidityMiningRecord, error) {
	var m LiquidityMiningRecord
	err := session.Database(ctx).QueryRow(ctx, `
SELECT user_id,asset_id,amount FROM liquidity_mining_record WHERE trace_id=$1
	`, id).Scan(&m.UserID, &m.AssetID, &m.Amount)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func ReceivedLiquidityMiningRecord(ctx context.Context, u *ClientUser, traceID string) error {
	mr, err := getLiquidityMiningRecordByTraceID(ctx, traceID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	if mr.UserID != u.UserID {
		return session.ForbiddenError(ctx)
	}
	memo := map[string]string{"type": SnapshotTypeMint}
	memoStr, _ := json.Marshal(memo)
	if err := createTransferPending(ctx, u.ClientID, traceID, mr.AssetID, u.UserID, string(memoStr), mr.Amount); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	return nil
}

func handelMintSnapshot(ctx context.Context, clientID string, s *mixin.Snapshot) error {
	return updateLiquidityMiningRecord(ctx, s.TraceID, LiquidityMiningRecordStatusSuccess)
}
