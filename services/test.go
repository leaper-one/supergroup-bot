package services

import (
	"context"
	"encoding/json"
	"image"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

type TestService struct{}

func (service *TestService) Run(ctx context.Context) error {
	str_obj := "Hello你好WorldGo语言真强"
	len, totalLen := tools.LanguageCount(str_obj, nil)
	log.Println("result=", len, "total", totalLen)
	return nil
}

func zxingTest() error {
	file, err := os.Open("test1.jpeg")
	if err != nil {
		return err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}

	// prepare BinaryBitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return err
	}

	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	log.Println(err)
	log.Println(result.GetText())

	return err
}

// 把 client_user 里的所有 user 添加到 users 表中
func initUserByClientUser(ctx context.Context) {
	clientUsers := make([]string, 0)
	session.Database(ctx).ConnQuery(ctx, `
SELECT distinct(user_id) FROM client_users
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var u string
			if err := rows.Scan(&u); err != nil {
				return err
			}
			clientUsers = append(clientUsers, u)
		}
		return nil
	})
	log.Println(len(clientUsers))
	users := make(map[string]bool)
	session.Database(ctx).ConnQuery(ctx, `
SELECT user_id FROM users
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var u string
			if err := rows.Scan(&u); err != nil {
				return err
			}
			users[u] = true
		}
		return nil
	})
	log.Println(len(users))

	for _, user := range clientUsers {
		if users[user] {
			continue
		}
		go func(user string) {
			log.Println(user)
			s.Add(1)
			_, err := models.SearchUser(ctx, user)
			if err != nil {
				log.Println(err)
			}
			s.Done()
		}(user)
	}
	s.Wait()
	return
}

// 统计大群人的状态的数量
var s sync.WaitGroup

type Smap struct {
	sync.RWMutex
	Map map[string]int
}

var i int

func (l *Smap) writeMap(key string) {
	l.Lock()
	l.Map[key] = l.Map[key] + 1
	i++
	log.Println(i)
	l.Unlock()
}

func (l *Smap) readAllMap() {
	for s, i := range l.Map {
		log.Printf("%s...%d", s, i)
	}
}

var mMap *Smap

// 检查所有用户的设别状态
func updateClientUserToUser(ctx context.Context) error {
	mMap = &Smap{
		Map: make(map[string]int),
	}
	users := make([]string, 0)
	if err := session.Database(ctx).ConnQuery(ctx, `
SELECT access_token FROM client_users WHERE client_id=$1
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var u string
			if err := rows.Scan(&u); err != nil {
				return err
			}
			users = append(users, u)
		}
		return nil
	}, "419c37ff-cce0-4073-9c26-6736ff3394e3"); err != nil {
		session.Logger(ctx).Println(err)
	}
	log.Println(len(users))
	for _, user := range users {
		s.Add(1)
		go getUsers(ctx, user)
	}
	s.Wait()
	mMap.readAllMap()
	return nil
}

func getUsers(ctx context.Context, token string) {
	if token == "" {
		log.Println(1)
		s.Done()
		return
	}
	u, err := mixin.UserMe(ctx, token)
	if err != nil {

		if strings.Contains(err.Error(), "401") {
			s.Done()
			return
		}
		session.Logger(ctx).Println(err)
		time.Sleep(time.Millisecond * 100)
		getUsers(ctx, token)
		return
	}
	mMap.writeMap(u.DeviceStatus)
	s.Done()
	return
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
