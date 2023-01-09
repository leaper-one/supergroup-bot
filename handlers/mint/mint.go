package mint

import (
	"context"
	"encoding/json"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

func GetLiquidityMiningRecordByMiningIDAndUserID(ctx context.Context, u *models.ClientUser, mintID, page, status string) ([]*models.LiquidityMiningRecord, error) {
	lmrs := make([]*models.LiquidityMiningRecord, 0)
	err := session.DB(ctx).Table("liquidity_mining_record as lmr").
		Select("a.symbol, lmr.record_id, lmr.amount, lmr.profit, to_char(lmr.created_at, 'YYYY-MM-DD') AS date, lmt.status").
		Joins("LEFT JOIN assets a ON a.asset_id = lmr.asset_id").
		Joins("LEFT JOIN liquidity_mining_tx lmt ON lmt.record_id = lmr.record_id").
		Where("lmr.mining_id = ? AND lmr.user_id = ?", mintID, u.UserID).
		Order("lmr.created_at DESC").
		Find(&lmrs).Error
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
		memo := map[string]string{"type": models.SnapshotTypeMint, "id": m.MiningID}
		memoStr, _ := json.Marshal(memo)
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
