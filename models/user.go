package models

import (
	"time"
)

type User struct {
	UserID         string    `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36)"`
	IdentityNumber string    `json:"identity_number,omitempty" gorm:"type:varchar(512);not null;index:users_identity_number_key,unique;"`
	FullName       string    `json:"full_name,omitempty" gorm:"type:varchar(512);not null;"`
	AvatarURL      string    `json:"avatar_url,omitempty" gorm:"type:varchar(1024);not null;"`
	IsScam         bool      `json:"is_scam,omitempty" gorm:"type:boolean;not null;"`
	CreatedAt      time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;not null;default:now()"`

	AuthenticationToken string `json:"authentication_token,omitempty" gorm:"-"`
	IsNew               bool   `json:"is_new,omitempty" gorm:"-"`
}

func (User) TableName() string {
	return "users"
}

const (
	DefaultAvatar = "https://images.mixin.one/E2y0BnTopFK9qey0YI-8xV3M82kudNnTaGw0U5SU065864SsewNUo6fe9kDF1HIzVYhXqzws4lBZnLj1lPsjk-0=s128"
)
