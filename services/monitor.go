package services

import (
	"context"
	"fmt"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"time"
)

type MonitorService struct{}

func (service *MonitorService) Run(ctx context.Context) error {
	// 最重要的指标是：
	// 1. 代发中 且 created_at 最远的一条消息 1分钟内一样。 说明这条消息卡了一分钟了...
	list, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}

	msgClient, err := mixin.NewFromKeystore(&mixin.Keystore{
		ClientID:   config.MonitorClientID,
		SessionID:  config.MonitorSessionID,
		PrivateKey: config.MonitorPrivateKeyID,
	})

	for _, client := range list {
		c, err := models.GetClientByID(ctx, client.ClientID)
		if err != nil {
			session.Logger(ctx).Println(err)
		}
		go monitor(ctx, msgClient, c)
	}
	select {}
}

func monitor(ctx context.Context, msgClient *mixin.Client, c models.Client) {
	var oriTime time.Time
	for {
		time.Sleep(time.Second * 5)
		curTime := models.GetRemotePendingMsg(ctx, c.ClientID)
		if curTime.IsZero() {
			continue
		}
		if !curTime.Equal(oriTime) {
			oriTime = curTime
			continue
		}
		msg := fmt.Sprintf("%s 有条消息卡了一分钟...时间为 %s", c.Name, oriTime)
		if err := msgClient.SendMessage(ctx, &mixin.MessageRequest{
			ConversationID: config.MonitorConversationID,
			Data:           tools.Base64Encode([]byte(msg)),
			Category:       mixin.MessageCategoryPlainText,
			MessageID:      tools.GetUUID(),
		}); err != nil {
			session.Logger(ctx).Println(err)
		}
	}
}
