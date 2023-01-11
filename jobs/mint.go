package jobs

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func StartMintJob() {
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("55 1 * * *", func() {
		handleMintStatistic(models.Ctx)
	})
	if err != nil {
		tools.Println(err)
		tools.SendMsgToDeveloper("定时任务StartMintJob。。。出问题了。。。")
		return
	}
	c.Start()
}

func handleMintStatistic(ctx context.Context) {
	ms := make([]*models.LiquidityMining, 0)
	err := session.DB(ctx).Find(&ms).Error
	if err != nil {
		tools.Println(err)
		return
	}
	if len(ms) == 0 {
		return
	}
	log.Println("start mint job")
	for _, m := range ms {
		// 1. 还没到 first_time 结束
		if m.FirstTime.After(time.Now()) {
			continue
		}
		// 2. 到了first_time，没到 first_end，这个时候是头矿奖励
		if m.FirstEnd.After(time.Now()) {
			// 处理头矿奖励
			if err := handleStatisticsAssets(ctx, m, models.LiquidityMiningFirst); err != nil {
				tools.Println(err)
			}
			continue
		}
		// 3. 到了first_end，没到 daily_time 结束
		if m.DailyTime.After(time.Now()) {
			continue
		}
		// 4. 到了daily_time，没到 daily_end，这个时候是每日奖励
		if m.DailyEnd.After(time.Now()) {
			// 处理挖矿奖励
			if err := handleStatisticsAssets(ctx, m, models.LiquidityMiningDaily); err != nil {
				tools.Println(err)
			}
			continue
		}
		// 5. 到了daily_end，结束
	}
}

func handleStatisticsAssets(ctx context.Context, m *models.LiquidityMining, mintStatus int) error {
	// 获取参与活动的用户
	users, err := GetLiquidityMiningUsersByID(ctx, m.ClientID, m.MiningID)
	if err != nil {
		return err
	}
	// 获取流动性资产
	lpAssets, err := common.GetClientAssetLPCheckMapByID(ctx, m.ClientID)
	if err != nil {
		return err
	}
	assetReward := m.FirstAmount
	extraReward := m.ExtraFirstAmount

	if mintStatus == models.LiquidityMiningDaily {
		assetReward = m.DailyAmount
		extraReward = m.ExtraDailyAmount
	}
	totalAmount, usersAmount, tmpData := statisticsUsersPartAndTotalAmount(ctx, m.MiningID, users, lpAssets)
	if totalAmount.IsZero() {
		tools.Println(m.MiningID + "... totalAmount is zero...")
		return nil
	}
	for userID, v := range usersAmount {
		if v.IsZero() {
			continue
		}
		// 份额
		part := v.Div(totalAmount)
		assetRewardAmount := assetReward.Mul(part).Truncate(8)
		extraAssetRewardAmount := extraReward.Mul(part).Truncate(8)
		if assetRewardAmount.IsZero() && extraAssetRewardAmount.IsZero() {
			continue
		}
		if err := models.RunInTransaction(ctx, func(tx *gorm.DB) error {
			recordID := tools.GetUUID()
			if !assetRewardAmount.IsZero() {
				if err := tx.Create(&models.LiquidityMiningTx{
					RecordID: recordID,
					MiningID: m.MiningID,
					AssetID:  m.RewardAssetID,
					UserID:   userID,
					Amount:   assetRewardAmount,
					TraceID:  tools.GetUUID(),
					Status:   models.LiquidityMiningRecordStatusPending,
				}).Error; err != nil {
					return err
				}
			}
			if m.ExtraAssetID != "" && !extraAssetRewardAmount.IsZero() {
				if err := tx.Create(&models.LiquidityMiningTx{
					RecordID: recordID,
					MiningID: m.MiningID,
					AssetID:  m.ExtraAssetID,
					UserID:   userID,
					Amount:   extraAssetRewardAmount,
					TraceID:  tools.GetUUID(),
					Status:   models.LiquidityMiningRecordStatusPending,
				}).Error; err != nil {
					return err
				}
			}

			for _, v := range tmpData[userID] {
				if err := tx.Create(&models.LiquidityMiningRecord{
					RecordID: recordID,
					MiningID: m.MiningID,
					UserID:   userID,
					AssetID:  v.AssetID,
					Amount:   v.Amount,
					Profit:   v.Profit,
				}).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			tools.Println(err)
			continue
		}
	}
	return nil
}

type tmpRecordData struct {
	AssetID string
	Amount  decimal.Decimal
	Profit  decimal.Decimal
	AddPart decimal.Decimal
}

func statisticsUsersPartAndTotalAmount(ctx context.Context, mintID string, users []*models.ClientUser, lpAssets map[string]decimal.Decimal) (decimal.Decimal, map[string]decimal.Decimal, map[string][]*tmpRecordData) {
	// 统计每个用户的流动性资产
	totalAmount := decimal.Zero
	usersAmount := make(map[string]decimal.Decimal)
	records := make(map[string][]*tmpRecordData)
	for _, u := range users {
		userAssets, err := common.GetUserAssets(ctx, u)
		if err != nil {
			if strings.Contains(err.Error(), "Forbidden") {
				// 取消授权的用户，添加一条未参与的记录
				if err := session.DB(ctx).Create(&models.LiquidityMiningTx{
					RecordID: tools.GetUUID(),
					MiningID: mintID,
					UserID:   u.UserID,
					AssetID:  "",
					Amount:   decimal.Zero,
					Status:   models.LiquidityMiningRecordStatusFailed,
					TraceID:  tools.GetUUID(),
				}).Error; err != nil {
					tools.Println(err)
				}
				continue
			}
			tools.Println(err)
			continue
		}
		// 检查流动性资产
		for _, a := range userAssets {
			if a.Balance.IsZero() {
				continue
			}
			if price, ok := lpAssets[a.AssetID]; ok {
				if price.IsZero() {
					a, err := common.GetAssetByID(ctx, nil, a.AssetID)
					if err != nil {
						tools.Println(err)
						continue
					}
					price = a.PriceUsd
				}
				addPart := a.Balance.Mul(price)
				// 用户的分数 和 总分数加
				if _, ok := usersAmount[u.UserID]; !ok {
					usersAmount[u.UserID] = decimal.Zero
					records[u.UserID] = make([]*tmpRecordData, 0)
				}
				usersAmount[u.UserID] = usersAmount[u.UserID].Add(addPart)
				totalAmount = totalAmount.Add(addPart)
				records[u.UserID] = append(records[u.UserID], &tmpRecordData{
					AssetID: a.AssetID,
					Amount:  a.Balance,
					AddPart: addPart,
				})
			}
		}
		if records[u.UserID] != nil {
			for _, v := range records[u.UserID] {
				v.Profit = v.AddPart.Div(usersAmount[u.UserID])
			}
		}
	}
	return totalAmount, usersAmount, records
}

func GetLiquidityMiningUsersByID(ctx context.Context, clientID, miningID string) ([]*models.ClientUser, error) {
	m := make([]*models.ClientUser, 0)
	err := session.DB(ctx).Find(&m,
		"user_id IN (SELECT user_id FROM liquidity_mining_users WHERE mining_id=?) AND client_id=?",
		miningID, clientID).Error
	return m, err
}
