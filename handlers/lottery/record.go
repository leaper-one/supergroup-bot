package lottery

import (
	"context"
	"errors"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"gorm.io/gorm"
)

func GetPowerRecordList(ctx context.Context, u *models.ClientUser, page int) ([]*models.PowerRecord, error) {
	if page < 1 {
		page = 1
	}
	var list []*models.PowerRecord
	if err := session.DB(ctx).
		Select("power_type, amount, to_char(created_at, 'YYYY-MM-DD') AS date").
		Order("created_at DESC").
		Offset((page-1)*20).
		Limit(20).
		Find(&list, "user_id = ?", u.UserID).Error; err != nil {
		return nil, err
	}
	return list, nil
}
func getLotteryRecord(ctx context.Context, userID string) *models.LotteryRecord {
	var r models.LotteryRecord

	if err := session.DB(ctx).Order("created_at").
		Where("is_received = false AND user_id = ?", userID).
		First(&r).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tools.Println("get client id error", err)
		}
		return nil
	}

	a, _ := common.GetAssetByID(ctx, nil, r.AssetID)
	r.IconURL = a.IconUrl
	r.Symbol = a.Symbol
	r.PriceUsd = a.PriceUsd
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

func GetLotteryRecordList(ctx context.Context, u *models.ClientUser, page int) ([]*models.LotteryRecord, error) {
	if page < 1 {
		page = 1
	}
	var list []*models.LotteryRecord
	if err := session.DB(ctx).
		Table("lottery_record as lr").
		Select("lr.asset_id, lr.amount, to_char(lr.created_at, 'YYYY-MM-DD') AS date, a.symbol, a.icon_url").
		Joins("LEFT JOIN asset a ON a.asset_id = lr.asset_id").
		Order("created_at DESC").
		Offset((page-1)*20).
		Limit(20).
		Where("lr.user_id = ?", u.UserID).
		Scan(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
