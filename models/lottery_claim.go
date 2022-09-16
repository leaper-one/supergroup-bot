package models

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
)

const claim_DDL = `
CREATE TABLE IF NOT EXISTS claim (
	user_id   VARCHAR(36) NOT NULL,
	date 		  DATE DEFAULT NOW(),
	ua 				VARCHAR DEFAULT '',
	addr			VARCHAR DEFAULT '',
	client_id VARCHAR DEFAULT '',
	PRIMARY KEY (user_id, date)
);
`

type Claim struct {
	UserID string    `json:"user_id"`
	Date   time.Time `json:"date"`
	UA     string    `json:"ua"`
}

type CliamPageResp struct {
	LastLottery     []LotteryRecord `json:"last_lottery"`
	LotteryList     []LotteryList   `json:"lottery_list"`
	Power           Power           `json:"power"`               // 当前能量 times
	IsClaim         bool            `json:"is_claim"`            // 是否已经签到
	Count           int             `json:"count"`               // 本周签到天数
	InviteCount     int64           `json:"invite_count"`        // 邀请人数
	Receiving       *LotteryRecord  `json:"receiving,omitempty"` // receviing 抽奖了没有领
	DoubleClaimList []*Client       `json:"double_claim_list"`   // 双倍签到
}

func GetClaimAndLotteryInitData(ctx context.Context, u *ClientUser) (*CliamPageResp, error) {
	doubleClaimList := make([]*Client, 0)
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
	resp := &CliamPageResp{
		LastLottery:     getLastLottery(ctx),
		LotteryList:     getLotteryList(ctx, u),
		Power:           getPower(ctx, u.UserID),
		IsClaim:         checkIsClaim(ctx, u.UserID),
		Count:           getWeekClaimDay(ctx, u.UserID),
		Receiving:       getReceivingLottery(ctx, u.UserID),
		InviteCount:     getInviteCountByUserID(ctx, u.UserID),
		DoubleClaimList: doubleClaimList,
	}
	return resp, nil
}

func PostClaim(ctx context.Context, u *ClientUser) error {
	if CheckIsBlockUser(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if checkIsClaim(ctx, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	isVip := checkUserIsVIP(ctx, u.UserID)
	isDouble := checkIsDoubleClaimClient(ctx, u.ClientID)

	if err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. 创建一个 claim
		if err := createClaim(ctx, tx, u); err != nil {
			return err
		}
		var addPower decimal.Decimal
		if isVip {
			addPower = decimal.NewFromInt(10)
		} else {
			addPower = decimal.NewFromInt(5)
		}
		if isDouble {
			addPower = addPower.Mul(decimal.NewFromInt(2))
		}

		// 2. 创建一条 power_record
		if err := createPowerRecord(ctx, tx, u.UserID, PowerTypeClaim, addPower); err != nil {
			return err
		}
		// 2.1 如果签到 4 天，则这一次加上额外的。
		if needAddExtraPower(ctx, u.UserID) {
			var extraPower decimal.Decimal
			if isVip {
				extraPower = decimal.NewFromInt(50)
			} else {
				extraPower = decimal.NewFromInt(25)
			}
			if err := createPowerRecord(ctx, tx, u.UserID, PowerTypeClaimExtra, extraPower); err != nil {
				return err
			}
			addPower = addPower.Add(extraPower)
		}
		// 3. 处理邀请奖励
		err := handleInvitationClaim(ctx, tx, u.UserID, isVip)
		if err != nil {
			return err
		}
		// 4. 更新 power balance
		if err := updatePowerBalanceWithAmount(ctx, tx, u.UserID, addPower); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func PostExchangeLottery(ctx context.Context, u *ClientUser) error {
	pow := getPower(ctx, u.UserID)
	if pow.Balance.LessThan(decimal.NewFromInt(100)) {
		return session.ForbiddenError(ctx)
	}
	b := pow.Balance.Sub(decimal.NewFromInt(100))
	return session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := createPowerRecord(ctx, tx, u.UserID, PowerTypeLottery, decimal.NewFromInt(-100)); err != nil {
			return err
		}
		if err := updatePower(ctx, tx, u.UserID, b.String(), pow.LotteryTimes+1); err != nil {
			return err
		}
		return nil
	})
}

func createClaim(ctx context.Context, tx pgx.Tx, u *ClientUser) error {
	query := durable.InsertQuery("claim", "client_id,user_id,ua,addr")
	req := session.Request(ctx)
	_, err := tx.Exec(ctx, query, u.ClientID, u.UserID, req.Header.Get("User-Agent"), req.RemoteAddr)
	return err
}

func checkIsClaim(ctx context.Context, userID string) bool {
	var count int
	if err := session.Database(ctx).QueryRow(ctx, "SELECT count(1) FROM claim WHERE user_id=$1 AND date=current_date", userID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func getWeekClaimDay(ctx context.Context, userID string) int {
	var count int
	if err := session.Database(ctx).QueryRow(
		ctx,
		fmt.Sprintf("SELECT count(1) FROM claim WHERE user_id=$1 AND date >= CURRENT_DATE-%d", getFirstDateOffsetOfWeek()),
		userID,
	).Scan(&count); err != nil {
		session.Logger(ctx).Println(err)
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

func getYesterdayClaim(ctx context.Context) (int, int, error) {
	var vipAmount int
	var normalAmount int
	if err := session.Database(ctx).QueryRow(ctx, "SELECT count(1) FROM power_record WHERE to_char(created_at, 'YYYY-MM-DD')= to_char(current_date-1, 'YYYY-MM-DD')  AND amount='10'").Scan(&vipAmount); err != nil {
		return 0, 0, err
	}
	if err := session.Database(ctx).QueryRow(ctx, "SELECT count(1) FROM power_record WHERE to_char(created_at, 'YYYY-MM-DD')= to_char(current_date-1, 'YYYY-MM-DD')  AND amount='5'").Scan(&normalAmount); err != nil {
		return 0, 0, err
	}
	return normalAmount + vipAmount, vipAmount, nil
}

func getUserTotalPower(ctx context.Context, userID string) (int, error) {
	var amount int
	err := session.Database(ctx).QueryRow(ctx, `
SELECT coalesce(SUM(amount::integer),0) FROM power_record 
WHERE user_id=$1 AND power_type='invitation'
`, userID).Scan(&amount)
	return amount, err
}

var ignoreDoubleList = make(map[string]bool)

func checkIsIgnoreDoubleClaim(ctx context.Context, clientID string) bool {
	if len(ignoreDoubleList) == 0 {
		ignoreList, err := session.Redis(ctx).QSMembers(ctx, "double_ignore")
		if err != nil {
			session.Logger(ctx).Println(err)
			return true
		}
		for _, v := range ignoreList {
			ignoreDoubleList[v] = true
		}
	}
	return ignoreDoubleList[clientID]
}
