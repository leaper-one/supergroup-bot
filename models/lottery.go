package models

import (
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/shopspring/decimal"
)

type LotteryRecord struct {
	LotteryID  string          `json:"lottery_id" gorm:"type:varchar(36);not null;"`
	UserID     string          `json:"user_id" gorm:"type:varchar(36);not null;"`
	AssetID    string          `json:"asset_id" gorm:"type:varchar(36);not null;"`
	TraceID    string          `json:"trace_id" gorm:"type:varchar(36);not null;"`
	SnapshotID string          `json:"snapshot_id" gorm:"type:varchar(36);default:'';"`
	IsReceived bool            `json:"is_received" gorm:"type:boolean;default:false;"`
	Amount     decimal.Decimal `json:"amount" gorm:"type:varchar;default:'0';"`
	CreatedAt  time.Time       `json:"created_at" gorm:"type:timestamp;default:now();"`

	IconURL     string          `json:"icon_url,omitempty" gorm:"-"`
	Symbol      string          `json:"symbol,omitempty" gorm:"-"`
	FullName    string          `json:"full_name,omitempty" gorm:"-"`
	PriceUsd    decimal.Decimal `json:"price_usd,omitempty" gorm:"-"`
	ClientID    string          `json:"client_id,omitempty" gorm:"-"`
	Date        string          `json:"date,omitempty" gorm:"-"`
	Description string          `json:"description,omitempty" gorm:"-"`
}

func (LotteryRecord) TableName() string {
	return "lottery_record"
}

type LotterySupply struct {
	SupplyID  string          `json:"supply_id" gorm:"primary_key;type:varchar(36);not null;"`
	LotteryID string          `json:"lottery_id" gorm:"type:varchar(36);not null;"`
	AssetID   string          `json:"asset_id" gorm:"type:varchar(36);not null;"`
	Inventory int             `json:"inventory" gorm:"type:integer;default:-1;"`
	Amount    decimal.Decimal `json:"amount" gorm:"type:varchar;not null"`
	ClientID  string          `json:"client_id" gorm:"type:varchar(36);not null;"`
	IconURL   string          `json:"icon_url" gorm:"type:varchar(256);not null;"`
	Status    int             `json:"status" gorm:"type:smallint;default:1;"`
	CreatedAt time.Time       `json:"created_at" gorm:"type:timestamp;default:now();"`
}

const (
	LotterySupplyStatusListing = 1
	LotterySupplyStatusEnd     = 2
)

func (LotterySupply) TableName() string {
	return "lottery_supply"
}

type LotterySupplyReceived struct {
	SupplyID  string    `json:"supply_id" gorm:"primary_key;type:varchar(36);not null;"`
	UserID    string    `json:"user_id" gorm:"primary_key;type:varchar(36);not null;"`
	TraceID   string    `json:"trace_id" gorm:"type:varchar(36);not null;"`
	Status    int       `json:"status" gorm:"type:smallint;default:1;"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:now();"`
}

func (LotterySupplyReceived) TableName() string {
	return "lottery_supply_received"
}

type LotteryList struct {
	config.Lottery
	Description string          `json:"description"`
	Symbol      string          `json:"symbol"`
	PriceUSD    decimal.Decimal `json:"price_usd"`
}
