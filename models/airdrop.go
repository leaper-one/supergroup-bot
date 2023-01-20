package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Airdrop struct {
	AirdropID string          `json:"airdrop_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID  string          `json:"client_id,omitempty" gorm:"type:varchar(36);not null;"`
	UserID    string          `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	AssetID   string          `json:"asset_id,omitempty" gorm:"type:varchar(36);not null;"`
	TraceID   string          `json:"trace_id,omitempty" gorm:"type:varchar(36);not null;"`
	Amount    decimal.Decimal `json:"amount,omitempty" gorm:"type:varchar;not null;"`
	// 1 waiting for claim, 2 pending, 3 success
	Status     int       `json:"status" gorm:"type:smallint;not null;default:1;"`
	AskAssetID string    `json:"ask_asset,omitempty" gorm:"type:varchar(36);default:'';"`
	AskAmount  string    `json:"ask_amount,omitempty" gorm:"type:varchar;default:'';"`
	CreatedAt  time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

const (
	AirdropStatusWaiting = iota + 1
	AirdropStatusPending
	AirdropStatusSuccess
	AirdropStatusAssetAuth
	AirdropStatusAssetCheck
)

func (Airdrop) TableName() string {
	return "airdrop"
}
