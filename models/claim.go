package models

import (
	"context"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const claim_DDL = `
CREATE TABLE IF NOT EXISTS claim (
	user_id   VARCHAR(36) NOT NULL,
	date 		  DATE NOT NULL DEFAULT NOW(),
	PRIMARY KEY (user_id, date)
);
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

type CliamPageResp struct {
	LastLottery []LotteryRecord `json:"last_lottery,omitempty"`
	LotteryList []Lottery       `json:"lottery_list"`
	Power       Power           `json:"power"`
	IsClaim     bool            `json:"is_claim"` // 是否已经签到
	Count       int             `json:"count"`    // 本周签到天数
	Receiving   *LotteryRecord  `json:"receiving,omitempty"`
}

func GetClaimAndLotteryInitData(ctx context.Context, u *ClientUser) (*CliamPageResp, error) {
	resp := &CliamPageResp{
		LastLottery: getLastLottery(ctx),
		LotteryList: getLotteryList(ctx),
		Power:       getPower(ctx, u.UserID),
		IsClaim:     checkIsClaim(ctx, u.UserID),
		Count:       getWeekClaimDay(ctx, u.UserID),
		Receiving:   getReceivingLottery(ctx),
	}
	return resp, nil
}

func PostClaim(ctx context.Context, u *ClientUser) error {
	if checkIsClaim(ctx, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. 创建一个 claim
		if err := createClaim(ctx, tx, u.UserID); err != nil {
			return err
		}
		// 2. 创建一条 power_record
		if err := createPowerRecord(ctx, tx, u.UserID, "claim", "10"); err != nil {
			return err
		}
		// 3. 拿到 power_balance
		pow := getPower(ctx, u.UserID)
		balance := pow.Balance.Add(decimal.NewFromInt(10))
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
		if err := createPowerRecord(ctx, tx, u.UserID, "lottery", "-100"); err != nil {
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
	query := durable.InsertQuery("claim", "user_id")
	_, err := tx.Exec(ctx, query, userID)
	return err
}

func checkIsClaim(ctx context.Context, userID string) bool {
	now := time.Now()
	nowStr := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
	var count int
	if err := session.Database(ctx).QueryRow(ctx, "SELECT count(1) FROM claim WHERE user_id=$1 AND date=$2", userID, nowStr).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func createPowerRecord(ctx context.Context, tx pgx.Tx, userID, powerType, amount string) error {
	query := durable.InsertQuery("power_record", "user_id,power_type,amount")
	_, err := tx.Exec(ctx, query, userID, powerType, amount)
	return err
}

func getPower(ctx context.Context, userID string) Power {
	var p Power
	if err := session.Database(ctx).QueryRow(ctx, "SELECT balance, lottery_times FROM power WHERE user_id=$1", userID).Scan(&p.Balance, &p.LotteryTimes); err != nil {
		if err == pgx.ErrNoRows {
			_, err := session.Database(ctx).Exec(ctx, durable.InsertQuery("power", "user_id,balance"), userID, "0")
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
