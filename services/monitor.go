package services

import (
	"context"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
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
		ClientID:   config.Config.Monitor.ClientID,
		SessionID:  config.Config.Monitor.SessionID,
		PrivateKey: config.Config.Monitor.PrivateKey,
	})

	for _, client := range list {
		c, err := models.GetClientByID(ctx, client.ClientID)
		if err != nil {
			session.Logger(ctx).Println(err)
		}
		models.SendMonitorGroupMsg(ctx, msgClient, fmt.Sprintf("%s 消息监控已开启...", c.Name))
		go monitor(ctx, msgClient, c)
	}
	select {}
}

func monitor(ctx context.Context, msgClient *mixin.Client, c models.Client) {
	var oriTime time.Time
	for {
		time.Sleep(time.Second * 60)
		curTime := models.GetRemotePendingMsg(ctx, c.ClientID)
		if curTime.IsZero() {
			continue
		}
		if !curTime.Equal(oriTime) {
			oriTime = curTime
			continue
		}
		models.SendMonitorGroupMsg(ctx, msgClient, fmt.Sprintf("%s 有条消息卡了一分钟...时间为 %s", c.Name, oriTime))
	}
}
