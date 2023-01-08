package common

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/shopspring/decimal"
)

// 获取某个社群的所有流动性资产
func GetClientAssetLPCheckMapByID(ctx context.Context, clientID string) (map[string]decimal.Decimal, error) {
	result := make(map[string]decimal.Decimal)
	var lps []*models.ClientAssetLpCheck
	err := session.DB(ctx).Table("client_asset_lp_check AS calc").
		Select("calc.client_id,calc.asset_id,a.price_usd").
		Joins("LEFT JOIN assets AS a ON calc.asset_id=a.asset_id").
		Where("calc.client_id=?", clientID).
		Find(&lps).Error
	if err != nil {
		return nil, err
	}
	for _, lp := range lps {
		result[lp.AssetID] = lp.PriceUsd
	}
	return result, nil
}
