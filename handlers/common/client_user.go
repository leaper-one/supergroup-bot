package common

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
)

func GetClientUserByClientIDAndUserID(ctx context.Context, clientID, userID string) (models.ClientUser, error) {
	key := fmt.Sprintf("client_user:%s:%s", clientID, userID)
	var u models.ClientUser
	if err := session.Redis(ctx).StructScan(ctx, key, &u); err != nil {
		if errors.Is(err, redis.Nil) {
			return cacheClientUser(ctx, clientID, userID)
		}
		if !errors.Is(err, context.Canceled) {
			tools.Println(err)
		}
		return u, err
	}
	return u, nil
}

func cacheClientUser(ctx context.Context, clientID, userID string) (models.ClientUser, error) {
	key := fmt.Sprintf("client_user:%s:%s", clientID, userID)
	var b models.ClientUser

	if err := session.DB(ctx).Table("client_users as cu").
		Select("cu.*, c.asset_id,c.speak_status").
		Joins("LEFT JOIN client c ON cu.client_id=c.client_id").
		Where("cu.client_id=? AND cu.user_id=?", clientID, userID).
		Scan(&b).Error; err != nil {
		return models.ClientUser{}, err
	}

	go func(key string, b models.ClientUser) {
		if err := session.Redis(models.Ctx).StructSet(models.Ctx, key, b); err != nil {
			tools.Println(err)
		}
	}(key, b)
	return b, nil
}

func GetDistributeMsgUser(ctx context.Context, clientID string, isJoinMsg, isBroadcast bool) ([]*models.ClientUser, error) {
	userList := make([]*models.ClientUser, 0)
	addQuery := " "
	if isJoinMsg {
		addQuery += "AND is_notice_join=true "
	}
	if !isBroadcast {
		addQuery += "AND is_received=true "
	}
	err := session.DB(ctx).
		Select("user_id, priority").
		Order("created_at").
		Find(&userList,
			"client_id=? AND priority IN (1,2) AND status IN (1,2,3,5,8,9)"+addQuery,
			clientID).Error

	if err != nil {
		return nil, err
	}
	return userList, nil
}

func UpdateClientUserPart(ctx context.Context, clientID, userID string, update map[string]interface{}) error {
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", clientID, userID))
	return session.DB(ctx).Model(&models.ClientUser{}).Where("client_id=? AND user_id=?", clientID, userID).Updates(update).Error
}

func GetClientUsersByClientIDAndStatus(ctx context.Context, clientID string, status int) ([]string, error) {
	users := make([]string, 0)
	addQuery := ""
	if status != 0 {
		addQuery = " AND status=" + strconv.Itoa(status)
	}
	err := session.DB(ctx).
		Model(&models.ClientUser{}).
		Where("client_id=?"+addQuery, clientID).
		Pluck("user_id", &users).Error
	return users, err
}
