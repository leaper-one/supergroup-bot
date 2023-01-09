package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/jackc/pgx/v4"
)

func GetAllClientNeedAssetsCheckUser(ctx context.Context, hasPayedUser bool) ([]*models.ClientUser, error) {
	allUser := make([]*models.ClientUser, 0)
	addQuery := ""
	if !hasPayedUser {
		addQuery = `AND cu.pay_expired_at<NOW()`
	}
	err := session.DB(ctx).Table("client_users as cu").
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

func UpdateClientUser(ctx context.Context, _u models.ClientUser, fullName string) (bool, error) {
	u, err := common.GetClientUserByClientIDAndUserID(ctx, _u.ClientID, _u.UserID)
	isNewUser := false
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 第一次入群
			isNewUser = true
		}
	}
	if _u.AccessToken != "" || _u.AuthorizationID != "" {
		go common.SendClientUserTextMsg(_u.ClientID, _u.UserID, config.Text.AuthSuccess, "")
		var msg string
		if _u.Status == models.ClientUserStatusLarge {
			if u.Status < models.ClientUserStatusLarge {
				msg = config.Text.AuthForLarge
			}
		} else if _u.Status != models.ClientUserStatusAudience {
			if u.Status < _u.Status {
				msg = config.Text.AuthForFresh
			}
		}
		go common.SendClientUserTextMsg(_u.ClientID, _u.UserID, msg, "")
	}
	if u.Status == models.ClientUserStatusAdmin || u.Status == models.ClientUserStatusGuest {
		_u.Status = u.Status
		_u.Priority = models.ClientUserPriorityHigh
	} else if u.PayStatus > models.ClientUserStatusAudience {
		_u.Status = u.PayStatus
		_u.Priority = models.ClientUserPriorityHigh
	}
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", _u.ClientID, _u.UserID))
	err = session.DB(ctx).Save(&_u).Error
	if isNewUser {
		cs := common.GetClientConversationStatus(ctx, _u.ClientID)
		// conversation 状态为普通的时候入群通知是打开的，就通知用户入群。
		if cs == models.ClientConversationStatusNormal &&
			common.GetClientNewMemberNotice(ctx, _u.ClientID) == models.ClientNewMemberNoticeOn {
			go common.SendClientTextMsg(_u.ClientID, strings.ReplaceAll(config.Text.JoinMsg, "{name}", tools.SplitString(fullName, 12)), _u.UserID, true)
		}
		go SendWelcomeAndLatestMsg(_u.ClientID, _u.UserID)
	}
	return isNewUser, err
}
