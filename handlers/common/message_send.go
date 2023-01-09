package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
)

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
		go sendMessages(client, msgList[start:end], &waitSync)
	}
	waitSync.Wait()
	return nil
}

func sendMessages(client *mixin.Client, msgList []*mixin.MessageRequest, waitSync *sync.WaitGroup) {
	if len(msgList) == 0 {
		waitSync.Done()
		return
	}
	err := client.SendMessages(context.Background(), msgList)
	if err != nil {
		time.Sleep(time.Millisecond)
		if !strings.Contains(err.Error(), "502 Bad Gateway") {
			data, _ := json.Marshal(msgList)
			log.Println("1...", err, string(data))
		}
		sendMessages(client, msgList, waitSync)
	} else {
		// 发送成功了
		msgIDs := make([]string, len(msgList))
		for i, msg := range msgList {
			msgIDs[i] = msg.MessageID
		}
		waitSync.Done()
	}
}

func SendMessage(ctx context.Context, client *mixin.Client, msg *mixin.MessageRequest, withCreate bool) error {
	err := client.SendMessage(context.Background(), msg)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			if withCreate {
				d, _ := json.Marshal(msg)
				tools.Println(err, string(d), client.ClientID)
				return nil
			}
			if _, err := client.CreateConversation(context.Background(), &mixin.CreateConversationInput{
				Category:       mixin.ConversationCategoryContact,
				ConversationID: mixin.UniqueConversationID(client.ClientID, msg.RecipientID),
				Participants:   []*mixin.Participant{{UserID: msg.RecipientID}},
			}); err != nil {
				return err
			}
			return SendMessage(ctx, client, msg, true)
		}
		if !strings.Contains(err.Error(), "502 Bad Gateway") {
			data, _ := json.Marshal(msg)
			log.Println("2...", err, string(data))
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
		if strings.Contains(err.Error(), "context deadline exceeded") ||
			errors.Is(err, context.Canceled) {
			ctx = models.Ctx
		} else if !strings.Contains(err.Error(), "502 Bad Gateway") &&
			!strings.Contains(err.Error(), "Internal Server Error") {
			data, _ := json.Marshal(msgs)
			log.Println("3...", string(data))
		}
		log.Println("4...", err)
		time.Sleep(time.Millisecond * 100)
		return SendMessages(ctx, client, msgs)
	}
	return nil
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

func getOriginMsgFromRedisResult(res string) (*models.DistributeMessage, error) {
	tmp := strings.Split(res, ",")
	if len(tmp) != 3 {
		tools.Println("invalid msg_origin_idx:", res)
		return nil, session.BadDataError(models.Ctx)
	}
	status, err := strconv.Atoi(tmp[2])
	if err != nil {
		return nil, err
	}
	return &models.DistributeMessage{
		OriginMessageID: tmp[0],
		UserID:          tmp[1],
		Status:          status,
	}, nil
}

func getMsgOriginFromRedisResult(res string) (*Message, error) {
	tmp := strings.Split(res, ",")
	if len(tmp) != 2 {
		tools.Println("invalid origin_msg_idx:", res)
		return nil, session.BadDataError(models.Ctx)
	}
	return &Message{
		MessageID: tmp[0],
		UserID:    tmp[1],
	}, nil
}

const maxLimit = 1024 * 1024

func HandleMsgWithLimit(messages []*mixin.MessageRequest) []*mixin.MessageRequest {
	total, _ := json.Marshal(messages)
	if len(total) < maxLimit {
		return messages
	}
	single, _ := json.Marshal(messages[0])
	msgCount := maxLimit / len(single)
	return messages[0:msgCount]
}
