package models

import (
	"time"
)

type ClientBlockUser struct {
	ClientID   string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;index:client_block_user_idx"`
	UserID     string    `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	OperatorID string    `json:"operator_id,omitempty" gorm:"type:varchar(36);default:''"`
	CreatedAt  time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (ClientBlockUser) TableName() string {
	return "client_block_user"
}

type BlockUser struct {
	UserID     string    `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	OperatorID string    `json:"operator_id,omitempty" gorm:"type:varchar(36);default:'';"`
	CreatedAt  time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
	Memo       string    `json:"memo,omitempty" gorm:"type:varchar(255);default:'';"`
}

func (BlockUser) TableName() string {
	return "block_user"
}
