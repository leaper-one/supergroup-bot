package models

import (
	"time"
)

type Voucher struct {
	Code      string    `json:"code" gorm:"primary_key;type:varchar(8);not null;"`
	ClientID  string    `json:"client_id" gorm:"type:varchar(36);default:''"`
	Status    int       `json:"status" gorm:"type:int2;not null;default:1"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);default:''"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamptz;not null;default:now()"`
	ExpiredAt time.Time `json:"expire_at" gorm:"type:timestamptz;not null;default:CURRENT_DATE + INTERVAL '7 day'"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamptz;default:now()"`
}

const (
	VoucherStatusUnused = 1
	VoucherStatusUsed   = 2
)

func (Voucher) TableName() string {
	return "voucher"
}
