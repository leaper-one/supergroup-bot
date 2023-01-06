package common

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const assets_DDL = `
-- asset
CREATE TABLE IF NOT EXISTS assets (
  asset_id            VARCHAR(36) NOT NULL PRIMARY KEY,
	chain_id						VARCHAR(36) NOT NULL,
  icon_url            VARCHAR(1024) NOT NULL,
  symbol              VARCHAR(128) NOT NULL,
	name								VARCHAR NOT NULL,
	price_usd						VARCHAR,
	change_usd					VARCHAR
);`

type Asset struct {
	AssetID   string          `json:"asset_id,omitempty"`
	ChainID   string          `json:"chain_id,omitempty"`
	IconUrl   string          `json:"icon_url,omitempty"`
	Symbol    string          `json:"symbol,omitempty"`
	Name      string          `json:"name,omitempty"`
	PriceUsd  decimal.Decimal `json:"price_usd,omitempty"`
	ChangeUsd string          `json:"change_usd,omitempty"`
}

const exinOTCAsset_DDL = `
CREATE TABLE IF NOT EXISTS exin_otc_asset(
  asset_id            VARCHAR(36) NOT NULL PRIMARY KEY,
  otc_id              VARCHAR NOT NULL,
  exchange            VARCHAR NOT NULL DEFAULT 'exchange',
  buy_max             VARCHAR NOT NULL,
  price_usd						VARCHAR,
  updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() -- 更新时间
);
`

type ExinOtcAsset struct {
	AssetID   string    `json:"asset_id,omitempty"`
	OtcID     string    `json:"otc_id,omitempty"`
	Exchange  string    `json:"exchange,omitempty"`
	BuyMax    string    `json:"buy_max,omitempty"`
	PriceUsd  string    `json:"price_usd,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

const exinLocalAsset_DDL = `
CREATE TABLE IF NOT EXISTS exin_local_asset(
  asset_id            VARCHAR(36) NOT NULL,
  price               VARCHAR NOT NULL,
  symbol              VARCHAR NOT NULL,
  buy_max             VARCHAR NOT NULL,
  updated_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW() -- 更新时间
);
`

type ExinLocalAsset struct {
	AssetID   string `json:"asset_id,omitempty"`
	Price     string `json:"price,omitempty"`
	Symbol    string `json:"symbol,omitempty"`
	BuyMax    string `json:"buy_max,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func GetAssetByID(ctx context.Context, client *mixin.Client, assetID string) (Asset, error) {
	if assetID == "" {
		return Asset{}, nil
	}
	var a Asset
	assetString, err := session.Redis(ctx).QGet(ctx, "asset:"+assetID).Result()
	if errors.Is(err, redis.Nil) {
		err = session.DB(ctx).QueryRow(ctx,
			"SELECT asset_id,chain_id,icon_url,symbol,name,price_usd,change_usd FROM assets WHERE asset_id=$1",
			assetID,
		).Scan(&a.AssetID, &a.ChainID, &a.IconUrl, &a.Symbol, &a.Name, &a.PriceUsd, &a.ChangeUsd)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_asset, err := SetAssetByID(ctx, client, assetID)
				if err != nil {
					return a, err
				}
				a = *_asset
			} else {
				return a, err
			}
		}
		assetByte, err := json.Marshal(a)
		if err != nil {
			return a, err
		}
		if err := session.Redis(ctx).QSet(ctx, "asset:"+assetID, string(assetByte), time.Minute*5); err != nil {
			tools.Println(err)
		}
	} else {
		err := json.Unmarshal([]byte(assetString), &a)
		if err != nil {
			return a, err
		}
	}
	return a, nil
}

func GetExinOtcAssetByID(ctx context.Context, assetID string) (*Swap, error) {
	var s Swap
	err := session.DB(ctx).QueryRow(ctx,
		`SELECT e.asset_id,e.otc_id,e.exchange,e.buy_max,e.price_usd,a.symbol,a.icon_url
FROM exin_otc_asset as e
LEFT JOIN assets as a ON a.asset_id=e.asset_id
WHERE e.asset_id=$1
`, assetID).Scan(&s.AssetID, &s.OtcID, &s.Exchange, &s.BuyMax, &s.PriceUsd, &s.Asset0Symbol, &s.IconURL)
	s.Type = SwapTypeExinOne
	return &s, err
}

func SetAssetByID(ctx context.Context, client *mixin.Client, assetID string) (*Asset, error) {
	if client == nil {
		client = GetFirstClient(ctx)
	}
	a, err := client.ReadAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}
	query := durable.InsertQueryOrUpdate("assets", "asset_id", "chain_id,icon_url,symbol,name,price_usd,change_usd")
	_, err = session.DB(ctx).Exec(ctx, query, assetID, a.ChainID, a.IconURL, a.Symbol, a.Name, a.PriceUSD, a.ChangeUsd)
	if err != nil {
		tools.Println(err)
		return nil, err
	}
	return &Asset{a.AssetID, a.ChainID, a.IconURL, a.Symbol, a.Name, a.PriceUSD, a.ChangeUsd.String()}, nil
}

func UpdateExinOtcAsset(ctx context.Context, a *ExinOtcAsset) error {
	query := durable.InsertQueryOrUpdate("exin_otc_asset", "asset_id", "otc_id,exchange,buy_max,price_usd,updated_at")
	_, err := session.DB(ctx).Exec(ctx, query, a.AssetID, a.OtcID, a.Exchange, a.BuyMax, a.PriceUsd, time.Now())
	return err
}

type exinLocal struct {
	Price           string `json:"price"`
	Symbol          string `json:"symbol"`
	MaxTradingLimit string `json:"maxTradingLimit"`
}

// exin Local 相关
func UpdateExinLocal(ctx context.Context, id string) {
	var e exinLocal
	err := session.Api(ctx).Get("https://hk.exinlocal.com/api/v1/mixin/advertisement?type=sell&asset_id="+id, &e)
	if err != nil {
		return
	}
	_, err = session.DB(ctx).Exec(ctx, "DELETE FROM exin_local_asset WHERE asset_id=$1", id)
	if err != nil {
		tools.Println(err)
		return
	}
	buyMax, err := decimal.NewFromString(e.MaxTradingLimit)
	if err != nil {
		return
	}

	if buyMax.GreaterThanOrEqual(decimal.NewFromInt(1000)) {
		query := durable.InsertQuery("exin_local_asset", "asset_id,price,symbol,buy_max")
		_, err := session.DB(ctx).Exec(ctx, query, id, e.Price, e.Symbol, buyMax.StringFixed(2))
		if err != nil {
			tools.Println(err)
		}
	}
}

func GetExinLocalByID(ctx context.Context, id string) (*Swap, error) {
	var e Swap
	err := session.DB(ctx).QueryRow(ctx,
		`SELECT e.asset_id,e.price,e.symbol as s,e.buy_max,a.icon_url,a.symbol
FROM exin_local_asset as e 
LEFT JOIN assets as a ON a.asset_id=e.asset_id
WHERE e.asset_id=$1`,
		id).Scan(&e.AssetID, &e.PriceUsd, &e.Asset1Symbol, &e.BuyMax, &e.IconURL, &e.Asset0Symbol)
	e.Type = SwapTypeExinLocal
	return &e, err
}

type ExinAd struct {
	ID                 int    `json:"id"`
	AvatarUrl          string `json:"avatarUrl"`
	Nickname           string `json:"nickname"`
	IsCertification    bool   `json:"isCertification"`
	IsLandun           bool   `json:"isLandun"`
	Price              string `json:"price"`
	MinPrice           string `json:"minPrice"`
	MaxPrice           string `json:"maxPrice"`
	MultisigOrderCount int    `json:"multisigOrderCount"`
	In5minRate         string `json:"in5minRate"`
	OrderSuccessRank   string `json:"orderSuccessRank"`
	AssetID            string `json:"assetId"`
	PayMethods         []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"payMethods"`
}

var cacheExin = make([]*ExinAd, 0)

func UpdateExinLocalAD() {
	if config.Config.ExinLocalKey == "" {
		return
	}
	for {
		if err := GetExinLocalAd(_ctx, &cacheExin); err != nil {
			session.Logger(_ctx).Println(err)
		}
		time.Sleep(time.Minute)
	}
}

func GetExinLocalAd(ctx context.Context, ad *[]*ExinAd) error {
	err := session.Api(ctx).Get(`https://www.tigaex.com/api/v1/mixin/usdt/advertisement?apiKey=`+config.Config.ExinLocalKey, ad)
	if err != nil {
		return err
	}
	return nil
}

func GetUserAssets(ctx context.Context, u *ClientUser) ([]*mixin.Asset, error) {
	var assets []*mixin.Asset
	var err error
	if u.AccessToken != "" {
		assets, err = mixin.ReadAssets(ctx, u.AccessToken)
	} else if u.AuthorizationID != "" {
		assets, err = getUserAssetsByClientUser(ctx, u)
	} else {
		return nil, session.ForbiddenError(ctx)
	}
	if err != nil {
		if strings.HasPrefix(err.Error(), "[202/403] Forbidden") ||
			strings.HasPrefix(err.Error(), "[202/401]") {
			return nil, session.ForbiddenError(ctx)
		} else if errors.Is(err, context.Canceled) {
			return nil, err
		} else {
			return GetUserAssets(ctx, u)
		}
	}
	return assets, nil
}

func getUserAssetsByClientUser(ctx context.Context, u *ClientUser) ([]*mixin.Asset, error) {
	client, err := getMixinOAuthClientByClientUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return client.ReadAssets(ctx)
}

func GetUserAsset(ctx context.Context, u *ClientUser, assetID string) (*mixin.Asset, error) {
	var asset *mixin.Asset
	var err error
	if u.AccessToken != "" {
		asset, err = mixin.ReadAsset(ctx, u.AccessToken, assetID)
	} else if u.AuthorizationID != "" {
		asset, err = getUserAssetByClientUser(ctx, u, assetID)
	} else {
		return nil, session.ForbiddenError(ctx)
	}
	if err != nil {
		if strings.HasPrefix(err.Error(), "[202/403] Forbidden") ||
			strings.HasPrefix(err.Error(), "[202/401]") {
			return nil, session.ForbiddenError(ctx)
		} else if errors.Is(err, context.Canceled) {
			return nil, err
		} else {
			tools.Println(err)
			return GetUserAsset(ctx, u, assetID)
		}
	}
	return asset, nil
}

func getUserAssetByClientUser(ctx context.Context, u *ClientUser, id string) (*mixin.Asset, error) {
	client, err := getMixinOAuthClientByClientUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return client.ReadAsset(ctx, id)
}

func GetUserSnapshots(ctx context.Context, u *ClientUser, assetID string, offset time.Time, order string, limit int) ([]*mixin.Snapshot, error) {
	var ss []*mixin.Snapshot
	var err error
	if u.AccessToken != "" {
		ss, err = mixin.ReadSnapshots(ctx, u.AccessToken, assetID, offset, order, limit)
	} else if u.AuthorizationID != "" {
		ss, err = getUserSnapshotByClientUser(ctx, u, assetID, offset, order, limit)
	}
	if err != nil {
		if strings.HasPrefix(err.Error(), "[202/403] Forbidden") ||
			strings.HasPrefix(err.Error(), "[202/401]") {
			return nil, session.ForbiddenError(ctx)
		} else if errors.Is(err, context.Canceled) {
			return nil, err
		} else {
			tools.Println(err)
			return GetUserSnapshots(ctx, u, assetID, offset, order, limit)
		}
	}
	return ss, err
}

func getUserSnapshotByClientUser(ctx context.Context, u *ClientUser, assetID string, offset time.Time, order string, limit int) ([]*mixin.Snapshot, error) {
	client, err := getMixinOAuthClientByClientUser(ctx, u)
	if err != nil {
		return nil, err
	}
	return client.ReadSnapshots(ctx, assetID, offset, order, limit)
}
