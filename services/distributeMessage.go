package services

import (
	"context"
	"errors"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"time"
)

type DistributeMessageService struct{}

func (service *DistributeMessageService) Run(ctx context.Context) error {

	clientList, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}

	for _, c := range clientList {
		client, err := mixin.NewFromKeystore(&mixin.Keystore{
			ClientID:   c.ClientID,
			SessionID:  c.SessionID,
			PrivateKey: c.PrivateKey,
			PinToken:   c.PinToken,
		})
		if err != nil {
			return err
		}
		go startDistributeMessageByClientID(ctx, client)
	}

	select {}
}

func startDistributeMessageByClientID(ctx context.Context, client *mixin.Client) {
	for {
		// 1. 删除超过三天的消息
		if err := models.RemoveOvertimeDistributeMessages(ctx); err != nil {
			session.Logger(ctx).Println(err)
			continue
		}
		// 2. 发送优先级高的消息
		if err := sendMsgByLevel(ctx, client, models.DistributeMessageLevelHigher); err == nil {
			// 发送消息成功，查看下次消息
			continue
		} else if errors.Is(err, pgx.ErrNoRows) {
			// 没有优先级高的消息了
			if err := sendMsgByLevel(ctx, client, models.DistributeMessageLevelLower); err == nil {
				continue
			} else if !errors.Is(err, pgx.ErrNoRows) {
				session.Logger(ctx).Println(err)
			} else {
				time.Sleep(time.Second)
			}
		} else {
			// 其他错误
			session.Logger(ctx).Println(err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func sendMsgByLevel(ctx context.Context, client *mixin.Client, level int) error {
	var msgStatus int
	if level == models.DistributeMessageLevelHigher {
		msgStatus = models.MessageStatusPrivilege
	} else {
		msgStatus = models.MessageStatusNormal
	}
	msg, err := models.GetLongestMessageByStatus(ctx, client.ClientID, msgStatus)
	if err != nil {
		return err
	}
	// 发送消息
	msgList, err := models.GetDistributeMessageByClientIDAndLevel(ctx, client.ClientID, msg, level)
	if err != nil {
		return err
	}
	if len(msgList) != 0 {
		if err := models.SendBatchMessages(ctx, client, msgList); err != nil {
			return err
		}
	}
	// 全部发送完毕
	var status int
	if level == models.DistributeMessageLevelHigher {
		status = models.MessageStatusNormal
	} else {
		status = models.MessageStatusFinished
	}
	_, err = session.Database(ctx).Exec(ctx, `UPDATE messages SET status=$3 WHERE client_id=$1 AND message_id=$2`, client.ClientID, msg.MessageID, status)
	if len(msgList) == 0 {
		return pgx.ErrNoRows
	}
	return err
}
