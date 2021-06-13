package models

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
)

// 检查管理员的消息 是否 quote 了 留言消息，如果是的话，就在这个函数里处理 return true
func checkIsQuoteLeaveMessage(ctx context.Context, clientUser *ClientUser, msg *mixin.MessageView) (bool, error) {
	if msg.QuoteMessageID == "" {
		return false, nil
	}
	dm, err := getDistributeMessageByClientIDAndMessageID(ctx, clientUser.ClientID, msg.QuoteMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if dm.Status != DistributeMessageStatusLeaveMessage {
		return false, nil
	}
	// 确定是 quote 的留言信息了, 转发给其他管理员和该用户
	go handleLeaveMsg(clientUser.ClientID, clientUser.UserID, dm.OriginMessageID, msg)
	return true, nil
}

// 通过 clientID 和 messageID 获取 distributeMessage
func getDistributeMessageByClientIDAndMessageID(ctx context.Context, clientID, messageID string) (*DistributeMessage, error) {
	var dm DistributeMessage
	err := session.Database(ctx).QueryRow(ctx, `
SELECT client_id,user_id,origin_message_id,message_id,quote_message_id,level,status,created_at
FROM distribute_messages
WHERE client_id=$1 AND message_id=$2
`, clientID, messageID).Scan(&dm.ClientID, &dm.UserID, &dm.OriginMessageID, &dm.MessageID, &dm.QuoteMessageID, &dm.Level, &dm.Status, &dm.CreatedAt)
	return &dm, err
}

// 检查 是否是 帮转/禁言/拉黑 的消息
func checkIsOperation(ctx context.Context, clientID string, msg *mixin.MessageView) (bool, error) {
	if msg.Category != mixin.MessageCategoryPlainText {
		return false, nil
	}
	data := string(tools.Base64Decode(msg.Data))
	if !strings.HasPrefix(data, "---operation") {
		return false, nil
	}
	// 确定是操作的内容了
	operationAction := strings.Split(data, ",")
	if len(operationAction) != 3 {
		return true, nil
	}
	originMsg, err := getMsgByClientIDAndMessageID(ctx, clientID, operationAction[2])
	if err != nil {
		return true, err
	}
	switch operationAction[1] {
	// 1. 帮转发
	case "forward":
		if err := createAndDistributeMessage(ctx, clientID, &mixin.MessageView{
			ConversationID: originMsg.ConversationID,
			UserID:         originMsg.UserID,
			MessageID:      originMsg.MessageID,
			Category:       originMsg.Category,
			Data:           originMsg.Data,
			CreatedAt:      msg.CreatedAt,
		}); err != nil {
			return true, err
		}
	// 2. 禁言
	case "mute":
		if err := muteClientUser(ctx, clientID, originMsg.UserID, "12"); err != nil {
			session.Logger(ctx).Println(err)
		}
	// 3. 拉黑
	case "block":
		if err := blockClientUser(ctx, clientID, originMsg.UserID, false); err != nil {
			session.Logger(ctx).Println(err)
		}
	}

	return true, nil
}

func checkIsOperationMsg(ctx context.Context, clientID string, msg *mixin.MessageView) (bool, error) {
	data := string(tools.Base64Decode(msg.Data))
	if data == "/mute open" || data == "/mute close" {
		muteStatus := data == "/mute open"
		muteClientOperation(muteStatus, clientID)
		return true, nil
	}
	if msg.Category != mixin.MessageCategoryPlainText {
		return false, nil
	}
	if msg.QuoteMessageID == "" {
		return false, nil
	}
	if data != "recall" && data != "block" && !strings.HasPrefix(data, "mute") {
		return false, nil
	}
	var msgID string
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT origin_message_id FROM distribute_messages WHERE client_id=$1 AND message_id=$2
`, clientID, msg.QuoteMessageID).Scan(&msgID); err != nil {
		return true, err
	}
	var uid string
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT user_id FROM messages WHERE client_id=$1 AND message_id=$2`, clientID, msgID).Scan(&uid); err != nil {
		session.Logger(ctx).Println(err)
		return true, err
	}

	isRecall := true

	if strings.HasPrefix(data, "mute") {
		muteTime := "12"
		tmp := strings.Split(data, " ")
		if len(tmp) > 1 {
			t, err := strconv.Atoi(tmp[1])
			if err == nil && t >= 0 {
				muteTime = tmp[1]
			}
			if t == 0 {
				isRecall = false
			}
		}

		if err := muteClientUser(ctx, clientID, uid, muteTime); err != nil {
			return true, err
		}
	}
	if data == "block" {
		if err := blockClientUser(ctx, clientID, uid, false); err != nil {
			return true, err
		}
	}

	if isRecall {
		if err := CreatedManagerRecallMsg(ctx, clientID, msgID, uid); err != nil {
			return true, err
		}
	}

	return true, nil
}

func muteClientOperation(muteStatus bool, clientID string) {
	// 1. 如果是关闭
	if !muteStatus {
		ClientMuteStatus[clientID] = false
		go SendClientTextMsg(clientID, "社群禁言已解除", "", false)
		return
	}
	// 2. 如果是打开
	ClientMuteStatus[clientID] = true
	go SendClientTextMsg(clientID, "社群已禁言", "", false)
}

func SendToClientManager(clientID string, msg *mixin.MessageView) {
	if msg.Category != mixin.MessageCategoryPlainText {
		return
	}
	users, err := getClientManager(_ctx, clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if len(users) <= 0 {
		session.Logger(_ctx).Println("该社群没有管理员", clientID)
		return
	}
	client := GetMixinClientByID(_ctx, clientID)
	msgList := make([]*mixin.MessageRequest, 0)
	data := config.PrefixLeaveMsg + string(tools.Base64Decode(msg.Data))

	for _, userID := range users {
		conversationID := mixin.UniqueConversationID(clientID, userID)
		msgList = append(msgList, &mixin.MessageRequest{
			ConversationID:   conversationID,
			RecipientID:      userID,
			MessageID:        tools.GetUUID(),
			Category:         msg.Category,
			Data:             tools.Base64Encode([]byte(data)),
			RepresentativeID: msg.UserID,
		})
	}
	if msg.UserID == "" {
		data, _ := json.Marshal(msg)
		session.Logger(_ctx).Println(string(data))
	}
	if err := createMessage(_ctx, clientID, msg, MessageStatusLeaveMessage); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := SendMessages(_ctx, client.Client, msgList); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	var insert [][]interface{}
	for _, _msg := range msgList {
		row := _createDistributeMessage(_ctx,
			clientID, _msg.RecipientID, msg.MessageID, _msg.MessageID, "", msg.Category, msg.Data, msg.RepresentativeID, ClientUserPriorityHigh, DistributeMessageStatusLeaveMessage, msg.CreatedAt)
		insert = append(insert, row)
	}
	if err := createDistributeMsgList(_ctx, insert); err != nil {
		session.Logger(_ctx).Println(err)
	}
}
