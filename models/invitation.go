package models

import (
	"time"

	"github.com/shopspring/decimal"
)

const MAX_POWER = 20000

type Invitation struct {
	InviteeID  string    `json:"invitee_id" gorm:"primary_key;type:varchar(36);not null;"`
	InviterID  string    `json:"inviter_id" gorm:"type:varchar(36);default:'';"`
	ClientID   string    `json:"client_id" gorm:"type:varchar(36);default:'';"`
	InviteCode string    `json:"invite_code" gorm:"type:varchar(6);not null;index:invitation_invite_code_key,unique;"`
	CreatedAt  time.Time `json:"created_at" gorm:"type:timestamp with time zone;default:now();"`
}

func (Invitation) TableName() string {
	return "invitation"
}

type InvitationPowerRecord struct {
	InviteeID string          `json:"invitee_id" gorm:"type:varchar(36);not null;"`
	InviterID string          `json:"inviter_id" gorm:"type:varchar(36);not null;"`
	Amount    decimal.Decimal `json:"amount" gorm:"type:varchar;not null;"`
	CreatedAt time.Time       `json:"created_at" gorm:"type:timestamp with time zone;default:now();"`
}

func (InvitationPowerRecord) TableName() string {
	return "invitation_power_record"
}
