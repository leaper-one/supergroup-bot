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

	FullName  string    `json:"full_name,omitempty" gorm:"-"`
	AvatarURL string    `json:"avatar_url,omitempty" gorm:"-"`
	TopAt     time.Time `json:"top_at,omitempty" gorm:"-"`
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
const distribute_messages_DDL = `
-- 分发的消息
CREATE TABLE IF NOT EXISTS distribute_messages (
	client_id           VARCHAR(36) NOT NULL,
	user_id             VARCHAR(36) NOT NULL,
  conversation_id     VARCHAR(36) NOT NULL,
  shard_id            VARCHAR(36) NOT NULL,
	origin_message_id   VARCHAR(36) NOT NULL,
	message_id          VARCHAR(36) NOT NULL,
	quote_message_id    VARCHAR(36) NOT NULL DEFAULT '',
  data                TEXT NOT NULL DEFAULT '',
  category            VARCHAR NOT NULL DEFAULT '',
  representative_id   VARCHAR(36) NOT NULL DEFAULT '',
	level               SMALLINT NOT NULL, -- 1 高优先级 2 低优先级 3 单独队列
	status              SMALLINT NOT NULL DEFAULT 1, -- 1 待分发 2 已分发
	created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
	PRIMARY KEY(client_id, user_id, origin_message_id)
);
CREATE INDEX IF NOT EXISTS distribute_messages_all_list_idx ON distribute_messages (client_id, shard_id, status, level, created_at);
CREATE INDEX IF NOT EXISTS distribute_messages_id_idx ON distribute_messages (message_id);
`

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
