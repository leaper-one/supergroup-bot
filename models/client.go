package models

import (
	"time"
)

type Client struct {
	ClientID       string    `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	IdentityNumber string    `json:"identity_number,omitempty" gorm:"type:varchar(11);not null;default:'';"`
	ClientSecret   string    `json:"client_secret,omitempty" gorm:"type:varchar;not null;"`
	SessionID      string    `json:"session_id,omitempty" gorm:"type:varchar(36);not null;"`
	PinToken       string    `json:"pin_token,omitempty" gorm:"type:varchar;not null;"`
	PrivateKey     string    `json:"private_key,omitempty" gorm:"type:varchar;not null;"`
	Pin            string    `json:"pin,omitempty" gorm:"type:varchar(6);default:'';"`
	Name           string    `json:"name,omitempty" gorm:"type:varchar;not null;"`
	Description    string    `json:"description,omitempty" gorm:"type:varchar;not null;"`
	Host           string    `json:"host,omitempty" gorm:"type:varchar;not null;"`
	Lang           string    `json:"lang,omitempty" gorm:"type:varchar;not null;default:'zh';"`
	AssetID        string    `json:"asset_id,omitempty" gorm:"type:varchar(36);not null;"`
	OwnerID        string    `json:"owner_id,omitempty" gorm:"type:varchar(36);not null;"`
	SpeakStatus    int       `json:"speak_status,omitempty" gorm:"type:smallint;not null;default:1;"`
	PayStatus      int       `json:"pay_status,omitempty" gorm:"type:smallint;not null;default:0;"`
	PayAmount      string    `json:"pay_amount,omitempty" gorm:"type:varchar;not null;default:'';"`
	CreatedAt      time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;not null;default:now();"`
	IconURL        string    `json:"icon_url,omitempty" gorm:"type:varchar;not null;default:'';"`
	Symbol         string    `json:"symbol,omitempty" gorm:"type:varchar;not null;default:'';"`
	AdminID        string    `json:"admin_id,omitempty" gorm:"type:varchar(36);not null;default:'';"`

	Welcome string `json:"welcome,omitempty" gorm:"-"`
	JoinMsg string `json:"join_msg,omitempty" gorm:"-"`
}

const (
	ClientSpeckStatusOpen  = 1 // 持仓发言打开，
	ClientSpeckStatusClose = 2 // 持仓发言关闭

	ClientPayStatusOpen = 1 // 入群开启，
)

const (
	ClientConversationStatusNormal    = "0"
	ClientConversationStatusMute      = "1"
	ClientConversationStatusAudioLive = "2"
)

func (Client) TableName() string {
	return "client"
}
