package services

import (
	"context"
	"encoding/json"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"log"
	"math/rand"
	"time"
)

type TestService struct{}

func (service *TestService) Run(ctx context.Context) error {
	uploadQiniu(ctx)
	//	cis := make([]models.ClientInfo, 0)
	//	log.Println(config.Config.ShowClientList)
	//	if err := session.Database(ctx).ConnQuery(ctx, `
	//SELECT client_id FROM client WHERE client_id=ANY($1)
	//`, func(rows pgx.Rows) error {
	//		for rows.Next() {
	//			var clientID string
	//			if err := rows.Scan(&clientID); err != nil {
	//				return err
	//			}
	//			if ci, err := models.GetClientInfoByHostOrID(ctx, "", clientID); err != nil {
	//				return err
	//			} else {
	//				cis = append(cis, *ci)
	//			}
	//		}
	//		return nil
	//	}, config.Config.ShowClientList); err != nil {
	//		return err
	//	}
	//	log.Println(len(cis))
	return nil
}

func uploadQiniu(ctx context.Context) {
	data := `eyJzaXplIjo4OTE2LCJhdHRhY2htZW50X2lkIjoiNzYyN2Y2ZTUtYmE2Yi00Yjk1LTgwZjQtOWIwZjAxNzQyZjk4Iiwid2F2ZWZvcm0iOiJBQUFMQndzTUJBVUpCUU1EQXdJREFnTURBZ0lDSEVETE8zNWJlRWhLRXBSaG16R0hOanRtVTA5NVZUUTRCcDBoV2xkWENBUURBd01EQkFRREJRSUUiLCJjcmVhdGVkX2F0IjoiMjAyMS0wNi0yOFQxMToyMzowMS4zNzk5NDM0OTVaIiwibWltZV90eXBlIjoiYXVkaW9cL29nZyIsImR1cmF0aW9uIjozOTM2fQ==`
	var audio mixin.AudioMessage
	if err := json.Unmarshal(tools.Base64Decode(data), &audio); err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	a, err := models.GetMixinClientByID(ctx, models.GetFirstClient(ctx).ClientID).ShowAttachment(ctx, audio.AttachmentID)
	if err != nil {
		log.Println(err)
	}
	fileBlob := session.Api(ctx).RawGet(a.ViewURL)
	log.Println(len(fileBlob))
}

func updateClientUser(ctx context.Context) {
	users := make([]string, 0)
	if err := session.Database(ctx).ConnQuery(ctx, `
SELECT user_id FROM client_users WHERE priority=1 LIMIT 5000
`, func(rows pgx.Rows) error {
		var u string
		for rows.Next() {
			if err := rows.Scan(&u); err != nil {
				return err
			}
			users = append(users, u)
		}
		return nil
	}); err != nil {
		session.Logger(ctx).Println(err)
		return
	}

	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=2 WHERE user_id=ANY($1)`, users)
	if err != nil {
		session.Logger(ctx).Println(err)
	}
}

func execAllClient(ctx context.Context) {
	list, err := models.GetClientList(ctx)
	if err != nil {
		return
	}

	for _, client := range list {
		for i := 0; i < 100000; i++ {
			log.Println(i)
			r := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(10)
			status := 3
			priority := 1
			if r <= 5 {
				status = 1
				priority = 2
			}
			if err := addClientUser(ctx, client.ClientID, tools.GetUUID(), status, priority); err != nil {
				session.Logger(ctx).Println(err)
			}
		}
	}
}

func addClientUser(ctx context.Context, clientID, userID string, status, priority int) error {
	query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,is_async,status")
	_, err := session.Database(ctx).Exec(ctx, query, clientID, userID, "", priority, true, status)
	return err
}

func addFavoriteApp(ctx context.Context, clientID string) error {
	_, err := models.GetMixinClientByID(ctx, clientID).FavoriteApp(ctx, config.Config.LuckCoinAppID)
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

func redisTest(ctx context.Context) error {
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
}
