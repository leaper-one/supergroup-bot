package models

import (
	"time"
)

type ClientUser struct {
	ClientID     string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;index:client_user_idx;index:client_user_priority_idx;"`
	UserID       string    `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	AccessToken  string    `json:"access_token,omitempty" gorm:"type:varchar(512);not null;"`
	Priority     int       `json:"priority,omitempty" gorm:"type:smallint;default:2;index:client_user_priority_idx;"`
	Status       int       `json:"status,omitempty" gorm:"type:smallint;default:0;"`
	PayStatus    int       `json:"pay_status,omitempty" gorm:"type:smallint;default:1;"`
	CreatedAt    time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
	IsReceived   bool      `json:"is_received,omitempty" gorm:"type:boolean;default:true;"`
	IsNoticeJoin bool      `json:"is_notice_join,omitempty" gorm:"type:boolean;default:true;"`
	MutedTime    string    `json:"muted_time,omitempty" gorm:"type:varchar;default:'';"`
	MutedAt      time.Time `json:"muted_at,omitempty" gorm:"type:timestamp with time zone;default:'1970-01-01 00:00:00+00';"`
	PayExpiredAt time.Time `json:"pay_expired_at,omitempty" gorm:"type:timestamp with time zone;default:'1970-01-01 00:00:00+00';"`
	ReadAt       time.Time `json:"read_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`
	DeliverAt    time.Time `json:"deliver_at,omitempty" gorm:"type:timestamp with time zone;default:now();"`

	AuthorizationID string `json:"authorization_id,omitempty" gorm:"type:varchar(36);default:'';"`
	Scope           string `json:"scope,omitempty" gorm:"type:varchar(512);default:'';"`
	PrivateKey      string `json:"private_key,omitempty" gorm:"type:varchar(128);default:'';"`
	Ed25519         string `json:"ed25519,omitempty" gorm:"type:varchar(128);default:'';"`

	AssetID     string `json:"asset_id,omitempty" gorm:"-"`
	SpeakStatus int    `json:"speak_status,omitempty" gorm:"-"`
}

func (ClientUser) TableName() string {
	return "client_users"
}

const (
	ClientUserPriorityHigh    = 1 // 高优先级
	ClientUserPriorityLow     = 2 // 低优先级
	ClientUserPriorityPending = 3 // 补发中
	ClientUserPriorityStop    = 4 // 暂停发送

	ClientUserStatusExit     = 0 // 退群
	ClientUserStatusAudience = 1 // 观众
	ClientUserStatusFresh    = 2 // 入门
	ClientUserStatusSenior   = 3 // 资深
	ClientUserStatusBlock    = 4 // 拉黑
	ClientUserStatusLarge    = 5 // 大户
	ClientUserStatusGuest    = 8 // 嘉宾
	ClientUserStatusAdmin    = 9 // 管理员
)

const (
	ClientNewMemberNoticeOn  = "1"
	ClientNewMemberNoticeOff = "0"

	ClientProxyStatusOn  = "1"
	ClientProxyStatusOff = "0"
)
