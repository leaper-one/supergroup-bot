package models

import (
	"time"
)

type ClientMemberAuth struct {
	ClientID        string    `json:"client_id" gorm:"primary_key;type:varchar(36);not null;"`
	UserStatus      int       `json:"user_status" gorm:"primary_key;type:smallint;not null;"`
	PlainText       bool      `json:"plain_text" gorm:"type:bool;default:false;"`
	PlainSticker    bool      `json:"plain_sticker" gorm:"type:bool;default:false;"`
	PlainImage      bool      `json:"plain_image" gorm:"type:bool;default:false;"`
	PlainVideo      bool      `json:"plain_video" gorm:"type:bool;default:false;"`
	PlainPost       bool      `json:"plain_post" gorm:"type:bool;default:false;"`
	PlainData       bool      `json:"plain_data" gorm:"type:bool;default:false;"`
	PlainLive       bool      `json:"plain_live" gorm:"type:bool;default:false;"`
	PlainContact    bool      `json:"plain_contact" gorm:"type:bool;default:false;"`
	PlainTranscript bool      `json:"plain_transcript" gorm:"type:bool;default:false;"`
	AppCard         bool      `json:"app_card" gorm:"type:bool;default:false;"`
	URL             bool      `json:"url" gorm:"type:bool;default:false;"`
	LuckyCoin       bool      `json:"lucky_coin" gorm:"type:bool;default:false;"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"type:timestamp;default:now();"`

	Limit int `json:"limit,omitempty" gorm:"-"`
}

func (ClientMemberAuth) TableName() string {
	return "client_member_auth"
}
