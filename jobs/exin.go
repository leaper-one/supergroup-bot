package jobs

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/asset"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

func UpdateExinLocalAD() {
	if config.Config.ExinLocalKey == "" {
		return
	}
	for {
		if err := GetExinLocalAd(models.Ctx, &asset.CacheExin); err != nil {
			session.Logger(models.Ctx).Println(err)
		}
		time.Sleep(time.Minute)
	}
}

func GetExinLocalAd(ctx context.Context, ad *[]*models.ExinAd) error {
	err := session.Api(ctx).Get(`https://www.tigaex.com/api/v1/mixin/usdt/advertisement?apiKey=`+config.Config.ExinLocalKey, ad)
	if err != nil {
		return err
	}
	return nil
}
