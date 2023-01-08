package lottery

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const MAX_POWER = 20000

func HandleInvitationClaim(ctx context.Context, tx *gorm.DB, userID, clientID string, isVip bool) error {
	// 1. 确认邀请关系是否是 30 天以内的
	i := common.GetInvitationByInviteeID(ctx, userID)
	if i == nil ||
		i.InviterID == "" ||
		time.Now().After(i.CreatedAt.Add(30*24*time.Hour)) {
		return nil
	}
	if !checkCanReceivedInvitationReward(ctx, clientID, i.InviterID) {
		return nil
	}
	addAmount := decimal.NewFromInt(1)
	if isVip {
		addAmount = decimal.NewFromInt(6)
	}

	if err := tx.Create(&models.InvitationPowerRecord{
		InviteeID: i.InviteeID,
		InviterID: i.InviterID,
		Amount:    addAmount,
	}).Error; err != nil {
		return err
	}

	if err := common.UpdatePower(ctx, tx, i.InviterID, addAmount, 0, models.PowerTypeInvitation); err != nil {
		return err
	}
	return nil
}

func checkCanReceivedInvitationReward(ctx context.Context, clientID, inviterID string) bool {
	// 1. 判断奖励到达上限
	totalPower, err := getUserTotalPower(ctx, inviterID)
	if err != nil {
		return false
	}
	if totalPower >= MAX_POWER {
		return false
	}
	// 2. 判断用户的状态
	if checkUserIsScam(ctx, clientID, inviterID) {
		return false
	}
	return true
}

func getUserTotalPower(ctx context.Context, userID string) (int, error) {
	var amount int
	err := session.DB(ctx).Table("power_record").
		Where("user_id = ? AND power_type = ?", userID, models.PowerTypeInvitation).
		Select("SUM(amount::integer)").Scan(&amount).Error

	return amount, err
}

func checkUserIsScam(ctx context.Context, clientID, userID string) bool {
	u, err := common.SearchUser(ctx, clientID, userID)
	if err != nil {
		tools.Println(err, userID)
		return false
	}
	return u.IsScam
}
