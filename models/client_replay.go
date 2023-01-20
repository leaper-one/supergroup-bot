package models

import (
	"time"
)

type ClientReplay struct {
	ClientID  string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	JoinMsg   string    `json:"join_msg,omitempty" gorm:"type:text;default:'';"`
	Welcome   string    `json:"welcome,omitempty" gorm:"type:text;default:'';"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (ClientReplay) TableName() string {
	return "client_replay"
}
