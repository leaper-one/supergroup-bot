package clients

import (
	"context"
	"errors"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
)

type ClientInfo struct {
	models.Client
	PriceUsd      decimal.Decimal      `json:"price_usd,omitempty" redis:"price_usd"`
	ChangeUsd     string               `json:"change_usd,omitempty" redis:"change_usd"`
	TotalPeople   int64                `json:"total_people" redis:"total_people"`
	WeekPeople    int64                `json:"week_people" redis:"week_people"`
	Activity      []*models.Activity   `json:"activity,omitempty"`
	HasReward     bool                 `json:"has_reward" redis:"has_reward"`
	NeedPayAmount decimal.Decimal      `json:"need_pay_amount,omitempty" redis:"need_pay_amount"`
	Amount        string               `json:"amount,omitempty" redis:"amount"`
	LargeAmount   string               `json:"large_amount,omitempty" redis:"large_amount"`
	Menus         []*models.ClientMenu `json:"menus,omitempty"`
}

const (
	ExinOneClientID = "47cdbc9e-e2b9-4d1f-b13e-42fec1d8853d"
	XinAssetID      = "c94ac88f-4671-3976-b60a-09064f1811e8"
	BtcAssetID      = "c6d0c728-2624-429b-8e0d-d9d19b6592fa"
)

func GetClientInfoByHostOrID(ctx context.Context, hostOrID string) (*ClientInfo, error) {
	mixinClient, err := common.GetMixinClientByIDOrHost(ctx, hostOrID)
	if err != nil {
		return nil, err
	}
	client := mixinClient.C
	var c ClientInfo
	if client.Pin != "" {
		c.HasReward = true
	}
	client.SessionID = ""
	client.PinToken = ""
	client.PrivateKey = ""
	client.Pin = ""
	c.Client = client
	assetID := client.AssetID
	if c.ClientID == ExinOneClientID {
		assetID = XinAssetID
	} else if assetID == "" {
		assetID = BtcAssetID
	}
	asset, err := common.GetAssetByID(ctx, mixinClient.Client, assetID)
	if err == nil {
		c.PriceUsd = asset.PriceUsd
		c.ChangeUsd = asset.ChangeUsd
		c.Symbol = asset.Symbol
		if client.AssetID != "" && c.IconURL == "" {
			c.IconURL = asset.IconUrl
		}
	}
	amount, err := common.GetClientAssetLevel(ctx, client.ClientID)
	if err == nil {
		c.Amount = amount.Fresh.String()
		c.LargeAmount = amount.Large.String()
	}

	c.TotalPeople, c.WeekPeople, err = getClientPeopleCount(ctx, client.ClientID)
	if err != nil {
		return nil, err
	}
	c.Activity, err = GetActivityByClientID(ctx, client.ClientID)
	if err != nil {
		return nil, err
	}
	if err := session.DB(ctx).Order("idx DESC").Find(&c.Menus, "client_id = ?", client.ClientID).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func GetAllConfigClientInfo(ctx context.Context) ([]ClientInfo, error) {
	cis := make([]ClientInfo, 0)
	for _, clientID := range config.Config.ShowClientList {
		if ci, err := GetClientInfoByHostOrID(ctx, clientID); err == nil {
			cis = append(cis, *ci)
		} else {
			tools.Println(err, clientID)
		}
	}
	return cis, nil
}

func getClientPeopleCount(ctx context.Context, clientID string) (int64, int64, error) {
	all, err := session.Redis(ctx).QGet(ctx, "people_count_all:"+clientID).Int64()
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
	}
	week, err := session.Redis(ctx).QGet(ctx, "people_count_week:"+clientID).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := session.DB(ctx).Table("client_users").
				Where("client_id=? AND status IN (1,2,3,5,8,9) AND NOW() - created_at < interval '7 days' ", clientID).
				Count(&week).Error; err != nil {
				return 0, 0, err
			}
			if err := session.Redis(ctx).QSet(ctx, "people_count_week:"+clientID, week, time.Minute); err != nil {
				tools.Println(err)
			}
		} else {
			return 0, 0, err
		}
	}
	return all, week, nil
}

func GetAllClient(ctx context.Context) ([]string, error) {
	cs := make([]string, 0)
	err := session.DB(ctx).Table("client").Pluck("client_id", &cs).Error
	return cs, err
}
