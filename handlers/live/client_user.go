package live

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
)

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
	allUser, err := common.GetClientUsersByClientIDAndStatus(ctx, clientID, 0)
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
	}
	return nil
}
