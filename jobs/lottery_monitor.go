package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/handlers/lottery"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

func DailyStatisticMsg() {
	// 每天凌晨0点统计资产
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("0 0 * * *", func() {
		LotteryStatistic(models.Ctx)
	})
	if err != nil {
		tools.Println(err)
	}
	c.Start()
}

func LotteryStatistic(ctx context.Context) {
	c := lottery.GetLotteryClient()
	if c == nil {
		return
	}
	yesterdayFinishedAssetMap := getYesterdaySendReward(ctx)
	times := getYesterdayLotteryTimes(ctx)
	var sendStr string
	sendStr += fmt.Sprintf("昨日抽奖次数:%d\n", times)
	totalClaim, vipClaim, err := getYesterdayClaim(ctx)
	if err == nil {
		sendStr += fmt.Sprintf("昨日会员签到人数:%d\n", vipClaim)
		sendStr += fmt.Sprintf("昨日总签到人数:%d\n", totalClaim)
	}
	sendStr += "\n昨日发放奖励\n"
	for assetID, amount := range yesterdayFinishedAssetMap {
		a, _ := common.GetAssetByID(ctx, nil, assetID)
		sendStr += fmt.Sprintf("%s:%s\n", a.Symbol, amount.String())
	}
	as, err := lottery.GetLotteryClient().ReadAssets(ctx)
	if err == nil {
		sendStr += "\n机器人余额\n"
		asMap := make(map[string]*mixin.Asset)
		for _, a := range as {
			asMap[a.AssetID] = a
		}
		lotteryList := getLotteryAssetMaxRewardMap(ctx)
		for assetID, maxReward := range lotteryList {
			a := asMap[assetID]
			if a != nil {
				sendStr += fmt.Sprintf("%s:%s\n", a.Symbol, a.Balance)
				if a.Balance.LessThan(maxReward.Mul(decimal.NewFromInt(config.NoticeLotteryTimes))) {
					sendStr += fmt.Sprintf("不足最大份数的%d倍，请及时充值...\n", config.NoticeLotteryTimes)
				}
			} else {
				a, _ := common.GetAssetByID(ctx, nil, assetID)
				sendStr += fmt.Sprintf("%s:0\n", a.Symbol)
				sendStr += fmt.Sprintf("不足最大份数的%d倍，请及时充值...\n", config.NoticeLotteryTimes)
			}
		}
	} else {
		tools.Println(err)
	}
	tools.SendMonitorGroupMsg(sendStr)
}

func getYesterdayLotteryTimes(ctx context.Context) int64 {
	var times int64
	if err := session.DB(ctx).Table("lottery_record").
		Where("created_at between CURRENT_DATE-1 and CURRENT_DATE").
		Count(&times).Error; err != nil {
		tools.Println(err)
	}
	return times
}

func getYesterdaySendReward(ctx context.Context) map[string]decimal.Decimal {
	finishedAssetMap := make(map[string]decimal.Decimal)
	var records []struct {
		AssetID string
		Amount  decimal.Decimal
	}

	if err := session.DB(ctx).Table("lottery_record").
		Select("asset_id, COALESCE(SUM(amount::decimal),0) AS amount").
		Where("is_received = true AND created_at between CURRENT_DATE-1 and CURRENT_DATE").
		Group("asset_id").
		Scan(&records).Error; err != nil {
		tools.Println(err)
	}

	for _, r := range records {
		finishedAssetMap[r.AssetID] = r.Amount
	}

	return finishedAssetMap
}

// 获取奖品列表及最大的奖品数量
func getLotteryAssetMaxRewardMap(ctx context.Context) map[string]decimal.Decimal {
	res := make(map[string]decimal.Decimal)
	for _, l := range config.Config.Lottery.List {
		if res[l.AssetID] == decimal.Zero {
			res[l.AssetID] = l.Amount
		} else {
			if res[l.AssetID].LessThan(l.Amount) {
				res[l.AssetID] = l.Amount
			}
		}
	}
	list, err := lottery.GetLotterySupply(ctx)
	if err == nil && len(list) > 0 {
		for _, l := range list {
			res[l.AssetID] = l.Amount
		}
	}
	return res
}

func getYesterdayClaim(ctx context.Context) (int64, int64, error) {
	var vipAmount int64
	var normalAmount int64
	if err := session.DB(ctx).Table("power_record").
		Where("to_char(created_at, 'YYYY-MM-DD')= to_char(current_date-1, 'YYYY-MM-DD')  AND amount='10'").
		Count(&vipAmount).Error; err != nil {
		return 0, 0, err
	}
	if err := session.DB(ctx).Table("power_record").
		Where("to_char(created_at, 'YYYY-MM-DD')= to_char(current_date-1, 'YYYY-MM-DD')  AND amount='5'").
		Count(&normalAmount).Error; err != nil {
		return 0, 0, err
	}
	return normalAmount + vipAmount, vipAmount, nil
}
