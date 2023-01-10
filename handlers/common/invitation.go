package common

import (
	"context"
	"errors"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const MAX_POWER = 20000

func CreateInvitation(ctx context.Context, userID, clientID, inviterID string) (string, error) {
	inviteCode := tools.GetRandomInvitedCode()
	err := session.DB(ctx).Create(&models.Invitation{
		InviteeID:  userID,
		InviterID:  inviterID,
		ClientID:   clientID,
		InviteCode: inviteCode,
	}).Error
	return inviteCode, err
}

func HandleUserInvite(inviteCode, clientID, userID string) {
	ctx := models.Ctx
	if inviteCode == "" {
		return
	}
	inviterID := GetUserByInviteCode(ctx, inviteCode)
	if inviterID == "" {
		return
	}
	if checkUserIsInSystem(ctx, userID) {
		return
	}
	// 创建邀请关系
	if _, err := CreateInvitation(ctx, userID, clientID, inviterID); err != nil {
		return
	}
	// 创建一条邀请记录
	if err := session.DB(ctx).Create(&models.InvitationPowerRecord{
		InviteeID: userID,
		InviterID: inviterID,
		Amount:    decimal.Zero,
	}).Error; err != nil {
		tools.Println(err)
	}
}

func checkUserIsInSystem(ctx context.Context, userID string) bool {
	var count int64
	if err := session.DB(ctx).Table("invitation").Where("invitee_id = ?", userID).Count(&count).Error; err != nil {
		tools.Println(err)
		return true
	}
	return count > 0
}

func GetInvitationByInviteeID(ctx context.Context, inviteeID string) *models.Invitation {
	var i models.Invitation
	if err := session.DB(ctx).Take(&i, "invitee_id = ?", inviteeID).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			tools.Println(err)
		}
		return nil
	}
	return &i
}

func GetUserByInviteCode(ctx context.Context, inviteCode string) string {
	var userID string
	if err := session.DB(ctx).Table("invitation").
		Select("invitee_id").
		Where("invite_code = ?", inviteCode).
		Scan(&userID).Error; err != nil {
		return ""
	}

	return userID
}

func GetInviteCountByUserID(ctx context.Context, userID string) int64 {
	var count int64
	if err := session.DB(ctx).Table("invitation").
		Where("inviter_id = ?", userID).
		Count(&count).Error; err != nil {
		tools.Println(err)
	}
	return count
}
