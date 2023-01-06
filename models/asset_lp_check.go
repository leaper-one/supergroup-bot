package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type ClientAssetLpCheck struct {
	ClientID  string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	AssetID   string    `json:"asset_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`

	PriceUsd decimal.Decimal `json:"price_usd,omitempty" gorm:"-"`
}

func (ClientAssetLpCheck) TableName() string {
	return "client_asset_lp_check"
}
