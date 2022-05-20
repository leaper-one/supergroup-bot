package models

import (
	"context"
	"errors"
	"time"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const voucher_DDL = `
CREATE TABLE IF NOT EXISTS voucher (
	code varchar(8) NOT NULL PRIMARY KEY,
	client_id varchar(36) DEFAULT '',
	user_id varchar(36) DEFAULT '',
	status int2 NOT NULL DEFAULT 1,
	updated_at timestamptz NOT NULL DEFAULT now(),
	created_at timestamptz NULL DEFAULT now()
);
`

type Voucher struct {
	Code      string    `json:"code"`
	ClientID  string    `json:"client_id"`
	Status    int       `json:"status"` // 1: unused, 2: used
	UserID    string    `json:"user_id"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// -1: limit, 0: not found, 2: used 3: success
func CheckVoucher(ctx context.Context, u *ClientUser, code string) (int, error) {
	if checkIsBlockUser(ctx, u.ClientID, u.UserID) {
		return -1, session.ForbiddenError(ctx)
	}
	date := time.Now().Format("2006-01-02")
	uKey := u.UserID + ":" + date
	b, err := session.Redis(ctx).QIncr(ctx, uKey, time.Hour*24)
	if err != nil {
		return 0, err
	}
	if b >= 10 {
		// return -1, nil
	}
	err = session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var status int
		if err := tx.QueryRow(ctx, "SELECT status FROM voucher WHERE code = $1", code).Scan(&status); err != nil {
			return err
		}
		if status == 2 {
			return errors.New("voucher used")
		}
		if err := createPowerRecord(ctx, tx, u.UserID, PowerTypeVoucher, decimal.NewFromInt(100)); err != nil {
			return err
		}
		if err := updatePowerBalanceWithAmount(ctx, tx, u.UserID, decimal.NewFromInt(100)); err != nil {
			return err
		}
		data, err := tx.Exec(ctx, `UPDATE voucher SET status=2, user_id=$1, client_id=$2, updated_at=now() WHERE code = $3`, u.UserID, u.ClientID, code)
		if err != nil {
			return err
		}
		if data.RowsAffected() == 1 {
			return nil
		}
		return errors.New("voucher used")
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		if err.Error() == "voucher used" {
			return 2, nil
		}
		return 0, err
	}
	return 3, nil
}
