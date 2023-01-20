package airdrop

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func GetAirdrop(ctx context.Context, u *models.ClientUser, airdropID string) (*models.Airdrop, error) {
	var a models.Airdrop
	err := session.DB(ctx).Select("asset_id,amount,status,trace_id,ask_amount,ask_asset_id").
		Where("airdrop_id = ? AND user_id = ?", airdropID, u.UserID).
		Take(&a).Error
	return &a, err
}

type ClaimAirdropResp struct {
	Symbol string          `json:"symbol,omitempty"`
	Amount decimal.Decimal `json:"amount,omitempty"`
	Status int             `json:"status"`
}

func ClaimAirdrop(ctx context.Context, u *models.ClientUser, airdropID string) (*ClaimAirdropResp, error) {
	a, err := GetAirdrop(ctx, u, airdropID)
	if err != nil || a.Status != models.AirdropStatusWaiting {
		return &ClaimAirdropResp{Status: a.Status}, err
	}
	if a.AskAmount != "" {
		var amount decimal.Decimal
		if a.AskAssetID == "" {
			amount, err = getClientUserUsdAmountByClientUser(ctx, u)
		} else {
			amount, err = getClientUserAssetAmountByClientUser(ctx, u, a.AskAssetID)
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
				asset, err := common.GetAssetByID(ctx, nil, a.AskAssetID)
				if err != nil {
					return nil, err
				}
				symbol = asset.Symbol
			}
			return &ClaimAirdropResp{Status: models.AirdropStatusAssetCheck, Amount: askAmount, Symbol: symbol}, nil
		}
	}

	memo := map[string]string{"type": models.SnapshotTypeAirdrop}
	memoStr, _ := json.Marshal(memo)
	if err := models.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Model(&models.Airdrop{}).
			Where("airdrop_id=? AND user_id=?", airdropID, u.UserID).
			Update("status", models.AirdropStatusPending).Error; err != nil {
			return err
		}

		if err := tx.Create(&models.Transfer{
			ClientID:   u.ClientID,
			TraceID:    a.TraceID,
			AssetID:    a.AssetID,
			OpponentID: u.UserID,
			Memo:       string(memoStr),
			Amount:     a.Amount,
		}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}
	// if err := common.CreateTransferPending(ctx, u.ClientID, a.TraceID, a.AssetID, u.UserID, string(memoStr), a.Amount); err != nil {
	// 	return nil, err
	// }

	return &ClaimAirdropResp{Status: models.AirdropStatusPending}, nil
}

func getClientUserUsdAmountByClientUser(ctx context.Context, u *models.ClientUser) (decimal.Decimal, error) {
	assets, err := common.GetUserAssets(ctx, u)
	if err != nil {
		return decimal.Zero, err
	}
	foxAsset, _ := common.GetAllUserFoxShares(ctx, []string{u.UserID})
	exinAsset, _ := common.GetAllUserExinShares(ctx, []string{u.UserID})
	return common.GetNoAssetUserStatus(ctx, assets, foxAsset[u.UserID], exinAsset[u.UserID])
}

func getClientUserAssetAmountByClientUser(ctx context.Context, u *models.ClientUser, assetID string) (decimal.Decimal, error) {
	assets, err := common.GetUserAssets(ctx, u)
	if err != nil {
		return decimal.Zero, err
	}
	foxAsset, _ := common.GetAllUserFoxShares(ctx, []string{u.UserID})
	exinAsset, _ := common.GetAllUserExinShares(ctx, []string{u.UserID})
	var res decimal.Decimal
	for _, a := range assets {
		if a.AssetID == assetID {
			res = a.Balance
		}
	}
	if foxAsset[u.UserID] != nil && !foxAsset[u.UserID][assetID].IsZero() {
		res = res.Add(foxAsset[u.UserID][assetID])
	}
	if exinAsset[u.UserID] != nil && !exinAsset[u.UserID][assetID].IsZero() {
		res = res.Add(exinAsset[u.UserID][assetID])
	}
	return res, nil
}
