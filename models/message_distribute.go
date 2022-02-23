package models

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
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

	DistributeMessageStatusPending      = 1  // 要发送的消息
	DistributeMessageStatusFinished     = 2  // 成功发送的消息
	DistributeMessageStatusLeaveMessage = 3  // 留言消息
	DistributeMessageStatusBroadcast    = 6  // 公告消息
	DistributeMessageStatusAloneList    = 9  // 单独处理的队列
	DistributeMessageStatusPINMessage   = 10 // PIN 的 message
)

// 删除超时的消息
func RemoveOvertimeDistributeMessages(ctx context.Context) error {
	_, err := session.Database(ctx).Exec(ctx, `DELETE FROM distribute_messages WHERE status IN (2,3,6) AND now()-created_at>interval '1 days'`)
	return err
}

// 获取指定的消息
func PendingActiveDistributedMessages(ctx context.Context, clientID, shardID string) ([]*mixin.MessageRequest, map[string]*DistributeMessage, error) {
	dms := make([]*mixin.MessageRequest, 0)
	msgOriginMsgIDMap := make(map[string]*DistributeMessage)
	msgIDs, err := session.Redis(ctx).ZRange(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), 0, 100).Result()
	if err != nil {
		return nil, nil, err
	}

	result := make([]*redis.StringStringMapCmd, 0, len(msgIDs))
	for _, v := range msgIDs {
		if _, err := session.Redis(ctx).Pipelined(ctx, func(p redis.Pipeliner) error {
			result = append(result, p.HGetAll(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, v)))
			return nil
		}); err != nil {
			return nil, nil, err
		}
	}
	userIDs := make(map[string]bool)
	for _, v := range result {
		msg, err := v.Result()
		if err != nil {
			return nil, nil, err
		}
		if userIDs[msg["user_id"]] {
			continue
		}
		userIDs[msg["user_id"]] = true
		originMsg, err := getMessageByMsgID(ctx, clientID, msg["origin_message_id"])
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				DeleteDistributeMsgByClientID(ctx, clientID)
				return nil, nil, nil
			}
			return nil, nil, err
		}
		if msg["data"] == "" {
			msg["data"] = originMsg.Data
		}
		l, _ := strconv.Atoi(msg["level"])
		msgOriginMsgIDMap[msg["message_id"]] = &DistributeMessage{
			Level:           l,
			OriginMessageID: msg["origin_message_id"],
		}
		mr := mixin.MessageRequest{
			RepresentativeID: originMsg.UserID,
			RecipientID:      msg["user_id"],
			ConversationID:   mixin.UniqueConversationID(msg["user_id"], clientID),
			MessageID:        msg["message_id"],
			Category:         originMsg.Category,
			Data:             msg["data"],
			QuoteMessageID:   msg["quote_message_id"],
		}
		if mr.Category == "MESSAGE_RECALL" {
			mr.RepresentativeID = ""
		}
		dms = append(dms, &mr)
	}
	return dms, msgOriginMsgIDMap, err
}

var cacheMessageData *tools.Mutex

func getMessageByMsgID(ctx context.Context, clientID, messageID string) (*Message, error) {
	data := cacheMessageData.Read(messageID)
	if m, ok := data.(*Message); ok && m != nil {
		return m, nil
	}
	msg, err := getMsgByClientIDAndMessageID(ctx, clientID, messageID)
	if err != nil {
		return nil, err
	}

	cacheMessageData.Write(messageID, msg)
	go func(msgID string) {
		time.Sleep(time.Hour)
		cacheMessageData.Delete(msgID)
	}(messageID)
	return msg, nil
}

func init() {
	cacheMessageData = tools.NewMutex()
}

func UpdateDistributeMessagesStatusToFinished(ctx context.Context, clientID, shardID string, msgIDs []string, msgOriginMsgIDMap map[string]*DistributeMessage) error {
	_, err := session.Redis(ctx).Pipelined(ctx, func(p redis.Pipeliner) error {
		members := make([]interface{}, 0, len(msgIDs))
		for _, msgID := range msgIDs {
			members = append(members, msgID)
			if err := p.Del(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, msgID)).Err(); err != nil {
				return err
			}
			msg := msgOriginMsgIDMap[msgID]
			msgBalance, err := session.Redis(ctx).Decr(ctx, fmt.Sprintf("l_msg:%s", msg.OriginMessageID)).Result()
			if err != nil {
				return err
			}
			if msgBalance < 0 {
				if err := p.Del(ctx, fmt.Sprintf("l_msg:%s", msg.OriginMessageID)).Err(); err != nil {
					return err
				}
				return nil
			}
			if msgBalance == 0 {
				msgStatusKey := "msg_status:" + msg.OriginMessageID
				status, err := session.Redis(ctx).Get(ctx, msgStatusKey).Int()
				if err != nil {
					return err
				}
				if status == MessageStatusFinished {
					if err := p.Set(ctx, msgStatusKey, strconv.Itoa(MessageRedisStatusFinished), time.Minute).Err(); err != nil {
						return err
					}
					if err := p.Del(ctx, "l_msg:"+msg.OriginMessageID).Err(); err != nil {
						return err
					}
					if err := p.Del(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, msg.OriginMessageID)).Err(); err != nil {
						return err
					}
					if err := updateMessageStatus(ctx, clientID, msg.OriginMessageID, MessageStatusFinished); err != nil {
						return err
					}
					originMsgIdx := "origin_msg_idx:" + msg.OriginMessageID
					if err := p.PExpire(ctx, originMsgIdx, time.Hour).Err(); err != nil {
						return err
					}
					resList, err := session.Redis(ctx).SMembers(ctx, originMsgIdx).Result()
					if err != nil {
						return err
					}
					for _, res := range resList {
						msg, err := getMsgOriginFromRedisResult(res)
						if err != nil {
							continue
						}
						if err := p.PExpire(ctx, "msg_origin_idx:"+msg.MessageID, time.Hour).Err(); err != nil {
							return err
						}
					}
				}
			}
		}
		if err := p.ZRem(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), members...).Err(); err != nil {
			return err
		}
		return nil
	})
	return err
}

func getDistributeMessageIDMapByOriginMsgID(ctx context.Context, clientID, originMsgID string) (map[string]string, string, error) {
	// 2. 用 origin_message_id 和 user_id 找出 对应会话 里的 message_id ，这个 message_id 就是要 quote 的 id
	mapList, err := getQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return nil, "", err
	}
	msg, err := getMsgByClientIDAndMessageID(ctx, clientID, originMsgID)
	if err == nil {
		mapList[msg.UserID] = originMsgID
		return mapList, msg.UserID, nil
	}
	return mapList, "", nil
}

func getDistributeMsgByMsgIDFromPsql(ctx context.Context, msgID string) (*DistributeMessage, error) {
	var m DistributeMessage
	err := session.Database(ctx).QueryRow(ctx, `
SELECT origin_message_id FROM distribute_messages WHERE message_id=$1
`, msgID).Scan(&m.OriginMessageID)
	return &m, err
}

func DeleteDistributeMsgByClientID(ctx context.Context, clientID string) {
	_, err := session.Database(ctx).Exec(ctx, `
DELETE FROM messages WHERE client_id=$1 AND status=ANY($2)`, clientID, []int{MessageStatusPending, MessageStatusPrivilege})
	if err != nil {
		session.Logger(ctx).Println(err)
		DeleteDistributeMsgByClientID(ctx, clientID)
		return
	}
	_, err = session.Redis(ctx).Pipelined(ctx, func(p redis.Pipeliner) error {
		dMsgs, err := session.Redis(ctx).Keys(ctx, fmt.Sprintf("d_msg:%s:*", clientID)).Result()
		if err != nil {
			return err
		}
		if err := p.Del(ctx, dMsgs...).Err(); err != nil {
			return err
		}
		sMsgs, err := session.Redis(ctx).Keys(ctx, fmt.Sprintf("s_msg:%s:*", clientID)).Result()
		if err != nil {
			return err
		}
		if err := p.Del(ctx, sMsgs...).Err(); err != nil {
			return err
		}
		oMsgIDs := make(map[string]bool)
		for _, res := range dMsgs {
			msgID := strings.Split(res, ":")[2]
			res, err := session.Redis(ctx).Get(ctx, "msg_origin_idx:"+msgID).Result()
			if err != nil {
				return err
			}
			msg, err := getOriginMsgFromRedisResult(res)
			if err != nil {
				return err
			}
			oMsgIDs[msg.OriginMessageID] = true
			if err := p.Del(ctx, "msg_origin_idx:"+msgID).Err(); err != nil {
				return err
			}
		}
		for msgID := range oMsgIDs {
			if err := p.Del(ctx, "l_msg:"+msgID).Err(); err != nil {
				return err
			}
			if err := p.Del(ctx, "origin_msg_idx:"+msgID).Err(); err != nil {
				return err
			}
			if err := p.Del(ctx, "msg_status:"+msgID).Err(); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		session.Logger(ctx).Println(err)
		DeleteDistributeMsgByClientID(ctx, clientID)
	}
}

func createFinishedDistributeMsg(ctx context.Context, clientID, userID, originMessageID, conversationID, shardID, messageID, quoteMessageID string, createdAt time.Time) error {
	return createDistributeMsgToRedis(ctx, []*DistributeMessage{{
		ClientID:        clientID,
		UserID:          userID,
		OriginMessageID: originMessageID,
		MessageID:       messageID,
		QuoteMessageID:  quoteMessageID,
		Status:          DistributeMessageStatusFinished,
		CreatedAt:       createdAt,
	}})
}
