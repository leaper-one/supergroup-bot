package models

import (
	"time"
)

type Message struct {
	ClientID       string    `json:"client_id,omitempty" gorm:"type:varchar(36);not null"`
	UserID         string    `json:"user_id,omitempty" gorm:"type:varchar(36);not null;index:idx_messages_user_id;"`
	ConversationID string    `json:"conversation_id,omitempty" gorm:"type:varchar(36);not null"`
	MessageID      string    `json:"message_id,omitempty" gorm:"primary_key;type:varchar(36);not null;index:idx_message_id"`
	QuoteMessageID string    `json:"quote_message_id,omitempty" gorm:"type:varchar(36);default:'';"`
	Category       string    `json:"category,omitempty" gorm:"type:varchar;default:'';"`
	Data           string    `json:"data,omitempty" gorm:"type:text;default:'';"`
	Status         int       `json:"status,omitempty" gorm:"type:smallint;not null;"`
	CreatedAt      time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;not null;"`
}

func (Message) TableName() string {
	return "messages"
}

const (
	MessageStatusPending       = 1 // 待分发
	MessageStatusFinished      = 4 // 消息创建到 redis
	MessageRedisStatusFinished = 5 // 消息发送完毕

	MessageStatusNormal = 3 // 临时发送的消息

	MessageStatusBroadcast = 6
	MessageStatusJoinMsg   = 7
	MessageStatusRecallMsg = 8
	MessageStatusClientMsg = 9  // 客户端发送的消息
	MessageStatusPINMsg    = 10 // PIN 消息
	MessageStatusRemoveMsg = 11 // 移除消息
)

type DistributeMessage struct {
	ClientID         string    `json:"client_id,omitempty" gorm:"type:varchar(36);not null;index:distribute_messages_all_list_idx;primary_key"`
	ShardID          string    `json:"shard_id,omitempty" gorm:"type:varchar(36);not null;index:distribute_messages_all_list_idx;"`
	Status           int       `json:"status,omitempty" gorm:"type:smallint;not null;index:distribute_messages_all_list_idx;"`
	Level            int       `json:"level,omitempty" gorm:"type:smallint;not null;index:distribute_messages_all_list_idx;"`
	CreatedAt        time.Time `json:"created_at,omitempty" gorm:"type:timestamp with time zone;not null;index:distribute_messages_all_list_idx;"`
	UserID           string    `json:"user_id,omitempty" gorm:"type:varchar(36);not null;primary_key"`
	ConversationID   string    `json:"conversation_id,omitempty" gorm:"type:varchar(36);not null;"`
	OriginMessageID  string    `json:"origin_message_id,omitempty" gorm:"type:varchar(36);not null;primary_key"`
	MessageID        string    `json:"message_id,omitempty" gorm:"type:varchar(36);not null;index:distribute_messages_id_idx;"`
	QuoteMessageID   string    `json:"quote_message_id,omitempty" gorm:"type:varchar(36);default:'';"`
	Data             string    `json:"data,omitempty" gorm:"type:text;default:'';"`
	Category         string    `json:"category,omitempty" gorm:"type:varchar;default:'';"`
	RepresentativeID string    `json:"representative_id,omitempty" gorm:"type:varchar(36);default:'';"`
}

func (DistributeMessage) TableName() string {
	return "distribute_messages"
}

const (
	DistributeMessageLevelHigher = 1
	DistributeMessageLevelLower  = 2
	DistributeMessageLevelAlone  = 3

	DistributeMessageStatusPending      = 1  // 要发送的消息
	DistributeMessageStatusFinished     = 2  // 成功发送的消息
	DistributeMessageStatusLeaveMessage = 3  // 留言消息
	DistributeMessageStatusBroadcast    = 6  // 公告消息
	DistributeMessageStatusAloneList    = 9  // 单独处理的队列
	DistributeMessageStatusPINMessage   = 10 // PIN 的 message
)
