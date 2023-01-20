package services

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

type VoucherService struct{}

const VoucherCount = 100

func (service *VoucherService) Run(ctx context.Context) error {
	vouchers := make([]string, 0, VoucherCount)
	for len(vouchers) < VoucherCount {
		code := tools.GetRandomVoucherCode()
		if err := session.DB(ctx).Create(&models.Voucher{Code: code}).Error; err != nil {
			return err
		}
		vouchers = append(vouchers, code)
	}
	tools.PrintJson(vouchers)
	return nil
}
