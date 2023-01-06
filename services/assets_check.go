package services

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/tools"
)

type AssetsCheckService struct{}

func (service *AssetsCheckService) Run(ctx context.Context) error {
	for {
		now := time.Now()
		if err := startAssetCheck(ctx); err != nil {
			tools.Println(err)
		}
		tools.PrintTimeDuration("资产检查...", now)
		time.Sleep(config.AssetsCheckTime)
	}
}

func startAssetCheck(ctx context.Context) error {
	// 获取所有的用户
	allClientUser, err := common.GetAllClientNeedAssetsCheckUser(ctx, true)
	if err != nil {
		return err
	}
	// 检查所有的用户是否活跃
	checkUserIsActive(ctx, allClientUser)
	allClientUser, err = common.GetAllClientNeedAssetsCheckUser(ctx, false)
	if err != nil {
		return err
	}
	if len(allClientUser) == 0 {
		return nil
	}
	var allUser []string
	_allUser := make(map[string]bool)
	for _, user := range allClientUser {
		_allUser[user.UserID] = true
	}
	for k := range _allUser {
		allUser = append(allUser, k)
	}
	foxUserAssetMap, _ := common.GetAllUserFoxShares(ctx, allUser)
	exinUserAssetMap, _ := common.GetAllUserExinShares(ctx, allUser)

	for _, user := range allClientUser {
		curStatus, err := common.GetClientUserStatus(ctx, user, foxUserAssetMap[user.UserID], exinUserAssetMap[user.UserID])
		if err != nil {
			tools.Println(err)
			if err := common.UpdateClientUserPriorityAndStatus(ctx, user.ClientID, user.UserID, common.ClientUserPriorityLow, common.ClientUserStatusAudience); err != nil {
				tools.Println(err)
			}
			return nil
		}
		priority := common.ClientUserPriorityLow
		if curStatus != common.ClientUserStatusAudience {
			priority = common.ClientUserPriorityHigh
		}
		if err := common.UpdateClientUserPriorityAndStatus(ctx, user.ClientID, user.UserID, priority, curStatus); err != nil {
			tools.Println(err)
		}
	}
	return nil
}

func checkUserIsActive(ctx context.Context, allClientUser []*common.ClientUser) {
	lms, err := common.GetClientLastMsg(ctx)
	if err != nil {
		tools.Println(err)
		return
	}
	for _, user := range allClientUser {
		if err := common.CheckUserIsActive(ctx, user, lms[user.ClientID]); err != nil {
			tools.Println(err)
		}
	}
}
