package clients

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
)

func GetActivityByClientID(ctx context.Context, clientID string) ([]*models.Activity, error) {
	as := make([]*models.Activity, 0)
	asString, err := session.Redis(ctx).QGet(ctx, "activity:"+clientID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := session.DB(ctx).
				Select("activity_index,img_url,expire_img_url,action,start_at,expire_at,created_at").
				Where("client_id=? AND status=2", clientID).
				Order("activity_index").
				Find(&as).Error; err != nil {
				return nil, err
			}
			if err := session.Redis(ctx).StructSet(ctx, "activity:"+clientID, as); err != nil {
				tools.Println(err)
			}
		} else {
			tools.Println(err)
		}
	} else {
		err = json.Unmarshal([]byte(asString), &as)
	}

	return as, err
}
