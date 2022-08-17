package models

import (
	"context"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

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
	PowerTypeInvitation = "invitation"
	PowerTypeVoucher    = "voucher"
)

func updatePowerBalanceWithAmount(ctx context.Context, tx pgx.Tx, userID string, amount decimal.Decimal) error {
	// 1. 拿到 power_balance
	pow := getPowerWithTx(ctx, tx, userID)
	balance := pow.Balance.Add(amount)
	// 2. 更新 power_balance
	return updatePower(ctx, tx, userID, balance.String(), pow.LotteryTimes)
}

func GetPowerRecordList(ctx context.Context, u *ClientUser, page int) ([]PowerRecord, error) {
	if page < 1 {
		page = 1
	}
	var list []PowerRecord
	if err := session.Database(ctx).ConnQuery(ctx,
		"SELECT power_type, amount, to_char(created_at, 'YYYY-MM-DD') AS date FROM power_record WHERE user_id = $1 ORDER BY created_at DESC OFFSET $2 LIMIT 20",
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

func createPowerRecord(ctx context.Context, tx pgx.Tx, userID, powerType string, amount decimal.Decimal) error {
	query := durable.InsertQuery("power_record", "user_id,power_type,amount")
	_, err := tx.Exec(ctx, query, userID, powerType, amount)
	return err
}

func getPower(ctx context.Context, userID string) Power {
	var p Power
	if err := session.Database(ctx).QueryRow(ctx, "SELECT balance, lottery_times FROM power WHERE user_id=$1", userID).Scan(&p.Balance, &p.LotteryTimes); err != nil {
		if err == pgx.ErrNoRows {
			if checkUserIsVIP(ctx, userID) {
				p.LotteryTimes = 1
			}
			_, err := session.Database(ctx).Exec(ctx, durable.InsertQuery("power", "user_id,balance,lottery_times"), userID, "0", p.LotteryTimes)
			if err != nil {
				session.Logger(ctx).Println(err)
			}
			return p
		}
	}
	return p
}

func getPowerWithTx(ctx context.Context, tx pgx.Tx, userID string) Power {
	var p Power
	if err := tx.QueryRow(ctx, "SELECT balance, lottery_times FROM power WHERE user_id=$1", userID).Scan(&p.Balance, &p.LotteryTimes); err != nil {
		if err == pgx.ErrNoRows {
			if checkUserIsVIP(ctx, userID) {
				p.LotteryTimes = 1
			}
			_, err := tx.Exec(ctx, durable.InsertQuery("power", "user_id,balance,lottery_times"), userID, "0", p.LotteryTimes)
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
