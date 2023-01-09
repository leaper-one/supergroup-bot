package jobs

import (
	"context"
	"time"

	clients "github.com/MixinNetwork/supergroup/handlers/client"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/tools"
)

func DailyUpdateProxyUserProfile() {
	for {
		updateAllProxyUserProfile(models.Ctx)
		time.Sleep(time.Hour * 24)
	}
}

func updateAllProxyUserProfile(ctx context.Context) error {
	clients, err := clients.GetAllClient(ctx)
	if err != nil {
		return err
	}

	for _, clientID := range clients {
		status := common.GetClientProxy(ctx, clientID)
		if status == models.ClientProxyStatusOff {
			continue
		}
		// 1. 拿到所有的 用户
		_users, err := common.GetClientUsersByClientIDAndStatus(ctx, clientID, 0)
		if err != nil {
			tools.Println(err)
			continue
		}
		for _, userID := range _users {
			u, err := common.SearchUser(ctx, clientID, userID)
			if err != nil {
				tools.Println(err)
				continue
			}
			if err := common.UpdateClientUserProxy(ctx, &models.ClientUser{ClientID: clientID, UserID: userID}, true, u.FullName, u.AvatarURL); err != nil {
				tools.Println(err)
				continue
			}
		}
	}

	return nil
}
