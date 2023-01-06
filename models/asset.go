package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Asset struct {
	AssetID   string          `json:"asset_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	ChainID   string          `json:"chain_id,omitempty" gorm:"type:varchar(36);not null;"`
	IconUrl   string          `json:"icon_url,omitempty" gorm:"type:varchar(1024);not null;"`
	Symbol    string          `json:"symbol,omitempty" gorm:"type:varchar(128);not null;"`
	Name      string          `json:"name,omitempty" gorm:"type:varchar;not null;"`
	PriceUsd  decimal.Decimal `json:"price_usd,omitempty" gorm:"type:varchar;default:'0';"`
	ChangeUsd string          `json:"change_usd,omitempty" gorm:"type:varchar;default:'0';"`
}

func (Asset) TableName() string {
	return "assets"
}

type ExinOtcAsset struct {
	AssetID   string    `json:"asset_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	OtcID     string    `json:"otc_id,omitempty" gorm:"type:varchar;not null;"`
	Exchange  string    `json:"exchange,omitempty" gorm:"type:varchar;default:'exchange';"`
	BuyMax    string    `json:"buy_max,omitempty" gorm:"type:varchar;default:'0';"`
	PriceUsd  string    `json:"price_usd,omitempty" gorm:"type:varchar;default:'0';"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (ExinOtcAsset) TableName() string {
	return "exin_otc_asset"
}

type ExinLocalAsset struct {
	AssetID   string `json:"asset_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Price     string `json:"price,omitempty" gorm:"type:varchar;default:'0';"`
	Symbol    string `json:"symbol,omitempty" gorm:"type:varchar;default:'exchange';"`
	BuyMax    string `json:"buy_max,omitempty" gorm:"type:varchar;default:'0';"`
	UpdatedAt string `json:"updated_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (ExinLocalAsset) TableName() string {
	return "exin_local_asset"
}
