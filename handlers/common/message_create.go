package common

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
)

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
	if status == models.MessageStatusPending {
		go session.Redis(models.Ctx).QPublish(models.Ctx, "create", clientID)
	}
	return nil
}

func CreatedManagerRecallMsg(ctx context.Context, clientID string, msgID, uid string) error {
	dataByte, _ := json.Marshal(map[string]string{"message_id": msgID})

	if err := CreateMessage(ctx, clientID, &mixin.MessageView{
		UserID:    uid,
		MessageID: tools.GetUUID(),
		Category:  mixin.MessageCategoryMessageRecall,
		Data:      tools.Base64Encode(dataByte),
	}, models.MessageStatusPending); err != nil {
		tools.Println(err)
	}

	return nil
}

func GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx context.Context, originMsgID string) (map[string]string, error) {
	recallMsgIDMap := make(map[string]string)
	resList, err := session.Redis(ctx).QSMembers(ctx, "origin_msg_idx:"+originMsgID)
	if err != nil {
		return nil, err
	}
	for _, res := range resList {
		msg, err := GetMsgOriginFromRedisResult(res)
		if err != nil {
			continue
		}
		recallMsgIDMap[msg.UserID] = msg.MessageID
	}
	return recallMsgIDMap, nil
}

func GetMsgOriginFromRedisResult(res string) (*models.Message, error) {
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

func CreateDistributeMsgToRedis(ctx context.Context, msgs []*models.DistributeMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	_, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, msg := range msgs {
			dMsgKey := fmt.Sprintf("d_msg:%s:%s", msg.ClientID, msg.MessageID)
			if err := p.HSet(ctx, dMsgKey,
				map[string]interface{}{
					"user_id":           msg.UserID,
					"origin_message_id": msg.OriginMessageID,
					"message_id":        msg.MessageID,
					"quote_message_id":  msg.QuoteMessageID,
					"data":              msg.Data,
					"representative_id": msg.RepresentativeID,
				},
			).Err(); err != nil {
				tools.Println(err)
				return err
			}
			if err := p.PExpire(ctx, dMsgKey, time.Hour*24).Err(); err != nil {
				return err
			}
			if msg.Status == models.DistributeMessageStatusPending {
				score := msg.CreatedAt.UnixNano()
				if msg.Level == models.ClientUserPriorityHigh {
					score = score / 2
				}
				if err := p.ZAdd(ctx, fmt.Sprintf("s_msg:%s:%s", msg.ClientID, getShardID(msg.ClientID, msg.UserID)), &redis.Z{
					Score:  float64(score),
					Member: msg.MessageID,
				}).Err(); err != nil {
					tools.Println(err)
					return err
				}
			} else {
				if err := p.PExpire(ctx, dMsgKey, config.QuoteMsgSavedTime).Err(); err != nil {
					return err
				}
			}
			if err := BuildOriginMsgAndMsgIndex(ctx, p, msg); err != nil {
				return err
			}
		}
		if msgs[0].Status == models.DistributeMessageStatusPending {
			lKey := fmt.Sprintf("l_msg:%s", msgs[0].OriginMessageID)
			cmcKey := fmt.Sprintf("client_msg_count:%s:%s", msgs[0].ClientID, tools.GetMinuteTime(time.Now()))
			if err := p.IncrBy(ctx, lKey, int64(len(msgs))).Err(); err != nil {
				return err
			}
			if err := p.IncrBy(ctx, cmcKey, int64(len(msgs))).Err(); err != nil {
				return err
			}
			if err := p.PExpire(ctx, lKey, time.Hour*24).Err(); err != nil {
				return err
			}
			if err := p.PExpire(ctx, cmcKey, time.Minute*2).Err(); err != nil {
				return err
			}
		}
		return nil
	})
	if msgs[0].Status == models.DistributeMessageStatusPending {
		if err := session.Redis(ctx).QPublish(ctx, "distribute", msgs[0].ClientID); err != nil {
			return err
		}
	}
	return err
}

func BuildOriginMsgAndMsgIndex(ctx context.Context, p redis.Pipeliner, msg *models.DistributeMessage) error {
	// 建立 message_id -> origin_message_id 的索引
	if err := p.Set(ctx, fmt.Sprintf("msg_origin_idx:%s", msg.MessageID), fmt.Sprintf("%s,%s,%d", msg.OriginMessageID, msg.UserID, msg.Status), config.QuoteMsgSavedTime).Err(); err != nil {
		return err
	}
	// 建立 origin_message_id -> message_id 的索引
	if err := p.SAdd(ctx, fmt.Sprintf("origin_msg_idx:%s", msg.OriginMessageID), fmt.Sprintf("%s,%s", msg.MessageID, msg.UserID)).Err(); err != nil {
		return err
	}
	if err := p.PExpire(ctx, fmt.Sprintf("origin_msg_idx:%s", msg.OriginMessageID), config.QuoteMsgSavedTime).Err(); err != nil {
		return err
	}
	return nil
}
