package jobs

import (
	"net/url"
	"time"

	clients "github.com/MixinNetwork/supergroup/handlers/client"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

func DailyUpdateClientWhiteURL() {
	ctx := models.Ctx
	for {
		cs, err := clients.GetAllClient(ctx)
		if err != nil {
			tools.Println(err)
			return
		}

		for _, c := range cs {
			client, err := common.GetMixinClientByIDOrHost(ctx, c)
			if err != nil {
				tools.Println(err)
				continue
			}
			me, err := client.UserMe(ctx)
			if err != nil {
				tools.Println(err)
				continue
			}
			for _, u := range me.App.ResourcePatterns {
				whiteURL, err := url.Parse(u)
				if err != nil {
					tools.Println(err)
					continue
				}
				session.DB(ctx).Save(&models.ClientWhiteURL{
					ClientID: c,
					WhiteURL: whiteURL.Host,
				})
				if err != nil {
					tools.Println(err)
				}
			}
		}
		time.Sleep(time.Hour * 24)
	}
}
