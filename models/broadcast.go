package models

import (
	"time"
)

type Broadcast struct {
	ClientID  string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	MessageID string    `json:"message_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Status    int       `json:"status,omitempty" gorm:"type:smallint;default:0;"`
	TopAt     time.Time `json:"top_at,omitempty" gorm:"type:timestamp with time zone;default:'1970-1-1';"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (Broadcast) TableName() string {
	return "broadcast"
}

var (
	BroadcastStatusPending        = 0 // 默认
	BroadcastStatusFinished       = 1
	BroadcastStatusRecallPending  = 2
	BroadcastStatusRecallFinished = 3
)
