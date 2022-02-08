package services

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/fox-one/mixin-sdk-go"
)

type BlazeService struct {
}

func (b *BlazeService) Run(ctx context.Context) error {
	go mixin.UseAutoFasterRoute()
	clientList, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}
	for _, client := range clientList {
		go connectMixinSDKClient(ctx, client)
	}
	select {}
}

type mixinBlazeHandler func(ctx context.Context, msg bot.MessageView, clientID string) error

func (f mixinBlazeHandler) OnAckReceipt(ctx context.Context, msg bot.MessageView, clientID string) error {
	go models.UpdateClientUserDeliverTime(ctx, clientID, msg.MessageId, msg.CreatedAt, msg.Status)
	return nil
}

func (f mixinBlazeHandler) SyncAck() bool {
	return false
}

func (f mixinBlazeHandler) OnMessage(ctx context.Context, msg bot.MessageView, clientID string) error {
	return f(ctx, msg, clientID)
}

func connectMixinSDKClient(ctx context.Context, c *models.Client) {
	batchAckMap := newAckMap()
	go batchAckMsg(ctx, batchAckMap, c.ClientID, c.SessionID, c.PrivateKey)
	h := func(ctx context.Context, botMsg bot.MessageView, clientID string) error {
		if botMsg.Category == mixin.MessageCategorySystemConversation {
			return nil
		}
		msg := mixin.MessageView{
			ConversationID:   botMsg.ConversationId,
			UserID:           botMsg.UserId,
			MessageID:        botMsg.MessageId,
			Category:         botMsg.Category,
			Data:             botMsg.Data,
			RepresentativeID: botMsg.RepresentativeId,
			QuoteMessageID:   botMsg.QuoteMessageId,
			Status:           botMsg.Status,
			Source:           botMsg.Source,
			CreatedAt:        botMsg.CreatedAt,
			UpdatedAt:        botMsg.UpdatedAt,
		}
		if botMsg.Category == mixin.MessageCategorySystemAccountSnapshot {
			if err := models.ReceivedSnapshot(ctx, clientID, &msg); err != nil {
				return err
			}
		} else if err := models.ReceivedMessage(ctx, clientID, &msg); err != nil {
			return err
		}
		batchAckMap.set(msg.MessageID)
		return nil
	}

	for {
		client := bot.NewBlazeClient(c.ClientID, c.SessionID, c.PrivateKey)
		if err := client.Loop(ctx, mixinBlazeHandler(h)); err != nil {
			if !ignoreLoopBlazeError(err) {
				log.Println("blaze", err)
			}
		}
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

func batchAckMsg(ctx context.Context, m *ackMap, uid, sid, key string) {
	for {
		if len(m.m) == 0 {
			time.Sleep(time.Second)
			continue
		}
		req := make([]*bot.ReceiptAcknowledgementRequest, 0)
		msgIDs := make([]string, 0)
		for msgID := range m.m {
			req = append(req, &bot.ReceiptAcknowledgementRequest{
				MessageId: msgID,
				Status:    "READ",
			})
			msgIDs = append(msgIDs, msgID)
		}
		if len(req) > 100 {
			req = req[0:100]
			msgIDs = msgIDs[0:100]
		}
		if err := bot.PostAcknowledgements(ctx, req, uid, sid, key); err == nil {
			m.remove(msgIDs)
		}
		if len(req) != 100 {
			time.Sleep(100 * time.Millisecond)
		}
	}
}

type ackMap struct {
	mutex sync.Mutex
	m     map[string]bool
}

func newAckMap() *ackMap {
	return &ackMap{
		m: make(map[string]bool),
	}
}

func (m *ackMap) set(key string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.m[key] = true
}

func (m *ackMap) remove(keys []string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, key := range keys {
		delete(m.m, key)
	}
}
