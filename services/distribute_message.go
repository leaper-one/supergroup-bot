package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
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
		go startDistributeMessageByClientID(ctx, client, c.PrivateKey)
	}

	for {
		if err := models.RemoveOvertimeDistributeMessages(ctx); err != nil {
			time.Sleep(time.Minute)
		}
	}
}

func startDistributeMessageByClientID(ctx context.Context, client *mixin.Client, pk string) {
	for i := 0; i < int(config.MessageShardSize); i++ {
		go pendingActiveDistributedMessages(ctx, client, i, pk)
	}
}

func pendingActiveDistributedMessages(ctx context.Context, client *mixin.Client, i int, pk string) {
	// 发送消息
	shardID := strconv.Itoa(i)
	isEncrypted := false
	if config.Config.Encrypted {
		me, err := client.UserMe(ctx)
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
		for _, v := range me.App.Capabilities {
			if v == "ENCRYPTED" {
				isEncrypted = true
			}
		}
	}
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
		now := time.Now().UnixNano()
		if isEncrypted {
			err = handleEncryptedDistributeMsg(ctx, pk, client, messages)
		} else {
			err = handleNormalDistributeMsg(ctx, client, messages)
		}
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages sendDistributedMessges ERROR:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		tools.PrintTimeDuration(fmt.Sprintf("%s:msg send %d...", shardID, len(messages)), now)
	}
}

func handleEncryptedDistributeMsg(ctx context.Context, pk string, client *mixin.Client, messages []*mixin.MessageRequest) error {
	var delivered []string
	results, err := models.SendEncryptedMessage(ctx, pk, client, messages)
	if err != nil {
		return err
	}
	var sessions []*models.Session
	for _, m := range results {
		if m.State == "SUCCESS" {
			delivered = append(delivered, m.MessageID)
		}
		if m.State == "FAILED" {
			for _, s := range m.Sessions {
				sessions = append(sessions, &models.Session{
					UserID:    m.RecipientID,
					SessionID: s.SessionID,
					PublicKey: s.PublicKey,
				})
			}
		}
	}
	err = models.UpdateDistributeMessagesStatusToFinished(ctx, delivered)
	if err != nil {
		return err
	}
	err = models.SyncSession(ctx, client.ClientID, sessions)
	if err != nil {
		return err
	}
	return nil
}

func handleNormalDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest) error {
	err := models.SendMessages(ctx, client, messages)
	if err != nil {
		return err
	}
	var delivered []string
	for _, v := range messages {
		delivered = append(delivered, v.MessageID)
	}
	err = models.UpdateDistributeMessagesStatusToFinished(ctx, delivered)
	if err != nil {
		return err
	}
	return nil
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
