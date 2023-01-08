package common

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
)

const messages_DDL = `
-- 消息
CREATE TABLE IF NOT EXISTS messages (
  client_id           VARCHAR(36) NOT NULL,
  user_id             VARCHAR(36) NOT NULL,
  conversation_id     VARCHAR(36) NOT NULL,
  message_id          VARCHAR(36) NOT NULL,
  quote_message_id    VARCHAR(36) NOT NULL DEFAULT '',
  category            VARCHAR,
  data                TEXT,
  status              SMALLINT NOT NULL, -- 1 pending 2 privilege 3 normal 4 finished
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY(client_id, message_id)
);
`

type Message struct {
	ClientID       string    `json:"client_id,omitempty"`
	UserID         string    `json:"user_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	MessageID      string    `json:"message_id,omitempty"`
	QuoteMessageID string    `json:"quote_message_id,omitempty"`
	Category       string    `json:"category,omitempty"`
	Data           string    `json:"data,omitempty"`
	Status         int       `json:"status,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`

	FullName  string `json:"full_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`

	TopAt time.Time `json:"top_at,omitempty"`
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

var statusLimitMap = map[int]int{
	models.ClientUserStatusAudience: 5,
	models.ClientUserStatusFresh:    10,
	models.ClientUserStatusSenior:   15,
	models.ClientUserStatusLarge:    20,
	models.ClientUserStatusAdmin:    30,
	models.ClientUserStatusGuest:    30,
}

func getMsgByClientIDAndMessageID(ctx context.Context, clientID, msgID string) (*models.Message, error) {
	var m models.Message
	err := session.DB(ctx).Take(&m, "client_id = ? AND message_id = ?", clientID, msgID).Error
	return &m, err
}

// 1. 存入 message 表中
func CreateMessage(ctx context.Context, clientID string, msg *mixin.MessageView, status int) error {
	if err := session.DB(ctx).Create(&models.Message{
		ClientID:       clientID,
		UserID:         msg.UserID,
		ConversationID: msg.ConversationID,
		MessageID:      msg.MessageID,
		Category:       msg.Category,
		Data:           msg.Data,
		QuoteMessageID: msg.QuoteMessageID,
		Status:         status,
		CreatedAt:      msg.CreatedAt,
	}).Error; err != nil {
		return err
	}
	if status == MessageStatusPending {
		go session.Redis(_ctx).QPublish(_ctx, "create", clientID)
	}
	return nil
}

func updateMessageStatus(ctx context.Context, clientID, messageID string, status int) error {
	return session.DB(ctx).Model(&models.Message{}).
		Where("client_id = ? AND message_id = ?", clientID, messageID).
		Update("status", status).Error
}
