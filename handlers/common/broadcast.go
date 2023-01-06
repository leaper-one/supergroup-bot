package common

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
)

const broadcast_DDL = `
CREATE TABLE IF NOT EXISTS broadcast (
  client_id           VARCHAR(36) NOT NULL,
  message_id          VARCHAR(36) NOT NULL,
  status              SMALLINT NOT NULL DEFAULT 0,
  top_at              TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT '1970-1-1',
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  PRIMARY KEY(client_id,message_id)
); 
`

type Broadcast struct {
	ClientID  string    `json:"client_id,omitempty"`
	MessageID string    `json:"message_id,omitempty"`
	Status    int       `json:"status,omitempty"`
	TopAt     time.Time `json:"top_at,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

var (
	BroadcastStatusPending        = 0 // 默认
	BroadcastStatusFinished       = 1
	BroadcastStatusRecallPending  = 2
	BroadcastStatusRecallFinished = 3
)

func GetBroadcast(ctx context.Context, u *ClientUser) ([]*Message, error) {
	broadcasts := make([]*Message, 0)
	if err := session.DB(ctx).ConnQuery(ctx, `
SELECT m.message_id, m.user_id, m.category, m.data, u.full_name, u.avatar_url, b.status, b.created_at, b.top_at
FROM messages as m
LEFT JOIN users as u ON m.user_id=u.user_id
LEFT JOIN broadcast as b ON m.message_id=b.message_id
WHERE m.message_id IN (
  SELECT message_id FROM broadcast WHERE m.client_id=$1
)
ORDER BY b.created_at DESC
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var b Message
			if err := rows.Scan(&b.MessageID, &b.UserID, &b.Category, &b.Data, &b.FullName, &b.AvatarURL, &b.Status, &b.CreatedAt, &b.TopAt); err != nil {
				return err
			}
			b.Data = string(tools.Base64Decode(b.Data))
			broadcasts = append(broadcasts, &b)
		}
		return nil
	}, u.ClientID); err != nil {
		return nil, err
	}
	return broadcasts, nil
}

func CreateBroadcast(ctx context.Context, u *ClientUser, data, category string) error {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	msgID := tools.GetUUID()
	now := time.Now()
	if category == "" {
		category = mixin.MessageCategoryPlainText
	}
	data = tools.Base64Encode([]byte(data))
	// 创建一条消息
	msg := &mixin.MessageView{
		ConversationID: mixin.UniqueConversationID(u.ClientID, u.UserID),
		UserID:         u.UserID,
		MessageID:      msgID,
		Category:       category,
		Data:           data,
		Status:         mixin.MessageStatusSent,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := createBroadcast(ctx, u.ClientID, msgID); err != nil {
		tools.Println(err)
		return err
	}
	if err := createMessage(ctx, u.ClientID, msg, MessageStatusBroadcast); err != nil {
		tools.Println(err)
		return err
	}
	go SendBroadcast(_ctx, u, msgID, category, data, now)
	return nil
}

func DeleteBroadcast(ctx context.Context, u *ClientUser, broadcastID string) error {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	// 发送一条 recall 的消息
	// 1. 找到之前的
	if err := updateBroadcast(ctx, u.ClientID, broadcastID, BroadcastStatusRecallPending); err != nil {
		return err
	}
	go recallBroadcastByID(_ctx, u.ClientID, broadcastID)

	return nil
}

func recallBroadcastByID(ctx context.Context, clientID, originMsgID string) {
	var status int
	if err := session.DB(ctx).QueryRow(ctx, `
SELECT status FROM broadcast WHERE client_id=$1 AND message_id=$2
`, clientID, originMsgID).Scan(&status); err != nil {
		tools.Println(err)
		return
	}
	if status != BroadcastStatusRecallPending {
		return
	}
	dms, err := getQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
	if err != nil {
		tools.Println(err)
		return
	}
	if len(dms) == 0 {
		return
	}
	// 构建 recall 消息请求
	msgs := make([]*mixin.MessageRequest, 0)
	for userID, MsgID := range dms {
		objData := map[string]string{"message_id": MsgID}
		byteData, _ := json.Marshal(objData)
		msgs = append(msgs, &mixin.MessageRequest{
			ConversationID: mixin.UniqueConversationID(clientID, userID),
			RecipientID:    userID,
			MessageID:      tools.GetUUID(),
			Category:       mixin.MessageCategoryMessageRecall,
			Data:           tools.Base64Encode(byteData),
		})
	}

	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}

	if err := SendBatchMessages(ctx, client.Client, msgs); err != nil {
		tools.Println(err)
		return
	}
	if err := updateBroadcast(ctx, clientID, originMsgID, BroadcastStatusRecallFinished); err != nil {
		tools.Println(err)
		return
	}
}

func createBroadcast(ctx context.Context, clientID, msgID string) error {
	query := durable.InsertQuery("broadcast", "client_id,message_id")
	_, err := session.DB(ctx).Exec(ctx, query, clientID, msgID)
	return err
}

func updateBroadcast(ctx context.Context, clientID, msgID string, status int) error {
	_, err := session.DB(ctx).Exec(ctx, `
UPDATE broadcast SET status=$3
WHERE client_id=$1 AND message_id=$2
`, clientID, msgID, status)
	return err
}

func SendBroadcast(ctx context.Context, u *ClientUser, msgID, category, data string, now time.Time) {
	users, err := GetDistributeMsgUser(ctx, u.ClientID, false, true)
	if err != nil {
		tools.Println(err)
		return
	}
	msgs := make([]*mixin.MessageRequest, 0)
	for _, _u := range users {
		if CheckIsBlockUser(ctx, u.ClientID, _u.UserID) {
			continue
		}
		_msgID := tools.GetUUID()
		msgs = append(msgs, &mixin.MessageRequest{
			ConversationID: mixin.UniqueConversationID(u.ClientID, _u.UserID),
			RecipientID:    _u.UserID,
			MessageID:      _msgID,
			Category:       category,
			Data:           data,
		})
	}
	client, err := GetMixinClientByIDOrHost(ctx, u.ClientID)
	if err != nil {
		return
	}
	if err := SendBatchMessages(ctx, client.Client, msgs); err != nil {
		tools.Println(err)
		return
	}
	if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, msg := range msgs {
			if err := buildOriginMsgAndMsgIndex(ctx, p, &DistributeMessage{
				UserID:          msg.RecipientID,
				OriginMessageID: msgID,
				MessageID:       msg.MessageID,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		tools.Println(err)
		return
	}
	if err := updateBroadcast(ctx, u.ClientID, msgID, BroadcastStatusFinished); err != nil {
		tools.Println(err)
	}
}
