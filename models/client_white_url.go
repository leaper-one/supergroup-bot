package models

import (
	"time"
)

type ClientWhiteURL struct {
	ClientID  string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	WhiteURL  string    `json:"white_url,omitempty" gorm:"primary_key;type:varchar;default:'';"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (ClientWhiteURL) TableName() string {
	return "client_white_url"
}
