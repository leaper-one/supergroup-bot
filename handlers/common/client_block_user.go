package common

import (
	"context"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var CacheBlockClientUserIDMap = tools.NewMutex()

// 检查是否是block的用户
func CheckIsBlockUser(ctx context.Context, clientID, userID string) bool {
	if r := CacheBlockClientUserIDMap.Read(userID); r == nil {
		if r := CacheBlockClientUserIDMap.Read(clientID + userID); r == nil {
			return false
		}
	}
	return true
}

// 禁言 一个用户 mutedTime=0 则为取消禁言
func MuteClientUser(ctx context.Context, clientID, userID, mutedTime string) error {
	var mutedAt time.Time
	CheckAndReplaceProxyUser(ctx, clientID, &userID)
	mute, _ := strconv.Atoi(mutedTime)
	mutedAt = time.Now().Add(time.Duration(int64(mute)) * time.Hour)
	if err := session.DB(ctx).Table("client_users").
		Where("client_id=? AND user_id=?", clientID, userID).
		Updates(models.ClientUser{MutedTime: mutedTime, MutedAt: mutedAt}).Error; err != nil {
		return err
	}
	cacheClientUser(ctx, clientID, userID)
	return nil
}

// 拉黑一个用户
func BlockClientUser(ctx context.Context, clientID, operatorID, userID string, isCancel bool) error {
	CheckAndReplaceProxyUser(ctx, clientID, &userID)
	var err error
	if isCancel {
		UpdateClientUserPart(ctx, clientID, userID, map[string]interface{}{
			"priority": models.ClientUserPriorityLow,
			"status":   models.ClientUserStatusAudience,
		})
		CacheBlockClientUserIDMap.Write(clientID+userID, nil)
		err = session.DB(ctx).Delete(&models.ClientBlockUser{ClientID: clientID, UserID: userID}).Error
	} else {
		UpdateClientUserPart(ctx, clientID, userID, map[string]interface{}{
			"status": models.ClientUserStatusBlock,
		})
		CacheBlockClientUserIDMap.Write(clientID+userID, true)
		go recallLatestMsg(clientID, userID)
		err = session.DB(ctx).Save(&models.ClientBlockUser{
			OperatorID: operatorID,
			ClientID:   clientID,
			UserID:     userID,
		}).Error
	}
	return err
}

// 撤回用户最近 1 小时的消息
func recallLatestMsg(clientID, uid string) {
	// 1. 找到该用户最近发的消息列表的ID
	msgIDList := make([]string, 0)
	if err := session.DB(models.Ctx).Table("messages").
		Where("client_id=? AND user_id=? AND status=? AND now()-created_at<interval '1 hours'", clientID, uid, models.MessageRedisStatusFinished).
		Pluck("message_id", &msgIDList).Error; err != nil {
		tools.Println(err)
		return
	}
	for _, msgID := range msgIDList {
		if err := CreatedManagerRecallMsg(models.Ctx, clientID, msgID, uid); err != nil {
			tools.Println(err)
			return
		}
	}
}

func SuperAddBlockUser(ctx context.Context, u *models.ClientUser, userID string) error {
	if u.UserID != "b26b9a74-40dd-4e8d-8e41-94d9fce0b5c0" {
		return session.ForbiddenError(ctx)
	}
	return AddBlockUser(ctx, "", u.ClientID, userID, "")
}

func AddBlockUser(ctx context.Context, operatorID, clientID, userID, memo string) error {
	_, err := uuid.FromString(userID)
	if err != nil {
		u, err := SearchUser(ctx, clientID, userID)
		if err != nil {
			return err
		}
		userID = u.UserID
	}
	CacheBlockClientUserIDMap.Write(userID, true)
	return models.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Save(&models.BlockUser{UserID: userID, OperatorID: operatorID, Memo: memo}).Error; err != nil {
			return err
		}
		return tx.Table("client_users").
			Where("user_id=?", userID).
			Updates(models.ClientUser{
				Status: models.ClientUserStatusBlock,
			}).Error
	})
}
