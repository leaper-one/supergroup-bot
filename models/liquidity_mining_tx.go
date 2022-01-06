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

const liquidity_mining_tx_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_mining_tx (
  trace_id VARCHAR(36) NOT NULL PRIMARY KEY,
	mining_id VARCHAR(36) NOT NULL,
	record_id VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	amount varchar NOT NULL DEFAULT '0',
	status smallint NOT NULL DEFAULT 0,
	created_at timestamp NOT NULL DEFAULT NOW()
);
`

type LiquidityMiningTx struct {
	MiningID  string          `json:"mining_id"`
	RecordID  string          `json:"record_id"`
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

func CreateLiquidityMiningTx(ctx context.Context, m *LiquidityMiningTx) error {
	query := durable.InsertQuery("liquidity_mining_tx", "mining_id, record_id, user_id, asset_id, trace_id, amount, status")
	_, err := session.Database(ctx).Exec(ctx, query, m.MiningID, m.RecordID, m.UserID, m.AssetID, m.TraceID, m.Amount, m.Status)
	return err
}

func getLiquidityMiningTxStatusByRecordID(ctx context.Context, id string) (int, error) {
	var status int
	err := session.Database(ctx).QueryRow(ctx, `
SELECT status FROM liquidity_mining_tx WHERE record_id=$1 LIMIT 1
	`, id).Scan(&status)
	if err != nil {
		return 0, err
	}
	return status, nil
}

func CreateLiquidityMiningTxWithTx(ctx context.Context, tx pgx.Tx, m *LiquidityMiningTx) error {
	query := durable.InsertQuery("liquidity_mining_tx", "mining_id, record_id, user_id, asset_id, trace_id, amount, status")
	_, err := tx.Exec(ctx, query, m.MiningID, m.RecordID, m.UserID, m.AssetID, m.TraceID, m.Amount, m.Status)
	return err
}

func updateLiquidityMiningTx(ctx context.Context, traceID string, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `
UPDATE liquidity_mining_tx SET status=$1 WHERE trace_id=$2
	`, status, traceID)
	return err
}
func getLiquidityMiningTxByRecordID(ctx context.Context, id string) ([]*LiquidityMiningTx, error) {
	ms := make([]*LiquidityMiningTx, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT mining_id,trace_id,user_id,asset_id,amount FROM liquidity_mining_tx WHERE record_id=$1
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var m LiquidityMiningTx
			if err := rows.Scan(&m.MiningID, &m.TraceID, &m.UserID, &m.AssetID, &m.Amount); err != nil {
				return err
			}
			ms = append(ms, &m)
		}
		return nil
	}, id)
	return ms, err
}

func ReceivedLiquidityMiningTx(ctx context.Context, u *ClientUser, recordID string) error {
	ms, err := getLiquidityMiningTxByRecordID(ctx, recordID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	if len(ms) == 0 {
		return session.BadDataError(ctx)
	}
	for _, m := range ms {
		memo := map[string]string{"type": SnapshotTypeMint, "id": m.MiningID}
		memoStr, _ := json.Marshal(memo)
		if err := createTransferPending(ctx, u.ClientID, m.TraceID, m.AssetID, m.UserID, string(memoStr), m.Amount); err != nil {
			session.Logger(ctx).Println(err)
			return err
		}
	}

	return nil
}

func handelMintSnapshot(ctx context.Context, clientID string, s *mixin.Snapshot) error {
	return updateLiquidityMiningTx(ctx, s.TraceID, LiquidityMiningRecordStatusSuccess)
}
