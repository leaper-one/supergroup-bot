package broadcast

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
)

func CreateBroadcast(ctx context.Context, u *models.ClientUser, data, category string) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	msgID := tools.GetUUID()
	now := time.Now()
	if category == "" {
		category = mixin.MessageCategoryPlainText
	}
	data = tools.Base64Encode([]byte(data))
	// 创建一条消息
	msg := &mixin.MessageView{
		ConversationID: mixin.UniqueConversationID(u.ClientID, u.UserID),
		UserID:         u.UserID,
		MessageID:      msgID,
		Category:       category,
		Data:           data,
		Status:         mixin.MessageStatusSent,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := session.DB(ctx).Create(&models.Broadcast{ClientID: u.ClientID, MessageID: msgID}).Error; err != nil {
		tools.Println(err)
		return err
	}

	if err := common.CreateMessage(ctx, u.ClientID, msg, models.MessageStatusBroadcast); err != nil {
		tools.Println(err)
		return err
	}
	go SendBroadcast(u, msgID, category, data, now)
	return nil
}
func SendBroadcast(u *models.ClientUser, msgID, category, data string, now time.Time) {
	ctx := models.Ctx
	users, err := common.GetDistributeMsgUser(ctx, u.ClientID, false, true)
	if err != nil {
		tools.Println(err)
		return
	}
	msgs := make([]*mixin.MessageRequest, 0)
	for _, _u := range users {
		if common.CheckIsBlockUser(ctx, u.ClientID, _u.UserID) {
			continue
		}
		_msgID := tools.GetUUID()
		msgs = append(msgs, &mixin.MessageRequest{
			ConversationID: mixin.UniqueConversationID(u.ClientID, _u.UserID),
			RecipientID:    _u.UserID,
			MessageID:      _msgID,
			Category:       category,
			Data:           data,
		})
	}
	client, err := common.GetMixinClientByIDOrHost(ctx, u.ClientID)
	if err != nil {
		return
	}
	if err := common.SendBatchMessages(ctx, client.Client, msgs); err != nil {
		tools.Println(err)
		return
	}
	if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, msg := range msgs {
			if err := common.BuildOriginMsgAndMsgIndex(ctx, p, &models.DistributeMessage{
				UserID:          msg.RecipientID,
				OriginMessageID: msgID,
				MessageID:       msg.MessageID,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		tools.Println(err)
		return
	}
	if err := UpdateBroadcast(ctx, u.ClientID, msgID, models.BroadcastStatusFinished); err != nil {
		tools.Println(err)
	}
}
