package services

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

type SwapService struct{}

var mu sync.WaitGroup

func (service *SwapService) Run(ctx context.Context) error {
	for {
		mu.Add(4)
		go UpdateAsset(ctx)
		go updateExinList(ctx)
		go updateFoxSwapList(ctx)
		go updateExinOtc(ctx)
		mu.Wait()
		time.Sleep(time.Minute * 5)
	}
}

// Exin 相关...

type exinAsset struct {
	UUID       string     `json:"uuid,omitempty"`
	Symbol     string     `json:"symbol,omitempty"`
	IconURL    string     `json:"iconUrl,omitempty"`
	PriceUsdt  string     `json:"priceUsdt,omitempty"`
	ChainAsset *exinAsset `json:"chainAsset,omitempty"`
}

type exinPair struct {
	Asset0                 *exinAsset `json:"asset0,omitempty"`
	Asset1                 *exinAsset `json:"asset1,omitempty"`
	LpAsset                *exinAsset `json:"lpAsset,omitempty"`
	Asset0Balance          string     `json:"asset0Balance,omitempty"`
	Asset1Balance          string     `json:"asset1Balance,omitempty"`
	UsdtTradeVolume24Hours string     `json:"usdtTradeVolume24Hours,omitempty"`
}

type exinStatistics struct {
	YearFloatingRate string `json:"yearFloatingRate,omitempty"`
}

type exinInfo struct {
	Pair       *exinPair
	Statistics *exinStatistics
}

type exinOtc struct {
	ID    int `json:"id"`
	Pair1 *struct {
		ExchangeID int `json:"exchangeId"`
	} `json:"pair1,omitempty"`
	AssetUUID string `json:"assetUuid"`

	//Asset struct {
	//	Data struct {
	//		UUID   string `json:"uuid"`
	//		Symbol string `json:"symbol"`
	//	} `json:"data"`
	//} `json:"asset"`
	//Pair1 struct {
	//	Exchange struct {
	//		En string `json:"en"`
	//	} `json:"exchange"`
	//} `json:"pair1"`
}

type exinPrice struct {
	CnyBuyMax float64 `json:"cnyBuyMax"`
	Pair1     struct {
		BuyPrice string `json:"buyPrice"`
	} `json:"pair1"`
}

func UpdateAsset(ctx context.Context) {
	var assets []*common.Asset
	err := session.DB(ctx).ConnQuery(ctx, "SELECT asset_id FROM assets", func(rows pgx.Rows) error {
		for rows.Next() {
			var a common.Asset
			err := rows.Scan(&a.AssetID)
			if err != nil {
				return err
			}
			assets = append(assets, &a)
		}
		return nil
	})
	if err != nil {
		tools.Println(err)
	}
	for _, asset := range assets {
		_, _ = common.SetAssetByID(ctx, nil, asset.AssetID)
		common.UpdateExinLocal(ctx, asset.AssetID)
	}
	mu.Done()
}

func updateExinList(ctx context.Context) {
	list, err := apiGetExinPairList(ctx)
	if err != nil {
		tools.Println(err)
		return
	}
	for _, pair := range list {
		updateExinSwapItem(ctx, pair.LpAsset.UUID)
	}
	mu.Done()
}

func updateExinSwapItem(ctx context.Context, id string) {
	info, err := apiGetExinStatistics(ctx, id)
	if err != nil || info == nil || info.Pair == nil {
		return
	}
	_, err = common.GetAssetByID(ctx, nil, id)
	if err != nil {
		return
	}
	pair := info.Pair
	lpAsset := pair.LpAsset
	asset0 := pair.Asset0
	asset1 := pair.Asset1

	_, _ = common.GetAssetByID(ctx, nil, asset0.UUID)
	_, _ = common.GetAssetByID(ctx, nil, asset1.UUID)

	asset0Price, asset0Balance, err := tools.CompareTwoString(asset0.PriceUsdt, pair.Asset0Balance)
	if err != nil {
		tools.Println("asset0 Price 出问题 了...", asset0.UUID)
		return
	}
	asset0Pool := asset0Price.Mul(asset0Balance)
	asset1Price, asset1Balance, err := tools.CompareTwoString(asset1.PriceUsdt, pair.Asset1Balance)
	if err != nil {
		tools.Println(err)
		return
	}
	asset1Pool := asset1Price.Mul(asset1Balance)
	pool := asset0Pool.Add(asset1Pool)
	err = common.UpdateSwap(ctx, &common.Swap{
		LpAsset:     lpAsset.UUID,
		Asset0:      asset0.UUID,
		Asset0Price: tools.NumberFixed(asset0.PriceUsdt, 8),
		Asset1:      asset1.UUID,
		Asset1Price: tools.NumberFixed(asset1.PriceUsdt, 8),
		Type:        common.SwapTypeExinSwap,
		Pool:        pool.StringFixed(2),
		Earn:        info.Statistics.YearFloatingRate + "%",
		Amount:      tools.NumberFixed(pair.UsdtTradeVolume24Hours, 2),
		UpdatedAt:   time.Now(),
	})
	if err != nil {
		tools.Println(err)
	}
}

func apiGetExinStatistics(ctx context.Context, id string) (*exinInfo, error) {
	var e exinInfo
	err := session.Api(ctx).Get("https://app.exinswap.com/api/v1/statistic/info?lpAssetUuid="+id, &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func apiGetExinPairList(ctx context.Context) ([]*exinPair, error) {
	var e []*exinPair
	err := session.Api(ctx).Get("https://app.exinswap.com/api/v1/pairs", &e)
	return e, err
}

// 更新 exin otc
func updateExinOtc(ctx context.Context) {
	otcList, err := apiGetExinOtcList(ctx)
	if err != nil {
		tools.Println(err)
	}
	for _, otc := range otcList {
		updateExinOtcItem(ctx, otc)
	}
	mu.Done()
}

var exchangeMap = map[int]string{
	2: "Binance",
	7: "Huobi",
}

// 更新 exin otc 的每一项
func updateExinOtcItem(ctx context.Context, otc *exinOtc) {
	price, err := apiGetExinPrice(ctx, strconv.Itoa(otc.ID))
	if err != nil {
		tools.Println(err)
		return
	}
	exchange := "MixSwap"
	if otc.Pair1 != nil &&
		otc.Pair1.ExchangeID != 0 &&
		exchangeMap[otc.Pair1.ExchangeID] != "" {
		exchange = exchangeMap[otc.Pair1.ExchangeID]
	}

	err = common.UpdateExinOtcAsset(ctx, &common.ExinOtcAsset{
		AssetID:  otc.AssetUUID,
		OtcID:    strconv.Itoa(otc.ID),
		Exchange: exchange,
		BuyMax:   decimal.NewFromFloat(price.CnyBuyMax).String(),
		PriceUsd: price.Pair1.BuyPrice,
	})
	if err != nil {
		tools.Println(err)
	}
}

func apiGetExinOtcList(ctx context.Context) ([]*exinOtc, error) {
	var e []*exinOtc
	err := session.Api(ctx).Get("https://eiduwejdk.com/api/v1/otc/", &e)
	return e, err
}

func apiGetExinPrice(ctx context.Context, id string) (*exinPrice, error) {
	var e exinPrice
	err := session.Api(ctx).Get("https://eiduwejdk.com/api/v1/otc/price?&otcId="+id, &e)
	return &e, err
}

// Fox 相关
type foxPair struct {
	BaseAmount       string `json:"base_amount,omitempty"`
	BaseAssetID      string `json:"base_asset_id,omitempty"`
	BaseValue        string `json:"base_value,omitempty"`
	Fee24h           string `json:"fee_24h,omitempty"`
	LiquidityAssetID string `json:"liquidity_asset_id,omitempty"`
	QuoteAmount      string `json:"quote_amount,omitempty"`
	QuoteAssetID     string `json:"quote_asset_id,omitempty"`
	QuoteValue       string `json:"quote_value,omitempty"`
	Volume24h        string `json:"volume_24h,omitempty"`
}

type foxResp struct {
	Pairs []*foxPair `json:"pairs,omitempty"`
}

func updateFoxSwapList(ctx context.Context) {
	mtgList, err := apiGetMtgFoxPairList(ctx)
	if err != nil {
		common.SendMsgToDeveloper(ctx, "", "获取 MtgFoxPairList 出问题了..."+err.Error())
		return
	}
	updateFoxSwapItem(ctx, common.SwapType4SwapMtg, mtgList)

	uniList, err := apiGetUniFoxPairList(ctx)
	if err != nil {
		common.SendMsgToDeveloper(ctx, "", "获取 UniFoxPairList 出问题了..."+err.Error())
		return
	}
	updateFoxSwapItem(ctx, common.SwapType4SwapNormal, uniList)
	mu.Done()
}

func updateFoxSwapItem(ctx context.Context, t string, list []*foxPair) {
	for _, pair := range list {
		_, _ = common.GetAssetByID(ctx, nil, pair.LiquidityAssetID)
		_, _ = common.GetAssetByID(ctx, nil, pair.BaseAssetID)
		_, _ = common.GetAssetByID(ctx, nil, pair.QuoteAssetID)
		_updateFoxSwapItem(ctx, t, pair)
	}
}

func _updateFoxSwapItem(ctx context.Context, t string, pair *foxPair) {
	bv, ba, err := tools.CompareTwoString(pair.BaseValue, pair.BaseAmount)
	if err != nil {
		return
	}
	qv, qa, err := tools.CompareTwoString(pair.QuoteValue, pair.QuoteAmount)
	if err != nil {
		return
	}
	fee, err := decimal.NewFromString(pair.Fee24h)
	if err != nil {
		return
	}
	pool := qv.Add(bv)

	asset0Price := "0"
	if !ba.IsZero() {
		asset0Price = bv.Div(ba).StringFixed(8)
	}

	asset1Price := "0"
	if !qa.IsZero() {
		asset1Price = qv.Div(qa).StringFixed(8)
	}

	earn := "0"
	if !pool.IsZero() {
		earn = fee.Mul(decimal.NewFromInt(36500)).Div(pool).StringFixed(2) + "%"
	}

	if pair.LiquidityAssetID == "" {
		pair.LiquidityAssetID = mixin.UniqueConversationID(pair.BaseAssetID, pair.QuoteAssetID)
	}

	err = common.UpdateSwap(ctx, &common.Swap{
		LpAsset:      pair.LiquidityAssetID,
		Asset0:       pair.BaseAssetID,
		Asset0Price:  asset0Price,
		Asset0Amount: ba.String(),
		Asset1:       pair.QuoteAssetID,
		Asset1Price:  asset1Price,
		Asset1Amount: qa.String(),
		Type:         t,
		Pool:         pool.StringFixed(2),
		Earn:         earn,
		Amount:       pair.Volume24h,
		UpdatedAt:    time.Now(),
	})
	if err != nil {
		tools.Println(err)
	}
}

var retry = 0

func apiGetMtgFoxPairList(ctx context.Context) ([]*foxPair, error) {
	var resp foxResp
	err := session.Api(ctx).Get("https://mtgswap-api.fox.one/api/pairs", &resp)
	if err != nil && retry <= 10 {
		retry++
		return apiGetMtgFoxPairList(ctx)
	}
	retry = 0
	return resp.Pairs, err
}

func apiGetUniFoxPairList(ctx context.Context) ([]*foxPair, error) {
	var resp foxResp
	err := session.Api(ctx).Get("https://legacy-api.4swap.org/api/pairs", &resp)
	if err != nil && retry <= 10 {
		retry++
		return apiGetUniFoxPairList(ctx)
	}
	retry = 0
	return resp.Pairs, err
}
