package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"gorm.io/gorm"
)

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

func UpdateClientUser(ctx context.Context, newUser models.ClientUser, fullName string) (bool, error) {
	go common.SendClientUserTextMsg(newUser.ClientID, newUser.UserID, config.Text.AuthSuccess, "")
	oldUser, err := common.GetClientUserByClientIDAndUserID(ctx, newUser.ClientID, newUser.UserID)
	go sendMemberStatusMsg(oldUser.Status, newUser.Status, newUser.ClientID, newUser.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 第一次入群
			newUser.CreatedAt = time.Now()
			err = session.DB(ctx).Save(&newUser).Error
			cs := common.GetClientConversationStatus(ctx, newUser.ClientID)
			// conversation 状态为普通的时候入群通知是打开的，就通知用户入群。
			if cs == models.ClientConversationStatusNormal &&
				common.GetClientNewMemberNotice(ctx, newUser.ClientID) == models.ClientNewMemberNoticeOn {
				go common.SendClientTextMsg(newUser.ClientID, strings.ReplaceAll(config.Text.JoinMsg, "{name}", tools.SplitString(fullName, 12)), newUser.UserID, true)
			}
			go SendWelcomeAndLatestMsg(newUser.ClientID, newUser.UserID)
			return true, err
		}
		return false, err
	}
	var status, priority int
	if oldUser.Status == models.ClientUserStatusAdmin || oldUser.Status == models.ClientUserStatusGuest {
		status = oldUser.Status
		priority = models.ClientUserPriorityHigh
	} else if oldUser.PayStatus > models.ClientUserStatusAudience {
		status = oldUser.PayStatus
		priority = models.ClientUserPriorityHigh
	}
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", newUser.ClientID, newUser.UserID))
	err = common.UpdateClientUserPart(ctx, newUser.ClientID, newUser.UserID, map[string]interface{}{
		"status":           status,
		"priority":         priority,
		"access_token":     newUser.AccessToken,
		"authorization_id": newUser.AuthorizationID,
		"scope":            newUser.Scope,
		"private_key":      newUser.PrivateKey,
		"ed25519":          newUser.Ed25519,
		"is_received":      true,
	})
	return false, err
}

func sendMemberStatusMsg(oldStatus, newStatus int, clientID, userID string) {
	if newStatus == models.ClientUserStatusLarge {
		if oldStatus < models.ClientUserStatusLarge {
			common.SendClientUserTextMsg(clientID, userID, config.Text.AuthForLarge, "")
		}
	} else if newStatus != models.ClientUserStatusAudience {
		if oldStatus < newStatus {
			common.SendClientUserTextMsg(clientID, userID, config.Text.AuthForFresh, "")
		}
	}
}
