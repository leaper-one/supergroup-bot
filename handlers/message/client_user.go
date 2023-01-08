package message

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
)

func getClientManager(ctx context.Context, clientID string) ([]string, error) {
	var users []string
	err := session.DB(ctx).
		Table("client_users").
		Where("client_id = ? AND status = ?", clientID, models.ClientUserStatusAdmin).
		Pluck("user_id", &users).Error
	return users, err
}

func UpdateClientUserActiveTimeToRedis(clientID, msgID string, deliverTime time.Time, status string) error {
	if status != "DELIVERED" && status != "READ" {
		return nil
	}
	ctx := models.Ctx
	dm, err := common.GetDistributeMsgByMsgIDFromRedis(ctx, msgID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		tools.Println(err)
		return err
	}
	user, err := common.GetClientUserByClientIDAndUserID(ctx, clientID, dm.UserID)
	if err != nil {
		return err
	}
	go common.ActiveUser(&user)
	if status == "READ" {
		if err := session.Redis(ctx).QSet(ctx, fmt.Sprintf("ack_msg:read:%s:%s", clientID, user.UserID), deliverTime, time.Hour*2); err != nil {
			return err
		}
	} else {
		if err := session.Redis(ctx).QSet(ctx, fmt.Sprintf("ack_msg:deliver:%s:%s", clientID, user.UserID), deliverTime, time.Hour*2); err != nil {
			return err
		}
	}
	return nil
}
