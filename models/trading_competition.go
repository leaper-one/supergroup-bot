package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type TradingCompetition struct {
	CompetitionID string          `json:"competition_id" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID      string          `json:"client_id" gorm:"type:varchar(36);not null;"`
	AssetID       string          `json:"asset_id" gorm:"type:varchar(36);not null;"`
	Amount        decimal.Decimal `json:"amount" gorm:"type:varchar;not null;"`
	Title         string          `json:"title" gorm:"type:varchar(255);not null;"`
	Tips          string          `json:"tips" gorm:"type:varchar(255);not null;"`
	Rules         string          `json:"rules" gorm:"type:varchar(255);not null;"`
	Reward        string          `json:"reward" gorm:"type:varchar;not null;"`
	StartAt       time.Time       `json:"start_at" gorm:"type:date;not null;"`
	EndAt         time.Time       `json:"end_at" gorm:"type:date;not null;"`
	CreatedAt     time.Time       `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
}

func (TradingCompetition) TableName() string {
	return "trading_competition"
}

type UserSnapshot struct {
	SnapshotID     string          `json:"snapshot_id" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID       string          `json:"client_id" gorm:"type:varchar(36);not null;"`
	UserID         string          `json:"user_id" gorm:"type:varchar(36);not null;"`
	OpponentID     string          `json:"opponent_id" gorm:"type:varchar(36);not null;"`
	AssetID        string          `json:"asset_id" gorm:"type:varchar(36);not null;"`
	OpeningBalance decimal.Decimal `json:"opening_balance,omitempty" gorm:"type:varchar;"`
	ClosingBalance decimal.Decimal `json:"closing_balance,omitempty" gorm:"type:varchar;"`
	Amount         decimal.Decimal `json:"amount" gorm:"type:varchar;not null;"`
	Source         string          `json:"source" gorm:"type:varchar(255);not null;"`
	CreatedAt      time.Time       `json:"created_at" gorm:"type:timestamp;not null;default:now()"`
}

func (UserSnapshot) TableName() string {
	return "user_snapshots"
}

type TradingRank struct {
	CompetitionID string          `json:"competition_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	UserID        string          `json:"user_id" gorm:"primary_key;type:varchar(36);not null;"`
	AssetID       string          `json:"asset_id,omitempty" gorm:"type:varchar(36);not null;"`
	Amount        decimal.Decimal `json:"amount" gorm:"type:varchar;not null;"`
	UpdatedAt     time.Time       `json:"-" gorm:"type:timestamp;not null;default:now()"`

	FullName       string `json:"full_name,omitempty" gorm:"-"`
	Avatar         string `json:"avatar,omitempty" gorm:"-"`
	IdentityNumber string `json:"identity_number,omitempty" gorm:"-"`
}

func (TradingRank) TableName() string {
	return "trading_rank"
}
