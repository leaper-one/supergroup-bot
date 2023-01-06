package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Liquidity struct {
	LiquidityID string          `json:"liquidity_id" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID    string          `json:"client_id" gorm:"type:varchar(36);not null;"`
	Title       string          `json:"title" gorm:"type:varchar;default:'';"`
	Description string          `json:"description" gorm:"type:varchar;default:'';"`
	StartAt     time.Time       `json:"start_at" gorm:"type:timestamp with time zone;not null;"`
	EndAt       time.Time       `json:"end_at" gorm:"type:timestamp with time zone;not null;"`
	AssetIDs    string          `json:"asset_ids" gorm:"type:varchar;default:'';"`
	MinAmount   decimal.Decimal `json:"min_amount" gorm:"type:varchar;default:'0';"`
	LpDesc      string          `json:"lp_desc" gorm:"type:varchar;default:'';"`
	LpURL       string          `json:"lp_url" gorm:"type:varchar;default:'';"`
	CreatedAt   time.Time       `json:"created_at" gorm:"type:timestamp with time zone;default:now();"`
}

func (Liquidity) TableName() string {
	return "liquidity"
}

type LiquidityDetail struct {
	LiquidityID string          `json:"liquidity_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Idx         int             `json:"idx,omitempty" gorm:"type:int;not null;"`
	StartAt     time.Time       `json:"start_at,omitempty" gorm:"type:timestamp with time zone;not null;"`
	EndAt       time.Time       `json:"end_at,omitempty" gorm:"type:timestamp with time zone;not null;"`
	AssetID     string          `json:"asset_id,omitempty" gorm:"type:varchar(36);not null;"`
	Amount      decimal.Decimal `json:"amount,omitempty" gorm:"type:varchar;default:'0';"`
	Symbol      string          `json:"symbol,omitempty" gorm:"type:varchar;default:'';"`
	CreatedAt   time.Time       `json:"-" gorm:"type:timestamp with time zone;default:now();"`
}

func (LiquidityDetail) TableName() string {
	return "liquidity_detail"
}

type LiquidityUser struct {
	LiquidityID string    `json:"liquidity_id" gorm:"primary_key;type:varchar(36);not null;"`
	UserID      string    `json:"user_id" gorm:"primary_key;type:varchar(36);not null;"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now();"`
}

func (LiquidityUser) TableName() string {
	return "liquidity_user"
}

type LiquiditySnapshot struct {
	UserID      string          `json:"user_id,omitempty" gorm:"type:varchar(36);not null;"`
	LiquidityID string          `json:"liquidity_id,omitempty" gorm:"type:varchar(36);not null;"`
	Idx         int             `json:"idx,omitempty" gorm:"type:int;not null;"`
	Date        string          `json:"date,omitempty" gorm:"type:date;not null;"`
	LpSymbol    string          `json:"lp_symbol,omitempty" gorm:"type:varchar;default:'';"`
	LpAmount    decimal.Decimal `json:"lp_amount,omitempty" gorm:"type:varchar;default:'0';"`
	UsdValue    decimal.Decimal `json:"usd_value,omitempty" gorm:"type:varchar;default:'0';"`
}

func (LiquiditySnapshot) TableName() string {
	return "liquidity_snapshot"
}

type LiquidityTx struct {
	LiquidityID string    `json:"liquidity_id" gorm:"type:varchar(36);not null;"`
	Month       time.Time `json:"month" gorm:"type:date;not null;"`
	Idx         int       `json:"idx" gorm:"type:int;not null;"`
	UserID      string    `json:"user_id" gorm:"type:varchar(36);not null;"`
	AssetID     string    `json:"asset_id" gorm:"type:varchar(36);not null;"`
	Amount      string    `json:"amount" gorm:"type:varchar;default:'0';"`
	TraceID     string    `json:"trace_id" gorm:"primary_key;type:varchar(36);not null;"`
	Status      string    `json:"status" gorm:"type:varchar;default:'W';"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now();"`
}

func (LiquidityTx) TableName() string {
	return "liquidity_tx"
}

const (
	LiquidityTxWait    = "W"
	LiquidityTxPending = "P"
	LiquidityTxSuccess = "S"
	LiquidityTxFail    = "F"
)
