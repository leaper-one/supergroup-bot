package common

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func CheckIsClaim(ctx context.Context, userID string) bool {
	var count int64
	session.DB(ctx).Table("claim").Where("user_id = ? AND date = CURRENT_DATE", userID).Count(&count)
	return count > 0
}

func UpdatePower(ctx context.Context, tx *gorm.DB, userID string, addAmount decimal.Decimal, addLotteryTime int, powerType string) error {
	// 1. 拿到 power_balance
	if powerType != "" {
		if err := tx.Create(&models.PowerRecord{
			UserID:    userID,
			PowerType: powerType,
			Amount:    addAmount,
		}).Error; err != nil {
			return err
		}
	}
	var pow models.Power
	if err := tx.Where("user_id = ?", userID).First(&pow).Error; err != nil {
		return err
	}
	pow.Balance = pow.Balance.Add(addAmount)
	pow.LotteryTimes += addLotteryTime
	// 2. 更新 power_balance
	return tx.Save(&pow).Error
}
