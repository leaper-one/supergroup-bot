package models

import (
	"time"
)

type Property struct {
	Key       string    `gorm:"primary_key;type:varchar(512);not null;"`
	Value     string    `gorm:"type:varchar(8192);not null;"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone;default:now();"`
}

func (Property) TableName() string {
	return "properties"
}
