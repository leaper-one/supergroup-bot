package models

import (
	"time"
)

type ClientMenu struct {
	ClientID  string    `json:"client_id" gorm:"type:varchar(36);not null;"`
	Icon      string    `json:"icon" gorm:"type:varchar(255);not null;"`
	NameZh    string    `json:"name_zh" gorm:"type:varchar(255);not null;"`
	NameEn    string    `json:"name_en" gorm:"type:varchar(255);not null;"`
	NameJa    string    `json:"name_ja" gorm:"type:varchar(255);not null;"`
	URL       string    `json:"url" gorm:"type:varchar(255);not null;"`
	Idx       int       `json:"idx" gorm:"type:int;default:0;"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp;default:now();"`
}

func (ClientMenu) TableName() string {
	return "client_menus"
}
