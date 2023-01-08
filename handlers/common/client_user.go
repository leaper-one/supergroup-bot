package common

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
)

func UpdateClientUser(ctx context.Context, user models.ClientUser, fullName string) (bool, error) {
	u, err := GetClientUserByClientIDAndUserID(ctx, user.ClientID, user.UserID)
	isNewUser := false
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 第一次入群
			isNewUser = true
		}
	}
	if user.AccessToken != "" || user.AuthorizationID != "" {
		go SendClientUserTextMsg(user.ClientID, user.UserID, config.Text.AuthSuccess, "")
		var msg string
		if user.Status == models.ClientUserStatusLarge {
			if u.Status < models.ClientUserStatusLarge {
				msg = config.Text.AuthForLarge
			}
		} else if user.Status != models.ClientUserStatusAudience {
			if u.Status < user.Status {
				msg = config.Text.AuthForFresh
			}
		}
		go SendClientUserTextMsg(user.ClientID, user.UserID, msg, "")
	}
	if u.Status == models.ClientUserStatusAdmin || u.Status == models.ClientUserStatusGuest {
		user.Status = u.Status
		user.Priority = models.ClientUserPriorityHigh
	} else if u.PayStatus > models.ClientUserStatusAudience {
		user.Status = u.PayStatus
		user.Priority = models.ClientUserPriorityHigh
	}
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", user.ClientID, user.UserID))
	err = session.DB(ctx).Save(&user).Error
	if isNewUser {
		cs := GetClientConversationStatus(ctx, user.ClientID)
		// conversation 状态为普通的时候入群通知是打开的，就通知用户入群。
		if cs == models.ClientConversationStatusNormal &&
			GetClientNewMemberNotice(ctx, user.ClientID) == ClientNewMemberNoticeOn {
			go SendClientTextMsg(user.ClientID, strings.ReplaceAll(config.Text.JoinMsg, "{name}", tools.SplitString(fullName, 12)), user.UserID, true)
		}
		go SendWelcomeAndLatestMsg(user.ClientID, user.UserID)
	}
	return isNewUser, err
}

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

	if err := session.DB(ctx).Table("client_users cu").
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
	addQuery := ""
	if isJoinMsg {
		addQuery = "AND is_notice_join=true"
	}
	if !isBroadcast {
		addQuery = fmt.Sprintf("%s %s", addQuery, "AND is_received=true")
	}

	err := session.DB(ctx).
		Select("user_id, priority").
		Order("created_at").
		Find(&userList,
			"client_id=? AND priority IN (1,2) AND status IN (1,2,3,5,8,9) "+addQuery,
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

func UpdateClientUserActiveTimeFromRedis(ctx context.Context, clientID string) error {
	if err := UpdateClientUserActiveTime(ctx, clientID, "deliver"); err != nil {
		return err
	}
	if err := UpdateClientUserActiveTime(ctx, clientID, "read"); err != nil {
		return err
	}
	return nil
}

func UpdateClientUserActiveTime(ctx context.Context, clientID, status string) error {
	allUser, err := GetClientUserByClientID(ctx, clientID, 0)
	if err != nil {
		return err
	}
	keys := make([]string, len(allUser))
	for _, userID := range allUser {
		keys = append(keys, fmt.Sprintf("ack_msg:%s:%s:%s", status, clientID, userID))
	}
	for {
		if len(keys) == 0 {
			break
		}
		var currentKeys []string
		if len(keys) > 500 {
			currentKeys = keys[:500]
			keys = keys[500:]
		} else {
			currentKeys = keys
			keys = nil
		}
		results := make([]*redis.StringCmd, 0, len(currentKeys))
		if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
			for _, key := range currentKeys {
				results = append(results, p.Get(ctx, key))
			}
			return nil
		}); err != nil {
			if !errors.Is(err, redis.Nil) {
				tools.Println(err)
			}
		}

		for _, v := range results {
			t, err := v.Result()
			if err != nil {
				if !errors.Is(err, redis.Nil) {
					tools.Println(err)
				}
				continue
			}
			key := v.Args()[1].(string)
			clientID := strings.Split(key, ":")[2]
			userID := strings.Split(key, ":")[3]
			if err := session.DB(ctx).Model(&models.ClientUser{}).
				Where("client_id=? AND user_id=?", clientID, userID).
				Update(fmt.Sprintf("%s_at", status), t).Error; err != nil {
				tools.Println(err)
				continue
			}
		}
		return nil

		time.Sleep(time.Second)
	}
	return nil
}

func getClientManager(ctx context.Context, clientID string) ([]string, error) {
	users, err := GetClientUserByClientID(ctx, clientID, models.ClientUserStatusAdmin)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func GetClientUserByClientID(ctx context.Context, clientID string, status int) ([]string, error) {
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

func getClientPeopleCount(ctx context.Context, clientID string) (int64, int64, error) {
	var all, week int64
	allString, err := session.Redis(ctx).QGet(ctx, "people_count_all:"+clientID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := session.DB(ctx).Table("client_users").
				Where("client_id=? AND status IN (1,2,3,5,8,9)", clientID).
				Count(&all).Error; err != nil {
				return 0, 0, err
			}
			if err := session.Redis(ctx).QSet(ctx, "people_count_all:"+clientID, all, time.Minute); err != nil {
				tools.Println(err)
			}
		} else {
			return 0, 0, err
		}
	} else {
		all, _ = strconv.ParseInt(allString, 10, 64)
	}
	weekString, err := session.Redis(ctx).QGet(ctx, "people_count_week:"+clientID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := session.DB(ctx).Table("client_users").
				Where("client_id=? AND status IN (1,2,3,5,8,9) AND NOW() - created_at < interval '7 days' ", clientID).
				Count(&all).Error; err != nil {
				return 0, 0, err
			}
			if err := session.Redis(ctx).QSet(ctx, "people_count_week:"+clientID, week, time.Minute); err != nil {
				tools.Println(err)
			}
		} else {
			return 0, 0, err
		}
	} else {
		week, _ = strconv.ParseInt(weekString, 10, 64)
	}
	return all, week, nil
}

func ActiveUser(u *models.ClientUser) {
	if u.Priority != models.ClientUserPriorityStop {
		return
	}
	var err error
	status := models.ClientUserStatusAudience
	priority := models.ClientUserPriorityLow
	if u.PayExpiredAt.After(time.Now()) {
		status = u.PayStatus
	} else {
		status, err = GetClientUserStatusByClientUser(models.Ctx, u)
		if err != nil {
			tools.Println(err)
		}
	}
	if status != models.ClientUserStatusAudience {
		priority = models.ClientUserPriorityHigh
	}
	if err := UpdateClientUserPart(models.Ctx, u.ClientID, u.UserID, map[string]interface{}{
		"priority":   priority,
		"status":     status,
		"deliver_at": time.Now(),
		"read_at":    time.Now(),
	}); err != nil {
		tools.Println(err)
	}
}

func CheckUserIsVIP(ctx context.Context, userID string) bool {
	var count int64
	if err := session.DB(ctx).Table("client_users").Where("user_id=? AND status>1", userID).Count(&count).Error; err != nil {
		tools.Println(err)
	}
	return count > 0
}
