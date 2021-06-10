package services

import (
	"context"
	"crypto/md5"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
	"math/big"
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

	for {
		if err := models.RemoveOvertimeDistributeMessages(ctx); err != nil {
			time.Sleep(time.Minute)
		}
	}
}

func startDistributeMessageByClientID(ctx context.Context, client *mixin.Client) {
	for i := int64(0); i < config.MessageShardSize; i++ {
		shard := shardId(config.MessageShardModifier, i)
		go pendingActiveDistributedMessages(ctx, client, shard)
	}

}

func shardId(modifier string, i int64) string {
	h := md5.New()
	h.Write([]byte(modifier))
	h.Write(new(big.Int).SetInt64(i).Bytes())
	s := h.Sum(nil)
	s[6] = (s[6] & 0x0f) | 0x30
	s[8] = (s[8] & 0x3f) | 0x80
	id, err := uuid.FromBytes(s)
	if err != nil {
		panic(err)
	}
	return id.String()
}

func pendingActiveDistributedMessages(ctx context.Context, client *mixin.Client, shardID string) {
	// 发送消息
	for {
		messages, err := models.PendingActiveDistributedMessages(ctx, client.ClientID, shardID)
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages ERROR:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if len(messages) < 1 {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		err = models.SendMessages(ctx, client, messages)
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages sendDistributedMessges ERROR:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
		err = models.UpdateMessagesStatusToFinished(ctx, messages)
		if err != nil {
			session.Logger(ctx).Println("PendingActiveDistributedMessages UpdateMessagesStatus ERROR:", err)
			time.Sleep(100 * time.Millisecond)
			continue
		}
	}
}
