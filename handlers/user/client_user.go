package user

import (
	"context"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

func GetAllClientNeedAssetsCheckUser(ctx context.Context, hasPayedUser bool) ([]*models.ClientUser, error) {
	allUser := make([]*models.ClientUser, 0)
	addQuery := ""
	if !hasPayedUser {
		addQuery = `AND cu.pay_expired_at<NOW()`
	}
	err := session.DB(ctx).Table("client_users AS cu").
		Select("cu.client_id, cu.user_id, cu.access_token, cu.priority, cu.status, coalesce(c.asset_id, '') as asset_id, c.speak_status, cu.deliver_at").
		Joins("LEFT JOIN client AS c ON c.client_id=cu.client_id").
		Where("cu.priority IN (1,2) AND cu.status IN (1,2,3,4,5) " + addQuery).
		Scan(&allUser).Error
	return allUser, err
}
func UpdateClientUserChatStatus(ctx context.Context, u *models.ClientUser, isReceived, isNoticeJoin bool) (models.ClientUser, error) {
	msg := ""
	if isReceived {
		msg = config.Text.OpenChatStatus
	} else {
		msg = config.Text.CloseChatStatus
		isNoticeJoin = false
	}

	if err := common.UpdateClientUserPart(ctx, u.ClientID, u.UserID,
		map[string]interface{}{"is_received": isReceived, "is_notice_join": isNoticeJoin}); err != nil {
		return models.ClientUser{}, err
	}

	if u.IsReceived != isReceived {
		go common.SendClientUserTextMsg(u.ClientID, u.UserID, msg, "")
	}
	return common.GetClientUserByClientIDAndUserID(ctx, u.ClientID, u.UserID)
}
