package common

import (
	"context"
	"errors"
	"strings"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var cacheClientAssetLevel *tools.Mutex

func init() {
	cacheClientAssetLevel = tools.NewMutex()
}

func GetClientAssetLevel(ctx context.Context, clientID string) (models.ClientAssetLevel, error) {
	l := cacheClientAssetLevel.Read(clientID)
	if l == nil {
		var cal models.ClientAssetLevel
		if err := session.DB(ctx).
			Select("client_id, fresh, senior, large, fresh_amount, large_amount").
			Where("client_id = ?", clientID).
			Take(&cal).Error; err != nil {
			return cal, err
		}
		cacheClientAssetLevel.Write(clientID, cal)
		return cal, nil
	} else {
		return l.(models.ClientAssetLevel), nil
	}
}

func GetClientUserStatusByClientUser(ctx context.Context, u *models.ClientUser) (int, error) {
	foxUserAssetMap, _ := GetAllUserFoxShares(ctx, []string{u.UserID})
	exinUserAssetMap, _ := GetAllUserExinShares(ctx, []string{u.UserID})
	return GetClientUserStatus(ctx, u, foxUserAssetMap[u.UserID], exinUserAssetMap[u.UserID])
}

// 更新每个社群的币资产数量
func GetClientUserStatus(ctx context.Context, u *models.ClientUser, foxAsset durable.AssetMap, exinAsset durable.AssetMap) (int, error) {
	client, err := GetMixinClientByIDOrHost(ctx, u.ClientID)
	if err != nil {
		return models.ClientUserStatusAudience, session.BadDataError(ctx)
	}
	assets, err := GetUserAssets(ctx, u)
	if err != nil {
		// 获取资产出现问题
		if strings.Contains(err.Error(), "Forbidden") {
			// 授权出问题了，降为 观众
			return models.ClientUserStatusAudience, nil
		}
		return models.ClientUserStatusAudience, err
	}
	assetLevel, err := GetClientAssetLevel(ctx, client.ClientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.ClientUserStatusAudience, nil
		}
		return models.ClientUserStatusAudience, err
	}
	totalAmount := decimal.Zero
	if client.C.AssetID != "" {
		totalAmount, err = getHasAssetUserStatus(ctx, client, assets, assetLevel, foxAsset, exinAsset)
	} else {
		totalAmount, err = GetNoAssetUserStatus(ctx, assets, foxAsset, exinAsset)
	}
	if err != nil {
		tools.Println(err)
		return models.ClientUserStatusAudience, nil
	}

	if assetLevel.Large.LessThanOrEqual(totalAmount) {
		return models.ClientUserStatusLarge, nil
	}
	if assetLevel.Senior.LessThanOrEqual(totalAmount) {
		return models.ClientUserStatusSenior, nil
	}
	if assetLevel.Fresh.LessThanOrEqual(totalAmount) {
		return models.ClientUserStatusFresh, nil
	}
	return models.ClientUserStatusAudience, nil
}

func getHasAssetUserStatus(ctx context.Context, client *MixinClient, assets []*mixin.Asset, assetLevel models.ClientAssetLevel, foxAsset durable.AssetMap, exinAsset durable.AssetMap) (decimal.Decimal, error) {
	lpPriceMap, err := GetClientAssetLPCheckMapByID(ctx, client.ClientID)
	if err != nil {
		return decimal.Zero, err
	}
	asset, err := GetAssetByID(ctx, client.Client, client.C.AssetID)
	if err != nil {
		return decimal.Zero, err
	}
	totalAmount := decimal.Zero
	for _, m := range assets {
		if !lpPriceMap[m.AssetID].IsZero() {
			if asset.PriceUsd.IsZero() {
				return assetLevel.Large, err
			}
			amount := lpPriceMap[m.AssetID].Mul(m.Balance).Div(asset.PriceUsd)
			totalAmount = totalAmount.Add(amount)
		}
		if m.AssetID == asset.AssetID {
			totalAmount = totalAmount.Add(m.Balance)
		}
	}
	if !foxAsset[asset.AssetID].IsZero() {
		totalAmount = totalAmount.Add(foxAsset[asset.AssetID])
	}
	if !exinAsset[asset.AssetID].IsZero() {
		totalAmount = totalAmount.Add(exinAsset[asset.AssetID])
	}
	return totalAmount, nil
}

func GetNoAssetUserStatus(ctx context.Context, assets []*mixin.Asset, foxAsset durable.AssetMap, exinAsset durable.AssetMap) (decimal.Decimal, error) {
	totalAmount := decimal.Zero
	for _, asset := range assets {
		if !asset.PriceUSD.IsZero() {
			totalAmount = totalAmount.Add(asset.PriceUSD.Mul(asset.Balance))
		}
	}
	for assetID, balance := range foxAsset {
		asset, err := GetAssetByID(ctx, nil, assetID)
		if err == nil && !asset.PriceUsd.IsZero() && !balance.IsZero() {
			totalAmount = totalAmount.Add(asset.PriceUsd.Mul(balance))
		}
	}
	for assetID, balance := range exinAsset {
		asset, err := GetAssetByID(ctx, nil, assetID)
		if err == nil && !asset.PriceUsd.IsZero() && !balance.IsZero() {
			totalAmount = totalAmount.Add(asset.PriceUsd.Mul(balance))
		}
	}
	return totalAmount, nil
}

func GetAllUserFoxShares(ctx context.Context, userIDs []string) (durable.UserSharesMap, error) {
	userSharesMap := make(durable.UserSharesMap)
	if config.Config.FoxToken != "" {
		err := session.Api(context.Background()).FoxSharesCheck(userIDs, &userSharesMap)
		if err != nil {
			return nil, err
		}
	}
	return userSharesMap, nil
}

func GetAllUserExinShares(ctx context.Context, userIDs []string) (durable.UserSharesMap, error) {
	userSharesMap := make(durable.UserSharesMap)
	if config.Config.ExinToken != "" {
		assetIDs, err := getAllCheckAssetID(ctx)
		if err != nil {
			return nil, err
		}
		times := len(userIDs)/30 + 1
		for i := 0; i < times; i++ {
			start := i * 30
			var end int
			if i == times-1 {
				end = len(userIDs)
			} else {
				end = (i + 1) * 30
			}
			if err := session.Api(context.Background()).ExinSharesCheck(userIDs[start:end], assetIDs, &userSharesMap); err != nil {
				tools.Println(err)
			}
		}
	}
	return userSharesMap, nil
}

func getAllCheckAssetID(ctx context.Context) ([]string, error) {
	_assetIDs := make([]string, 0)
	err := session.DB(ctx).
		Table("client as c").
		Distinct("c.asset_id").
		Joins("LEFT JOIN assets AS a ON c.asset_id=a.asset_id").
		Find(&_assetIDs).Error
	assetIDs := make([]string, 0, len(_assetIDs))
	for _, assetID := range _assetIDs {
		if assetID != "" {
			assetIDs = append(assetIDs, assetID)
		}
	}
	return assetIDs, err
}
