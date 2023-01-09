package jobs

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/jackc/pgx/v4"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func StartLiquidityDailyJob() {
	ctx := models.Ctx
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("1 0 * * *", func() {
		log.Println("start liquidity job")
		StatisticLiquidityDaily(ctx)
		// StatisticLiquidityMonth(ctx)
	})
	if err != nil {
		session.Logger(ctx).Println(err)
		tools.SendMsgToDeveloper("定时任务StartLiquidityJob。。。出问题了。。。")
		return
	}
	c.Start()
}

// 统计每日的情况
func StatisticLiquidityDaily(ctx context.Context) error {
	// 1. 获取所有的 liquidity
	var liquidities []*models.Liquidity
	if err := session.DB(ctx).Find(&liquidities, "start_at < now() AND end_at > now()").Error; err != nil {
		return err
	}

	// 2. 选择一个 liquidity，然后获取所有的用户
	for _, l := range liquidities {
		var users []string
		if err := session.DB(ctx).Table("liquidity_user").
			Where("liquidity_id = ?", l.LiquidityID).Pluck("user_id", &users).Error; err != nil {
			tools.Println(err)
			continue
		}

		// 10 9 9:00 9:24
		now := time.Now().UTC()
		endAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		startAt := endAt.AddDate(0, 0, -1)
		asset, err := common.GetAssetByID(ctx, nil, l.AssetIDs)
		if err != nil {
			tools.Println(err)
			continue
		}
		for _, uid := range users {
			// 3.0 判断上一次是否达标，如果不达标则跳过
			if time.Now().UTC().Day() != 1 {
				amount, err := getRecentSnapshot(ctx, l.LiquidityID, uid)
				if err != nil {
					tools.Println(err)
					continue
				}
				if !amount.IsZero() && amount.LessThan(l.MinAmount) {
					continue
				}
			}

			// 3.1 获取该用户的资产, 使指定 asset 为初始值
			u, err := common.GetClientUserByClientIDAndUserID(ctx, l.ClientID, uid)
			if err != nil {
				tools.Println(err)
				continue
			}
			a, err := common.GetUserAsset(ctx, &u, l.AssetIDs)
			if err != nil {
				tools.Println(err)
				if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "401") {
					a, err := common.GetAssetByID(ctx, nil, l.AssetIDs)
					if err != nil {
						tools.Println(err)
						continue
					}
					// if _, err := session.DB(ctx).Exec(ctx,
					// 	durable.InsertQuery("liquidity_snapshot",
					// 		"user_id,liquidity_id,idx,date,lp_symbol,lp_amount,usd_value"),
					// 	// TODO
					// 	u.UserID, l.LiquidityID, 1, startAt, a.Symbol, "0", "0"); err != nil {
					// 	tools.Println(err)
					// 	continue
					// }
					if err := session.DB(ctx).Create(&models.LiquiditySnapshot{
						UserID:      u.UserID,
						LiquidityID: l.LiquidityID,
						Idx:         1,
						Date:        startAt.String(),
						LpSymbol:    a.Symbol,
						LpAmount:    decimal.Zero,
						UsdValue:    decimal.Zero,
					}).Error; err != nil {
						tools.Println(err)
						continue
					}

					// if _, err := session.DB(ctx).Exec(ctx,
					// 	durable.InsertQuery("liquidity_tx",
					// 		"trace_id,month,liquidity_id,idx,user_id,asset_id,status"),
					// 	tools.GetUUID(), time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), l.LiquidityID, 0, u.UserID, "", models.LiquidityTxFail); err != nil {
					// 	tools.Println(err)
					// 	continue
					// }
					if err := session.DB(ctx).Create(&models.LiquidityTx{
						TraceID:     tools.GetUUID(),
						Month:       time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
						LiquidityID: l.LiquidityID,
						Idx:         0,
						UserID:      u.UserID,
						AssetID:     "",
						Status:      models.LiquidityTxFail,
					}).Error; err != nil {
						tools.Println(err)
						continue
					}

				}
				continue
			}
			minAmount, err := GetMinAmount(ctx, &u, l.AssetIDs, startAt, endAt, a.Balance, true)
			if err != nil {
				tools.Println(err)
				continue
			}
			// 5. 保存该值，结束
			if err := session.DB(ctx).Create(&models.LiquiditySnapshot{
				UserID:      u.UserID,
				LiquidityID: l.LiquidityID,
				Idx:         1,
				Date:        startAt.String(),
				LpSymbol:    a.Symbol,
				LpAmount:    minAmount,
				UsdValue:    asset.PriceUsd.Mul(minAmount),
			}).Error; err != nil {
				tools.Println(err)
				continue
			}

			if minAmount.LessThan(l.MinAmount) {
				if err := session.DB(ctx).Create(&models.LiquidityTx{
					TraceID:     tools.GetUUID(),
					Month:       time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
					LiquidityID: l.LiquidityID,
					Idx:         0,
					UserID:      u.UserID,
					AssetID:     "",
					Status:      models.LiquidityTxFail,
				}).Error; err != nil {
					tools.Println(err)
					continue
				}
			}
		}
	}

	return nil
}

func StatisticLiquidityMonth(ctx context.Context) error {
	now := time.Now().UTC()
	if now.Day() != 1 {
		return nil
	}
	lastMonthFirstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
	lastMonthLastDay := lastMonthFirstDay.AddDate(0, 1, 0).Add(-time.Second)
	// 1. 获取上个月的 liquidity_detail
	var lds []*models.LiquidityDetail
	if err := session.DB(ctx).Find(&lds, "start_at>=? AND end_at<?", lastMonthFirstDay, lastMonthLastDay).Error; err != nil {
		return err
	}

	for _, ld := range lds {
		// 2. 获取该 liquidity 的所有参与者
		var users []string
		if err := session.DB(ctx).Model(&models.LiquidityUser{}).
			Where("liquidity_id=?", ld.LiquidityID).Pluck("user_id", &users).Error; err != nil {
			return err
		}

		// 3. 遍历所有参与者，获得 lp_amount_map
		lpUserAmountMap := make(map[string]decimal.Decimal)
		totalAmount := decimal.Zero
		for _, uid := range users {
			// 3.1 判断该用户是否存在 liquidity_tx
			var txStatus string

			if err := session.DB(ctx).
				Select("status").
				Where("liquidity_id=? AND user_id=? AND month=?", ld.LiquidityID, uid, lastMonthFirstDay).
				Scan(&txStatus).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
			}

			if txStatus == models.LiquidityTxFail {
				continue
			}

			// 3.2 获取该用户上月的 lp_amount，
			var lss []*models.LiquiditySnapshot
			if err := session.DB(ctx).Find(&lss, "liquidity_id=? AND user_id=? AND date>=? AND date<?",
				ld.LiquidityID, uid, lastMonthFirstDay, lastMonthLastDay).Error; err != nil {
				return err
			}
			for _, ls := range lss {
				lpUserAmountMap[uid] = lpUserAmountMap[uid].Add(ls.LpAmount)
				totalAmount = totalAmount.Add(ls.LpAmount)
			}
		}

		for uid, amount := range lpUserAmountMap {
			// 4. 计算每个用户的分成
			share := amount.Div(totalAmount)
			// 5. 计算每个用户的分成金额
			amount = share.Mul(ld.Amount).Truncate(8)
			// 6. 插入 liquidity_tx
			if err := session.DB(ctx).Create(&models.LiquidityTx{
				TraceID:     tools.GetUUID(),
				Month:       lastMonthFirstDay,
				LiquidityID: ld.LiquidityID,
				Idx:         0,
				UserID:      uid,
				AssetID:     ld.AssetID,
				Status:      models.LiquidityTxSuccess,
				Amount:      amount.String(),
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func getRecentSnapshot(ctx context.Context, lid, uid string) (decimal.Decimal, error) {
	var amount decimal.Decimal
	if err := session.DB(ctx).Model(&models.LiquiditySnapshot{}).
		Where("liquidity_id=? AND user_id=?", lid, uid).
		Order("date DESC").
		Limit(1).
		Scan(&amount).Error; err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return decimal.Zero, nil
		}
		return decimal.Zero, err
	}
	return amount, nil
}

func GetMinAmount(ctx context.Context, u *models.ClientUser, assetID string, startAt, endAt time.Time, minAmount decimal.Decimal, isStart bool) (decimal.Decimal, error) {
	// 4. 获取该用户的 snapshot，遍历最近一天的 snapshot，取 asset 的最低值
	ss, err := common.GetUserSnapshots(ctx, u, assetID, endAt, "DESC", 500)
	if err != nil {
		tools.Println(err)
		return decimal.Zero, err
	}
	if !isStart {
		if len(ss) == 1 {
			return minAmount, nil
		}
		ss = ss[1:]
	}
	for _, s := range ss {
		if s.CreatedAt.Before(startAt) {
			return minAmount, nil
		}
		if s.ClosingBalance.LessThan(minAmount) {
			minAmount = s.ClosingBalance
		}
	}
	return GetMinAmount(ctx, u, assetID, startAt, ss[len(ss)-1].CreatedAt, minAmount, false)
}
