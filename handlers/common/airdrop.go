package common

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/shopspring/decimal"
)

func GetAirdrop(ctx context.Context, u *ClientUser, airdropID string) (*models.Airdrop, error) {
	var a models.Airdrop
	err := session.DB(ctx).Select("asset_id,amount,status,trace_id,ask_amount,ask_asset_id").
		Where("airdrop_id = ? AND user_id = ?", airdropID, u.UserID).
		First(&a).Error
	return &a, err
}

type ClaimAirdropResp struct {
	Symbol string          `json:"symbol,omitempty"`
	Amount decimal.Decimal `json:"amount,omitempty"`
	Status int             `json:"status"`
}

func ClaimAirdrop(ctx context.Context, u *ClientUser, airdropID string) (*ClaimAirdropResp, error) {
	a, err := GetAirdrop(ctx, u, airdropID)
	if err != nil || a.Status != models.AirdropStatusWaiting {
		return &ClaimAirdropResp{Status: a.Status}, err
	}
	if a.AskAmount != "" {
		var amount decimal.Decimal
		if a.AskAssetID == "" {
			amount, err = GetClientUserUsdAmountByClientUser(ctx, u)
		} else {
			amount, err = GetClientUserAssetAmountByClientUser(ctx, u, a.AskAssetID)
		}
		if err != nil {
			if strings.Contains(err.Error(), "Forbidden") {
				// 没有资产授权
				return &ClaimAirdropResp{Status: models.AirdropStatusAssetAuth}, nil
			}
		}

		askAmount, _ := decimal.NewFromString(a.AskAmount)
		if amount.LessThan(askAmount) {
			symbol := "USD"
			if a.AskAssetID != "" {
				asset, err := GetAssetByID(ctx, nil, a.AskAssetID)
				if err != nil {
					return nil, err
				}
				symbol = asset.Symbol
			}
			return &ClaimAirdropResp{Status: models.AirdropStatusAssetCheck, Amount: askAmount, Symbol: symbol}, nil
		}
	}

	if err := UpdateAirdropStatus(ctx, airdropID, u.UserID, AirdropStatusPending); err != nil {
		return nil, err
	}

	memo := map[string]string{"type": SnapshotTypeAirdrop}
	memoStr, _ := json.Marshal(memo)
	if err := createTransferPending(ctx, u.ClientID, a.TraceID, a.AssetID, u.UserID, string(memoStr), a.Amount); err != nil {
		return nil, err
	}

	return &ClaimAirdropResp{Status: AirdropStatusPending}, nil
}

func UpdateAirdropStatus(ctx context.Context, airdropID, userID string, status int) error {
	_, err := session.DB(ctx).Exec(ctx, `
UPDATE airdrop SET status = $1 WHERE airdrop_id = $2 AND user_id=$3
`, status, airdropID, userID)
	return err
}
func UpdateAirdropToSuccess(ctx context.Context, traceID string) error {
	_, err := session.DB(ctx).Exec(ctx, `
UPDATE airdrop SET status = $1 WHERE trace_id = $2
`, AirdropStatusSuccess, traceID)
	return err
}
