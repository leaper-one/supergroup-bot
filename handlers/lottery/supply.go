package lottery

import (
	"context"
	"strconv"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

func GetLotterySupply(ctx context.Context) (map[string]*models.LotterySupply, error) {
	var lss []*models.LotterySupply
	err := session.DB(ctx).Find(&lss, "status = ?", 1).Error
	res := make(map[string]*models.LotterySupply)
	for _, ls := range lss {
		if ls.Inventory > 0 {
			res[ls.LotteryID] = ls
		}
	}
	return res, err
}

func getLotterySupplyBySupplyID(ctx context.Context, supplyID string) (*models.LotterySupply, error) {
	var ls models.LotterySupply
	err := session.DB(ctx).Take(&ls, "supply_id=?", supplyID).Error
	return &ls, err
}

func getLotteryByTrace(ctx context.Context, traceID string) (*config.Lottery, error) {
	var lsr models.LotterySupplyReceived
	session.DB(ctx).Take(&lsr, "trace_id=?", traceID)
	if lsr.SupplyID != "" {
		supply, err := getLotterySupplyBySupplyID(ctx, lsr.SupplyID)
		if err != nil {
			return nil, err
		}
		return &config.Lottery{
			LotteryID: supply.LotteryID,
			AssetID:   supply.AssetID,
			Amount:    supply.Amount,
			IconURL:   supply.IconURL,
			ClientID:  supply.ClientID,
			SupplyID:  supply.SupplyID,
			Inventory: supply.Inventory,
		}, nil
	}
	var lotteryID string
	if err := session.DB(ctx).Table("lottery_record").
		Select("lottery_id").
		Where("trace_id=?", traceID).
		Scan(&lotteryID).Error; err != nil {
		return nil, err
	}
	list := getInitListingLottery()
	i, _ := strconv.Atoi(lotteryID)
	return &list[i], nil
}
