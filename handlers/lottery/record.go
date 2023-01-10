package lottery

import (
	"context"
	"errors"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PowerRecord struct {
	models.PowerRecord
	Date string `json:"date"`
}

func GetPowerRecordList(ctx context.Context, u *models.ClientUser, page int) ([]*PowerRecord, error) {
	if page < 1 {
		page = 1
	}
	var list []*PowerRecord
	if err := session.DB(ctx).
		Table("power_record").
		Select("power_type, amount, to_char(created_at, 'YYYY-MM-DD') AS date").
		Where("user_id = ?", u.UserID).
		Order("created_at DESC").
		Offset((page - 1) * 20).
		Limit(20).
		Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

type LotteryRecordTx struct {
	TraceID     string          `json:"trace_id"`
	Amount      decimal.Decimal `json:"amount"`
	PriceUsd    decimal.Decimal `json:"price_usd"`
	Symbol      string          `json:"symbol"`
	Description string          `json:"description"`
}

func getLotteryRecord(ctx context.Context, userID string) *LotteryRecordTx {
	var r LotteryRecordTx

	if err := session.DB(ctx).Table("lottery_record lr").
		Select("lr.trace_id, lr.amount, a.asset_id, a.symbol, a.icon_url, a.price_usd").
		Joins("LEFT JOIN assets a ON a.asset_id = lr.asset_id").
		Where("lr.is_received = false AND lr.user_id = ?", userID).
		Limit(1).
		Scan(&r).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tools.Println("get client id error", err)
		}
		return nil
	}
	if r.TraceID == "" {
		return nil
	}

	clientID := ""
	lottery, err := getLotteryByTrace(ctx, r.TraceID)
	if err != nil {
		return nil
	}
	clientID = lottery.ClientID
	if clientID != "" {
		c, err := common.GetClientByIDOrHost(ctx, clientID)
		if err != nil {
			tools.Println("get client error", err)
			return nil
		}
		r.Description = c.Description
	}
	return &r
}

func GetLotteryRecordList(ctx context.Context, u *models.ClientUser, page int) ([]*LotteryRecordView, error) {
	if page < 1 {
		page = 1
	}
	var list []*LotteryRecordView
	if err := session.DB(ctx).
		Table("lottery_record as lr").
		Select("lr.asset_id, lr.amount, to_char(lr.created_at, 'YYYY-MM-DD') AS date, a.symbol, a.icon_url").
		Joins("LEFT JOIN assets a ON a.asset_id = lr.asset_id").
		Order("created_at DESC").
		Offset((page-1)*20).
		Limit(20).
		Where("lr.user_id = ?", u.UserID).
		Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
