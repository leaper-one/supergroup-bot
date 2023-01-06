package models

import (
	"time"
)

type Message struct {
	ClientID       string    `json:"client_id,omitempty" gorm:"type:varchar(36);not null"`
	UserID         string    `json:"user_id,omitempty" gorm:"type:varchar(36);not null"`
	ConversationID string    `json:"conversation_id,omitempty" gorm:"type:varchar(36);not null"`
	MessageID      string    `json:"message_id,omitempty" gorm:"primary_key;type:varchar(36);not null;index:idx_message_id"`
	QuoteMessageID string    `json:"quote_message_id,omitempty" gorm:"type:varchar(36);default:'';"`
	Category       string    `json:"category,omitempty" gorm:"type:varchar;default:'';"`
	Data           string    `json:"data,omitempty" gorm:"type:text;default:'';"`
	Status         int       `json:"status,omitempty" gorm:"type:smallint;not null;"`
	CreatedAt      time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;not null;"`

	FullName  string `json:"full_name,omitempty" gorm:"-"`
	AvatarURL string `json:"avatar_url,omitempty" gorm:"-"`

	TopAt time.Time `json:"top_at,omitempty" gorm:"-"`
}

func (Message) TableName() string {
	return "messages"
}

const (
	MessageStatusPending       = 1 // 待分发
	MessageStatusFinished      = 4 // 消息创建到 redis
	MessageRedisStatusFinished = 5 // 消息发送完毕

	MessageStatusNormal = 3 // 临时发送的消息

	MessageStatusLeaveMessage = 5
	MessageStatusBroadcast    = 6
	MessageStatusJoinMsg      = 7
	MessageStatusRecallMsg    = 8
	MessageStatusClientMsg    = 9  // 客户端发送的消息
	MessageStatusPINMsg       = 10 // PIN 消息
	MessageStatusRemoveMsg    = 11 // 移除消息
)
