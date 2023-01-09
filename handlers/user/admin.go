package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"gorm.io/gorm"
)

func GetClientUserList(ctx context.Context, u *models.ClientUser, page int, status string) ([]*clientUserView, error) {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	cs, err := getMuteOrBlockClientUserList(ctx, u, status)
	if err != nil {
		return nil, err
	}
	if cs == nil {
		cs, err = getAllOrGuestOrAdminClientUserList(ctx, u, page, status)
	}
	return cs, err
}

var clientUserStatusMap = map[string][]int{
	"all": {
		models.ClientUserStatusAudience,
		models.ClientUserStatusFresh,
		models.ClientUserStatusSenior,
		models.ClientUserStatusLarge,
	},
	"guest": {models.ClientUserStatusGuest},
	"admin": {models.ClientUserStatusAdmin},
	// "all": fmt.Sprintf("%d,%d,%d,%d",
	// 	models.ClientUserStatusAudience,
	// 	models.ClientUserStatusFresh,
	// 	models.ClientUserStatusSenior,
	// 	models.ClientUserStatusLarge,
	// ),
	// "guest": fmt.Sprintf("%d", models.ClientUserStatusGuest),
	// "admin": fmt.Sprintf("%d", models.ClientUserStatusAdmin),
}
var clientUserViewPrefix = `SELECT u.user_id,avatar_url,full_name,identity_number,status,deliver_at,cu.created_at
FROM client_users cu
LEFT JOIN users u ON cu.user_id=u.user_id 
WHERE client_id=$1 `

func getMuteOrBlockClientUserList(ctx context.Context, u *models.ClientUser, status string) ([]*clientUserView, error) {
	if status == "mute" {
		// 获取禁言用户列表
		return getClientUserList(ctx, "AND muted_at>NOW()", u.ClientID)
	}
	if status == "block" {
		// 获取拉黑用户列表
		return getClientUserList(ctx,
			fmt.Sprintf("AND status=%d", models.ClientUserStatusBlock),
			u.ClientID)
	}
	return nil, nil
}

type clientUserView struct {
	UserID         string    `json:"user_id,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	FullName       string    `json:"full_name,omitempty"`
	IdentityNumber string    `json:"identity_number,omitempty"`
	Status         int       `json:"status,omitempty"`
	ActiveAt       time.Time `json:"active_at,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}

func getClientUserList(ctx context.Context, addQuery, clientID string) ([]*clientUserView, error) {
	cus := make([]*clientUserView, 0)
	err := getBaseClientUserList(ctx).
		Where("cu.client_id=? "+addQuery, clientID).
		Scan(&cus).Error
	return cus, err
}

func getAllOrGuestOrAdminClientUserList(ctx context.Context, u *models.ClientUser, page int, status string) ([]*clientUserView, error) {
	if status == "mute" || status == "block" {
		tools.Println("status::", status)
		return nil, nil
	}
	statusList := clientUserStatusMap[status]
	if statusList == nil {
		tools.Println("status::", status)
		return nil, nil
	}
	var cus []*clientUserView
	if err := getBaseClientUserList(ctx).
		Order("cu.created_at ASC").
		Where("cu.client_id=? AND status IN ?", u.ClientID, statusList).
		Offset((page - 1) * 20).
		Limit(20).
		Scan(&cus).Error; err != nil {
		return nil, err
	}

	if page == 1 && status == "all" {
		var users []*clientUserView
		if err := getBaseClientUserList(ctx).
			Order("status DESC").
			Where("cu.client_id=? AND status IN (8,9)", u.ClientID).
			Scan(&users).Error; err != nil {
			return nil, err
		}
		c, _ := common.GetClientByIDOrHost(ctx, u.ClientID)
		for i, v := range users {
			if v.UserID == c.OwnerID {
				users[0], users[i] = users[i], users[0]
				break
			}
		}
		cus = append(users, cus...)

	}
	return cus, nil
}

func GetAdminAndGuestUserList(ctx context.Context, u *models.ClientUser) ([]*clientUserView, error) {
	var cus []*clientUserView
	//	getBaseClientUserList(ctx, clientUserViewPrefix+`
	//
	// AND status IN (8,9)
	// ORDER BY status DESC
	// `, u.ClientID)
	err := getBaseClientUserList(ctx).Order("status DESC").Where("cu.client_id=? AND status IN (8,9)", u.ClientID).Scan(&cus).Error
	return cus, err
}

func getBaseClientUserList(ctx context.Context) *gorm.DB {
	return session.DB(ctx).Table("client_user as cu").
		Select("cu.user_id, u.avatar_url, u.full_name, u.identity_number, cu.status, cu.deliver_at,cu.created_at").
		Joins("LEFT JOIN users u ON cu.user_id=u.user_id")
}

// 获取 全部用户数量/禁言用户数量/拉黑用户数量/嘉宾数量/管理员数量
func GetClientUserStat(ctx context.Context, u *models.ClientUser) (map[string]int64, error) {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	var allUserCount, muteUserCount, blockUserCount, guestUserCount, adminUserCount int64
	if err := session.DB(ctx).Model(&models.ClientUser{}).Where("client_id=?", u.ClientID).Count(&allUserCount).Error; err != nil {
		return nil, err
	}

	if err := session.DB(ctx).Model(&models.ClientBlockUser{}).Where("client_id=?", u.ClientID).Count(&blockUserCount).Error; err != nil {
		return nil, err
	}

	if err := session.DB(ctx).Model(&models.ClientUser{}).Where("client_id=? AND muted_at>NOW()", u.ClientID).Count(&muteUserCount).Error; err != nil {
		return nil, err
	}

	if err := session.DB(ctx).Model(&models.ClientUser{}).Where("client_id=? AND status=8", u.ClientID).Count(&guestUserCount).Error; err != nil {
		return nil, err
	}

	if err := session.DB(ctx).Model(&models.ClientUser{}).Where("client_id=? AND status=9", u.ClientID).Count(&adminUserCount).Error; err != nil {
		return nil, err
	}

	return map[string]int64{
		"all":   allUserCount,
		"mute":  muteUserCount,
		"block": blockUserCount,
		"guest": guestUserCount,
		"admin": adminUserCount,
	}, nil
}

func GetClientUserByIDOrName(ctx context.Context, u *models.ClientUser, key string) ([]*clientUserView, error) {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	var cus []*clientUserView
	err := getBaseClientUserList(ctx).
		Where("cu.client_id=? AND (u.identity_number ILIKE '%' || ? || '%' OR u.full_name ILIKE '%' || ? || '%')", u.ClientID, key, key).
		Limit(20).
		Scan(&cus).Error
	return cus, err
}

func UpdateClientUserStatus(ctx context.Context, u *models.ClientUser, userID string, status int, isCancel bool) error {
	if status == models.ClientUserStatusAdmin {
		if !common.CheckIsOwner(ctx, u.ClientID, u.UserID) {
			return session.ForbiddenError(ctx)
		}
	} else {
		if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
			return session.ForbiddenError(ctx)
		}
	}
	var s string
	var msg string
	if status == models.ClientUserStatusAdmin {
		s = config.Text.StatusAdmin
	} else if status == models.ClientUserStatusGuest {
		s = config.Text.StatusGuest
	}
	if isCancel {
		msg = config.Text.StatusCancel
		status = models.ClientUserStatusLarge
	} else {
		msg = config.Text.StatusSet
	}

	if err := common.UpdateClientUserPart(ctx, u.ClientID, userID, map[string]interface{}{
		"status": status,
	}); err != nil {
		return err
	}

	user, err := common.SearchUser(ctx, u.ClientID, userID)
	if err != nil {
		tools.Println("设置用户状态的时候没找到用户...", err)
		return err
	}
	msg = strings.ReplaceAll(msg, "{full_name}", user.FullName)
	msg = strings.ReplaceAll(msg, "{identity_number}", user.IdentityNumber)
	msg = strings.ReplaceAll(msg, "{status}", s)
	if !isCancel && status == models.ClientUserStatusGuest {
		go common.SendClientUserTextMsg(u.ClientID, userID, msg, "")
	}
	go common.SendToClientManager(u.ClientID, &mixin.MessageView{
		ConversationID: mixin.UniqueConversationID(u.ClientID, userID),
		UserID:         userID,
		MessageID:      tools.GetUUID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           tools.Base64Encode([]byte(msg)),
		CreatedAt:      time.Now(),
	}, false, false)
	return nil
}

func BlockUserByID(ctx context.Context, u *models.ClientUser, userID string, isCancel bool) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	return common.BlockClientUser(ctx, u.ClientID, u.UserID, userID, isCancel)
}
func MuteUserByID(ctx context.Context, u *models.ClientUser, userID, muteTime string) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	return common.MuteClientUser(ctx, u.ClientID, userID, muteTime)
}
