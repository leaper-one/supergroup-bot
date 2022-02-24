package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MixinNetwork/supergroup/models"
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
	list, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}
	needReInit = SafeUpdater{v: make(map[string]time.Time)}

	for _, client := range list {
		needReInit.v[client.ClientID] = time.Now()
		createMutex.Write(client.ClientID, false)
		if err := models.InitShardID(ctx, client.ClientID); err != nil {
			session.Logger(ctx).Println(err)
		} else {
			go mutexCreateMsg(ctx, client.ClientID)
		}
	}

	pubsub := session.Redis(ctx).Subscribe(ctx, "create")
	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}
		if msg.Channel == "create" {
			go mutexCreateMsg(ctx, msg.Payload)
		} else {
			session.Logger(ctx).Println(msg.Channel, msg.Payload)
		}
	}
}

func (s *SafeUpdater) Update(ctx context.Context, clientID string, t time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.v[clientID] = t
	models.InitShardID(ctx, clientID)
}

var needReInit SafeUpdater
var createMutex *tools.Mutex

func reInitShardID(ctx context.Context, clientID string) {
	if needReInit.v[clientID].Add(time.Hour).Before(time.Now()) {
		needReInit.Update(ctx, clientID, time.Now())
	}
}

func mutexCreateMsg(ctx context.Context, clientID string) {
	m := createMutex.Read(clientID)
	if m == nil {
		return
	}
	if m.(bool) {
		return
	}
	createMutex.Write(clientID, true)
	createMsg(ctx, clientID)
	createMutex.Write(clientID, false)
}

func createMsg(ctx context.Context, clientID string) {
	for {
		count := createMsgByPriority(ctx, clientID, models.MessageStatusPending)
		if count != 0 {
			continue
		}
		count = createMsgByPriority(ctx, clientID, models.MessageStatusPrivilege)
		if count != 0 {
			continue
		}
		reInitShardID(ctx, clientID)
		return
	}
}

func createMsgByPriority(ctx context.Context, clientID string, msgStatus int) int {
	now := time.Now()
	msgs, err := models.GetPendingMessageByClientID(ctx, clientID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return 0
	}
	if len(msgs) == 0 {
		return 0
	}
	for _, msg := range msgs {
		status, err := session.Redis(ctx).Get(ctx, "msg_status:"+msg.MessageID).Int()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				status = 0
			} else {
				session.Logger(ctx).Println(err)
				return 0
			}
		}
		if msgStatus == models.MessageStatusPending {
			// 要创建优先级高的消息
			if status == 0 {
				if err := createDistributeMsg(ctx, clientID, models.ClientUserPriorityHigh, msg); err != nil {
					session.Logger(ctx).Println(err)
					return 0
				}
				tools.PrintTimeDuration(clientID+"创建消息...", now)
				return 1
			}
			if status == models.MessageStatusPrivilege ||
				status == models.MessageRedisStatusFinished ||
				status == models.MessageStatusFinished {
				// 已经创建了优先级高的消息了
				continue
			}
			session.Logger(ctx).Println("unknown msg status", msgStatus, status)
		} else if msgStatus == models.MessageStatusPrivilege {
			// 要创建优先级低的消息了
			if status == models.MessageStatusPrivilege {
				if err := createDistributeMsg(ctx, clientID, models.ClientUserPriorityLow, msg); err != nil {
					session.Logger(ctx).Println(err)
					return 0
				}
				tools.PrintTimeDuration(clientID+"创建消息...", now)
				return 1
			}
			if status == models.MessageStatusFinished ||
				status == models.MessageRedisStatusFinished {
				// 已经创建了优先级低的消息了
				continue
			}
			session.Logger(ctx).Println("unknown msg status", msgStatus, status) // 2 0
		}
	}
	return 0
}

func createDistributeMsg(ctx context.Context, clientID string, level int, msg *models.Message) error {
	return models.CreateDistributeMsgAndMarkStatus(ctx, clientID, &mixin.MessageView{
		UserID:         msg.UserID,
		MessageID:      msg.MessageID,
		Category:       msg.Category,
		Data:           msg.Data,
		QuoteMessageID: msg.QuoteMessageID,
	}, []int{level})
}
