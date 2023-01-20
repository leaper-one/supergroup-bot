package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type ClientAssetLevel struct {
	ClientID    string          `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Fresh       decimal.Decimal `json:"fresh,omitempty" gorm:"type:varchar;default:'0';"`
	Senior      decimal.Decimal `json:"senior,omitempty" gorm:"type:varchar;default:'0';"`
	Large       decimal.Decimal `json:"large,omitempty" gorm:"type:varchar;default:'0';"`
	FreshAmount decimal.Decimal `json:"fresh_amount,omitempty" gorm:"type:varchar;default:'0';"`
	LargeAmount decimal.Decimal `json:"large_amount,omitempty" gorm:"type:varchar;default:'0';"`
	CreatedAt   time.Time       `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (ClientAssetLevel) TableName() string {
	return "client_asset_level"
}
