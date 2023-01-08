package jobs

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

type ExinAd struct {
	ID                 int    `json:"id"`
	AvatarUrl          string `json:"avatarUrl"`
	Nickname           string `json:"nickname"`
	IsCertification    bool   `json:"isCertification"`
	IsLandun           bool   `json:"isLandun"`
	Price              string `json:"price"`
	MinPrice           string `json:"minPrice"`
	MaxPrice           string `json:"maxPrice"`
	MultisigOrderCount int    `json:"multisigOrderCount"`
	In5minRate         string `json:"in5minRate"`
	OrderSuccessRank   string `json:"orderSuccessRank"`
	AssetID            string `json:"assetId"`
	PayMethods         []struct {
		ID     int    `json:"id"`
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"payMethods"`
}

var cacheExin = make([]*ExinAd, 0)

func UpdateExinLocalAD() {
	if config.Config.ExinLocalKey == "" {
		return
	}
	for {
		if err := GetExinLocalAd(models.Ctx, &cacheExin); err != nil {
			session.Logger(models.Ctx).Println(err)
		}
		time.Sleep(time.Minute)
	}
}

func GetExinLocalAd(ctx context.Context, ad *[]*ExinAd) error {
	err := session.Api(ctx).Get(`https://www.tigaex.com/api/v1/mixin/usdt/advertisement?apiKey=`+config.Config.ExinLocalKey, ad)
	if err != nil {
		return err
	}
	return nil
}
