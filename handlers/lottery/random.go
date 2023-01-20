package lottery

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
)

// 获取随机的抽奖奖励
func GetRandomLottery(ctx context.Context, u *models.ClientUser) *config.Lottery {
	// 通过转账获取一个随机数
	rand.Seed(time.Now().UnixNano())
	random := decimal.NewFromInt(int64(rand.Intn(10000)))
	for lotteryID, rate := range config.Config.Lottery.Rate {
		if random.LessThanOrEqual(rate) {
			return getLotteryByID(ctx, lotteryID, u.UserID)
		} else {
			random = random.Sub(rate)
		}
	}
	tools.Println("get random lottery error")
	return &config.Config.Lottery.List[0]
}

func getLotteryByID(ctx context.Context, id, userID string) *config.Lottery {
	list := getUserListingLottery(ctx, userID)
	i, _ := strconv.Atoi(id)
	return &list[i]
}
