package message

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
		if originMsg.Status == models.MessageStatusRemoveMsg {
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

func getMessageByMsgID(ctx context.Context, clientID, messageID string) (*models.Message, error) {
	data := cacheMessageData.Read(messageID)
	if m, ok := data.(*models.Message); ok && m != nil {
		return m, nil
	}
	msg, err := getMsgByClientIDAndMessageID(ctx, clientID, messageID)
	if err != nil {
		return nil, err
	}

	cacheMessageData.WriteWithTTL(messageID, msg, time.Minute*3)
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
				if status == models.MessageStatusFinished {
					if err := p.Set(ctx, msgStatusKey, strconv.Itoa(models.MessageRedisStatusFinished), time.Minute).Err(); err != nil {
						return err
					}
					if err := p.Unlink(ctx, "l_msg:"+msgID).Err(); err != nil {
						return err
					}
					if err := p.Unlink(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, msgID)).Err(); err != nil {
						return err
					}
					if err := updateMessageStatus(ctx, clientID, msgID, models.MessageRedisStatusFinished); err != nil {
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
func getMsgOriginFromRedisResult(res string) (*models.Message, error) {
	tmp := strings.Split(res, ",")
	if len(tmp) != 2 {
		tools.Println("invalid origin_msg_idx:", res)
		return nil, session.BadDataError(models.Ctx)
	}
	return &models.Message{
		MessageID: tmp[0],
		UserID:    tmp[1],
	}, nil
}
