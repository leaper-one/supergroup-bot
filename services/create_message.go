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
	"github.com/jackc/pgx/v4"
)

type CreateDistributeMsgService struct{}

type SafeUpdater struct {
	mu sync.Mutex
	v  map[string]time.Time
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

func mutexCreateMsg(ctx context.Context, clientID string) {
	if createMutex.Read(clientID).(bool) {
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
	now := time.Now().UnixNano()
	msg, err := models.GetLongestMessageByStatus(ctx, clientID, msgStatus)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			session.Logger(ctx).Println(err)
		}
		return 0
	}
	// 1.1 拿到消息没有错误
	var level int
	if msgStatus == models.MessageStatusPending {
		level = models.ClientUserPriorityHigh
	} else if msgStatus == models.MessageStatusPrivilege {
		level = models.ClientUserPriorityLow
	}
	if err := models.CreateDistributeMsgAndMarkStatus(ctx, clientID, &mixin.MessageView{
		UserID:         msg.UserID,
		MessageID:      msg.MessageID,
		Category:       msg.Category,
		Data:           msg.Data,
		QuoteMessageID: msg.QuoteMessageID,
	}, []int{level}); err != nil {
		session.Logger(ctx).Println(err)
	}
	tools.PrintTimeDuration(clientID+"创建消息...", now)
	return 1
}
