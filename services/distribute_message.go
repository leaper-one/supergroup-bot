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
	"github.com/go-redis/redis/v8"
	"github.com/panjf2000/ants/v2"
)

type DistributeMessageService struct{}

var distributeMutex = tools.NewMutex()
var distributeWait = make(map[string]*sync.WaitGroup)
var distributeAntsPool, _ = ants.NewPool(500, ants.WithPreAlloc(true), ants.WithMaxBlockingTasks(50))

func (service *DistributeMessageService) Run(ctx context.Context) error {
	go mixin.UseAutoFasterRoute()
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
	clients, err := models.GetAllClient(ctx)
	if err != nil {
		return err
	}
	canClean := true
	for _, clientID := range clients {
		for i := 0; i < int(config.MessageShardSize); i++ {
			count, err := session.Redis(ctx).R.ZCard(ctx, fmt.Sprintf("s_msg:%s:%d", clientID, i)).Result()
			if err != nil {
				return err
			}
			if count > 0 {
				canClean = false
				go startDistributeMessageByClientID(ctx, clientID)
			}
		}
	}
	if canClean {
		go _cleanMsg(ctx)
	}
	return nil
}

var cleanMutex = tools.NewMutex()

func _cleanMsg(ctx context.Context) {
	if cleanMutex.Read("is_clean") != nil {
		return
	}
	cleanMutex.Write("is_clean", true)
	defer cleanMutex.Write("is_clean", nil)
	lMsgKeyList, err := session.Redis(ctx).QKeys(ctx, "l_msg:*")
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	if len(lMsgKeyList) == 0 {
		return
	}
	results := make(map[string]*redis.StringCmd)
	if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, key := range lMsgKeyList {
			results[key] = p.Get(ctx, key)
		}
		return nil
	}); err != nil {
		session.Logger(ctx).Println(err)
		return
	}

	if _, err = session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for key, _lMsgCount := range results {
			lMsgCountStr, err := _lMsgCount.Result()
			if err != nil {
				session.Logger(ctx).Println(err)
				continue
			}
			lMsgCount, err := strconv.Atoi(lMsgCountStr)
			if err != nil {
				session.Logger(ctx).Println(err)
				continue
			}
			if lMsgCount <= 0 {
				if err := p.Unlink(ctx, key).Err(); err != nil {
					session.Logger(ctx).Println(err)
					continue
				}
			}
		}
		return nil
	}); err != nil {
		session.Logger(ctx).Println(err)
	}
}

func startDistributeMessageByClientID(ctx context.Context, clientID string) {
	m := distributeMutex.Read(clientID)
	if m == nil || m.(bool) {
		return
	}
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
	distributeMutex.Write(client.ClientID, true)
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

func handleEncryptedDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest, pk, shardID string, msgOriginMsgIDMap map[string]string) error {
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

func handleNormalDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest, shardID string, msgOriginMsgIDMap map[string]string) error {
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
