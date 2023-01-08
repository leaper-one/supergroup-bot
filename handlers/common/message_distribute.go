package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
)

const (
	DistributeMessageLevelHigher = 1
	DistributeMessageLevelLower  = 2
	DistributeMessageLevelAlone  = 3
)

// 删除超时的消息
func RemoveOvertimeDistributeMessages(ctx context.Context) error {
	_, err := session.DB(ctx).Exec(ctx, `DELETE FROM distribute_messages WHERE status IN (2,3,6) AND now()-created_at>interval '1 days'`)
	return err
}

// 获取指定的消息
func PendingActiveDistributedMessages(ctx context.Context, clientID, shardID string) ([]*mixin.MessageRequest, map[string]string, error) {
	msgIDs, err := session.Redis(ctx).QZRangeByScore(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), &redis.ZRangeBy{
		Min:    "-inf",
		Max:    "+inf",
		Offset: 0,
		Count:  120,
	})
	if err != nil {
		return nil, nil, err
	}
	result := make([]*redis.StringStringMapCmd, 0, len(msgIDs))
	for _, v := range msgIDs {
		if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
			result = append(result, p.HGetAll(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, v)))
			return nil
		}); err != nil {
			return nil, nil, err
		}
	}
	userIDs := make(map[string]bool)
	msgOriginMsgIDMap := make(map[string]string)
	dms := make([]*mixin.MessageRequest, 0)
	for i, v := range result {
		msg, err := v.Result()
		if msg["user_id"] == "" {
			tools.Println("d_msg:未找到...", fmt.Sprintf("s_msg:%s:%s", clientID, shardID), msgIDs[i], err)
			if err := session.Redis(ctx).W.ZRem(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), msgIDs[i]).Err(); err != nil {
				tools.Println(err)
			}
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		if userIDs[msg["user_id"]] {
			continue
		}
		userIDs[msg["user_id"]] = true
		originMsg, err := getMessageByMsgID(ctx, clientID, msg["origin_message_id"])
		if strings.HasPrefix(originMsg.Category, "SYSTEM_") {
			tools.Println("系统信息", fmt.Sprintf("s_msg:%s:%s", clientID, shardID), msgIDs[i], err)
			if err := session.Redis(ctx).W.ZRem(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), msgIDs[i]).Err(); err != nil {
				tools.Println(err)
			}
			continue
		}
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil, nil
			}
			return nil, nil, err
		}
		if originMsg.Status == MessageStatusRemoveMsg {
			if err := session.Redis(ctx).W.ZRem(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), msgIDs[i]); err != nil {
				tools.Println(err)
			}
			continue
		}

		if msg["data"] == "" {
			msg["data"] = originMsg.Data
		}
		msgOriginMsgIDMap[msg["message_id"]] = msg["origin_message_id"]
		mr := mixin.MessageRequest{
			RepresentativeID: msg["representative_id"],
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
		if len(dms) >= 100 {
			break
		}
	}
	return dms, msgOriginMsgIDMap, err
}

var cacheMessageData = tools.NewMutex()

func getMessageByMsgID(ctx context.Context, clientID, messageID string) (*Message, error) {
	data := cacheMessageData.Read(messageID)
	if m, ok := data.(*Message); ok && m != nil {
		return m, nil
	}
	msg, err := getMsgByClientIDAndMessageID(ctx, clientID, messageID)
	if err != nil {
		return nil, err
	}

	cacheMessageData.WriteWithTTL(messageID, msg, time.Minute)
	return msg, nil
}

func UpdateDistributeMessagesStatusToFinished(ctx context.Context, clientID, shardID string, delivered []string, msgOriginMsgIDMap map[string]string) error {
	msgIDs := make(map[string]bool)

	_, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		members := make([]interface{}, 0, len(delivered))
		for _, msgID := range delivered {
			members = append(members, msgID)
			if err := p.Unlink(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, msgID)).Err(); err != nil {
				return err
			}
			originMsgID := msgOriginMsgIDMap[msgID]
			if !msgIDs[originMsgID] {
				msgIDs[originMsgID] = true
			}
			err := p.Decr(ctx, fmt.Sprintf("l_msg:%s", originMsgID)).Err()
			if err != nil {
				return err
			}
		}
		if len(members) <= 0 {
			return nil
		}
		if err := p.ZRem(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), members...).Err(); err != nil {
			return err
		}
		return nil
	})
	for msgID := range msgIDs {
		msgBalance, _ := session.Redis(ctx).SyncGet(ctx, "l_msg:"+msgID).Int()
		if msgBalance == 0 {
			if _, err = session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
				msgStatusKey := "msg_status:" + msgID
				status, err := session.Redis(ctx).QGet(ctx, msgStatusKey).Int()
				if err != nil {
					return err
				}
				if status == MessageStatusFinished {
					if err := p.Set(ctx, msgStatusKey, strconv.Itoa(MessageRedisStatusFinished), time.Minute).Err(); err != nil {
						return err
					}
					if err := p.Unlink(ctx, "l_msg:"+msgID).Err(); err != nil {
						return err
					}
					if err := p.Unlink(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, msgID)).Err(); err != nil {
						return err
					}
					if err := updateMessageStatus(ctx, clientID, msgID, MessageRedisStatusFinished); err != nil {
						return err
					}
					originMsgIdx := "origin_msg_idx:" + msgID
					if err := p.PExpire(ctx, originMsgIdx, config.QuoteMsgSavedTime).Err(); err != nil {
						return err
					}
					resList, err := session.Redis(ctx).QSMembers(ctx, originMsgIdx)
					if err != nil {
						return err
					}
					for _, res := range resList {
						msg, err := getMsgOriginFromRedisResult(res)
						if err != nil {
							continue
						}
						if err := p.PExpire(ctx, "msg_origin_idx:"+msg.MessageID, config.QuoteMsgSavedTime).Err(); err != nil {
							return err
						}
					}
				}
				return nil
			}); err != nil {
				return err
			}
			log.Printf("消息%s已经完成", msgID)
		} else {
			log.Printf("消息剩余...%s:%d", msgID, msgBalance)
		}
	}
	if err != nil {
		tools.Println(err)
	}
	return nil
}

func GetDistributeMessageIDMapByOriginMsgID(ctx context.Context, clientID, originMsgID string) (map[string]string, string, error) {
	// 2. 用 origin_message_id 和 user_id 找出 对应会话 里的 message_id ，这个 message_id 就是要 quote 的 id
	mapList, err := GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
	if err != nil {
		tools.Println(err)
		return nil, "", err
	}
	msg, err := getMsgByClientIDAndMessageID(ctx, clientID, originMsgID)
	if err == nil {
		mapList[msg.UserID] = originMsgID
		return mapList, msg.UserID, nil
	}
	return mapList, "", nil
}

func getDistributeMsgByMsgIDFromPsql(ctx context.Context, msgID string) (*models.DistributeMessage, error) {
	var m models.DistributeMessage
	err := session.DB(ctx).QueryRow(ctx, `
SELECT origin_message_id FROM distribute_messages WHERE message_id=$1
`, msgID).Scan(&m.OriginMessageID)
	return &m, err
}

func DeleteDistributeMsgByClientID(ctx context.Context, clientID string) {
	_, err := session.DB(ctx).Exec(ctx, `
UPDATE messages SET status=$2 WHERE client_id=$1 AND status=1`, clientID, MessageStatusRemoveMsg)
	if err != nil {
		tools.Println(err)
		DeleteDistributeMsgByClientID(ctx, clientID)
		return
	}
	for i := 0; i < int(config.MessageShardSize); i++ {
		_, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
			if err := p.Del(ctx, fmt.Sprintf("s_msg:%s:%d", clientID, i)).Err(); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			tools.Println(err)
		}
	}
	go func() {
		ctx = _ctx
		dMsgs, err := session.Redis(ctx).QKeys(ctx, fmt.Sprintf("d_msg:%s:*", clientID))
		if err != nil {
			return
		}
		if len(dMsgs) > 0 {
			if err := session.Redis(ctx).W.Unlink(ctx, dMsgs...).Err(); err != nil {
				return
			}
		}
		oMsgIDs := make(map[string]bool)
		for _, res := range dMsgs {
			msgID := strings.Split(res, ":")[2]
			res, err := session.Redis(ctx).QGet(ctx, "msg_origin_idx:"+msgID).Result()
			if err != nil {
				return
			}
			msg, err := getOriginMsgFromRedisResult(res)
			if err != nil {
				return
			}
			oMsgIDs[msg.OriginMessageID] = true
			if err := session.Redis(ctx).W.Unlink(ctx, "msg_origin_idx:"+msgID).Err(); err != nil {
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
		for msgID := range oMsgIDs {
			if err := session.Redis(ctx).W.Unlink(ctx, "l_msg:"+msgID).Err(); err != nil {
				return
			}
			if err := session.Redis(ctx).W.Unlink(ctx, "origin_msg_idx:"+msgID).Err(); err != nil {
				return
			}
			if err := session.Redis(ctx).W.Unlink(ctx, "msg_status:"+msgID).Err(); err != nil {
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()
	if err != nil {
		tools.Println(err)
		DeleteDistributeMsgByClientID(ctx, clientID)
	}
}

func createFinishedDistributeMsg(ctx context.Context,
	clientID, userID, originMessageID,
	conversationID, shardID, messageID,
	quoteMessageID string, createdAt time.Time) error {
	return createDistributeMsgToRedis(ctx, []*models.DistributeMessage{{
		ClientID:        clientID,
		UserID:          userID,
		OriginMessageID: originMessageID,
		MessageID:       messageID,
		QuoteMessageID:  quoteMessageID,
		Status:          models.DistributeMessageStatusFinished,
		CreatedAt:       createdAt,
	}})
}
