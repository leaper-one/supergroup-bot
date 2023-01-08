package clients

import (
	"context"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
)

type VipResp struct {
	Level models.ClientAssetLevel         `json:"level,omitempty"`
	Auth  map[int]models.ClientMemberAuth `json:"auth,omitempty"`
}

func GetClientVipAmount(ctx context.Context, host string) (*VipResp, error) {
	c, err := common.GetMixinClientByIDOrHost(ctx, host)
	if err != nil {
		return nil, err
	}
	var vr VipResp
	vr.Auth, err = GetClientMemberAuth(ctx, c.ClientID)
	if err != nil {
		return nil, err
	}
	vr.Level, err = common.GetClientAssetLevel(ctx, c.ClientID)
	if err != nil {
		return nil, err
	}
	return &vr, nil
}
