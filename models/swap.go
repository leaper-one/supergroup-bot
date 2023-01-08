package models

import (
	"time"
)

type Swap struct {
	LpAsset     string    `json:"lp_asset,omitempty" gorm:"primary_key;varchar(36);not null"`
	Asset0      string    `json:"asset0,omitempty" gorm:"varchar(36);not null;index:swap_asset0_idx"`
	Asset0Price string    `json:"asset0_price,omitempty" gorm:"varchar;not null"`
	Asset1      string    `json:"asset1,omitempty" gorm:"varchar(36);not null;index:swap_asset1_idx"`
	Asset1Price string    `json:"asset1_price,omitempty" gorm:"varchar;not null"`
	Type        string    `json:"type,omitempty" gorm:"varchar(1);not null"`
	Pool        string    `json:"pool,omitempty" gorm:"varchar;not null"`
	Earn        string    `json:"earn,omitempty" gorm:"varchar;not null"`
	Amount      string    `json:"amount,omitempty" gorm:"varchar;not null"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`

	IconURL      string `json:"icon_url,omitempty" gorm:"-"`
	Asset0Symbol string `json:"asset0_symbol,omitempty" gorm:"-"`
	Asset0Amount string `json:"asset0_amount,omitempty" gorm:"-"`

	Asset1Symbol string `json:"asset1_symbol,omitempty" gorm:"-"`
	Asset1Amount string `json:"asset1_amount,omitempty" gorm:"-"`
	PriceUsd     string `json:"price_usd,omitempty" gorm:"-"`
	ChangeUsd    string `json:"change_usd,omitempty" gorm:"-"`

	// otc
	ExinOtcAsset *ExinOtcAsset `json:"exin_otc_asset,omitempty" gorm:"-"`
}

const (
	SwapTypeExinSwap    = "0"
	SwapType4SwapMtg    = "1"
	SwapTypeExinOne     = "2"
	SwapTypeExinLocal   = "3"
	SwapType4SwapNormal = "4"
)

func (Swap) TableName() string {
	return "swap"
}
