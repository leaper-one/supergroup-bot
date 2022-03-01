package models

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
	"github.com/MixinNetwork/supergroup/session"
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
		if !errors.Is(err, context.Canceled) ||
			!errors.Is(err, context.DeadlineExceeded) {
			data, _ := json.Marshal(msgList)
			log.Println(err, string(data))
		}
		sendMessages(ctx, client, msgList, waitSync, end)
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
		if !errors.Is(err, context.Canceled) ||
			!errors.Is(err, context.DeadlineExceeded) {
			data, _ := json.Marshal(msgs)
			log.Println(err, string(data))
		}
		time.Sleep(time.Millisecond)
		return SendMessages(ctx, client, msgs)
	}
	return nil
}

type EncryptedMessageResp struct {
	MessageID   string `json:"message_id"`
	RecipientID string `json:"recipient_id"`
	State       string `json:"state"`
	Sessions    []struct {
		SessionID string `json:"session_id"`
		PublicKey string `json:"public_key"`
	} `json:"sessions"`
}

func SendEncryptedMessage(ctx context.Context, pk string, client *mixin.Client, msgs []*mixin.MessageRequest) ([]*EncryptedMessageResp, error) {
	var resp []*EncryptedMessageResp
	var userIDs []string
	for _, m := range msgs {
		userIDs = append(userIDs, m.RecipientID)
	}
	sessionSet, err := ReadSessionSetByUsers(ctx, client.ClientID, userIDs)
	if err != nil {
		return nil, err
	}
	var body []map[string]interface{}
	for _, message := range msgs {
		if message.RepresentativeID == client.ClientID {
			message.RepresentativeID = ""
		}
		if message.Category == mixin.MessageCategoryMessageRecall {
			message.RepresentativeID = ""
		}
		m := map[string]interface{}{
			"conversation_id":   message.ConversationID,
			"recipient_id":      message.RecipientID,
			"message_id":        message.MessageID,
			"quote_message_id":  message.QuoteMessageID,
			"category":          message.Category,
			"data_base64":       message.Data,
			"silent":            false,
			"representative_id": message.RepresentativeID,
		}
		recipient := sessionSet[message.RecipientID]
		category := readEncrypteCategory(message.Category, recipient)
		m["category"] = category
		if recipient != nil {
			m["checksum"] = GenerateUserChecksum(recipient.Sessions)
			var sessions []map[string]string
			for _, s := range recipient.Sessions {
				sessions = append(sessions, map[string]string{"session_id": s.SessionID})
			}
			m["recipient_sessions"] = sessions
			if strings.Contains(category, "ENCRYPTED") {
				data, err := encryptMessageData(message.Data, pk, recipient.Sessions)
				if err != nil {
					return nil, err
				}
				m["data_base64"] = data
			}
		}
		body = append(body, m)
	}
	if err := client.Post(ctx, "/encrypted_messages", body, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func readEncrypteCategory(category string, user *SimpleUser) string {
	if user == nil {
		return strings.Replace(category, "ENCRYPTED_", "PLAIN_", -1)
	}
	switch user.Category {
	case UserCategoryPlain:
		return strings.Replace(category, "ENCRYPTED_", "PLAIN_", -1)
	case UserCategoryEncrypted:
		return strings.Replace(category, "PLAIN_", "ENCRYPTED_", -1)
	default:
		return category
	}
}

func createDistributeMsgToRedis(ctx context.Context, msgs []*DistributeMessage) error {
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
					"message_id":        msg.MessageID,
					"quote_message_id":  msg.QuoteMessageID,
					"data":              msg.Data,
					"representative_id": msg.RepresentativeID,
					"level":             msg.Level,
				},
			).Err(); err != nil {
				session.Logger(ctx).Println(err)
				return err
			}
			if msg.Status == DistributeMessageStatusPending {
				score := msg.CreatedAt.UnixNano()
				if msg.Level == ClientUserPriorityHigh {
					score = score / 2
				}
				if err := p.Incr(ctx, fmt.Sprintf("l_msg:%s", msg.OriginMessageID)).Err(); err != nil {
					return err
				}
				if err := p.ZAdd(ctx, fmt.Sprintf("s_msg:%s:%s", msg.ClientID, getShardID(msg.ClientID, msg.UserID)), &redis.Z{
					Score:  float64(score),
					Member: msg.MessageID,
				}).Err(); err != nil {
					session.Logger(ctx).Println(err)
					return err
				}
			} else {
				if err := p.PExpire(ctx, dMsgKey, config.QuoteMsgSavedTime).Err(); err != nil {
					return err
				}
			}
			if err := buildOriginMsgAndMsgIndex(ctx, p, msg); err != nil {
				return err
			}
		}
		return nil
	})
	if msgs[0].Status == DistributeMessageStatusPending {
		if err := session.Redis(ctx).QPublish(ctx, "distribute", msgs[0].ClientID); err != nil {
			return err
		}
	}
	return err
}

func buildOriginMsgAndMsgIndex(ctx context.Context, p redis.Pipeliner, msg *DistributeMessage) error {
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

func getOriginMsgFromRedisResult(res string) (*DistributeMessage, error) {
	tmp := strings.Split(res, ",")
	if len(tmp) != 3 {
		session.Logger(_ctx).Println("invalid msg_origin_idx:", res)
		return nil, session.BadDataError(_ctx)
	}
	status, err := strconv.Atoi(tmp[2])
	if err != nil {
		return nil, err
	}
	return &DistributeMessage{
		OriginMessageID: tmp[0],
		UserID:          tmp[1],
		Status:          status,
	}, nil
}

func getMsgOriginFromRedisResult(res string) (*Message, error) {
	tmp := strings.Split(res, ",")
	if len(tmp) != 2 {
		session.Logger(_ctx).Println("invalid origin_msg_idx:", res)
		return nil, session.BadDataError(_ctx)
	}
	return &Message{
		MessageID: tmp[0],
		UserID:    tmp[1],
	}, nil
}
