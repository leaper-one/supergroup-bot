package models

import (
	"time"
)

type ClientUserProxy struct {
	ClientID    string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;index:client_user_proxy_user_idx;"`
	ProxyUserID string    `json:"proxy_user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	UserID      string    `json:"user_id,omitempty" gorm:"type:varchar(36);not null;index:client_user_proxy_user_idx;"`
	FullName    string    `json:"full_name,omitempty" gorm:"type:varchar(255);not null;"`
	SessionID   string    `json:"session_id,omitempty" gorm:"type:varchar(36);not null;"`
	PinToken    string    `json:"pin_token,omitempty" gorm:"type:varchar;not null;"`
	PrivateKey  string    `json:"private_key,omitempty" gorm:"type:varchar;not null;"`
	Status      int       `json:"status,omitempty" gorm:"type:smallint;default:1;"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
}

func (ClientUserProxy) TableName() string {
	return "client_user_proxy"
}

const (
	ClientUserProxyStatusInactive = 1
	ClientUserProxyStatusActive   = 2
)
