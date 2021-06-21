package services

import (
	"context"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/go-redis/redis/v8"
	"log"
	"math/rand"
	"time"
)

type TestService struct{}

func (service *TestService) Run(ctx context.Context) error {
	t := []string{"1"}
	t = t[1:]
	log.Println(t)
	sub := session.Redis(ctx).Subscribe(ctx, "test")
	for {
		iface, err := sub.Receive(ctx)
		if err != nil {
			session.Logger(ctx).Println(err)
			return err
		}

		switch iface.(type) {
		case *redis.Subscription:
			// subscribe success
			log.Println("subscribe success")
		case *redis.Message:
			log.Println(iface.(*redis.Message).Payload)
		case *redis.Pong:
			log.Println("pong...")
		default:
			_ = sub.Unsubscribe(ctx, "test")

		}
		//sub.Channel()
	}
	//session.Redis(ctx).PubSubChannels(ctx, "test")
	//session.Redis(ctx).Publish(ctx, "test", "msg1").Err()
	return nil
}

func execAllClient(ctx context.Context) {
	list, err := models.GetClientList(ctx)
	if err != nil {
		return
	}

	for _, client := range list {
		if err := addClientUser(ctx, client.ClientID, "f59b9309-70c2-4b69-8fd8-5773dbd10018", models.ClientUserStatusGuest, models.ClientUserPriorityHigh); err != nil {
			session.Logger(ctx).Println(err)
		}
		if err := addClientUser(ctx, client.ClientID, "b847a455-aa41-4f7d-8038-0aefbe40dcaa", models.ClientUserStatusGuest, models.ClientUserPriorityHigh); err != nil {
			session.Logger(ctx).Println(err)
		}
	}
}

func addClientUser(ctx context.Context, clientID, userID string, status, priority int) error {
	query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,is_async,status")
	_, err := session.Database(ctx).Exec(ctx, query, clientID, userID, "", priority, true, status)
	return err
}

func addFavoriteApp(ctx context.Context, clientID string) error {
	_, err := models.GetMixinClientByID(ctx, clientID).FavoriteApp(ctx, config.LuckCoinAppID)
	return err
}
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(charset))]
	}
	return string(b)
}
