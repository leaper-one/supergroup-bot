package services

import (
	"context"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

type TestService struct{}

func (service *TestService) Run(ctx context.Context) error {

	list, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}

	for _, client := range list {
		if err := addClientUser(ctx, client.ClientID, "f59b9309-70c2-4b69-8fd8-5773dbd10018", models.ClientUserStatusGuest, models.ClientUserPriorityHigh); err != nil {
			session.Logger(ctx).Println(err)
		}
		if err := addClientUser(ctx, client.ClientID, "b847a455-aa41-4f7d-8038-0aefbe40dcaa", models.ClientUserStatusGuest, models.ClientUserPriorityHigh); err != nil {
			session.Logger(ctx).Println(err)
		}
	}

	return nil
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
