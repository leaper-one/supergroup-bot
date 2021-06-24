package services

import (
	"context"
	"encoding/json"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"strconv"
	"time"
)

type DistributeMessageService struct{}

func (service *DistributeMessageService) Run(ctx context.Context) error {

	clientList, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}

	for _, c := range clientList {
		client, err := mixin.NewFromKeystore(&mixin.Keystore{
			ClientID:   c.ClientID,
			SessionID:  c.SessionID,
			PrivateKey: c.PrivateKey,
			PinToken:   c.PinToken,
		})
		if err != nil {
			return err
		}
		go startDistributeMessageByClientID(ctx, client)
	}

	for {
		if err := models.RemoveOvertimeDistributeMessages(ctx); err != nil {
			time.Sleep(time.Minute)
		}
	}
}

func startDistributeMessageByClientID(ctx context.Context, client *mixin.Client) {
	for i := 0; i < int(config.MessageShardSize); i++ {
		go pendingActiveDistributedMessages(ctx, client, i)
	}
}

func pendingActiveDistributedMessages(ctx context.Context, client *mixin.Client, i int) {
	// 发送消息
	shardID := strconv.Itoa(i)
	for {
		messages, err := models.PendingActiveDistributedMessages(ctx, client.ClientID, shardID)
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages ERROR:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if len(messages) < 1 {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		messages = handleMsg(messages)
		err = models.SendMessages(ctx, client, messages)
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages sendDistributedMessges ERROR:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		err = models.UpdateMessagesStatusToFinished(ctx, messages)
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages UpdateMessagesStatus ERROR:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
	}
}

const maxLimit = 1024 * 1024

func handleMsg(messages []*mixin.MessageRequest) []*mixin.MessageRequest {
	total, _ := json.Marshal(messages)
	if len(total) < maxLimit {
		return messages
	}
	single, _ := json.Marshal(messages[0])
	msgCount := maxLimit / len(single)
	return messages[0:msgCount]
}
