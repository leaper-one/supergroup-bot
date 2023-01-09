package clients

import (
	"context"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

type ClientAdvanceSetting struct {
	ConversationStatus string `json:"conversation_status"`
	NewMemberNotice    string `json:"new_member_notice"`
	ProxyStatus        string `json:"proxy_status"`
}

func GetClientAdvanceSetting(ctx context.Context, u *models.ClientUser) (*ClientAdvanceSetting, error) {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	var sr ClientAdvanceSetting
	sr.ConversationStatus = common.GetClientConversationStatus(ctx, u.ClientID)
	sr.NewMemberNotice = common.GetClientNewMemberNotice(ctx, u.ClientID)
	sr.ProxyStatus = common.GetClientProxy(ctx, u.ClientID)
	return &sr, nil
}

func UpdateClientAdvanceSetting(ctx context.Context, u *models.ClientUser, sr ClientAdvanceSetting) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if sr.ConversationStatus == "0" || sr.ConversationStatus == "1" {
		common.MuteClientOperation(sr.ConversationStatus != models.ClientConversationStatusNormal, u.ClientID)
		return nil
	}
	if sr.NewMemberNotice == "0" || sr.NewMemberNotice == "1" {
		return common.SetClientNewMemberNoticeByIDAndStatus(ctx, u.ClientID, sr.NewMemberNotice)
	}
	if sr.ProxyStatus == "0" || sr.ProxyStatus == "1" {
		return common.SetClientProxyStatusByIDAndStatus(ctx, u.ClientID, sr.ProxyStatus)
	}
	return session.BadRequestError(ctx)
}
