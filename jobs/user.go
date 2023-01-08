package jobs

import (
	"time"

	clients "github.com/MixinNetwork/supergroup/handlers/client"
	"github.com/MixinNetwork/supergroup/handlers/live"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/tools"
)

func TaskUpdateActiveUserToPsql() {
	ctx := models.Ctx
	list, err := clients.GetClientList(ctx)
	if err != nil {
		tools.Println(err)
	}
	for {
		time.Sleep(time.Hour)
		for _, client := range list {
			live.UpdateClientUserActiveTimeFromRedis(ctx, client.ClientID)
		}
	}
}
