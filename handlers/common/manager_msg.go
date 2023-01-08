package common

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

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

// 检查 是否是 帮转/禁言/拉黑 的消息
func checkIsButtonOperation(ctx context.Context, clientID string, msg *mixin.MessageView) (bool, error) {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != "ENCRYPTED_TEXT" {
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
		if err := CreateAndDistributeMessage(ctx, clientID, &mixin.MessageView{
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
		if err := MuteClientUser(ctx, clientID, originMsg.UserID, "12"); err != nil {
			tools.Println(err)
		}
	// 3. 拉黑
	case "block":
		if err := BlockClientUser(ctx, clientID, msg.UserID, originMsg.UserID, false); err != nil {
			tools.Println(err)
		}
	}

	return true, nil
}

func checkIsOperationMsg(ctx context.Context, u *models.ClientUser, msg *mixin.MessageView) (bool, error) {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != "ENCRYPTED_TEXT" {
		return false, nil
	}
	data := string(tools.Base64Decode(msg.Data))
	if data == "/mute open" || data == "/mute close" {
		muteStatus := data == "/mute open"
		MuteClientOperation(muteStatus, u.ClientID)
		return true, nil
	}
	if isOperation, err := handleUnmuteAndUnblockMsg(ctx, data, u); err != nil {
		tools.Println(err)
	} else if isOperation {
		return true, nil
	}

	return handleRecallOrMuteOrBlockOrInfoMsg(ctx, data, u.ClientID, msg)
}

func handleRecallOrMuteOrBlockOrInfoMsg(ctx context.Context, data, clientID string, msg *mixin.MessageView) (bool, error) {
	if msg.QuoteMessageID == "" {
		return false, nil
	}
	if data != "/info" && data != "ban" && data != "kick" && data != "delete" && data != "/recall" && data != "/block" && !strings.HasPrefix(data, "/mute") {
		return false, nil
	}
	dm, err := GetDistributeMsgByMsgIDFromRedis(ctx, msg.QuoteMessageID)
	if err != nil {
		return true, err
	}
	m, err := getMsgByClientIDAndMessageID(ctx, clientID, dm.OriginMessageID)
	if err != nil {
		tools.Println(err)
		return true, err
	}
	if data == "/recall" || data == "delete" {
		if err := CreatedManagerRecallMsg(ctx, clientID, dm.OriginMessageID, m.UserID); err != nil {
			return true, err
		}
	}
	// 针对用户的操作
	if data == "/info" {
		checkAndReplaceProxyUser(ctx, clientID, &m.UserID)
		objData := map[string]string{"user_id": m.UserID}
		byteData, _ := json.Marshal(objData)
		client, err := GetMixinClientByIDOrHost(ctx, clientID)
		if err != nil {
			return true, err
		}
		go SendMessage(_ctx, client.Client, &mixin.MessageRequest{
			ConversationID: msg.ConversationID,
			RecipientID:    msg.RepresentativeID,
			MessageID:      tools.GetUUID(),
			Category:       mixin.MessageCategoryPlainContact,
			Data:           tools.Base64Encode(byteData),
		}, false)
		return true, nil
	}
	if strings.HasPrefix(data, "/mute") || data == "kick" {
		muteTime := "12"
		tmp := strings.Split(data, " ")
		if len(tmp) > 1 {
			t, err := strconv.Atoi(tmp[1])
			if err == nil && t >= 0 {
				muteTime = tmp[1]
			}
		}
		if err := MuteClientUser(ctx, clientID, m.UserID, muteTime); err != nil {
			return true, err
		}
	}
	if data == "/block" || data == "ban" {
		if err := BlockClientUser(ctx, clientID, msg.UserID, m.UserID, false); err != nil {
			return true, err
		}
	}
	return true, nil
}

func handleUnmuteAndUnblockMsg(ctx context.Context, data string, u *models.ClientUser) (bool, error) {
	operation := strings.Split(data, " ")
	if len(operation) < 2 || len(operation[1]) <= 4 {
		return false, nil
	}
	if strings.HasPrefix(data, "/unmute") {
		_u, err := SearchUser(ctx, u.ClientID, operation[1])
		if err != nil {
			tools.Println(err)
			return true, nil
		}
		if err := MuteClientUser(ctx, u.ClientID, _u.UserID, "0"); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	if strings.HasPrefix(data, "/unblock") {
		_u, err := SearchUser(ctx, u.ClientID, operation[1])
		if err != nil {
			tools.Println(err)
			return true, nil
		}
		if err := BlockClientUser(ctx, u.ClientID, u.UserID, _u.UserID, true); err != nil {
			tools.Println(err)
		}
		return true, nil
	}

	if strings.HasPrefix(data, "/blockall") {
		if checkIsSuperManager(u.UserID) {
			_u, err := SearchUser(ctx, u.ClientID, operation[1])
			if err != nil {
				tools.Println(err)
				return true, nil
			}
			memo := ""
			if len(operation) == 3 {
				memo = operation[2]
			}
			if err := AddBlockUser(ctx, u.UserID, u.ClientID, _u.UserID, memo); err != nil {
				tools.Println(err)
			}
			if err := SendClientUserTextMsg(ctx, u.ClientID, u.UserID, "success", ""); err != nil {
				tools.Println(err)
			}
		}
		return true, nil
	}
	return false, nil
}

func MuteClientOperation(muteStatus bool, clientID string) {
	if muteStatus {
		// 1. 如果是关闭
		if err := SetClientConversationStatusByIDAndStatus(_ctx, clientID, models.ClientConversationStatusMute); err != nil {
			tools.Println(err)
		} else {
			DeleteDistributeMsgByClientID(_ctx, clientID)
			go SendClientTextMsg(clientID, config.Text.MuteOpen, "", false)
		}

	} else {
		// 2. 如果是打开
		if err := SetClientConversationStatusByIDAndStatus(_ctx, clientID, models.ClientConversationStatusNormal); err != nil {
			tools.Println(err)
		} else {
			go SendClientTextMsg(clientID, config.Text.MuteClose, "", false)
		}
	}
}

func SendToClientManager(clientID string, msg *mixin.MessageView, isLeaveMsg, hasRepresentativeID bool) {
	if msg.Category != mixin.MessageCategoryPlainText &&
		msg.Category != mixin.MessageCategoryPlainImage &&
		msg.Category != mixin.MessageCategoryPlainVideo {
		return
	}
	managers, err := getClientManager(_ctx, clientID)
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
	if err := CreateMessage(_ctx, clientID, msg, MessageStatusLeaveMessage); err != nil {
		tools.Println(err)
		return
	}
	client, err := GetMixinClientByIDOrHost(_ctx, clientID)
	if err != nil {
		return
	}
	if err := SendMessages(_ctx, client.Client, msgList); err != nil {
		tools.Println(err)
		return
	}
	if _, err := session.Redis(_ctx).QPipelined(_ctx, func(p redis.Pipeliner) error {
		for _, _msg := range msgList {
			dm := &models.DistributeMessage{
				MessageID:       _msg.MessageID,
				UserID:          _msg.RecipientID,
				OriginMessageID: msg.MessageID,
			}
			if isLeaveMsg {
				dm.Status = models.DistributeMessageStatusLeaveMessage
			}
			if err := BuildOriginMsgAndMsgIndex(_ctx, p, dm); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		tools.Println(err)
	}
}
