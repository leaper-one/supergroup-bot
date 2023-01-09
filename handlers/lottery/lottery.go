package lottery

import (
	"context"
	"strconv"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 获取抽奖列表
func getLotteryList(ctx context.Context, u *models.ClientUser) []models.LotteryList {
	ls := make([]models.LotteryList, 0)
	list := getUserListingLottery(ctx, u.UserID)
	for _, lottery := range list {
		var l models.LotteryList
		l.Lottery = lottery
		if lottery.ClientID != "" {
			client, _ := common.GetClientByIDOrHost(ctx, lottery.ClientID)
			l.Description = client.Description
		}
		if lottery.AssetID != "" {
			asset, _ := common.GetAssetByID(ctx, nil, lottery.AssetID)
			l.Symbol = asset.Symbol
			l.PriceUSD = asset.PriceUsd
		}
		if l.Inventory != 0 {
			l.Inventory = 0
		}
		ls = append(ls, l)
	}
	return ls
}

func getUserListingLottery(ctx context.Context, userID string) [16]config.Lottery {
	lotteryList := getInitListingLottery()
	lss, err := GetLotterySupply(ctx)
	if err != nil {
		tools.Println(err)
		return [16]config.Lottery{}
	}
	if len(lss) != 0 {
		lssID := make([]string, 0)
		for id := range lss {
			lssID = append(lssID, id)
		}
		var lsr []*models.LotterySupplyReceived
		if err := session.DB(ctx).Select("supply_id").Find(&lsr, "user_id=? AND supply_id IN ?", userID, lssID).Error; err != nil {
			tools.Println(err)
		}
		for _, v := range lsr {
			delete(lss, v.SupplyID)
		}
	}
	for _, ls := range lss {
		id, _ := strconv.Atoi(ls.LotteryID)
		lotteryList[id] = config.Lottery{
			LotteryID: ls.LotteryID,
			AssetID:   ls.AssetID,
			Amount:    ls.Amount,
			IconURL:   ls.IconURL,
			ClientID:  ls.ClientID,
			SupplyID:  ls.SupplyID,
			Inventory: ls.Inventory,
		}
	}
	return lotteryList
}

func getInitListingLottery() [16]config.Lottery {
	initLottery := [16]config.Lottery{}
	for i, v := range config.Config.Lottery.List {
		initLottery[i] = v
	}
	return initLottery
}

func getLastLottery(ctx context.Context) []models.LotteryRecord {
	list := make([]models.LotteryRecord, 0)

	session.DB(ctx).Table("lottery_record as lr").
		Select("lr.asset_id, lr.amount, u.full_name,a.symbol,a.icon_url,a.price_usd").
		Joins("LEFT JOIN users u ON u.user_id = lr.user_id").
		Joins("LEFT JOIN assets a ON a.asset_id = lr.asset_id").
		Order("lr.created_at DESC").
		Limit(5).
		Find(&list)

	for i := range list {
		list[i].PriceUsd = list[i].PriceUsd.Mul(list[i].Amount).Round(2)
	}

	return list
}

func PostExchangeLottery(ctx context.Context, u *models.ClientUser) error {
	pow := getPower(ctx, u.UserID)
	if pow.Balance.LessThan(decimal.NewFromInt(100)) {
		return session.ForbiddenError(ctx)
	}
	return models.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := common.UpdatePower(ctx, tx, u.UserID, decimal.NewFromInt(-100), 1, models.PowerTypeLottery); err != nil {
			return err
		}
		return nil
	})
}
func PostLottery(ctx context.Context, u *models.ClientUser) (string, error) {
	if common.CheckIsBlockUser(ctx, u.ClientID, u.UserID) {
		return "", session.ForbiddenError(ctx)
	}
	lotteryID := ""
	err := models.RunInTransaction(ctx, func(tx *gorm.DB) error {
		// 1. 检查是否有足够的能量
		var pow models.Power
		if err := tx.Where(models.Power{UserID: u.UserID}).
			Attrs(models.Power{Balance: decimal.Zero, LotteryTimes: 0}).
			FirstOrCreate(&pow).Error; err != nil {
			return err
		}

		if pow.LotteryTimes < 1 {
			return session.ForbiddenError(ctx)
		}
		// 2. 根据概率获取 lottery
		lottery := GetRandomLottery(ctx, u)
		// 3. 保存当前的 power
		if err := tx.Model(&models.Power{}).
			Where("user_id = ?", u.UserID).
			Update("lottery_times", pow.LotteryTimes-1).Error; err != nil {
			return err
		}
		// 6. 保存 lottery 记录
		traceID := tools.GetUUID()
		if err := tx.Create(&models.LotteryRecord{
			LotteryID:  lottery.LotteryID,
			UserID:     u.UserID,
			AssetID:    lottery.AssetID,
			Amount:     lottery.Amount,
			TraceID:    traceID,
			SnapshotID: "",
		}).Error; err != nil {
			return err
		}

		if lottery.SupplyID != "" {
			// 记录抽到了项目方的奖品
			if err := tx.Create(&models.LotterySupplyReceived{
				SupplyID: lottery.SupplyID,
				UserID:   u.UserID,
				TraceID:  traceID,
			}).Error; err != nil {
				return err
			}

			if lottery.Inventory == 0 {
				return session.ForbiddenError(ctx)
			}
			if lottery.Inventory == 1 {
				if err := tx.Model(&models.LotterySupply{}).
					Where("supply_id = ?", lottery.SupplyID).
					Updates(map[string]interface{}{
						"inventory": 0,
						"status":    3,
					}).Error; err != nil {
					return err
				}
			} else if lottery.Inventory > 1 {
				if err := tx.Model(&models.LotterySupply{}).
					Where("supply_id = ?", lottery.SupplyID).
					Update("inventory", lottery.Inventory-1).Error; err != nil {
					return err
				}
			}
		}
		lotteryID = lottery.LotteryID
		return nil
	})
	if lotteryID == "" {
		return "", session.ForbiddenError(ctx)
	}
	return lotteryID, err
}
