package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
)

func CacheAllClientUser() {
	for {
		lastTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local)
		count := -1
		total := 0
		for {
			count, lastTime = _cacheAllClientUser(models.Ctx, lastTime)
			total += count
			if count == 0 {
				break
			}
		}
		time.Sleep(5 * time.Minute)
	}
}

func _cacheAllClientUser(ctx context.Context, lastTime time.Time) (int, time.Time) {
	cus := make([]models.ClientUser, 0, 1000)

	if err := session.DB(ctx).Table("client_users cu").
		Select("cu.*, c.asset_id,c.speak_status").
		Joins("LEFT JOIN client c ON cu.client_id=c.client_id").
		Order("cu.created_at ASC").
		Where("cu.created_at>?", lastTime).
		Limit(1000).
		Find(&cus).Error; err != nil {
		tools.Println(err)
	}

	if len(cus) == 0 {
		return 0, lastTime
	}
	if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, u := range cus {
			key := fmt.Sprintf("client_user:%s:%s", u.ClientID, u.UserID)
			uStr, err := json.Marshal(u)
			if err != nil {
				tools.Println(err)
			}
			if err := p.Set(ctx, key, uStr, 15*time.Minute).Err(); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		tools.Println(err)
	}
	return len(cus), cus[len(cus)-1].CreatedAt
}
