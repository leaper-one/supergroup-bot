package lottery

import (
	"context"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func getPower(ctx context.Context, userID string) models.Power {
	var p models.Power
	err := session.DB(ctx).Take(&p, "user_id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		p = models.Power{
			UserID:       userID,
			Balance:      decimal.Zero,
			LotteryTimes: 0,
		}
		if common.CheckUserIsVIP(ctx, userID) {
			p.LotteryTimes = 1
		}
		if err := session.DB(ctx).Create(&p).Error; err != nil {
			tools.Println(err)
		}
		return p
	}
	return p
}
