package lottery

import (
	"context"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type LotteryList struct {
	config.Lottery
	Description string          `json:"description"`
	Symbol      string          `json:"symbol"`
	PriceUSD    decimal.Decimal `json:"price_usd"`
}

type ClaimPageResp struct {
	LastLottery     []*LotteryRecordView `json:"last_lottery"`
	LotteryList     []LotteryList        `json:"lottery_list"`
	Power           models.Power         `json:"power"`               // 当前能量 times
	IsClaim         bool                 `json:"is_claim"`            // 是否已经签到
	Count           int64                `json:"count"`               // 本周签到天数
	InviteCount     int64                `json:"invite_count"`        // 邀请人数
	Receiving       *LotteryRecordTx     `json:"receiving,omitempty"` // receiving 抽奖了没有领
	DoubleClaimList []*models.Client     `json:"double_claim_list"`   // 双倍签到
}

func GetClaimAndLotteryInitData(ctx context.Context, u *models.ClientUser) (*ClaimPageResp, error) {
	doubleClaimList := make([]*models.Client, 0)
	_double, err := getDoubleClaimClientList(ctx)
	if err != nil {
		return nil, err
	}
	if !checkIsIgnoreDoubleClaim(ctx, u.ClientID) {
		doubleClaimList = _double
	} else {
		for _, v := range _double {
			if v.ClientID == u.ClientID {
				doubleClaimList = append(doubleClaimList, v)
				break
			}
		}
	}
	resp := &ClaimPageResp{
		LastLottery:     getLastLottery(ctx),
		LotteryList:     getLotteryList(ctx, u),
		Power:           getPower(ctx, u.UserID),
		IsClaim:         common.CheckIsClaim(ctx, u.UserID),
		Count:           getWeekClaimDay(ctx, u.UserID),
		Receiving:       getLotteryRecord(ctx, u.UserID),
		InviteCount:     common.GetInviteCountByUserID(ctx, u.UserID),
		DoubleClaimList: doubleClaimList,
	}
	return resp, nil
}

func getWeekClaimDay(ctx context.Context, userID string) int64 {
	var count int64
	if err := session.DB(ctx).Table("claim").
		Where(fmt.Sprintf("user_id = ? AND date >= CURRENT_DATE - %d", getFirstDateOffsetOfWeek()), userID).
		Count(&count).Error; err != nil {
		tools.Println(err)
		return 0
	}
	return count
}

func getFirstDateOffsetOfWeek() int {
	todayWeekday := time.Now().Weekday()
	if config.Config.Lang == "zh" {
		if todayWeekday == time.Sunday {
			todayWeekday = 7
		}
		return int(todayWeekday) - 1
	} else {
		return int(todayWeekday)
	}
}

func PostClaim(ctx context.Context, u *models.ClientUser) error {
	if common.CheckIsBlockUser(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if common.CheckIsClaim(ctx, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	isVip := CheckUserIsVIP(ctx, u.UserID)

	if err := models.RunInTransaction(ctx, func(tx *gorm.DB) error {
		// 1. 创建一个 claim
		if err := tx.Create(&models.Claim{
			ClientID: u.ClientID,
			UserID:   u.UserID,
			UA:       session.Request(ctx).Header.Get("User-Agent"),
			Addr:     session.Request(ctx).RemoteAddr,
		}).Error; err != nil {
			return err
		}
		var addPower decimal.Decimal
		if isVip {
			addPower = decimal.NewFromInt(10)
		} else {
			addPower = decimal.NewFromInt(5)
		}
		if checkIsDoubleClaimClient(ctx, u.ClientID) {
			addPower = addPower.Mul(decimal.NewFromInt(2))
		}

		// 2. 创建一条 power_record
		if err := tx.Create(&models.PowerRecord{
			UserID:    u.UserID,
			PowerType: models.PowerTypeClaim,
			Amount:    addPower,
		}).Error; err != nil {
			return err
		}
		// 2.1 如果签到 4 天，则这一次加上额外的。
		if needAddExtraPower(tx, u.UserID) {
			var extraPower decimal.Decimal
			if isVip {
				extraPower = decimal.NewFromInt(50)
			} else {
				extraPower = decimal.NewFromInt(25)
			}
			if err := tx.Create(&models.PowerRecord{
				UserID:    u.UserID,
				PowerType: models.PowerTypeClaimExtra,
				Amount:    extraPower,
			}).Error; err != nil {
				return err
			}
			addPower = addPower.Add(extraPower)
		}
		// 3. 处理邀请奖励
		err := HandleInvitationClaim(ctx, tx, u.UserID, u.ClientID, isVip)
		if err != nil {
			return err
		}
		// 4. 更新 power balance
		if err := common.UpdatePower(ctx, tx, u.UserID, addPower, 0, ""); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
