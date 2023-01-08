package clients

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

func UpdateClientSetting(ctx context.Context, u *models.ClientUser, desc, welcome string) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if desc != "" {
		if err := session.DB(ctx).Model(&models.Client{}).
			Where("client_id = ?", u.ClientID).
			Update("description", desc).Error; err != nil {
			return err
		}
	}
	if welcome != "" {
		if err := session.DB(ctx).Model(&models.ClientReplay{}).
			Where("client_id = ?", u.ClientID).
			Update("welcome", welcome).Error; err != nil {
			return err
		}
		go func() {
			// 给管理员发两条消息
			common.SendToClientManager(u.ClientID, &mixin.MessageView{
				ConversationID: mixin.UniqueConversationID(u.ClientID, u.UserID),
				UserID:         u.UserID,
				MessageID:      tools.GetUUID(),
				Category:       mixin.MessageCategoryPlainText,
				Data:           tools.Base64Encode([]byte(config.Text.WelcomeUpdate)),
				CreatedAt:      time.Now(),
			}, false, false)
			common.SendToClientManager(u.ClientID, &mixin.MessageView{
				ConversationID: mixin.UniqueConversationID(u.ClientID, u.UserID),
				UserID:         u.UserID,
				MessageID:      tools.GetUUID(),
				Category:       mixin.MessageCategoryPlainText,
				Data:           tools.Base64Encode([]byte(welcome)),
				CreatedAt:      time.Now(),
			}, false, false)
			go session.Redis(models.Ctx).QDel(ctx, "client:"+u.ClientID)
		}()
	}
	common.CacheClient(ctx, u.ClientID)
	return nil
}

func GetClientList(ctx context.Context) ([]*models.Client, error) {
	clientList := make([]*models.Client, 0)
	err := session.DB(ctx).Find(&clientList, "client_id in (?)", config.Config.ClientList).Error
	return clientList, err
}
