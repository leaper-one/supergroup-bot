package services

import (
	"context"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"time"
)

type AssetsCheckService struct{}

func (service *AssetsCheckService) Run(ctx context.Context) error {
	for {
		//handlePendingUsers(ctx)
		if err := startAssetCheck(ctx); err != nil {
			session.Logger(ctx).Println(err)
		}
		time.Sleep(config.AssetsCheckTime)
	}
}

func startAssetCheck(ctx context.Context) error {
	// 获取所有的用户
	allClientUser, err := models.GetAllClientUser(ctx)
	if err != nil {
		return err
	}
	// 检查所有的用户是否活跃
	go models.CheckUserIsActive(ctx, allClientUser)
	var allUser []string
	_allUser := make(map[string]bool)
	for _, user := range allClientUser {
		_allUser[user.UserID] = true
	}
	for k := range _allUser {
		allUser = append(allUser, k)
	}
	foxUserAssetMap, _ := models.GetAllUserFoxShares(ctx, allUser)
	exinUserAssetMap, _ := models.GetAllUserExinShares(ctx, allUser)

	for _, user := range allClientUser {
		if status, err := models.GetClientUserStatus(ctx, user, foxUserAssetMap[user.UserID], exinUserAssetMap[user.UserID]); err != nil {
			session.Logger(ctx).Println(err)
			if err := models.UpdateClientUserPriorityAndStatus(ctx, user.ClientID, user.UserID, models.ClientUserPriorityLow, models.ClientUserStatusAudience); err != nil {
				session.Logger(ctx).Println(err)
			}
		} else {
			// 如果之前是低状态，现在是高状态，那么先 pending 之前的消息
			if user.SpeakStatus == models.ClientSpeckStatusOpen && user.Priority == models.ClientUserPriorityLow && status != models.ClientUserStatusAudience {
				if err := models.UpdateClientUserAndMessageToPending(ctx, user.ClientID, user.UserID); err != nil {
					session.Logger(ctx).Println(err)
				}
				// TODO
				//go models.SendClientUserPendingMessages(ctx, user.ClientID, user.UserID)
			}
			// 如果之前是高状态，现在是低状态
			if user.Priority == models.ClientUserPriorityHigh && status == models.ClientUserStatusAudience {
				if err := models.UpdateClientUserPriority(ctx, user.ClientID, user.UserID, models.ClientUserPriorityLow); err != nil {
					return err
				}
			}
			// 如果之前是高状态，现在是低状态
			if err := models.UpdateClientUserStatus(ctx, user.ClientID, user.UserID, status); err != nil {
				session.Logger(ctx).Println(err)
			}
		}
	}
	return nil
}

//func handlePendingUsers(ctx context.Context) {
//	clientList, err := models.GetClientList(ctx)
//	if err != nil {
//		session.Logger(ctx).Println(err)
//	}
//	for _, client := range clientList {
//		users, err := models.GetClientUserByPriority(ctx, client.ClientID, []int{models.ClientUserPriorityPending})
//		if err != nil {
//			session.Logger(ctx).Println(err)
//			continue
//		}
//		for _, user := range users {
//			log.Println(user)
//			go models.SendClientUserPendingMessages(ctx, client.ClientID, user)
//		}
//	}
//}
