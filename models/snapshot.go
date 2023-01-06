package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Snapshot struct {
	SnapshotID string          `json:"snapshot_id" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID   string          `json:"client_id" gorm:"type:varchar(36);not null;"`
	TraceID    string          `json:"trace_id" gorm:"type:varchar(36);not null;"`
	UserID     string          `json:"user_id" gorm:"type:varchar(36);not null;"`
	AssetID    string          `json:"asset_id" gorm:"type:varchar(36);not null;"`
	Amount     decimal.Decimal `json:"amount" gorm:"type:varchar;not null;"`
	Memo       string          `json:"memo" gorm:"type:varchar;default:'';"`
	CreatedAt  time.Time       `json:"created_at" gorm:"type:timestamp with time zone;not null;"`
}

const (
	SnapshotTypeReward  = "reward"
	SnapshotTypeJoin    = "join"
	SnapshotTypeVip     = "vip"
	SnapshotTypeAirdrop = "airdrop"
	SnapshotTypeMint    = "mint"
)

func (Snapshot) TableName() string {
	return "snapshots"
}

type Transfer struct {
	TraceID    string          `json:"trace_id" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID   string          `json:"client_id" gorm:"type:varchar(36);not null;"`
	AssetID    string          `json:"asset_id" gorm:"type:varchar(36);not null;"`
	OpponentID string          `json:"opponent_id" gorm:"type:varchar(36);not null;"`
	Amount     decimal.Decimal `json:"amount" gorm:"type:varchar;not null;"`
	Memo       string          `json:"memo" gorm:"type:varchar;default:'';"`
	Status     int             `json:"status" gorm:"type:smallint;not null;default:1;"`
	CreatedAt  time.Time       `json:"created_at" gorm:"type:timestamp with time zone;not null;"`
}

const (
	TransferStatusPending = 1
	TransferStatusSucceed = 2
)

func (Transfer) TableName() string {
	return "transfer_pending"
}
