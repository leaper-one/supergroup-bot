package lottery

import (
	"context"
	"errors"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// -1: limit, 0: not found, 2: used 3: expired 9: success
func CheckVoucher(ctx context.Context, u *models.ClientUser, code string) (int, error) {
	if common.CheckIsBlockUser(ctx, u.ClientID, u.UserID) {
		return -1, session.ForbiddenError(ctx)
	}
	date := time.Now().Format("2006-01-02")
	uKey := u.UserID + ":" + date
	b, err := session.Redis(ctx).QIncr(ctx, uKey, time.Hour*24)
	if err != nil {
		return 0, err
	}
	if b >= 10 {
		return -1, nil
	}
	err = models.RunInTransaction(ctx, func(tx *gorm.DB) error {
		var voucher models.Voucher
		if err := tx.Take(&voucher, "code = ?", code).Error; err != nil {
			return err
		}
		if voucher.Status == 2 {
			return errors.New("voucher used")
		}
		if voucher.ExpiredAt.Before(time.Now()) {
			return errors.New("voucher expired")
		}
		if err := common.UpdatePower(ctx, tx, u.UserID, decimal.NewFromInt(100), 0, models.PowerTypeVoucher); err != nil {
			return err
		}
		voucher.Status = 2
		voucher.UserID = u.UserID
		voucher.ClientID = u.ClientID
		voucher.UpdatedAt = time.Now()
		return tx.Save(&voucher).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		if err.Error() == "voucher used" {
			return 2, nil
		}
		if err.Error() == "voucher expired" {
			return 3, nil
		}
		return 0, err
	}
	return 9, nil
}
