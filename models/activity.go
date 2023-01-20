package models

import "time"

type Activity struct {
	ActivityIndex int       `json:"activity_index,omitempty" gorm:"primary_key;type:smallint;not null;"`
	ClientID      string    `json:"client_id,omitempty" gorm:"type:varchar(36);not null;"`
	Status        int       `json:"status,omitempty" gorm:"type:smallint;default:1;"`
	ImgURL        string    `json:"img_url,omitempty" gorm:"type:varchar(512);default:'';"`
	ExpireImgURL  string    `json:"expire_img_url,omitempty" gorm:"type:varchar(512);default:'';"`
	Action        string    `json:"action,omitempty" gorm:"type:varchar(512);default:'';"`
	StartAt       time.Time `json:"start_at,omitempty" gorm:"type:timestamp with time zone;not null;"`
	ExpireAt      time.Time `json:"expire_at,omitempty" gorm:"type:timestamp with time zone;not null;"`
	CreatedAt     time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;not null;default:now();"`
}

func (Activity) TableName() string {
	return "activity"
}
