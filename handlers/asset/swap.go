package asset

import (
	"context"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

var CacheExin = make([]*models.ExinAd, 0)

type SwapResp struct {
	List  []*models.Swap   `json:"list"`
	Asset *models.Asset    `json:"asset"`
	Ad    []*models.ExinAd `json:"ad"`
}

func GetSwapList(ctx context.Context, id string) (*SwapResp, error) {
	var ss []*models.Swap
	// CNB 特殊处理...
	addQuery := ""
	if id != "965e5c6e-434c-3fa9-b780-c50f43cd955c" {
		addQuery = " AND s.pool::real > 10000"
	}
	if err := session.DB(ctx).Table("swap as s").
		Select("s.*, lp.icon_url, a0.symbol as asset0_symbol, a1.symbol as asset1_symbol").
		Joins("LEFT JOIN assets as lp ON lp.asset_id=s.lp_asset").
		Joins("LEFT JOIN assets as a0 ON a0.asset_id=s.asset0").
		Joins("LEFT JOIN assets as a1 ON a1.asset_id=s.asset1").
		Where("(s.asset0=? OR s.asset1=?) AND lp.icon_url IS NOT null"+addQuery, id, id).
		Order("s.pool::real DESC").
		Scan(&ss).Error; err != nil {
		return nil, err
	}

	var exin models.Swap
	err := session.DB(ctx).Table("exin_otc_asset as e").
		Select("e.asset_id,e.otc_id,e.exchange,e.buy_max,e.price_usd,a.symbol as asset0_symbol,a.icon_url").
		Joins("LEFT JOIN assets as a ON a.asset_id=e.asset_id").
		Where("e.asset_id=?", id).Take(&exin).Error
	if err == nil {
		exin.Type = models.SwapTypeExinOne
		ss = append(ss, &exin)
	}

	asset, _ := common.GetAssetByID(ctx, nil, id)
	return &SwapResp{
		List:  ss,
		Asset: &asset,
		Ad:    CacheExin,
	}, nil
}
