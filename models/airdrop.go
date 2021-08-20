package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
)

const airdrop_DDL = `
CREATE TABLE IF NOT EXISTS airdrop (
	airdrop_id          VARCHAR(36) NOT NULL,
	client_id           VARCHAR(36) NOT NULL,
	user_id             VARCHAR(36) NOT NULL,
	asset_id						VARCHAR(36) NOT NULL,
	trace_id						VARCHAR(36) NOT NULL,
	amount              VARCHAR NOT NULL,
	status              SMALLINT DEFAULT 1, -- 1 等待领取 2 正在发放 3 已完成
	created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	PRIMARY KEY (airdrop_id, user_id)
);
`

type Airdrop struct {
	AirdropID string          `json:"airdrop_id,omitempty"`
	ClientID  string          `json:"client_id,omitempty"`
	UserID    string          `json:"user_id,omitempty"`
	AssetID   string          `json:"asset_id,omitempty"`
	TraceID   string          `json:"trace_id,omitempty"`
	Amount    decimal.Decimal `json:"amount,omitempty"`
	Status    int             `json:"status"` // 1 waiting for claim, 2 pending, 3 success
	CreatedAt time.Time       `json:"created_at,omitempty"`
}

const (
	AirdropStatusWaiting = iota + 1
	AirdropStatusPending
	AirdropStatusSuccess
)

func CreateAirdrop(ctx context.Context, a *Airdrop) error {
	query := durable.InsertQuery("airdrop", "airdrop_id,client_id,user_id,asset_id,trace_id,amount")
	_, err := session.Database(ctx).Exec(ctx, query, a.AirdropID, a.ClientID, a.UserID, a.AssetID, tools.GetUUID(), a.Amount)
	if err != nil {
		return err
	}
	return nil
}

func GetAirdrop(ctx context.Context, u *ClientUser, airdropID string) (*Airdrop, error) {
	var a Airdrop
	err := session.Database(ctx).QueryRow(ctx, `
SELECT asset_id,amount,status,trace_id FROM airdrop WHERE airdrop_id = $1 AND user_id = $2
	 `, airdropID, u.UserID).Scan(&a.AssetID, &a.Amount, &a.Status, &a.TraceID)
	if durable.CheckEmptyError(err) != nil {
		return &a, err
	}
	return &a, nil
}

func ClaimAirdrop(ctx context.Context, u *ClientUser, airdropID string) (int, error) {
	a, err := GetAirdrop(ctx, u, airdropID)
	if err != nil || a.Status != AirdropStatusWaiting {
		return a.Status, err
	}

	if err := UpdateAirdropStatus(ctx, airdropID, u.UserID, AirdropStatusPending); err != nil {
		return 0, err
	}

	memo := map[string]string{"type": "airdrop"}
	memoStr, _ := json.Marshal(memo)
	if err := createTransferPending(ctx, u.ClientID, a.TraceID, a.AssetID, u.UserID, string(memoStr), a.Amount); err != nil {
		return 0, err
	}

	return AirdropStatusPending, nil
}

func UpdateAirdropStatus(ctx context.Context, airdropID, userID string, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `
UPDATE airdrop SET status = $1 WHERE airdrop_id = $2 AND user_id=$3
`, status, airdropID, userID)
	return err
}
func UpdateAirdropToSuccess(ctx context.Context, traceID string) error {
	_, err := session.Database(ctx).Exec(ctx, `
UPDATE airdrop SET status = $1 WHERE trace_id = $2
`, AirdropStatusSuccess, traceID)
	return err
}
