package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Guess struct {
	ClientId  string    `json:"client_id" gorm:"type:varchar(36);not null;"`
	GuessId   string    `json:"guess_id" gorm:"type:varchar(36);not null;"`
	Symbol    string    `json:"symbol" gorm:"type:varchar;not null;"`
	AssetID   string    `json:"asset_id" gorm:"type:varchar(36);not null;"`
	PriceUsd  string    `json:"price_usd" gorm:"type:varchar;not null;"`
	Rules     string    `json:"rules" gorm:"type:varchar;not null;"`
	Explain   string    `json:"explain" gorm:"type:varchar;not null;"`
	StartTime string    `json:"start_time" gorm:"type:varchar;not null;"`
	EndTime   string    `json:"end_time" gorm:"type:varchar;not null;"`
	StartAt   time.Time `json:"start_at" gorm:"type:timestamp;not null;"`
	EndAt     time.Time `json:"end_at" gorm:"type:timestamp;not null;"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:now();"`
}

func (Guess) TableName() string {
	return "guess"
}

type GuessRecord struct {
	GuessId   string `json:"guess_id" gorm:"primary_key;type:varchar(36);not null;"`
	UserId    string `json:"user_id" gorm:"primary_key;type:varchar(36);not null;"`
	Date      string `json:"date" gorm:"primary_key;type:date;not null;"`
	GuessType int    `json:"guess_type" gorm:"type:smallint;not null;"`
	Result    int    `json:"result" gorm:"type:smallint;not null;default:0;"`
}

func (GuessRecord) TableName() string {
	return "guess_record"
}

type GuessResult struct {
	AssetID string          `json:"asset_id" gorm:"type:varchar(36);not null;"`
	Price   decimal.Decimal `json:"price" gorm:"type:varchar;not null;"`
	Date    time.Time       `json:"date" gorm:"type:date;default:now();"`
}

func (GuessResult) TableName() string {
	return "guess_result"
}
