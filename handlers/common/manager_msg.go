package common

import (
	"context"
	"encoding/json"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
)

// 通过 clientID 和 messageID 获取 distributeMessage
func GetDistributeMsgByMsgIDFromRedis(ctx context.Context, msgID string) (*models.DistributeMessage, error) {
	res, err := session.Redis(ctx).SyncGet(ctx, "msg_origin_idx:"+msgID).Result()
	if err != nil {
		return nil, err
	}
	return getOriginMsgFromRedisResult(res)
}

func SendToClientManager(clientID string, msg *mixin.MessageView, isLeaveMsg, hasRepresentativeID bool) {
	ctx := models.Ctx
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != mixin.MessageCategoryPlainImage &&
		msg.Category != mixin.MessageCategoryPlainVideo {
		return
	}
	managers, err := GetClientUsersByClientIDAndStatus(ctx, clientID, models.ClientUserStatusAdmin)
	if err != nil {
		tools.Println(err)
		return
	}
	if len(managers) <= 0 {
		tools.Println("该社群没有管理员", clientID)
		return
	}
	msgList := make([]*mixin.MessageRequest, 0)
	var data string
	if isLeaveMsg && msg.Category == mixin.MessageCategoryPlainText {
		data = tools.Base64Encode([]byte(config.Text.PrefixLeaveMsg + string(tools.Base64Decode(msg.Data))))
	} else {
		data = msg.Data
	}

	for _, userID := range managers {
		conversationID := mixin.UniqueConversationID(clientID, userID)
		_msg := mixin.MessageRequest{
			ConversationID:   conversationID,
			RecipientID:      userID,
			MessageID:        mixin.UniqueConversationID(msg.MessageID, userID),
			Category:         msg.Category,
			Data:             data,
			RepresentativeID: msg.UserID,
		}
		if !hasRepresentativeID {
			_msg.RepresentativeID = ""
		}
		msgList = append(msgList, &_msg)
	}
	if msg.UserID == "" {
		data, _ := json.Marshal(msg)
		tools.Println(string(data))
	}
	if err := CreateMessage(ctx, clientID, msg, MessageStatusLeaveMessage); err != nil {
		tools.Println(err)
		return
	}
	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}
	if err := SendMessages(client.Client, msgList); err != nil {
		tools.Println(err)
		return
	}
	if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, _msg := range msgList {
			dm := &models.DistributeMessage{
				MessageID:       _msg.MessageID,
				UserID:          _msg.RecipientID,
				OriginMessageID: msg.MessageID,
			}
			if isLeaveMsg {
				dm.Status = models.DistributeMessageStatusLeaveMessage
			}
			if err := BuildOriginMsgAndMsgIndex(ctx, p, dm); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		tools.Println(err)
	}
}
