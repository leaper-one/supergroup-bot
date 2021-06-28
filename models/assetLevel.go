package models

import (
	"context"
	"errors"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
	"strings"
	"time"
)

const client_asset_level_DDL = `
-- 用户的持仓等级
CREATE TABLE IF NOT EXISTS client_asset_level (
  client_id          VARCHAR(36) PRIMARY KEY,
  fresh              VARCHAR NOT NULL DEFAULT '0',
  senior             VARCHAR NOT NULL DEFAULT '0',
  large              VARCHAR NOT NULL DEFAULT '0',
  created_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type ClientAssetLevel struct {
	ClientID  string          `json:"client_id,omitempty"`
	Fresh     decimal.Decimal `json:"fresh,omitempty"`
	Senior    decimal.Decimal `json:"senior,omitempty"`
	Large     decimal.Decimal `json:"large,omitempty"`
	CreatedAt time.Time       `json:"created_at,omitempty"`
}

func UpdateClientAssetLevel(ctx context.Context, level *ClientAssetLevel) error {
	query := durable.InsertQueryOrUpdate("client_asset_level", "client_id", "fresh,senior,large")
	_, err := session.Database(ctx).Exec(ctx, query, level.ClientID, level.Fresh, level.Senior, level.Large)
	return err
}

var cacheClientAssetLevel = make(map[string]ClientAssetLevel)
var nilAssetLevel = ClientAssetLevel{}

func GetClientAssetLevel(ctx context.Context, clientID string) (ClientAssetLevel, error) {
	if cacheClientAssetLevel[clientID] == nilAssetLevel {
		var cal ClientAssetLevel
		if err := session.Database(ctx).QueryRow(ctx, `
SELECT client_id, fresh, senior, large 
FROM client_asset_level
WHERE client_id=$1
`, clientID).Scan(&cal.ClientID, &cal.Fresh, &cal.Senior, &cal.Large); err != nil {
			return cal, err
		}
		cacheClientAssetLevel[clientID] = cal
	}
	return cacheClientAssetLevel[clientID], nil
}

func GetClientUserStatusByClientUser(ctx context.Context, clientUser *ClientUser) (int, error) {
	foxUserAssetMap, _ := GetAllUserFoxShares(ctx, []string{clientUser.UserID})
	exinUserAssetMap, _ := GetAllUserExinShares(ctx, []string{clientUser.UserID})
	return GetClientUserStatus(ctx, clientUser, foxUserAssetMap[clientUser.UserID], exinUserAssetMap[clientUser.UserID])
}

// 更新每个社群的币资产数量
func GetClientUserStatus(ctx context.Context, clientUser *ClientUser, foxAsset durable.AssetMap, exinAsset durable.AssetMap) (int, error) {
	client := GetMixinClientByID(ctx, clientUser.ClientID)
	if client.ClientID == "" {
		return ClientUserStatusAudience, session.BadDataError(ctx)
	}
	assets, err := GetUserAssets(ctx, clientUser.AccessToken)
	if err != nil {
		// 获取资产出现问题
		if strings.Contains(err.Error(), "Forbidden") {
			// 授权出问题了，降为 观众
			return ClientUserStatusAudience, nil
		}
		return ClientUserStatusAudience, err
	}
	lpPriceMap, err := GetClientAssetLPCheckMapByID(ctx, clientUser.ClientID)
	if err != nil {
		return ClientUserStatusAudience, err
	}
	assetLevel, err := GetClientAssetLevel(ctx, clientUser.ClientID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ClientUserStatusAudience, nil
		}
		return ClientUserStatusAudience, err
	}
	asset, err := GetAssetByID(ctx, client.Client, client.AssetID)
	if err != nil {
		return ClientUserStatusAudience, err
	}
	totalAmount := decimal.Zero
	for _, m := range assets {
		if !lpPriceMap[m.AssetID].IsZero() {
			if asset.PriceUsd.IsZero() {
				continue
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
	if assetLevel.Large.LessThanOrEqual(totalAmount) {
		return ClientUserStatusLarge, nil
	}
	if assetLevel.Senior.LessThanOrEqual(totalAmount) {
		return ClientUserStatusSenior, nil
	}
	if assetLevel.Fresh.LessThanOrEqual(totalAmount) {
		return ClientUserStatusFresh, nil
	}

	return ClientUserStatusAudience, nil
}

func GetAllUserFoxShares(ctx context.Context, userIDs []string) (durable.UserSharesMap, error) {
	var userSharesMap durable.UserSharesMap

	err := session.Api(context.Background()).FoxSharesCheck(userIDs, &userSharesMap)
	if err != nil {
		return nil, err
	}
	return userSharesMap, nil
}

func GetAllUserExinShares(ctx context.Context, userIDs []string) (durable.UserSharesMap, error) {
	userSharesMap := make(durable.UserSharesMap)
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
			session.Logger(ctx).Println(err, userIDs)
		}
	}

	return userSharesMap, nil
}

func getAllCheckAssetID(ctx context.Context) ([]string, error) {
	assetIDs := make([]string, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT distinct(c.asset_id) FROM client as c
LEFT JOIN assets AS a ON c.asset_id=a.asset_id
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var assetID string
			if err := rows.Scan(&assetID); err != nil {
				return err
			}
			assetIDs = append(assetIDs, assetID)
		}
		return nil
	})
	return assetIDs, err
}
