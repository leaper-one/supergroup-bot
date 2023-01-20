package models

import "time"

type LoginLog struct {
	UserID    string    `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID  string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Addr      string    `json:"addr,omitempty" gorm:"type:varchar(255);not null;"`
	UA        string    `json:"ua,omitempty" gorm:"type:varchar(255);not null;"`
	IpAddr    string    `json:"ip_addr,omitempty" gorm:"type:varchar;default:'';"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (LoginLog) TableName() string {
	return "login_log"
}
