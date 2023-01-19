package common

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
)

// 给客户端的每一个人发送一条消息，userID表示代表发送的用户，可以为空。
func SendClientTextMsg(clientID, msg, userID string, isJoinMsg bool) {
	if isJoinMsg && CheckIsBlockUser(models.Ctx, clientID, userID) {
		return
	}
	c, err := GetMixinClientByIDOrHost(models.Ctx, clientID)
	if err != nil {
		return
	}
	msgList := make([]*mixin.MessageRequest, 0)
	users, err := GetDistributeMsgUser(models.Ctx, clientID, isJoinMsg, false)
	if err != nil {
		tools.Println(err)
	}
	if len(users) <= 0 {
		return
	}
	msgBase64 := tools.Base64Encode([]byte(msg))
	originMsgID := tools.GetUUID()
	if isJoinMsg {
		if err := CreateMessage(models.Ctx, clientID, &mixin.MessageView{
			ConversationID: mixin.UniqueConversationID(clientID, userID),
			UserID:         userID,
			MessageID:      originMsgID,
			Category:       mixin.MessageCategoryPlainText,
			Data:           msgBase64,
		}, models.MessageStatusJoinMsg); err != nil {
			tools.Println(err)
		}
	}
	dms := make([]*models.DistributeMessage, 0, len(users))
	for _, u := range users {
		msgID := tools.GetUUID()
		if isJoinMsg {
			dms = append(dms, &models.DistributeMessage{
				ClientID:        clientID,
				UserID:          u.UserID,
				OriginMessageID: originMsgID,
				MessageID:       msgID,
				Category:        mixin.MessageCategoryPlainText,
				Level:           models.DistributeMessageLevelHigher,
				Status:          models.DistributeMessageStatusFinished,
			})
		}
		msgList = append(msgList, &mixin.MessageRequest{
			ConversationID: mixin.UniqueConversationID(clientID, u.UserID),
			RecipientID:    u.UserID,
			MessageID:      msgID,
			Category:       mixin.MessageCategoryPlainText,
			Data:           msgBase64,
		})
	}
	if isJoinMsg {
		if err := CreateDistributeMsgToRedis(models.Ctx, dms); err != nil {
			tools.Println(err)
		}
	}

	if err := SendBatchMessages(models.Ctx, c.Client, msgList); err != nil {
		tools.Println(err)
		return
	}
}

// 给社群里的每个人发送一条普通消息
func SendClientMsg(clientID, category, data string) {
	c, err := GetMixinClientByIDOrHost(models.Ctx, clientID)
	if err != nil {
		return
	}
	msgList := make([]*mixin.MessageRequest, 0)
	users, err := GetDistributeMsgUser(models.Ctx, clientID, false, false)
	if err != nil {
		tools.Println(err)
	}
	if len(users) <= 0 {
		return
	}
	originMsgID := tools.GetUUID()
	if err := CreateMessage(models.Ctx, clientID, &mixin.MessageView{
		MessageID: originMsgID,
		Category:  category,
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, models.MessageStatusClientMsg); err != nil {
		tools.Println(err)
	}

	for _, u := range users {
		msgList = append(msgList, &mixin.MessageRequest{
			ConversationID: mixin.UniqueConversationID(clientID, u.UserID),
			RecipientID:    u.UserID,
			MessageID:      tools.GetUUID(),
			Category:       category,
			Data:           data,
		})
	}

	if err := SendBatchMessages(models.Ctx, c.Client, msgList); err != nil {
		tools.Println(err)
		return
	}
}

// 指定大群给指定用户发送一条文本消息
func SendClientUserTextMsg(clientID, userID, data, quoteMsgID string) {
	uid, err := uuid.FromString(userID)
	if err != nil || uid.IsNil() {
		return
	}

	ctx := models.Ctx
	if data == "" {
		return
	}
	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return
	}

	representativeID := ""

	if data != config.Text.AuthSuccess {
		admin, err := GetClientAdminOrOwner(ctx, clientID)
		if err != nil {
			tools.Println(err)
			return
		}

		if representativeID != userID {
			representativeID = admin.UserID
		}
	}

	msg := &mixin.MessageRequest{
		ConversationID:   mixin.UniqueConversationID(client.ClientID, userID),
		RecipientID:      userID,
		MessageID:        tools.GetUUID(),
		Category:         mixin.MessageCategoryPlainText,
		Data:             tools.Base64Encode([]byte(data)),
		QuoteMessageID:   quoteMsgID,
		RepresentativeID: representativeID,
	}
	if err := SendMessage(ctx, client.Client, msg, false); err != nil {
		tools.Println(err)
		tools.PrintJson(msg)
		return
	}
}

func SendBtnMsg(ctx context.Context, clientID, userID string, data mixin.AppButtonGroupMessage) error {
	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return errors.New("client is nil")
	}
	conversationID := mixin.UniqueConversationID(client.ClientID, userID)
	if err := SendMessage(ctx, client.Client, &mixin.MessageRequest{
		ConversationID: conversationID,
		RecipientID:    userID,
		MessageID:      tools.GetUUID(),
		Category:       mixin.MessageCategoryAppButtonGroup,
		Data:           getBtnMsg(data),
	}, false); err != nil {
		return err
	}
	return nil
}

func getBtnMsg(data mixin.AppButtonGroupMessage) string {
	btnData, err := json.Marshal(data)
	if err != nil {
		tools.Println(err)
		return ""
	}
	return tools.Base64Encode(btnData)
}

func GetPlainCategory(category string) string {
	if strings.HasPrefix(category, "ENCRYPTED_") {
		category = strings.Replace(category, "ENCRYPTED_", "PLAIN_", 1)
	}
	return category
}
