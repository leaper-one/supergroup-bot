package mint

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func GetLiquidityMiningRespByID(ctx context.Context, u *models.ClientUser, id string) (*models.LiquidityMining, error) {
	var m models.LiquidityMining
	err := session.DB(ctx).Where("mining_id=?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	a, err := common.GetAssetByID(ctx, nil, m.AssetID)
	if err != nil {
		return nil, err
	}
	m.Symbol = a.Symbol
	// 如果没有token则跳授权页
	m.Status = models.LiquidityMiningStatusAuth

	rewardAsset, err := common.GetAssetByID(ctx, nil, m.RewardAssetID)
	if err != nil {
		return nil, err
	}
	m.RewardSymbol = rewardAsset.Symbol
	extraAsset, err := common.GetAssetByID(ctx, nil, m.ExtraAssetID)
	if err != nil {
		return nil, err
	}
	m.ExtraSymbol = extraAsset.Symbol
	// 检查token是否有资产权限
	assets, err := common.GetUserAssets(ctx, u)
	if err == nil && len(assets) > 0 {
		// 有授权资产则跳已参与活动页面
		m.Status = models.LiquidityMiningStatusPending
		lpAssets, err := common.GetClientAssetLPCheckMapByID(ctx, u.ClientID)
		if err != nil {
			return nil, err
		}
		for _, a := range assets {
			if _, ok := lpAssets[a.AssetID]; ok {
				if a.Balance.GreaterThan(decimal.Zero) {
					m.Status = models.LiquidityMiningStatusDone
					break
				}
			}
		}
		if m.Status == models.LiquidityMiningStatusDone {
			// 添加到已参与活动用户
			var u models.LiquidityMiningUser
			err := session.DB(ctx).Where("mining_id=? AND user_id=?", m.MiningID, u.UserID).First(&u).Error
			if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
				if err := session.DB(ctx).Create(&models.LiquidityMiningUser{
					MiningID: m.MiningID,
					UserID:   u.UserID,
				}).Error; err != nil {
					tools.Println(err)
					return nil, err
				}
			}
		}
	}
	return &m, nil
}

type Mint struct {
	models.LiquidityMiningRecord
	Status int64  `json:"status"`
	Symbol string `json:"symbol"`
	Date   string `json:"date"`
}

func GetLiquidityMiningRecordByMiningIDAndUserID(ctx context.Context, u *models.ClientUser, mintID, page, status string) ([]*Mint, error) {
	lmrs := make([]*Mint, 0)
	err := session.DB(ctx).Table("liquidity_mining_record as lmr").
		Select("a.symbol, lmr.record_id, lmr.amount, lmr.profit, to_char(lmr.created_at, 'YYYY-MM-DD') AS date, lmt.status").
		Joins("LEFT JOIN assets a ON a.asset_id = lmr.asset_id").
		Joins("LEFT JOIN liquidity_mining_tx lmt ON lmt.record_id = lmr.record_id").
		Where("lmr.mining_id = ? AND lmr.user_id = ?", mintID, u.UserID).
		Order("lmr.created_at DESC").
		Scan(&lmrs).Error
	return lmrs, err
}

func getLiquidityMiningTxByRecordID(ctx context.Context, id string) ([]*models.LiquidityMiningTx, error) {
	ms := make([]*models.LiquidityMiningTx, 0)
	err := session.DB(ctx).Table("liquidity_mining_tx").
		Select("mining_id,trace_id,user_id,asset_id,amount").
		Where("record_id = ?", id).
		Find(&ms).Error
	return ms, err
}

func ReceivedLiquidityMiningTx(ctx context.Context, u *models.ClientUser, recordID string) error {
	ms, err := getLiquidityMiningTxByRecordID(ctx, recordID)
	if err != nil {
		tools.Println(err)
		return err
	}
	if len(ms) == 0 {
		return session.BadDataError(ctx)
	}
	for _, m := range ms {
		memoStr, _ := json.Marshal(map[string]string{"type": models.SnapshotTypeMint, "id": m.MiningID})
		if err := session.DB(ctx).Create(&models.Transfer{
			ClientID:   u.ClientID,
			TraceID:    m.TraceID,
			AssetID:    m.AssetID,
			OpponentID: m.UserID,
			Memo:       string(memoStr),
			Amount:     m.Amount,
		}).Error; err != nil {
			if !durable.CheckIsPKRepeatError(err) {
				tools.Println(err)
			}
			return err
		}
	}
	return nil
}
