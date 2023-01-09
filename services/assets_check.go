package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	clients "github.com/MixinNetwork/supergroup/handlers/client"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/handlers/message"
	"github.com/MixinNetwork/supergroup/handlers/user"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"gorm.io/gorm"
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

	allClientUser, err := user.GetAllClientNeedAssetsCheckUser(ctx, true)
	if err != nil {
		return err
	}
	// 检查所有的用户是否活跃
	checkUserIsActive(ctx, allClientUser)
	allClientUser, err = user.GetAllClientNeedAssetsCheckUser(ctx, false)
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
			if err := common.UpdateClientUserPart(ctx, user.ClientID, user.UserID, map[string]interface{}{
				"priority": models.ClientUserPriorityLow,
				"status":   models.ClientUserStatusAudience,
			}); err != nil {
				tools.Println(err)
			}
			return nil
		}
		priority := models.ClientUserPriorityLow
		if curStatus != models.ClientUserStatusAudience {
			priority = models.ClientUserPriorityHigh
		}
		if err := common.UpdateClientUserPart(ctx, user.ClientID, user.UserID, map[string]interface{}{
			"priority": priority,
			"status":   curStatus,
		}); err != nil {
			tools.Println(err)
		}
	}
	return nil
}

func checkUserIsActive(ctx context.Context, allClientUser []*models.ClientUser) {
	lms, err := GetClientLastMsg(ctx)
	if err != nil {
		tools.Println(err)
		return
	}
	for _, user := range allClientUser {
		if err := CheckUserIsActive(ctx, user, lms[user.ClientID]); err != nil {
			tools.Println(err)
		}
	}
}

func GetClientLastMsg(ctx context.Context) (map[string]time.Time, error) {
	clients, err := clients.GetAllClient(ctx)
	if err != nil {
		return nil, err
	}
	lms := make(map[string]time.Time)
	for _, client := range clients {
		var lm models.Message
		if err := session.DB(ctx).
			Select("created_at").
			Order("created_at desc").
			Limit(1).
			Take(&lm, "client_id = ?", client).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, err
			}
			lm.CreatedAt = time.Now()
		}
		lms[client] = lm.CreatedAt
	}
	return lms, nil
}

func CheckUserIsActive(ctx context.Context, user *models.ClientUser, lastMsgCreatedAt time.Time) error {
	if lastMsgCreatedAt.IsZero() {
		return nil
	}
	if lastMsgCreatedAt.Sub(user.DeliverAt).Hours() > config.NotActiveCheckTime {
		// 标记用户为不活跃，停止消息推送
		go SendStopMsg(user.ClientID, user.UserID)
		if err := common.UpdateClientUserPart(ctx, user.ClientID, user.UserID, map[string]interface{}{"priority": models.ClientUserPriorityStop}); err != nil {
			tools.Println(err)
			return err
		}
	} else if user.Priority == models.ClientUserPriorityStop {
		message.ActiveUser(user)
	}
	return nil
}

func SendStopMsg(clientID, userID string) {
	client, err := common.GetMixinClientByIDOrHost(models.Ctx, clientID)
	if err != nil {
		return
	}
	common.SendClientUserTextMsg(clientID, userID, config.Text.StopMessage, "")
	if err := common.SendBtnMsg(models.Ctx, clientID, userID, mixin.AppButtonGroupMessage{
		{Label: config.Text.StopClose, Action: "input:/received_message", Color: "#5979F0"},
		{Label: config.Text.StopBroadcast, Action: fmt.Sprintf("%s/news", client.C.Host), Color: "#5979F0"},
	}); err != nil {
		tools.Println(err)
		return
	}
}
