package models

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"log"
	"strings"
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
	err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT client_id,user_id,origin_message_id,message_id,quote_message_id,level,status,created_at
FROM distribute_messages
WHERE client_id=$1 AND message_id=$2
`, func(row pgx.Row) error {
		return row.Scan(&dm.ClientID, &dm.UserID, &dm.OriginMessageID, &dm.MessageID, &dm.QuoteMessageID, &dm.Level, &dm.Status, &dm.CreatedAt)
	}, clientID, messageID)
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
		if err := distributeMsg(ctx, clientID, ClientUserStatusManager, &mixin.MessageView{
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
		log.Println("mute")
	// 3. 拉黑
	case "block":
		log.Println("block")
	}

	return true, nil
}

func checkIsRecallMsg(ctx context.Context, clientID string, msg *mixin.MessageView) (bool, error) {
	if msg.Category != mixin.MessageCategoryPlainText {
		return false, nil
	}
	if msg.QuoteMessageID == "" {
		return false, nil
	}
	data := string(tools.Base64Decode(msg.Data))
	if data != "recall" {
		return false, nil
	}
	// 确认是管理员的 recall 消息
	// 把管理员 quote 的 msg 的 origin msgid 标记为 data
	if err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT origin_message_id FROM distribute_messages WHERE client_id=$1 AND message_id=$2
`, func(row pgx.Row) error {
		return row.Scan(&data)
	}, clientID, msg.QuoteMessageID); err != nil {
		return true, err
	} else {
		msg.Category = mixin.MessageCategoryMessageRecall
		dataByte, _ := json.Marshal(map[string]string{"message_id": data})
		msg.Data = tools.Base64Encode(dataByte)
	}
	go SendRecallMsg(clientID, msg)
	return false, nil
}

func SendToClientManager(clientID string, msg *mixin.MessageView) {
	//if msg.Category != mixin.MessageCategoryPlainImage && msg.Category != mixin.MessageCategoryPlainText {
	if msg.Category != mixin.MessageCategoryPlainText {
		return
	}
	users, err := getClientManager(_ctx, clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if len(users) <= 0 {
		log.Println("该社群没有管理员", clientID)
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
			QuoteMessageID:   msg.QuoteMessageID,
		})
	}
	if err := SendMessages(_ctx, client.Client, msgList); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := createMessage(_ctx, clientID, msg, MessageStatusLeaveMessage); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	for _, _msg := range msgList {
		if err := _createDistributeMessage(_ctx, clientID, _msg.RecipientID, msg.MessageID, _msg.MessageID, _msg.QuoteMessageID, ClientUserPriorityHigh, DistributeMessageStatusLeaveMessage, msg.CreatedAt); err != nil {
			session.Logger(_ctx).Println(err)
			continue
		}
	}
}
