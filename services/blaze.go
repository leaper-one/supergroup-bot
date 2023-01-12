package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/clients"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/handlers/message"
	"github.com/MixinNetwork/supergroup/handlers/snapshot"
	"github.com/MixinNetwork/supergroup/jobs"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/panjf2000/ants/v2"
)

var ackAntsPool *ants.Pool

type BlazeService struct {
}

var i uint64

func (b *BlazeService) Run(ctx context.Context) error {
	ackAntsPool, _ = ants.NewPool(1000, ants.WithPreAlloc(true), ants.WithMaxBlockingTasks(2000))
	go tools.UseAutoFasterRoute()
	jobs.CacheAllBlockUser()
	go func() {
		for {
			runningCount := ackAntsPool.Running()
			if runningCount == 1000 {
				log.Println("ackAntsPool running:", runningCount, i)
			}
			time.Sleep(time.Second)
		}
	}()
	clientList, err := clients.GetClientList(ctx)
	if err != nil {
		return err
	}
	for _, client := range clientList {
		go connectFoxSDKClient(ctx, client)
	}
	select {}
}

type blazeHandler func(ctx context.Context, msg *mixin.MessageView, clientID string) error

func (f blazeHandler) OnAckReceipt(ctx context.Context, msg *mixin.MessageView, clientID string) error {
	go UpdateClientUserActiveTimeToRedis(clientID, msg.MessageID, msg.CreatedAt, msg.Status)
	return nil
}

func (f blazeHandler) OnMessage(ctx context.Context, msg *mixin.MessageView, clientID string) error {
	return f(ctx, msg, clientID)
}
func connectFoxSDKClient(ctx context.Context, c *models.Client) {
	client, err := mixin.NewFromKeystore(&mixin.Keystore{
		ClientID:   c.ClientID,
		SessionID:  c.SessionID,
		PrivateKey: c.PrivateKey,
		PinToken:   c.PinToken,
	})
	if err != nil {
		log.Panicln(err)
	}

	h := func(ctx context.Context, _msg *mixin.MessageView, clientID string) error {
		msg := *_msg
		if msg.Category == mixin.MessageCategorySystemConversation {
			return nil
		}
		if msg.Category == mixin.MessageCategorySystemAccountSnapshot {
			if err := snapshot.ReceivedSnapshot(ctx, clientID, &msg); err != nil {
				return err
			}
			return nil
		}
		if err := message.ReceivedMessage(ctx, clientID, &msg); err != nil {
			tools.Println(err)
			return err
		}
		return nil
	}

	for {
		if err := client.LoopBlaze(ctx, blazeHandler(h)); err != nil {
			if !ignoreLoopBlazeError(err) {
				log.Printf("LoopBlaze: %s, id: %s", err.Error(), c.ClientID)
			}
		}
		time.Sleep(time.Second)
	}
}

var ignoreMessage = []string{"1006", "timeout", "connection reset by peer"}

func ignoreLoopBlazeError(err error) bool {
	for _, s := range ignoreMessage {
		if strings.Contains(err.Error(), s) {
			return true
		}
	}
	return false
}

func UpdateClientUserActiveTimeToRedis(clientID, msgID string, deliverTime time.Time, status string) error {
	if status != "DELIVERED" && status != "READ" {
		return nil
	}
	ctx := models.Ctx
	dm, err := message.GetDistributeMsgByMsgIDFromRedis(ctx, msgID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		tools.Println(err)
		return err
	}
	user, err := common.GetClientUserByClientIDAndUserID(ctx, clientID, dm.UserID)
	if err != nil {
		return err
	}
	go message.ActiveUser(&user)
	if status == "READ" {
		if err := session.Redis(ctx).QSet(ctx, fmt.Sprintf("ack_msg:read:%s:%s", clientID, user.UserID), deliverTime, time.Hour*2); err != nil {
			return err
		}
	} else {
		if err := session.Redis(ctx).QSet(ctx, fmt.Sprintf("ack_msg:deliver:%s:%s", clientID, user.UserID), deliverTime, time.Hour*2); err != nil {
			return err
		}
	}
	return nil
}
