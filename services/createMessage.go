package services

import (
	"context"
	"errors"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"time"
)

type CreateDistributeMsgService struct{}

func (service *CreateDistributeMsgService) Run(ctx context.Context) error {
	list, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}

	for _, client := range list {
		if err := models.InitShardID(ctx, client.ClientID); err != nil {
			session.Logger(ctx).Println(err)
		} else {
			go createMsg(ctx, client.ClientID)
		}
	}

	select {}
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
		time.Sleep(time.Second)
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
