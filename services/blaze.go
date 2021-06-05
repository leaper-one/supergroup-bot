package services

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/fox-one/mixin-sdk-go"
)

type BlazeService struct {
}

type blazeHandler func(ctx context.Context, msg *mixin.MessageView, clientID string) error

func (f blazeHandler) OnAckReceipt(ctx context.Context, msg *mixin.MessageView, clientID string) error {
	if msg.Status == "DELIVERED" {
		if err := models.UpdateClientUserDeliverTime(ctx, clientID, msg.MessageID, msg.CreatedAt); err != nil {
			return err
		}
	}
	return nil
}

func (f blazeHandler) OnMessage(ctx context.Context, msg *mixin.MessageView, clientID string) error {
	return f(ctx, msg, clientID)
}

func (b *BlazeService) Run(ctx context.Context) error {
	clientList, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}
	for _, client := range clientList {
		go connectClient(ctx, client)
	}
	select {}
}

func connectClient(ctx context.Context, c *models.Client) {
	client, err := mixin.NewFromKeystore(&mixin.Keystore{
		ClientID:   c.ClientID,
		SessionID:  c.SessionID,
		PrivateKey: c.PrivateKey,
		PinToken:   c.PinToken,
	})
	if err != nil {
		log.Panicln(err)
	}

	h := func(ctx context.Context, msg *mixin.MessageView, clientID string) error {
		if msg.Category == mixin.MessageCategorySystemConversation {
			return nil
		}
		if msg.Category == mixin.MessageCategorySystemAccountSnapshot {
			return nil
		}
		if err := models.ReceivedMessage(ctx, clientID, *msg); err != nil {
			return err
		}
		//if userID, _ := uuid.FromString(msg.UserID); userID == uuid.Nil {
		//	return nil
		//}
		//_, _ = models.SearchUserByID(ctx, msg.UserID, c.ClientID)
		return nil
	}

	for {
		if err := client.LoopBlaze(ctx, blazeHandler(h)); err != nil {
			if !ignoreLoopBlazeError(err) {
				log.Printf("LoopBlaze: %s", err.Error())
			}
		} else {
			log.Println(err, "啥情况")
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
