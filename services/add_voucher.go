package services

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

type VoucherService struct{}

const VoucherCount = 100

func (service *VoucherService) Run(ctx context.Context) error {
	vouchers := make([]string, 0, VoucherCount)
	for len(vouchers) < VoucherCount {
		code := tools.GetRandomVoucherCode()
		if _, err := session.DB(ctx).Exec(ctx, durable.InsertQuery("voucher", "code"), code); err != nil {
			log.Println(err)
		}
		vouchers = append(vouchers, code)
	}
	tools.PrintJson(vouchers)
	return nil
}
