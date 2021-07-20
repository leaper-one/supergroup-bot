package models

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup/session"
)

func Test(ctx context.Context) {
	clients, err := getAllClient(ctx)
	if err != nil {
		session.Logger(ctx).Println(err)
	}

	for _, clientID := range clients {
		client := GetMixinClientByID(ctx, clientID)
		me, err := client.UserMe(ctx)
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
		log.Println(clientID, me.App.CreatorID)
		if _, err := session.Database(ctx).Exec(ctx, `
UPDATE client SET owner_id = $2 WHERE client_id = $1`, clientID, me.App.CreatorID); err != nil {
			session.Logger(ctx).Println(err)
		}
	}

}
