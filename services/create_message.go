package services

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
)

type CreateDistributeMsgService struct{}

type SafeUpdater struct {
	mu sync.Mutex
	v  map[string]time.Time
}

func (service *CreateDistributeMsgService) Run(ctx context.Context) error {
	createMutex = tools.NewMutex()
	list, err := common.GetClientList(ctx)

	go common.CacheAllBlockUser()

	if err != nil {
		return err
	}
	needReInit = SafeUpdater{v: make(map[string]time.Time)}

	for _, client := range list {
		needReInit.Update(ctx, client.ClientID, time.Now())
		createMutex.Write(client.ClientID, false)
		if err := common.InitShardID(ctx, client.ClientID); err != nil {
			tools.Println(err)
		} else {
			go mutexCreateMsg(ctx, client.ClientID)
		}
	}

	pubsub := session.Redis(ctx).QSubscribe(ctx, "create")
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}
		if msg.Channel == "create" {
			go mutexCreateMsg(ctx, msg.Payload)
		} else {
			tools.Println(msg.Channel, msg.Payload)
		}
	}
}

func (s *SafeUpdater) Update(ctx context.Context, clientID string, t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.v[clientID] = t
	common.InitShardID(ctx, clientID)
}

func (s *SafeUpdater) Get(ctx context.Context, clientID string) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.v[clientID]
}

var needReInit SafeUpdater
var createMutex *tools.Mutex

func reInitShardID(ctx context.Context, clientID string) {
	if needReInit.Get(ctx, clientID).Add(time.Hour).Before(time.Now()) {
		needReInit.Update(ctx, clientID, time.Now())
	}
}

func mutexCreateMsg(ctx context.Context, clientID string) {
	m := createMutex.Read(clientID)
	if m == nil || m.(bool) {
		return
	}
	createMutex.Write(clientID, true)
	defer createMutex.Write(clientID, false)
	createMsg(ctx, clientID)
}

func createMsg(ctx context.Context, clientID string) {
	for {
		min := tools.GetMinuteTime(time.Now())
		_count, err := session.Redis(ctx).SyncGet(ctx, fmt.Sprintf("client_msg_count:%s:%s", clientID, min)).Int()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				tools.Println(err)
			}
		} else {
			if _count >= 100000 {
				time.Sleep(time.Duration(tools.GetNextMinuteTime(min)))
			}
		}
		count := createMsgByPriority(ctx, clientID)
		if count != 0 {
			continue
		}
		reInitShardID(ctx, clientID)
		return
	}
}

func createMsgByPriority(ctx context.Context, clientID string) int {
	now := time.Now()
	msgs, err := common.GetPendingMessageByClientID(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return 0
	}
	if len(msgs) == 0 {
		return 0
	}
	for _, msg := range msgs {
		status, err := session.Redis(ctx).SyncGet(ctx, "msg_status:"+msg.MessageID).Int()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				status = 0
			} else {
				tools.Println(err)
				return 0
			}
		}
		if status == 0 {
			if err := common.CreateDistributeMsgAndMarkStatus(ctx, clientID, &mixin.MessageView{
				UserID:         msg.UserID,
				MessageID:      msg.MessageID,
				Category:       msg.Category,
				Data:           msg.Data,
				QuoteMessageID: msg.QuoteMessageID,
			}); err != nil {
				tools.Println(err)
				return 0
			}
			tools.PrintTimeDuration(clientID+"创建消息...", now)
			return 1
		}
		if status == common.MessageStatusFinished ||
			status == common.MessageRedisStatusFinished {
			// 已经创建了优先级高的消息了
			continue
		}
		tools.Println("unknown msg status::", status)
	}
	return 0
}
