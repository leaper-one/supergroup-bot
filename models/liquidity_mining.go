package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type LiquidityMining struct {
	MiningID         string          `json:"mining_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Title            string          `json:"title,omitempty" gorm:"type:varchar;not null;"`
	Bg               string          `json:"bg,omitempty" gorm:"type:varchar;not null;"`
	Description      string          `json:"description,omitempty" gorm:"type:varchar;not null;"`
	Faq              string          `json:"faq,omitempty" gorm:"type:varchar;not null;"`
	JoinTips         string          `json:"join_tips,omitempty" gorm:"type:varchar;not null;default:'';"`
	JoinURL          string          `json:"join_url,omitempty" gorm:"type:varchar;not null;default:'';"`
	AssetID          string          `json:"asset_id,omitempty" gorm:"type:varchar(36);not null;"`
	ClientID         string          `json:"client_id,omitempty" gorm:"type:varchar(36);not null;"`
	FirstTime        time.Time       `json:"first_time,omitempty" gorm:"type:timestamp;not null;default:now();"`
	FirstEnd         time.Time       `json:"first_end,omitempty" gorm:"type:timestamp;not null;default:now();"`
	FirstDesc        string          `json:"first_desc,omitempty" gorm:"type:varchar;not null;default:'';"`
	DailyTime        time.Time       `json:"daily_time,omitempty" gorm:"type:timestamp;not null;default:now();"`
	DailyEnd         time.Time       `json:"daily_end,omitempty" gorm:"type:timestamp;not null;default:now();"`
	DailyDesc        string          `json:"daily_desc,omitempty" gorm:"type:varchar;not null;default:'';"`
	RewardAssetID    string          `json:"reward_asset_id,omitempty" gorm:"type:varchar(36);not null;"`
	FirstAmount      decimal.Decimal `json:"first_amount,omitempty" gorm:"type:varchar;not null;default:'0';"`
	DailyAmount      decimal.Decimal `json:"daily_amount,omitempty" gorm:"type:varchar;not null;default:'0';"`
	ExtraAssetID     string          `json:"extra_asset_id,omitempty" gorm:"type:varchar(36);not null;default:'';"`
	ExtraFirstAmount decimal.Decimal `json:"extra_first_amount,omitempty" gorm:"type:varchar;not null;default:'0';"`
	ExtraDailyAmount decimal.Decimal `json:"extra_daily_amount,omitempty" gorm:"type:varchar;not null;default:'0';"`
	CreatedAt        time.Time       `json:"created_at,omitempty" gorm:"type:timestamp;not null;default:now();"`

	Symbol string `json:"symbol,omitempty" gorm:"-"`
	Status string `json:"status,omitempty" gorm:"-"`

	RewardSymbol string `json:"reward_symbol,omitempty" gorm:"-"`
	ExtraSymbol  string `json:"extra_symbol,omitempty" gorm:"-"`
}

func (LiquidityMining) TableName() string {
	return "liquidity_mining"
}

const (
	LiquidityMiningFirst = 1 // 头矿挖矿
	LiquidityMiningDaily = 2 // 日矿挖矿

	LiquidityMiningStatusAuth    = "auth"    // 跳认证弹框
	LiquidityMiningStatusPending = "pending" // 跳未参与活动弹框
	LiquidityMiningStatusDone    = "done"    // 跳已参与活动页面
)

type LiquidityMiningUser struct {
	MiningID  string    `json:"mining_id" gorm:"primary_key;type:varchar(36);not null;"`
	UserID    string    `json:"user_id" gorm:"primary_key;type:varchar(36);not null;"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:now();"`
}

func (LiquidityMiningUser) TableName() string {
	return "liquidity_mining_users"
}

type LiquidityMiningTx struct {
	TraceID   string          `json:"trace_id" gorm:"primary_key;type:varchar(36);not null;"`
	MiningID  string          `json:"mining_id" gorm:"type:varchar(36);not null;"`
	RecordID  string          `json:"record_id" gorm:"type:varchar(36);not null;"`
	UserID    string          `json:"user_id" gorm:"type:varchar(36);not null;"`
	AssetID   string          `json:"asset_id" gorm:"type:varchar(36);not null;"`
	Amount    decimal.Decimal `json:"amount" gorm:"type:varchar;default:'0';"`
	Status    int             `json:"status" gorm:"type:smallint;default:0;"`
	CreatedAt time.Time       `json:"created_at" gorm:"type:timestamp;default:now();"`
}

func (LiquidityMiningTx) TableName() string {
	return "liquidity_mining_tx"
}

const (
	LiquidityMiningRecordStatusPending = 1 // 等待点击领取
	LiquidityMiningRecordStatusSuccess = 2 // 转账成功
	LiquidityMiningRecordStatusFailed  = 3 // 资产检查失败
)

type LiquidityMiningRecord struct {
	MiningID  string          `json:"mining_id,omitempty" gorm:"type:varchar(36);not null;"`
	RecordID  string          `json:"record_id,omitempty" gorm:"type:varchar(36);not null;"`
	UserID    string          `json:"user_id,omitempty" gorm:"type:varchar(36);not null;"`
	AssetID   string          `json:"asset_id,omitempty" gorm:"type:varchar(36);not null;"`
	Amount    decimal.Decimal `json:"amount,omitempty" gorm:"type:varchar;default:'0';"`
	Profit    decimal.Decimal `json:"profit,omitempty" gorm:"type:varchar;default:'0';"`
	CreatedAt time.Time       `json:"created_at,omitempty" gorm:"type:timestamp;default:now();"`
}

func (LiquidityMiningRecord) TableName() string {
	return "liquidity_mining_record"
}
