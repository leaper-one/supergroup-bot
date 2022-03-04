package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

type DistributeMessageService struct{}

var distributeMutex *tools.Mutex

var distributeWait map[string]*sync.WaitGroup

func (service *DistributeMessageService) Run(ctx context.Context) error {
	distributeMutex = tools.NewMutex()
	distributeWait = make(map[string]*sync.WaitGroup)
	go mixin.UseAutoFasterRoute()
	for _, clientID := range config.Config.ClientList {
		distributeMutex.Write(clientID, false)
		distributeWait[clientID] = &sync.WaitGroup{}
		go startDistributeMessageByClientID(ctx, clientID)
	}

	// 每天删除过期的大群消息
	go func() {
		for {
			time.Sleep(time.Hour * 24)
			if err := models.RemoveOvertimeDistributeMessages(ctx); err != nil {
				session.Logger(ctx).Println(err)
			}
		}
	}() // 每秒重试未完成的消息服务
	go func() {
		for {
			time.Sleep(time.Second)
			if err := startDistributeMessageIfUnfinished(ctx); err != nil {
				session.Logger(ctx).Println(err)
			}
		}
	}()
	pubsub := session.Redis(ctx).Subscribe(ctx, "distribute")
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			pubsub = session.Redis(ctx).Subscribe(ctx, "distribute")
			session.Logger(ctx).Println(err)
			continue
		}
		if msg.Channel == "distribute" {
			go startDistributeMessageByClientID(ctx, msg.Payload)
		} else {
			session.Logger(ctx).Println(msg.Channel, msg.Payload)
		}
	}
}

func startDistributeMessageIfUnfinished(ctx context.Context) error {
	cs := make(map[string]bool)
	keys, err := session.Redis(ctx).Keys(ctx, "s_msg:*").Result()
	if err != nil {
		return err
	}
	for _, key := range keys {
		tmp := strings.Split(key, ":")
		if len(tmp) != 3 {
			session.Logger(ctx).Println(key)
			continue
		}
		clientID := tmp[1]
		cs[clientID] = true
	}
	for clientID := range cs {
		go startDistributeMessageByClientID(ctx, clientID)
	}
	time.Sleep(time.Second * 10)
	return nil
}

func startDistributeMessageByClientID(ctx context.Context, clientID string) {
	m := distributeMutex.Read(clientID)
	if m == nil || m.(bool) {
		return
	}
	client, err := models.GetClientByID(ctx, clientID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	mixinClient, err := mixin.NewFromKeystore(&mixin.Keystore{
		ClientID:   client.ClientID,
		SessionID:  client.SessionID,
		PrivateKey: client.PrivateKey,
		PinToken:   client.PinToken,
	})
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	distributeMutex.Write(client.ClientID, true)
	for i := 0; i < int(config.MessageShardSize); i++ {
		distributeWait[client.ClientID].Add(1)
		go pendingActiveDistributedMessages(ctx, mixinClient, i, client.PrivateKey)
	}
	distributeWait[client.ClientID].Wait()
	distributeMutex.Write(client.ClientID, false)
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
		messages, msgOriginMsgIDMap, err := models.PendingActiveDistributedMessages(ctx, client.ClientID, shardID)
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages ERROR:", err)
			time.Sleep(time.Duration(i) * time.Millisecond * 100)
			continue
		}
		if len(messages) < 1 {
			distributeWait[client.ClientID].Done()
			return
		}
		messages = handleMsg(messages)
		now := time.Now()
		if isEncrypted {
			err = handleEncryptedDistributeMsg(ctx, client, messages, pk, shardID, msgOriginMsgIDMap)
		} else {
			err = handleNormalDistributeMsg(ctx, client, messages, shardID, msgOriginMsgIDMap)
		}
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages sendDistributedMessges ERROR:", err)
			time.Sleep(time.Duration(i) * time.Millisecond * 100)
			continue
		}
		tools.PrintTimeDuration(fmt.Sprintf("%s:%s:msg send %d...", client.ClientID, shardID, len(messages)), now)
	}
}

func handleEncryptedDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest, pk, shardID string, msgOriginMsgIDMap map[string]*models.DistributeMessage) error {
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
	if err := models.UpdateDistributeMessagesStatusToFinished(ctx, client.ClientID, shardID, delivered, msgOriginMsgIDMap); err != nil {
		return err
	}
	if err := models.SyncSession(ctx, client.ClientID, sessions); err != nil {
		return err
	}
	return nil
}

func handleNormalDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest, shardID string, msgOriginMsgIDMap map[string]*models.DistributeMessage) error {
	if err := models.SendMessages(ctx, client, messages); err != nil {
		return err
	}
	var delivered []string
	for _, v := range messages {
		delivered = append(delivered, v.MessageID)
	}
	if err := models.UpdateDistributeMessagesStatusToFinished(ctx, client.ClientID, shardID, delivered, msgOriginMsgIDMap); err != nil {
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
