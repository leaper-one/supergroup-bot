package models

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"log"
	"strings"
	"sync"
	"time"
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
CREATE INDEX distribute_messages_list_idx ON distribute_messages(client_id, origin_message_id, level);
CREATE INDEX distribute_message_idx ON distribute_messages(client_id, message_id);
CREATE INDEX distribute_messages_all_list_idx ON distribute_messages(shard_id, level, created_at, client_id, status);
`

type DistributeMessage struct {
	ClientID         string    `json:"client_id,omitempty"`
	UserID           string    `json:"user_id,omitempty"`
	ConversationID   string    `json:"conversation_id,omitempty"`
	ShardID          string    `json:"shard_id,omitempty"`
	OriginMessageID  string    `json:"origin_message_id,omitempty"`
	MessageID        string    `json:"message_id,omitempty"`
	QuoteMessageID   string    `json:"quote_message_id,omitempty"`
	Data             string    `json:"data,omitempty"`
	Category         string    `json:"category,omitempty"`
	RepresentativeID string    `json:"representative_id,omitempty"`
	Level            int       `json:"level,omitempty"`
	Status           int       `json:"status,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
}

const (
	DistributeMessageLevelHigher = 1
	DistributeMessageLevelLower  = 2
	DistributeMessageLevelAlone  = 3

	DistributeMessageStatusPending      = 1 // 要发送的消息
	DistributeMessageStatusFinished     = 2 // 成功发送的消息
	DistributeMessageStatusLeaveMessage = 3 // 留言消息
	DistributeMessageStatusBroadcast    = 6 // 公告消息
)

//func SendClientUserPendingMessages(ctx context.Context, clientID, userID string) {
//	// 1. 将剩余的消息发送完
//	if err := sendPendingDistributeMessage(ctx, clientID, userID); err != nil {
//		session.Logger(ctx).Println(err)
//	}
//	if err := checkIsAsyncAndSendPendingMessage(ctx, clientID, userID); err != nil {
//		session.Logger(ctx).Println(err)
//	}
//}

// 删除超时的消息
func RemoveOvertimeDistributeMessages(ctx context.Context) error {
	_, err := session.Database(ctx).Exec(ctx, `DELETE FROM distribute_messages WHERE now()-created_at>interval '3 hours' AND status=$1`, DistributeMessageStatusFinished)
	return err
}

// 获取指定的消息
func PendingActiveDistributedMessages(ctx context.Context, clientID, shardID string) ([]*mixin.MessageRequest, error) {
	dms := make([]*mixin.MessageRequest, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT representative_id, user_id, conversation_id, message_id, category, data, quote_message_id FROM distribute_messages
WHERE client_id=$1 AND shard_id=$2 AND status=$3
ORDER BY level, created_at
LIMIT 80
`, func(rows pgx.Rows) error {
		repeatUser := make(map[string]bool)
		for rows.Next() {
			var dm mixin.MessageRequest
			if err := rows.Scan(&dm.RepresentativeID, &dm.RecipientID, &dm.ConversationID, &dm.MessageID, &dm.Category, &dm.Data, &dm.QuoteMessageID); err != nil {
				return err
			}
			if repeatUser[dm.RecipientID] {
				continue
			}
			repeatUser[dm.RecipientID] = true
			dms = append(dms, &dm)
		}
		return nil
	}, clientID, shardID, DistributeMessageStatusPending)
	return dms, err
}

func SendBatchMessages(ctx context.Context, client *mixin.Client, msgList []*mixin.MessageRequest) error {
	sendTimes := len(msgList)/80 + 1
	var waitSync sync.WaitGroup
	for i := 0; i < sendTimes; i++ {
		start := i * 80
		var end int
		if i == sendTimes-1 {
			end = len(msgList)
		} else {
			end = (i + 1) * 80
		}
		waitSync.Add(1)
		go sendMessages(ctx, client, msgList[start:end], &waitSync, end)
	}
	waitSync.Wait()
	return nil
}

func sendMessages(ctx context.Context, client *mixin.Client, msgList []*mixin.MessageRequest, waitSync *sync.WaitGroup, end int) {
	if len(msgList) == 0 {
		waitSync.Done()
		return
	}
	err := client.SendMessages(ctx, msgList)
	if err != nil {
		time.Sleep(time.Millisecond)
		data, _ := json.Marshal(msgList[0])
		session.Logger(ctx).Println(err, string(data))
		sendMessages(ctx, client, msgList, waitSync, end)
	} else {
		// 发送成功了
		for _, m := range msgList {
			_, err := session.Database(ctx).Exec(ctx, `UPDATE distribute_messages SET status=$3 WHERE client_id=$1 AND message_id=$2`, client.ClientID, m.MessageID, DistributeMessageStatusFinished)
			if err != nil {
				session.Logger(ctx).Println(err)
				return
			}
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
		if strings.Contains(err.Error(), "403") {
			return nil
		}
		data, _ := json.Marshal(msgs[0])
		log.Println(err, string(data))
		time.Sleep(time.Millisecond)
		return SendMessages(ctx, client, msgs)
	}
	return nil
}

func UpdateMessagesStatusToFinished(ctx context.Context, msgs []*mixin.MessageRequest) error {
	ids := make([]string, len(msgs))
	for i, m := range msgs {
		ids[i] = m.MessageID
	}
	_, err := session.Database(ctx).Exec(ctx, `UPDATE distribute_messages SET status=$2 WHERE message_id=ANY($1)`, ids, DistributeMessageStatusFinished)
	return err
}

//func checkIsAsyncAndSendPendingMessage(ctx context.Context, clientID, userID string) error {
//	// 1. 检查用户的 is_async, 为 true 则结束
//	isAsync, err := GetClientUserIsAsync(ctx, clientID, userID)
//	if err != nil {
//		return err
//	}
//	if isAsync {
//		return nil
//	}
//	// 2. 将 messages 表中的消息添加的 distribute_messages 中，然后再发
//	if num, err := addNoPendingMessageToDistributeMessages(ctx, clientID, userID); err != nil {
//		session.Logger(ctx).Println(err)
//	} else if num == -1 {
//		// 先变高优先级
//		if err := UpdateClientUserPriorityAndAsync(ctx, clientID, userID, ClientUserPriorityHigh, true); err != nil {
//			return err
//		}
//		return nil
//	} else if num == 0 {
//		// 3. 当 messages 表中没有新消息时，将用户状态的 priority 变为优先级高
//		if err := UpdateClientUserPriorityAndAsync(ctx, clientID, userID, ClientUserPriorityHigh, true); err != nil {
//			return err
//		}
//		return nil
//	} else {
//		// 添加了消息进入
//		if err := sendPendingDistributeMessage(ctx, clientID, userID); err != nil {
//			return err
//		}
//	}
//	time.Sleep(time.Millisecond * 500)
//	return checkIsAsyncAndSendPendingMessage(ctx, clientID, userID)
//}

// 将 messages 表中的消息添加的 distribute_messages 中
//func addNoPendingMessageToDistributeMessages(ctx context.Context, clientID, userID string) (int, error) {
//	// 1. 获取 distribute_messages 表中的最后一条
//	var lastMsgTime time.Time
//	if err := session.Database(ctx).QueryRow(ctx, `
//SELECT created_at FROM distribute_messages
//WHERE client_id=$1 AND user_id=$2
//ORDER BY created_at DESC limit 1
//`, clientID, userID).Scan(&lastMsgTime); err != nil {
//		if errors.Is(err, pgx.ErrNoRows) {
//			// 如果 distribute_messages 中没有消息，说明消息很低，直接 return -1 结束
//			return -1, nil
//		}
//		return 0, err
//	}
//	if lastMsgTime.IsZero() {
//		return -1, nil
//	}
//	// 获取 messages 表中最后一条
//	msgCount := 0
//	if err := session.Database(ctx).ConnQuery(ctx, `
//SELECT message_id, quote_message_id, created_at
//FROM messages
//WHERE client_id=$1 AND created_at>$2
//ORDER BY created_at ASC
//`, func(rows pgx.Rows) error {
//		for rows.Next() {
//			msgCount++
//			var msg mixin.MessageView
//			if err := rows.Scan(&msg.MessageID, &msg.QuoteMessageID, &msg.CreatedAt); err != nil {
//				return err
//			}
//			if err := createDistributeMessageByLevel(ctx, clientID, userID, &msg, nil, DistributeMessageLevelAlone); err != nil {
//				return err
//			}
//		}
//		return nil
//	}, clientID, lastMsgTime); err != nil {
//		if durable.CheckIsPKRepeatError(err) {
//			return -1, nil
//		}
//		return 0, err
//	}
//	return msgCount, nil
//}

//func sendPendingDistributeMessage(ctx context.Context, clientID, userID string) error {
//	client := GetMixinClientByID(ctx, clientID)
//	if client.ClientID == "" {
//		return session.BadDataError(ctx)
//	}
//	msgs := make([]*mixin.MessageRequest, 0)
//	if err := session.Database(ctx).ConnQuery(ctx, `
//SELECT m.user_id, dm.user_id, dm.message_id, m.category, m.data, dm.quote_message_id
//FROM distribute_messages AS dm
//LEFT JOIN messages AS m ON m.message_id=dm.origin_message_id
//WHERE dm.client_id=$1 AND dm.user_id=$2 AND level=$3 AND dm.status=$4
//ORDER BY dm.created_at ASC
//`, func(rows pgx.Rows) error {
//		for rows.Next() {
//			var dm mixin.MessageRequest
//			if err := rows.Scan(&dm.RepresentativeID, &dm.RecipientID, &dm.MessageID, &dm.Category, &dm.Data, &dm.QuoteMessageID); err != nil {
//				return err
//			}
//			dm.ConversationID = mixin.UniqueConversationID(clientID, dm.RecipientID)
//			msgs = append(msgs, &dm)
//		}
//		return nil
//	}, clientID, userID, DistributeMessageLevelAlone, DistributeMessageStatusPending); err != nil {
//		return err
//	}
//	for _, msg := range msgs {
//		if err := SendMessage(ctx, client.Client, msg, false); err != nil {
//			session.Logger(ctx).Println(err)
//		}
//		_, err := session.Database(ctx).Exec(ctx, `UPDATE distribute_messages SET status=$3 WHERE client_id=$1 AND message_id=$2`, clientID, msg.MessageID, DistributeMessageStatusFinished)
//		if err != nil {
//			session.Logger(ctx).Println(err)
//		}
//	}
//	return nil
//}

var cacheOriginMsgID = make(map[string]string)

func getOriginDistributeMsgID(ctx context.Context, clientID, messageID string) (string, error) {
	// 1. 用自己会话里 quote 的 message_id  找出 origin_message_id
	if cacheOriginMsgID[messageID] == "" {
		var originQuoteMsgID string
		if err := session.Database(ctx).QueryRow(ctx, `
SELECT origin_message_id FROM distribute_messages 
WHERE client_id=$1 AND message_id=$2
`, clientID, messageID).Scan(&originQuoteMsgID); err != nil {
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

func getDistributeMessageIDMapByOriginMsgID(ctx context.Context, clientID, originMsgID string) (map[string]string, error) {
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
			return cacheQuoteMsgID[originMsgID], err
		}
		var uid, mid string
		if err := session.Database(ctx).QueryRow(ctx, `
SELECT user_id, message_id FROM messages WHERE message_id=$1
`, originMsgID).Scan(&uid, &mid); err == nil {
			cacheQuoteMsgID[originMsgID][uid] = mid
		}
	}
	return cacheQuoteMsgID[originMsgID], nil
}

func getUserByDistributeMessageID(ctx context.Context, msgID string) (string, error) {
	var userID string
	err := session.Database(ctx).QueryRow(ctx, `SELECT user_id FROM distribute_messages WHERE message_id=$1`, msgID).Scan(&userID)
	return userID, err
}

func GetMsgStatistics(ctx context.Context) ([][]map[string]int, error) {
	sss := make([][]map[string]int, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT c.name, count(1)
FROM distribute_messages  d
LEFT JOIN  client c ON d.client_id = c.client_id
WHERE d.status=1
GROUP BY (c.name)
`, func(rows pgx.Rows) error {
		ss := make([]map[string]int, 0)
		for rows.Next() {
			var name string
			var count int
			if err := rows.Scan(&name, &count); err != nil {
				return err
			}
			ss = append(ss, map[string]int{name: count})
		}
		sss = append(sss, ss)
		return nil
	})
	err = session.Database(ctx).ConnQuery(ctx, `
SELECT c.name, count(1) 
FROM distribute_messages as d 
LEFT JOIN client c ON d.client_id = c.client_id 
GROUP BY (c.name)
`, func(rows pgx.Rows) error {
		ss := make([]map[string]int, 0)
		for rows.Next() {
			var name string
			var count int
			if err := rows.Scan(&name, &count); err != nil {
				return err
			}
			ss = append(ss, map[string]int{name: count})
		}
		sss = append(sss, ss)
		return nil
	})
	return sss, err
}

func DeleteDistributeMsgByClientID(ctx context.Context, clientID string) error {
	_, err := session.Database(ctx).Exec(ctx, `DELETE FROM distribute_messages WHERE client_id=$1 AND status=1`, clientID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return DeleteDistributeMsgByClientID(ctx, clientID)
	}
	return nil
}

func GetRemotePendingMsg(ctx context.Context, clientID string) time.Time {
	var t time.Time
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT created_at FROM distribute_messages WHERE client_id=$1 AND status=1 ORDER BY created_at ASC LIMIT 1
`, clientID).Scan(&t); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			session.Logger(ctx).Println(err)
		}
	}
	return t
}
