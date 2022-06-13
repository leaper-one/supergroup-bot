package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/panjf2000/ants/v2"
)

type DistributeMessageService struct{}

var distributeMutex = tools.NewMutex()
var distributeWait = make(map[string]*sync.WaitGroup)
var distributeAntsPool, _ = ants.NewPool(500, ants.WithPreAlloc(true), ants.WithMaxBlockingTasks(50))

var encrypClientMutex = tools.NewMutex()

func (service *DistributeMessageService) Run(ctx context.Context) error {
	mixin.GetRestyClient().SetTimeout(3 * time.Second)
	go tools.UseAutoFasterRoute()
	go models.CacheAllBlockUser()

	for _, clientID := range config.Config.ClientList {
		distributeMutex.Write(clientID, false)
		distributeWait[clientID] = &sync.WaitGroup{}
		go startDistributeMessageByClientID(ctx, clientID)
	}

	go func() {
		for {
			runningCount := distributeAntsPool.Running()
			if runningCount > 300 {
				log.Println("distributeAntsPool running:", runningCount)
			}
			time.Sleep(time.Second)
		}
	}()

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
			time.Sleep(time.Second * 5)
			if err := startDistributeMessageIfUnfinished(ctx); err != nil {
				session.Logger(ctx).Println(err)
			}
		}
	}()
	pubsub := session.Redis(ctx).QSubscribe(ctx, "distribute")
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			pubsub = session.Redis(ctx).QSubscribe(ctx, "distribute")
			session.Logger(ctx).Println(err)
			time.Sleep(time.Second * 2)
			continue
		}
		if msg.Channel == "distribute" {
			go startDistributeMessageByClientID(ctx, msg.Payload)
		} else {
			time.Sleep(time.Second * 2)
			session.Logger(ctx).Println(msg.Channel, msg.Payload)
		}
	}
}

func startDistributeMessageIfUnfinished(ctx context.Context) error {
	clients, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}
	for _, client := range clients {
		for i := 0; i < int(config.MessageShardSize); i++ {
			count, err := session.Redis(ctx).R.ZCard(ctx, fmt.Sprintf("s_msg:%s:%d", client.ClientID, i)).Result()
			if err != nil {
				return err
			}
			if count > 0 {
				log.Println("startDistributeMessageIfUnfinished", client.ClientID, count)
				go startDistributeMessageByClientID(ctx, client.ClientID)
				break
			}
		}
	}
	return nil
}

func startDistributeMessageByClientID(ctx context.Context, clientID string) {
	m := distributeMutex.Read(clientID)
	if m == nil || m.(bool) {
		return
	}
	distributeMutex.Write(clientID, true)
	defer distributeMutex.Write(clientID, false)
	client, err := models.GetClientByIDOrHost(ctx, clientID)
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
	if config.Config.Encrypted && encrypClientMutex.Read(client.ClientID) == nil {
		me, err := mixinClient.UserMe(ctx)
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
		isEncrypted := false
		for _, v := range me.App.Capabilities {
			if v == "ENCRYPTED" {
				isEncrypted = true
				encrypClientMutex.Write(client.ClientID, true)
				break
			}
		}
		if !isEncrypted {
			encrypClientMutex.Write(client.ClientID, false)
		}
	} else {
		encrypClientMutex.Write(client.ClientID, false)
	}
	fn := func(i int) func() {
		return func() {
			pendingActiveDistributedMessages(ctx, mixinClient, i, client.PrivateKey)
		}
	}
	for i := 0; i < int(config.MessageShardSize); i++ {
		distributeWait[client.ClientID].Add(1)
		distributeAntsPool.Submit(fn(i))
	}
	distributeWait[client.ClientID].Wait()
}

func pendingActiveDistributedMessages(ctx context.Context, client *mixin.Client, i int, pk string) {
	// 发送消息
	shardID := strconv.Itoa(i)
	isEncrypted := encrypClientMutex.Read(client.ClientID).(bool)
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

func handleEncryptedDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest, pk, shardID string, msgOriginMsgIDMap map[string]string) error {
	var delivered []string
	var unfinishedMsg []*mixin.MessageRequest
	results, err := models.SendEncryptedMessage(ctx, pk, client, messages)
	if err != nil {
		return err
	}
	var sessions []*models.Session
	for i, m := range results {
		if m.State == "SUCCESS" {
			delivered = append(delivered, m.MessageID)
		}
		if m.State == "FAILED" {
			if m.Sessions == nil {
				delivered = append(delivered, m.MessageID)
				continue
			}
			for _, s := range m.Sessions {
				sessions = append(sessions, &models.Session{
					UserID:    m.RecipientID,
					SessionID: s.SessionID,
					PublicKey: s.PublicKey,
				})
			}
			unfinishedMsg = append(unfinishedMsg, messages[i])
		}
	}
	if err := models.UpdateDistributeMessagesStatusToFinished(ctx, client.ClientID, shardID, delivered, msgOriginMsgIDMap); err != nil {
		return err
	}
	if err := models.SyncSession(ctx, client.ClientID, sessions); err != nil {
		return err
	}
	if len(unfinishedMsg) > 0 && sessions != nil {
		return handleEncryptedDistributeMsg(ctx, client, unfinishedMsg, pk, shardID, msgOriginMsgIDMap)
	}
	return nil
}

func handleNormalDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest, shardID string, msgOriginMsgIDMap map[string]string) error {
	if err := models.SendMessages(ctx, client, messages); err != nil {
		return err
	}
	delivered := make([]string, 0, len(messages))
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
	msgStr, _ := json.Marshal(messages)
	if len(msgStr) < maxLimit {
		return messages
	}
	totalSize := 0
	size := 0
	for ; size < len(messages); size++ {
		msgStr, _ := json.Marshal(messages[size])
		totalSize += len(msgStr)
		if totalSize > maxLimit || size == 100 {
			break
		}
	}
	return messages[0:size]
}
