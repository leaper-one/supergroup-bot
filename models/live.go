package models

import (
	"time"
)

type Live struct {
	LiveID      string    `json:"live_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID    string    `json:"client_id,omitempty" gorm:"type:varchar(36);not null;"`
	ImgURL      string    `json:"img_url,omitempty" gorm:"type:varchar(512);default:'';"`
	Category    int       `json:"category,omitempty" gorm:"type:smallint;default:1;"`
	Title       string    `json:"title,omitempty" gorm:"type:varchar;not null;"`
	Description string    `json:"description,omitempty" gorm:"type:varchar;not null;"`
	Status      int       `json:"status" gorm:"type:smallint;default:1;"`
	TopAt       time.Time `json:"top_at,omitempty" gorm:"type:timestamp with time zone;not null;default:'1970-1-1';"`
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;not null;default:now();"`
}

func (Live) TableName() string {
	return "lives"
}

type LiveData struct {
	LiveID       string    `json:"live_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	ReadCount    int       `json:"read_count" gorm:"type:integer;default:0;"`
	DeliverCount int       `json:"deliver_count" gorm:"type:integer;default:0;"`
	MsgCount     int       `json:"msg_count" gorm:"type:integer;default:0;"`
	UserCount    int       `json:"user_count" gorm:"type:integer;default:0;"`
	StartAt      time.Time `json:"start_at,omitempty" gorm:"type:timestamp with time zone;not null;default:now();"`
	EndAt        time.Time `json:"end_at,omitempty" gorm:"type:timestamp with time zone;not null;default:now();"`
}

func (LiveData) TableName() string {
	return "live_data"
}

type LiveReplay struct {
	MessageID string    `json:"message_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	ClientID  string    `json:"client_id,omitempty" gorm:"type:varchar(36);not null;"`
	LiveID    string    `json:"live_id,omitempty" gorm:"type:varchar(36);not null;default:'';"`
	Category  string    `json:"category,omitempty" gorm:"type:varchar;not null;"`
	Data      string    `json:"data,omitempty" gorm:"type:varchar;not null;"`
	CreatedAt time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;not null;default:now();"`
}

func (LiveReplay) TableName() string {
	return "live_replay"
}

const (
	LiveStatusBefore   = 0
	LiveStatusLiving   = 1
	LiveStatusFinished = 2

	LiveCategoryVideo         = 1
	LiveCategoryAudioAndImage = 2
)
