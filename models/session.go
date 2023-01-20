package models

const (
	UserCategoryPlain     = "PLAIN"
	UserCategoryEncrypted = "ENCRYPTED"
)

type Session struct {
	ClientID  string `json:"client_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	UserID    string `json:"user_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	SessionID string `json:"session_id,omitempty" gorm:"primary_key;type:varchar(36);not null;"`
	PublicKey string `json:"public_key,omitempty" gorm:"type:varchar(128);not null;"`
}

func (Session) TableName() string {
	return "session"
}
