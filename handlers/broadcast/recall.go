package broadcast

import (
	"context"
	"encoding/json"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

func DeleteBroadcast(ctx context.Context, u *models.ClientUser, broadcastID string) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	// 发送一条 recall 的消息
	// 1. 找到之前的
	if err := UpdateBroadcast(ctx, u.ClientID, broadcastID, models.BroadcastStatusRecallPending); err != nil {
		return err
	}
	go recallBroadcastByID(u.ClientID, broadcastID)

	return nil
}

func recallBroadcastByID(clientID, originMsgID string) {
	ctx := models.Ctx
	var status int
	if err := session.DB(ctx).
		Table("broadcast").
		Select("status").
		Where("client_id=? AND message_id=?", clientID, originMsgID).
		Scan(&status).Error; err != nil {
		tools.Println(err)
		return
	}

	if status != models.BroadcastStatusRecallPending {
		return
	}
	dms, err := common.GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
	if err != nil {
		tools.Println(err)
		return
	}
	if len(dms) == 0 {
		return
	}
	// 构建 recall 消息请求
	msgs := make([]*mixin.MessageRequest, 0)
	for userID, MsgID := range dms {
		objData := map[string]string{"message_id": MsgID}
		byteData, _ := json.Marshal(objData)
		msgs = append(msgs, &mixin.MessageRequest{
			ConversationID: mixin.UniqueConversationID(clientID, userID),
			RecipientID:    userID,
			MessageID:      tools.GetUUID(),
			Category:       mixin.MessageCategoryMessageRecall,
			Data:           tools.Base64Encode(byteData),
		})
	}

	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}

	if err := common.SendBatchMessages(ctx, client.Client, msgs); err != nil {
		tools.Println(err)
		return
	}
	if err := UpdateBroadcast(ctx, clientID, originMsgID, models.BroadcastStatusRecallFinished); err != nil {
		tools.Println(err)
		return
	}
}
