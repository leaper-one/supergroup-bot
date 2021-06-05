package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"time"
)

type CreateDistributeMsgService struct{}

func (service *CreateDistributeMsgService) Run(ctx context.Context) error {
	list, err := models.GetClientList(ctx)
	if err != nil {
		return err
	}

	for _, client := range list {
		go createMsg(ctx, client.ClientID)
	}

	select {}
}

func createMsg(ctx context.Context, clientID string) {
	for {
		now := time.Now().UnixNano()
		msg, err := models.GetLongestMessageByStatus(ctx, clientID, models.MessageStatusPending)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				session.Logger(ctx).Println(err)
			}
			time.Sleep(time.Second)
			continue
		}
		clientUser, err := models.GetClientUserByClientIDAndUserID(ctx, clientID, msg.UserID)
		if err != nil {
			session.Logger(ctx).Println(err)
			continue
		}
		if err := models.CreateDistributeMsgAndMarkPrivilege(ctx, clientID, clientUser.Status, &mixin.MessageView{
			UserID:         msg.UserID,
			MessageID:      msg.MessageID,
			Category:       msg.Category,
			Data:           msg.Data,
			QuoteMessageID: msg.QuoteMessageID,
		}); err != nil {
			session.Logger(ctx).Println(err)
		}
		tools.PrintTimeDuration(fmt.Sprintf("创建消息 %s", msg.MessageID), now)
		time.Sleep(100 * time.Millisecond)
	}

}
