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
	"github.com/MixinNetwork/supergroup/handlers/clients"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/handlers/message"
	"github.com/MixinNetwork/supergroup/jobs"
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

var encryptClientMutex = tools.NewMutex()

func (service *DistributeMessageService) Run(ctx context.Context) error {
	mixin.GetRestyClient().SetTimeout(3 * time.Second)
	go tools.UseAutoFasterRoute()
	jobs.CacheAllBlockUser()

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
			// 删除超时的消息
			if err := session.DB(ctx).Delete(&models.DistributeMessage{},
				"status IN (2,3,6) AND now()-created_at>interval '1 days'").Error; err != nil {
				tools.Println(err)
			}
		}
	}() // 每秒重试未完成的消息服务
	go func() {
		for {
			time.Sleep(time.Second * 5)
			if err := startDistributeMessageIfUnfinished(ctx); err != nil {
				tools.Println(err)
			}
		}
	}()
	pubsub := session.Redis(ctx).QSubscribe(ctx, "distribute")
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			pubsub = session.Redis(ctx).QSubscribe(ctx, "distribute")
			tools.Println(err)
			time.Sleep(time.Second * 2)
			continue
		}
		if msg.Channel == "distribute" {
			go startDistributeMessageByClientID(ctx, msg.Payload)
		} else {
			time.Sleep(time.Second * 2)
			tools.Println(msg.Channel, msg.Payload)
		}
	}
}

var unfinishedNoticeMap = make(map[string]int)

func startDistributeMessageIfUnfinished(ctx context.Context) error {
	clients, err := clients.GetClientList(ctx)
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
				if unfinishedNoticeMap[client.ClientID] > 10 {
					go tools.SendMonitorGroupMsg(fmt.Sprintf("distribute message unfinished %s:%d ,left message: %d", client.ClientID, i, count))
					unfinishedNoticeMap[client.ClientID] = 0
				}
				if count > 200 {
					unfinishedNoticeMap[client.ClientID] = unfinishedNoticeMap[client.ClientID] + 1
				}
				log.Println("startDistributeMessageIfUnfinished", client.ClientID, count)
				go startDistributeMessageByClientID(ctx, client.ClientID)
				break
			}
		}
		unfinishedNoticeMap[client.ClientID] = 0
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
	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return
	}
	if err != nil {
		tools.Println(err)
		return
	}
	if encryptClientMutex.Read(client.ClientID) == nil {
		if !config.Config.Encrypted {
			encryptClientMutex.Write(client.ClientID, false)
		} else {
			me, err := client.UserMe(ctx)
			if err != nil {
				tools.Println(err)
				return
			}
			isEncrypted := false
			for _, v := range me.App.Capabilities {
				if v == "ENCRYPTED" {
					isEncrypted = true
					encryptClientMutex.Write(client.ClientID, true)
					break
				}
			}
			if !isEncrypted {
				encryptClientMutex.Write(client.ClientID, false)
			}
		}
	}
	for i := 0; i < int(config.MessageShardSize); i++ {
		distributeWait[client.ClientID].Add(1)
		i := i
		distributeAntsPool.Submit(func() {
			pendingActiveDistributedMessages(ctx, client.Client, i, client.C.PrivateKey)
		})
	}
	distributeWait[client.ClientID].Wait()
}

func pendingActiveDistributedMessages(ctx context.Context, client *mixin.Client, i int, pk string) {
	// 发送消息
	shardID := strconv.Itoa(i)
	isEncrypted := encryptClientMutex.Read(client.ClientID).(bool)
	for {
		messages, msgOriginMsgIDMap, err := message.PendingActiveDistributedMessages(ctx, client.ClientID, shardID)
		if err != nil {
			tools.Println("PendingActiveDistributedMessages ERROR:", err)
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
			tools.Println("PendingActiveDistributedMessages sendDistributedMessages ERROR:", err, client.ClientID, shardID)
			time.Sleep(time.Duration(i) * time.Millisecond * 100)
			continue
		}
		tools.PrintTimeDuration(fmt.Sprintf("%s:%s:msg send %d...", client.ClientID, shardID, len(messages)), now)
	}
}

func handleEncryptedDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest, pk, shardID string, msgOriginMsgIDMap map[string]string) error {
	var delivered []string
	var unfinishedMsg []*mixin.MessageRequest
	results, err := message.SendEncryptedMessage(ctx, pk, client, messages)
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
	if err := message.SyncSession(ctx, sessions); err != nil {
		return err
	}
	if err := message.UpdateDistributeMessagesStatusToFinished(ctx, client.ClientID, shardID, delivered, msgOriginMsgIDMap); err != nil {
		return err
	}
	if len(unfinishedMsg) > 0 && sessions != nil {
		return handleEncryptedDistributeMsg(ctx, client, unfinishedMsg, pk, shardID, msgOriginMsgIDMap)
	}
	return nil
}

func handleNormalDistributeMsg(ctx context.Context, client *mixin.Client, messages []*mixin.MessageRequest, shardID string, msgOriginMsgIDMap map[string]string) error {
	if err := common.SendMessages(client, messages); err != nil {
		return err
	}
	delivered := make([]string, 0, len(messages))
	for _, v := range messages {
		delivered = append(delivered, v.MessageID)
	}
	if err := message.UpdateDistributeMessagesStatusToFinished(ctx, client.ClientID, shardID, delivered, msgOriginMsgIDMap); err != nil {
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
