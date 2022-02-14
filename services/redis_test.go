package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
)

type RedisTestService struct{}

func (service *RedisTestService) Run(ctx context.Context) error {
	clientIDList := []string{
		"1",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
		"9",
	}

	go func() {
		for {
			for _, clientID := range clientIDList {
				go insertMsg(ctx, clientID, 10000)
			}
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			for _, clientID := range clientIDList {
				for shardID := int64(0); shardID < config.MessageShardSize; shardID++ {
					_shardID := strconv.Itoa(int(shardID))
					distributeMsgFromRedis(ctx, clientID, _shardID)
				}
			}
			time.Sleep(time.Second)
		}
	}()
	select {}
}

func insertMsg(ctx context.Context, clientID string, count int) {
	msgs := make([]*models.DistributeMessage, 0)
	for i := 0; i < count; i++ {
		msgs = append(msgs, &models.DistributeMessage{
			ClientID:         clientID,
			UserID:           tools.GetUUID(),
			MessageID:        tools.GetUUID(),
			OriginMessageID:  tools.GetUUID(),
			QuoteMessageID:   tools.GetUUID(),
			Category:         "PLAIN_TEXT",
			Data:             "hello world",
			RepresentativeID: tools.GetUUID(),
			Level:            1,
			Status:           models.DistributeMessageStatusPending,
			CreatedAt:        time.Now(),
		})
	}
	if err := CreateDistributeMsgToRedis(ctx, msgs); err != nil {
		log.Println(err)
	}
}

func CreateDistributeMsgToRedis(ctx context.Context, msgs []*models.DistributeMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	_, err := session.Redis(ctx).Pipelined(ctx, func(p redis.Pipeliner) error {
		for _, msg := range msgs {
			dMsgKey := fmt.Sprintf("d_msg:%s:%s", msg.ClientID, msg.MessageID)
			if err := p.HSet(ctx, dMsgKey,
				map[string]interface{}{
					"user_id":           msg.UserID,
					"origin_message_id": msg.OriginMessageID,
					"conversation_id":   mixin.UniqueConversationID(msg.ClientID, msg.UserID),
					"message_id":        msg.MessageID,
					"quote_message_id":  msg.QuoteMessageID,
					"category":          msg.Category,
					"data":              msg.Data,
					"representative_id": msg.RepresentativeID,
					"created_at":        msg.CreatedAt,
				},
			).Err(); err != nil {
				session.Logger(ctx).Println(err)
				return err
			}
			if msg.Status == models.DistributeMessageStatusPending {
				score := msg.CreatedAt.Unix()
				if msg.Level == models.ClientUserPriorityHigh {
					score = score / 2
				}
				if err := p.ZAdd(ctx, fmt.Sprintf("s_msg:%s:%s", msg.ClientID, getShardID(msg.ClientID, msg.UserID)), &redis.Z{
					Score:  float64(score),
					Member: msg.MessageID,
				}).Err(); err != nil {
					session.Logger(ctx).Println(err)
					return err
				}
			} else {
				if err := p.PExpire(ctx, dMsgKey, time.Hour).Err(); err != nil {
					return err
				}
			}
			if err := buildOriginMsgAndMsgIndex(ctx, p, msg); err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func getShardID(clientID, userID string) string {
	shardID := models.ClientShardIDMap[clientID][userID]
	if shardID == "" {
		shardID = strconv.Itoa(rand.Intn(int(config.MessageShardSize)))
	}
	return shardID
}

func buildOriginMsgAndMsgIndex(ctx context.Context, p redis.Pipeliner, msg *models.DistributeMessage) error {
	// 建立 message_id -> origin_message_id 的索引
	if err := p.Set(ctx, fmt.Sprintf("msg_origin_idx:%s", msg.MessageID), fmt.Sprintf("%s,%s,%d", msg.OriginMessageID, msg.UserID, msg.Status), time.Hour*24).Err(); err != nil {
		return err
	}
	// 建立 origin_message_id -> message_id 的索引
	if err := p.SAdd(ctx, fmt.Sprintf("origin_msg_idx:%s", msg.OriginMessageID), fmt.Sprintf("%s,%s", msg.MessageID, msg.UserID)).Err(); err != nil {
		return err
	}
	if err := p.PExpire(ctx, fmt.Sprintf("origin_msg_idx:%s", msg.OriginMessageID), time.Hour*2).Err(); err != nil {
		return err
	}
	return nil
}

func distributeMsgFromRedis(ctx context.Context, clientID, shardID string) error {
	dms := make([]*mixin.MessageRequest, 0)
	msgIDs, err := session.Redis(ctx).ZRange(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), 0, 100).Result()
	if err != nil {
		return err
	}
	result := make([]*redis.StringStringMapCmd, 0, len(msgIDs))
	for _, msgID := range msgIDs {
		if _, err := session.Redis(ctx).Pipelined(ctx, func(p redis.Pipeliner) error {
			result = append(result, p.HGetAll(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, msgID)))
			return nil
		}); err != nil {
			return err
		}
	}
	for _, v := range result {
		msg, err := v.Result()
		if err != nil {
			return err
		}
		mr := mixin.MessageRequest{
			RepresentativeID: msg["representative_id"],
			RecipientID:      msg["user_id"],
			ConversationID:   msg["conversation_id"],
			MessageID:        msg["message_id"],
			Category:         msg["category"],
			Data:             msg["data"],
			QuoteMessageID:   msg["quote_message_id"],
		}
		dms = append(dms, &mr)
		msgIDs = append(msgIDs, msg["message_id"])
	}
	return UpdateDistributeMessagesStatusToFinished(ctx, clientID, shardID, msgIDs)
}

func UpdateDistributeMessagesStatusToFinished(ctx context.Context, clientID, shardID string, msgIDs []string) error {
	_, err := session.Redis(ctx).Pipelined(ctx, func(p redis.Pipeliner) error {
		members := make([]interface{}, 0, len(msgIDs))
		for _, v := range msgIDs {
			members = append(members, v)
			if err := p.PExpire(ctx, fmt.Sprintf("d_msg:%s:%s", clientID, v), time.Hour).Err(); err != nil {
				return err
			}
		}
		if err := p.ZRem(ctx, fmt.Sprintf("s_msg:%s:%s", clientID, shardID), members...).Err(); err != nil {
			return err
		}
		return nil
	})
	return err
}
