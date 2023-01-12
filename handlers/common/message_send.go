package common

import (
	"context"
	"encoding/json"
	"errors"
	"log"
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

func SendToClientManager(clientID string, msg *mixin.MessageView, isLeaveMsg, hasRepresentativeID bool) {
	ctx := models.Ctx
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != mixin.MessageCategoryPlainImage &&
		msg.Category != mixin.MessageCategoryPlainVideo {
		return
	}
	managers, err := GetClientUsersByClientIDAndStatus(ctx, clientID, models.ClientUserStatusAdmin)
	if err != nil {
		tools.Println(err)
		return
	}
	if len(managers) <= 0 {
		tools.Println("该社群没有管理员", clientID)
		return
	}
	msgList := make([]*mixin.MessageRequest, 0)
	var data string
	if isLeaveMsg && msg.Category == mixin.MessageCategoryPlainText {
		data = tools.Base64Encode([]byte(config.Text.PrefixLeaveMsg + string(tools.Base64Decode(msg.Data))))
	} else {
		data = msg.Data
	}

	for _, userID := range managers {
		conversationID := mixin.UniqueConversationID(clientID, userID)
		_msg := mixin.MessageRequest{
			ConversationID:   conversationID,
			RecipientID:      userID,
			MessageID:        mixin.UniqueConversationID(msg.MessageID, userID),
			Category:         msg.Category,
			Data:             data,
			RepresentativeID: msg.UserID,
		}
		if !hasRepresentativeID {
			_msg.RepresentativeID = ""
		}
		msgList = append(msgList, &_msg)
	}
	if msg.UserID == "" {
		data, _ := json.Marshal(msg)
		tools.Println(string(data))
	}
	if err := CreateMessage(ctx, clientID, msg, models.MessageStatusLeaveMessage); err != nil {
		tools.Println(err)
		return
	}
	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}
	if err := SendMessages(client.Client, msgList); err != nil {
		tools.Println(err)
		return
	}
	if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, _msg := range msgList {
			dm := &models.DistributeMessage{
				MessageID:       _msg.MessageID,
				UserID:          _msg.RecipientID,
				OriginMessageID: msg.MessageID,
			}
			if isLeaveMsg {
				dm.Status = models.DistributeMessageStatusLeaveMessage
			}
			if err := BuildOriginMsgAndMsgIndex(ctx, p, dm); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		tools.Println(err)
	}
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
		if LogWithNotNetworkError(err) {
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
		if LogWithNotNetworkError(err) {
			data, _ := json.Marshal(msg)
			log.Println("2...", err, string(data))
		}
		time.Sleep(time.Millisecond)
		return SendMessage(ctx, client, msg, false)
	}
	return nil
}

func SendMessages(client *mixin.Client, msgs []*mixin.MessageRequest) error {
	err := client.SendMessages(context.Background(), msgs)
	if err != nil {
		if strings.Contains(err.Error(), "403") {
			return nil
		}
		if LogWithNotNetworkError(err) {
			data, _ := json.Marshal(msgs)
			log.Println("3...", err, string(data))
		}
		log.Println("4...", err)
		time.Sleep(time.Millisecond * 100)
		return SendMessages(client, msgs)
	}
	return nil
}

func LogWithNotNetworkError(err error) bool {
	if strings.Contains(err.Error(), "502 Bad Gateway") ||
		strings.Contains(err.Error(), "Internal Server Error") ||
		strings.Contains(err.Error(), "context deadline exceeded") ||
		errors.Is(err, context.Canceled) {
		return false
	}
	return true
}

const maxLimit = 1024 * 1024

func HandleMsgWithLimit(msgs []*mixin.MessageRequest) []*mixin.MessageRequest {
	total, _ := json.Marshal(msgs)
	if len(total) < maxLimit {
		return msgs
	}
	single, _ := json.Marshal(msgs[0])
	msgCount := maxLimit / len(single)
	return msgs[0:msgCount]
}
