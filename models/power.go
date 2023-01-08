package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Claim struct {
	UserID   string `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Date     string `json:"date,omitempty" gorm:"primary_key;type:date;default:now();"`
	UA       string `json:"ua,omitempty" gorm:"type:varchar;default:'';"`
	Addr     string `json:"addr,omitempty" gorm:"type:varchar;default:'';"`
	ClientID string `json:"client_id,omitempty" gorm:"type:varchar;default:'';"`
}

func (Claim) TableName() string {
	return "claim"
}

type Power struct {
	UserID       string          `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Balance      decimal.Decimal `json:"balance" gorm:"type:varchar;default:'0';"`
	LotteryTimes int             `json:"lottery_times" gorm:"type:integer;default:0;"`
}

func (Power) TableName() string {
	return "power"
}

type PowerRecord struct {
	UserID    string          `json:"user_id" gorm:"type:varchar(36);not null;"`
	PowerType string          `json:"power_type" gorm:"type:varchar(128);not null;"`
	Amount    decimal.Decimal `json:"amount" gorm:"type:varchar;default:'0';"`
	CreatedAt time.Time       `json:"created_at" gorm:"type:timestamp;default:now();"`

	Date string `json:"date" gorm:"-"`
}

func (PowerRecord) TableName() string {
	return "power_record"
}

const (
	PowerTypeClaim      = "claim"
	PowerTypeClaimExtra = "claim_extra"
	PowerTypeLottery    = "lottery"
	PowerTypeInvitation = "invitation"
	PowerTypeVoucher    = "voucher"
)

type PowerExtra struct {
	ClientID    string          `json:"client_id" gorm:"primary_key;type:varchar(36);not null;"`
	Description string          `json:"description" gorm:"type:varchar;default:'';"`
	Multiplier  decimal.Decimal `json:"multiplier" gorm:"type:varchar;default:'2';"`
	StartAt     time.Time       `json:"start_at" gorm:"type:date;default:'1970-01-01';"`
	EndAt       time.Time       `json:"end_at" gorm:"type:date;default:'1970-01-01';"`
	CreatedAt   time.Time       `json:"created_at" gorm:"type:timestamp with time zone;default:now();"`
}

func (PowerExtra) TableName() string {
	return "power_extra"
}
