package services

import (
	"context"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"log"
	"math/rand"
	"time"
)

type TestService struct{}

func (service *TestService) Run(ctx context.Context) error {
	//session.Database(ctx).CopyFrom(ctx, "", "", "")
	var dataToInsert [][]interface{}
	const InsertCount = 1000
	for i := 0; i < InsertCount; i++ {
		var row []interface{}

		row = append(row, RandomString(36), RandomString(36), 1)
		//row = append(row, RandomString(36))
		//row = append(row, 1)

		dataToInsert = append(dataToInsert, row)
	}

	now := time.Now().UnixNano()
	//log.Println("Begin Insert Rows:", len(dataToInsert))
	//var ident = pgx.Identifier{"broadcast"}
	//var cols = []string{"client_id", "message_id", "status"}
	//numRowsCopied, err := session.Database(ctx).CopyFrom(ctx, ident, cols, pgx.CopyFromRows(dataToInsert))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println("pgx says we worte:", numRowsCopied)

	query := durable.InsertQueryOrUpdate("broadcast", "client_id,message_id", "status")
	for _, i := range dataToInsert {
		if _, err := session.Database(ctx).Exec(ctx, query, i[0], i[1], i[2]); err != nil {
			log.Println(err)
		}
	}

	tools.PrintTimeDuration("end Insert rows", now)
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
