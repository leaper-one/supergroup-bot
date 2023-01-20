package models

type DailyData struct {
	ClientID    string `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	Date        string `json:"date,omitempty" gorm:"primary_key;type:date;not null;"`
	Users       int64  `json:"users" gorm:"type:integer;not null;default:0;"`
	ActiveUsers int64  `json:"active_users" gorm:"type:integer;not null;default:0;"`
	Messages    int64  `json:"messages" gorm:"type:integer;not null;default:0;"`
}

func (DailyData) TableName() string {
	return "daily_data"
}
