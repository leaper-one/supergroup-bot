package models

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"strings"
	"sync"
	"time"
)

const distribute_messages_DDL = `
-- 分发的消息
CREATE TABLE IF NOT EXISTS distribute_messages (
    client_id           VARCHAR(36) NOT NULL,
    user_id             VARCHAR(36) NOT NULL,
    origin_message_id   VARCHAR(36) NOT NULL,
    message_id          VARCHAR(36) NOT NULL,
    quote_message_id    VARCHAR(36) NOT NULL DEFAULT '',
    level               SMALLINT NOT NULL, -- 1 高优先级 2 低优先级 3 单独队列
    status              SMALLINT NOT NULL DEFAULT 1, -- 1 待分发 2 已分发
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    PRIMARY KEY(client_id, user_id, origin_message_id)
);
CREATE INDEX distribute_messages_list_idx ON distribute_messages(client_id, origin_message_id, level);
CREATE INDEX distribute_message_idx ON distribute_messages(client_id, message_id);
CREATE INDEX distribute_messages_all_list_idx ON distribute_messages(client_id, origin_message_id);
`

type DistributeMessage struct {
	ClientID        string    `json:"client_id,omitempty"`
	UserID          string    `json:"user_id,omitempty"`
	OriginMessageID string    `json:"origin_message_id,omitempty"`
	MessageID       string    `json:"message_id,omitempty"`
	QuoteMessageID  string    `json:"quote_message_id,omitempty"`
	Level           int       `json:"level,omitempty"`
	Status          int       `json:"status,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
}

const (
	DistributeMessageLevelHigher = 1
	DistributeMessageLevelLower  = 2
	DistributeMessageLevelAlone  = 3

	DistributeMessageStatusPending      = 1 // 要发送的消息
	DistributeMessageStatusFinished     = 2 // 成功发送的消息
	DistributeMessageStatusLeaveMessage = 3 // 留言消息
)

func createDistributeMessageList(ctx context.Context, clientID string, msg *mixin.MessageView, isNew, isManager bool, ) error {
	high, low, err := GetClientUser(ctx, clientID, isManager)
	if err != nil {
		return err
	}

	for _, s := range high {
		if s == msg.UserID {
			continue
		}
		if err := createDistributeMessage(ctx, clientID, s, msg, DistributeMessageLevelHigher); err == nil {
			continue
		} else {
			if !durable.CheckIsPKRepeatError(err) {
				session.Logger(ctx).Println(err)
				continue
			}
			if !isNew {
				continue
			}
			// 新消息提示重复，则将 `is_async` 标记为 `true`
			if err := UpdateClientUserIsAsync(ctx, clientID, s, true); err != nil {
				session.Logger(ctx).Println(err)
				continue
			}
		}
	}
	for _, s := range low {
		if s == msg.UserID {
			continue
		}
		err := createDistributeMessage(ctx, clientID, s, msg, DistributeMessageLevelLower)
		if err != nil && !durable.CheckIsPKRepeatError(err) {
			session.Logger(ctx).Println(err)
			continue
		}
	}

	return nil
}
func getRecallOriginMsgID(ctx context.Context, msgData string) string {
	data := tools.Base64Decode(msgData)
	var msg struct{ MessageID string `json:"message_id"` }
	err := json.Unmarshal(data, &msg)
	if err != nil {
		session.Logger(ctx).Println(err)
		return ""
	}
	return msg.MessageID
}

func SendClientUserPendingMessages(ctx context.Context, clientID, userID string) {
	// 1. 将剩余的消息发送完
	if err := sendPendingDistributeMessage(ctx, clientID, userID); err != nil {
		session.Logger(ctx).Println(err)
	}
	if err := checkIsAsyncAndSendPendingMessage(ctx, clientID, userID); err != nil {
		session.Logger(ctx).Println(err)
	}
}
func createDistributeMessage(ctx context.Context, clientID, userID string, msg *mixin.MessageView, level int) error {
	if userID == msg.UserID {
		return nil
	}

	if msg.QuoteMessageID == "" {
		return _createDistributeMessage(ctx, clientID, userID, msg.MessageID, tools.GetUUID(), "", level, DistributeMessageStatusPending, msg.CreatedAt)
	} else {
		// 1. 获取 originMessageID
		originMsgID, err := getOriginDistributeMsgID(ctx, clientID, msg.QuoteMessageID)
		if err != nil {
			return err
		}
		var quoteMessageID string
		if originMsgID != "" {
			quoteMessageID, err = getDistributeMessageIDByOriginMsgID(ctx, clientID, userID, originMsgID)
		}
		return _createDistributeMessage(ctx, clientID, userID, msg.MessageID, tools.GetUUID(), quoteMessageID, level, DistributeMessageStatusPending, msg.CreatedAt)
	}
}

var insertDistributeMessageQuery = durable.InsertQuery("distribute_messages", "client_id,user_id,origin_message_id,message_id,quote_message_id,level,status,created_at")

func _createDistributeMessage(ctx context.Context, clientID, userID, originMsgID, messageID, quoteMsgID string, level, status int, createdAt time.Time) error {
	_, err := session.Database(ctx).Exec(ctx, insertDistributeMessageQuery,
		clientID, userID, originMsgID, messageID, quoteMsgID, level, status, createdAt)
	return err
}

// 删除超时的消息
func RemoveOvertimeDistributeMessages(ctx context.Context) error {
	_, err := session.Database(ctx).Exec(ctx, `DELETE FROM distribute_messages WHERE now()-created_at>interval '3 hours'`)
	return err
}

// 获取指定的消息
func GetDistributeMessageByClientIDAndLevel(ctx context.Context, clientID string, msg *Message, level int) ([]*mixin.MessageRequest, error) {
	recallMsgIDMap := make(map[string]string)
	var originMsgID string
	if msg.Category == mixin.MessageCategoryMessageRecall {
		originMsgID = getRecallOriginMsgID(ctx, msg.Data)
		var count int
		if err := session.Database(ctx).ConnQuery(ctx, `
		SELECT message_id, user_id
		FROM distribute_messages
		WHERE client_id=$1 AND origin_message_id=$2`, func(rows pgx.Rows) error {
			for rows.Next() {
				var msgID, userID string
				if err := rows.Scan(&msgID, &userID); err != nil {
					return err
				}
				recallMsgIDMap[userID] = msgID
				count++
			}
			return nil
		}, clientID, originMsgID); err != nil {
			return nil, err
		}
		if count == 0 {
			// 消息已经被删除...
			return nil, nil
		}
	}
	dms := make([]*mixin.MessageRequest, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT m.user_id, dm.user_id, dm.message_id, m.category, m.data, dm.quote_message_id FROM distribute_messages AS dm
LEFT JOIN messages AS m 
ON dm.origin_message_id=m.message_id AND dm.client_id=m.client_id
WHERE dm.client_id=$1 AND dm.origin_message_id=$2 AND dm.level=$3
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var dm mixin.MessageRequest
			if err := rows.Scan(&dm.RepresentativeID, &dm.RecipientID, &dm.MessageID, &dm.Category, &dm.Data, &dm.QuoteMessageID); err != nil {
				return err
			}
			dm.ConversationID = mixin.UniqueConversationID(clientID, dm.RecipientID)
			if msg.Category == mixin.MessageCategoryMessageRecall {
				if recallMsgIDMap[dm.RecipientID] == "" {
					continue
				}
				data, err := json.Marshal(map[string]string{"message_id": recallMsgIDMap[dm.RecipientID]})
				if err != nil {
					return err
				}
				dm.RepresentativeID = ""
				dm.QuoteMessageID = ""
				dm.Data = tools.Base64Encode(data)
			}
			if dm.RepresentativeID == dm.RecipientID {
				continue
			}
			dms = append(dms, &dm)
		}
		return nil
	}, clientID, msg.MessageID, level)
	return dms, err
}

func SendBatchMessages(ctx context.Context, client *mixin.Client, msgList []*mixin.MessageRequest) error {
	sendTimes := len(msgList)/80 + 1
	var waitSync sync.WaitGroup

	for i := 0; i < sendTimes; i++ {
		waitSync.Add(1)
		start := i * 80
		var end int
		if i == sendTimes-1 {
			end = len(msgList)
		} else {
			end = (i + 1) * 80
		}
		go sendMessages(ctx, client, msgList[start:end], &waitSync)
	}
	waitSync.Wait()
	return nil
}

func sendMessages(ctx context.Context, client *mixin.Client, msgList []*mixin.MessageRequest, waitSync *sync.WaitGroup) {
	err := client.SendMessages(ctx, msgList)
	if err != nil {
		time.Sleep(time.Millisecond)
		data, _ := json.Marshal(msgList)
		session.Logger(ctx).Println(err, string(data))
		sendMessages(ctx, client, msgList, waitSync)
	} else {
		// 发送成功了
		if err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
			for _, m := range msgList {
				_, err := tx.Exec(ctx, `UPDATE distribute_messages SET status=$3 WHERE client_id=$1 AND message_id=$2`, client.ClientID, m.MessageID, DistributeMessageStatusFinished)
				if err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			session.Logger(ctx).Println(err)
		}
		waitSync.Done()
	}
}

func SendMessage(ctx context.Context, client *mixin.Client, msg *mixin.MessageRequest, withCreate bool) error {
	err := client.SendMessage(ctx, msg)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			if withCreate {
				d, _ := json.Marshal(msg)
				session.Logger(ctx).Println(err, string(d), client.ClientID)
				return nil
			}
			if _, err := client.CreateConversation(ctx, &mixin.CreateConversationInput{
				Category:       mixin.ConversationCategoryContact,
				ConversationID: mixin.UniqueConversationID(client.ClientID, msg.RecipientID),
				Participants:   []*mixin.Participant{{UserID: msg.RecipientID}},
			}); err != nil {
				return err
			}
			return SendMessage(ctx, client, msg, true)
		}
		time.Sleep(time.Millisecond)
		return SendMessage(ctx, client, msg, false)
	}
	return nil
}

func SendMessages(ctx context.Context, client *mixin.Client, msgs []*mixin.MessageRequest) error {
	err := client.SendMessages(ctx, msgs)
	if err != nil {
		session.Logger(ctx).Println(err)
		if strings.Contains(err.Error(), "403") {
			return nil
		}
		time.Sleep(time.Millisecond)
		return SendMessages(ctx, client, msgs)
	}
	return nil
}

func checkIsAsyncAndSendPendingMessage(ctx context.Context, clientID, userID string) error {
	// 1. 检查用户的 is_async, 为 true 则结束
	isAsync, err := GetClientUserIsAsync(ctx, clientID, userID)
	if err != nil {
		return err
	}
	if isAsync {
		return nil
	}
	// 2. 将 messages 表中的消息添加的 distribute_messages 中，然后再发
	if num, err := addNoPendingMessageToDistributeMessages(ctx, clientID, userID); err != nil {
		session.Logger(ctx).Println(err)
	} else if num == -1 {
		// 先变高优先级
		if err := UpdateClientUserPriorityAndAsync(ctx, clientID, userID, ClientUserPriorityHigh, true); err != nil {
			return err
		}
		return nil
	} else if num == 0 {
		// 3. 当 messages 表中没有新消息时，将用户状态的 priority 变为优先级高
		if err := UpdateClientUserPriorityAndAsync(ctx, clientID, userID, ClientUserPriorityHigh, true); err != nil {
			return err
		}
		return nil
	} else {
		// 添加了消息进入
		if err := sendPendingDistributeMessage(ctx, clientID, userID); err != nil {
			return err
		}
	}
	time.Sleep(time.Millisecond * 500)
	return checkIsAsyncAndSendPendingMessage(ctx, clientID, userID)
}

// 将 messages 表中的消息添加的 distribute_messages 中
func addNoPendingMessageToDistributeMessages(ctx context.Context, clientID, userID string) (int, error) {
	// 1. 获取 distribute_messages 表中的最后一条
	var lastMsgTime time.Time
	if err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT created_at FROM distribute_messages 
WHERE client_id=$1 AND user_id=$2 
ORDER BY created_at DESC limit 1
`, func(row pgx.Row) error {
		return row.Scan(&lastMsgTime)
	}, clientID, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 如果 distribute_messages 中没有消息，说明消息很低，直接 return -1 结束
			return -1, nil
		}
		return 0, err
	}
	if lastMsgTime.IsZero() {
		return -1, nil
	}
	// 获取 messages 表中最后一条
	msgCount := 0
	if err := session.Database(ctx).ConnQuery(ctx, `
SELECT message_id, quote_message_id, created_at
FROM messages
WHERE client_id=$1 AND created_at>$2
ORDER BY created_at ASC
`, func(rows pgx.Rows) error {
		for rows.Next() {
			msgCount++
			var msg mixin.MessageView
			if err := rows.Scan(&msg.MessageID, &msg.QuoteMessageID, &msg.CreatedAt); err != nil {
				return err
			}
			if err := createDistributeMessage(ctx, clientID, userID, &msg, DistributeMessageLevelAlone); err != nil {
				return err
			}
		}
		return nil
	}, clientID, lastMsgTime); err != nil {
		if durable.CheckIsPKRepeatError(err) {
			return -1, nil
		}
		return 0, err
	}
	return msgCount, nil
}

func sendPendingDistributeMessage(ctx context.Context, clientID, userID string) error {
	client := GetMixinClientByID(ctx, clientID)
	if client == nil {
		return session.BadDataError(ctx)
	}
	msgs := make([]*mixin.MessageRequest, 0)
	if err := session.Database(ctx).ConnQuery(ctx, `
SELECT m.user_id, dm.user_id, dm.message_id, m.category, m.data, dm.quote_message_id 
FROM distribute_messages AS dm
LEFT JOIN messages AS m ON m.message_id=dm.origin_message_id
WHERE dm.client_id=$1 AND dm.user_id=$2 AND level=$3 AND dm.status=$4
ORDER BY dm.created_at ASC
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var dm mixin.MessageRequest
			if err := rows.Scan(&dm.RepresentativeID, &dm.RecipientID, &dm.MessageID, &dm.Category, &dm.Data, &dm.QuoteMessageID); err != nil {
				return err
			}
			dm.ConversationID = mixin.UniqueConversationID(clientID, dm.RecipientID)
			msgs = append(msgs, &dm)
		}
		return nil
	}, clientID, userID, DistributeMessageLevelAlone, DistributeMessageStatusPending); err != nil {
		return err
	}
	for _, msg := range msgs {
		if err := SendMessage(ctx, client.Client, msg, false); err != nil {
			session.Logger(ctx).Println(err)
		}
		_, err := session.Database(ctx).Exec(ctx, `UPDATE distribute_messages SET status=$3 WHERE client_id=$1 AND message_id=$2`, clientID, msg.MessageID, DistributeMessageStatusFinished)
		if err != nil {
			session.Logger(ctx).Println(err)
		}
	}
	return nil
}

var cacheOriginMsgID = make(map[string]string)

func getOriginDistributeMsgID(ctx context.Context, clientID, messageID string) (string, error) {
	// 1. 用自己会话里 quote 的 message_id  找出 origin_message_id
	if cacheOriginMsgID[messageID] == "" {
		var originQuoteMsgID string
		if err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT origin_message_id FROM distribute_messages 
WHERE client_id=$1 AND message_id=$2
`,
			func(row pgx.Row) error {
				return row.Scan(&originQuoteMsgID)
			}, clientID, messageID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				cacheOriginMsgID[messageID] = messageID
			} else {
				return "", err
			}
		} else {
			cacheOriginMsgID[messageID] = originQuoteMsgID
		}
	}
	return cacheOriginMsgID[messageID], nil
}

var cacheQuoteMsgID = make(map[string]map[string]string)

func getDistributeMessageIDByOriginMsgID(ctx context.Context, clientID, userID, originMsgID string) (string, error) {
	// 2. 用 origin_message_id 和 user_id 找出 对应会话 里的 message_id ，这个 message_id 就是要 quote 的 id
	if cacheQuoteMsgID[originMsgID] == nil {
		cacheQuoteMsgID[originMsgID] = make(map[string]string)
		if err := session.Database(ctx).ConnQuery(ctx, `
SELECT user_id, message_id FROM distribute_messages
WHERE client_id=$1 AND origin_message_id=$2
`,
			func(rows pgx.Rows) error {
				for rows.Next() {
					var uid, mid string
					if err := rows.Scan(&uid, &mid); err != nil {
						return err
					}
					cacheQuoteMsgID[originMsgID][uid] = mid
				}
				return nil
			}, clientID, originMsgID); err != nil {
			return "", err
		}
		var uid, mid string
		if err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT user_id, message_id FROM messages WHERE message_id=$1
`, func(row pgx.Row) error {
			return row.Scan(&uid, &mid)
		}, originMsgID); err == nil {
			cacheQuoteMsgID[originMsgID][uid] = mid
		}
	}
	return cacheQuoteMsgID[originMsgID][userID], nil
}

func getUserByDistributeMessageID(ctx context.Context, msgID string) (string, error) {
	var userID string
	err := session.Database(ctx).ConnQueryRow(ctx, `SELECT user_id FROM distribute_messages WHERE message_id=$1`, func(row pgx.Row) error {
		return row.Scan(&userID)
	}, msgID)
	return userID, err
}
