package clients

import (
	"context"
	"errors"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
)

func GetActivityByClientID(ctx context.Context, clientID string) ([]*models.Activity, error) {
	as := make([]*models.Activity, 0)
	err := session.Redis(ctx).StructScan(ctx, "activity:"+clientID, &as)
	if err == nil || !errors.Is(err, redis.Nil) {
		return as, err
	}
	defer func() {
		if err := session.Redis(ctx).StructSet(ctx, "activity:"+clientID, as); err != nil {
			tools.Println(err)
		}
	}()
	if err := session.DB(ctx).
		Select("activity_index,img_url,expire_img_url,action,start_at,expire_at,created_at").
		Where("client_id=? AND status=2", clientID).
		Order("activity_index").
		Find(&as).Error; err != nil {
		return nil, err
	}
	return as, nil
}
