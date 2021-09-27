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
	PRIMARY KEY (user_id, date)
);
alter table claim add if not exists ua varchar DEFAULT '';
`

const power_DDL = `
CREATE TABLE IF NOT EXISTS power (
	user_id   VARCHAR(36) NOT NULL PRIMARY KEY,
	balance   VARCHAR NOT NULL DEFAULT '0',
	lottery_times 	 INTEGER NOT NULL DEFAULT 0
);
`

const power_record_DDL = `
CREATE TABLE IF NOT EXISTS power_record (
	power_type VARCHAR(128) NOT NULL,
	user_id    VARCHAR(36) NOT NULL,
	amount 	   VARCHAR NOT NULL DEFAULT '0',
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

type Claim struct {
	UserID string    `json:"user_id"`
	Date   time.Time `json:"date"`
	UA     string    `json:"ua"`
}

type Power struct {
	UserID       string          `json:"user_id,omitempty"`
	Balance      decimal.Decimal `json:"balance"`
	LotteryTimes int             `json:"lottery_times"`
}

type PowerRecord struct {
	UserID    string    `json:"user_id"`
	PowerType string    `json:"power_type"`
	Amount    string    `json:"amount"`
	CreatedAt time.Time `json:"created_at"`

	Date string `json:"date"`
}

const (
	PowerTypeClaim      = "claim"
	PowerTypeClaimExtra = "claim_extra"
	PowerTypeLottery    = "lottery"
)

type CliamPageResp struct {
	LastLottery []LotteryRecord  `json:"last_lottery"`
	LotteryList []config.Lottery `json:"lottery_list"`
	Power       Power            `json:"power"`               // 当前能量 times
	IsClaim     bool             `json:"is_claim"`            // 是否已经签到
	Count       int              `json:"count"`               // 本周签到天数
	Receiving   *LotteryRecord   `json:"receiving,omitempty"` // receviing 抽奖了没有领
}

func GetClaimAndLotteryInitData(ctx context.Context, u *ClientUser) (*CliamPageResp, error) {
	resp := &CliamPageResp{
		LastLottery: getLastLottery(ctx),
		LotteryList: getLotteryList(ctx),
		Power:       getPower(ctx, u.UserID),
		IsClaim:     checkIsClaim(ctx, u.UserID),
		Count:       getWeekClaimDay(ctx, u.UserID),
		Receiving:   getReceivingLottery(ctx, u.UserID),
	}
	return resp, nil
}

func PostClaim(ctx context.Context, u *ClientUser) error {
	if checkIsBlockUser(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if checkIsClaim(ctx, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	isVip := checkUserIsVIP(ctx, u.UserID)
	if err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. 创建一个 claim
		if err := createClaim(ctx, tx, u.UserID); err != nil {
			return err
		}
		var addPower decimal.Decimal
		if isVip {
			addPower = decimal.NewFromInt(10)
		} else {
			addPower = decimal.NewFromInt(5)
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
		// 3. 拿到 power_balance
		pow := getPower(ctx, u.UserID)
		balance := pow.Balance.Add(addPower)
		// 4. 更新 power_balance
		if err := updatePower(ctx, tx, u.UserID, balance.String(), pow.LotteryTimes); err != nil {
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

func GetPowerRecordList(ctx context.Context, u *ClientUser, page int) ([]PowerRecord, error) {
	if page < 1 {
		page = 1
	}
	var list []PowerRecord
	if err := session.Database(ctx).ConnQuery(ctx,
		"SELECT power_type, amount, to_char(created_at, 'YYYY-MM-DD') AS created_at FROM power_record WHERE user_id = $1 ORDER BY created_at DESC OFFSET $2 LIMIT 20",
		func(rows pgx.Rows) error {
			for rows.Next() {
				var r PowerRecord
				if err := rows.Scan(&r.PowerType, &r.Amount, &r.Date); err != nil {
					return err
				}
				list = append(list, r)
			}
			return nil
		}, u.UserID, (page-1)*20); err != nil {
		return nil, err
	}
	return list, nil
}

func createClaim(ctx context.Context, tx pgx.Tx, userID string) error {
	query := durable.InsertQuery("claim", "user_id,ua")
	_, err := tx.Exec(ctx, query, userID, session.Request(ctx).Header.Get("User-Agent"))
	return err
}

func checkIsClaim(ctx context.Context, userID string) bool {
	var count int
	if err := session.Database(ctx).QueryRow(ctx, "SELECT count(1) FROM claim WHERE user_id=$1 AND date=current_date", userID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func createPowerRecord(ctx context.Context, tx pgx.Tx, userID, powerType string, amount decimal.Decimal) error {
	query := durable.InsertQuery("power_record", "user_id,power_type,amount")
	_, err := tx.Exec(ctx, query, userID, powerType, amount)
	return err
}

func getPower(ctx context.Context, userID string) Power {
	var p Power
	if err := session.Database(ctx).QueryRow(ctx, "SELECT balance, lottery_times FROM power WHERE user_id=$1", userID).Scan(&p.Balance, &p.LotteryTimes); err != nil {
		if err == pgx.ErrNoRows {
			_, err := session.Database(ctx).Exec(ctx, durable.InsertQuery("power", "user_id,balance,lottery_times"), userID, "0", 1)
			if err != nil {
				session.Logger(ctx).Println(err)
			}
			return p
		}
	}
	return p
}

func updatePower(ctx context.Context, tx pgx.Tx, userID, balance string, times int) error {
	_, err := tx.Exec(ctx, "UPDATE power SET balance=$1, lottery_times=$2 WHERE user_id=$3", balance, times, userID)
	return err
}
func getPowerRecord(ctx context.Context, userID string) (string, error) {
	var amount string
	if err := session.Database(ctx).QueryRow(ctx, "SELECT balance FROM power_record WHERE user_id=$1", userID).Scan(&amount); err != nil {
		return "", err
	}
	return amount, nil
}

func getWeekClaimDay(ctx context.Context, userID string) int {
	var count int
	if err := session.Database(ctx).QueryRow(ctx, "SELECT count(1) FROM claim WHERE user_id=$1 AND date >= $2", userID, getFirstDateOfWeek()).Scan(&count); err != nil {
		session.Logger(ctx).Println(err)
		return 0
	}
	return count
}

func getFirstDateOfWeek() (weekMonday string) {
	now := time.Now()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	weekStartDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, offset)
	weekMonday = weekStartDate.Format("2006-01-02")
	return
}

func needAddExtraPower(ctx context.Context, userID string) bool {
	passDays := int(time.Now().Weekday())
	if config.Config.Lang == "zh" {
		if passDays == 0 {
			passDays = 7
		}
		if passDays < 5 {
			return false
		}

		var count int
		if err := session.Database(ctx).QueryRow(ctx,
			fmt.Sprintf("SELECT count(1) FROM claim WHERE user_id=$1 AND date>CURRENT_DATE-%d", passDays),
			userID,
		).Scan(&count); err != nil {
			session.Logger(ctx).Println(err)
			return false
		}
		return count == 4
	} else {
		if passDays < 4 {
			return false
		}
		var count int
		if err := session.Database(ctx).QueryRow(ctx,
			fmt.Sprintf("SELECT count(1) FROM claim WHERE user_id=$1 AND date>CURRENT_DATE-%d", passDays+1),
			userID,
		).Scan(&count); err != nil {
			session.Logger(ctx).Println(err)
			return false
		}
		return count == 4
	}
}
