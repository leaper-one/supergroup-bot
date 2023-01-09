package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
)

func GetPendingMessageByClientID(ctx context.Context, clientID string) ([]*models.Message, error) {
	ms := make([]*models.Message, 0)
	err := session.DB(ctx).Order("created_at").Find(&ms, "client_id=? AND status=?", clientID, models.MessageStatusPending).Error
	return ms, err
}

func createPendingMessage(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	if err := common.CreateMessage(ctx, clientID, msg, models.MessageStatusPending); err != nil && !durable.CheckIsPKRepeatError(err) {
		tools.Println(err)
		return err
	}
	if err := common.CreateDistributeMsgToRedis(ctx, []*models.DistributeMessage{{
		ClientID:        clientID,
		UserID:          msg.UserID,
		OriginMessageID: msg.MessageID,
		MessageID:       msg.MessageID,
		QuoteMessageID:  msg.QuoteMessageID,
		Status:          models.DistributeMessageStatusFinished,
		CreatedAt:       msg.CreatedAt,
	}}); err != nil {
		tools.Println(err)
		return err
	}
	return nil
}
func CreateDistributeMsgAndMarkStatus(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	userList, err := common.GetDistributeMsgUser(ctx, clientID, false, false)
	if err != nil {
		return err
	}
	// 处理 撤回 消息
	recallMsgIDMap := make(map[string]string)
	if msg.Category == mixin.MessageCategoryMessageRecall {
		recallMsgIDMap, err = getOriginMsgIDMapAndUpdateMsg(ctx, clientID, msg)
		if err != nil {
			return err
		}
		if len(recallMsgIDMap) == 0 {
			if err := updateMessageStatus(ctx, clientID, msg.MessageID, models.MessageStatusFinished); err != nil {
				tools.Println(err)
				return err
			}
			return nil
		}
	}
	// 处理 PIN 消息
	var action string
	var pinMsgIDs map[string][]string
	if msg.Category == "MESSAGE_PIN" {
		pinMsgIDs, action, err = getPINMsgIDMapAndUpdateMsg(ctx, msg, clientID)
		if err != nil {
			tools.Println(err)
			return err
		}
		if pinMsgIDs == nil {
			// 没有 pin 消息（可能被删除了）
			if err := updateMessageStatus(ctx, clientID, msg.MessageID, models.MessageStatusFinished); err != nil {
				tools.Println(err)
				return err
			}
			go common.SendClientUserTextMsg(clientID, msg.UserID, config.Text.PINMessageError, "")
			return nil
		}
		defer func() {
			msgIDs := make([]string, 0)
			for _, pinMsg := range pinMsgIDs {
				for _, v := range pinMsg {
					if v != "" {
						msgIDs = append(msgIDs, v)
					}
				}
			}
			if action == "UNPIN" {
				go session.DB(ctx).Table("distribute_messages").Where("message_id in ?", msgIDs).Update("status", models.MessageStatusFinished)
			} else if action == "PIN" {
				go createdPINDistributeMsg(models.Ctx, clientID, msgIDs)
			}
		}()
	}

	// 处理 quote 消息
	quoteMessageIDMap := make(map[string]string)
	if msg.QuoteMessageID != "" {
		originMsg, _ := common.GetDistributeMsgByMsgIDFromRedis(ctx, msg.QuoteMessageID)
		if originMsg != nil && originMsg.OriginMessageID != "" {
			quoteMessageIDMap, _, err = common.GetDistributeMessageIDMapByOriginMsgID(ctx, clientID, originMsg.OriginMessageID)
			if err != nil {
				tools.Println(err)
			}
		}
	}

	// 创建消息
	msgs := make([]*models.DistributeMessage, 0, len(userList))
	now := time.Now()
	// 处理 用户代理
	sendUserID := msg.UserID
	if sendUserID != config.Config.LuckCoinAppID &&
		sendUserID != "b523c28b-1946-4b98-a131-e1520780e8af" {
		if common.GetClientProxy(ctx, clientID) == models.ClientProxyStatusOn {
			u, err := common.GetClientUserByClientIDAndUserID(ctx, clientID, sendUserID)
			if err != nil {
				tools.Println(err)
				return nil
			}
			if u.Status != models.ClientUserStatusAdmin &&
				u.Status != models.ClientUserStatusGuest {
				proxy, err := common.GetClientUserProxyByProxyID(ctx, clientID, sendUserID)
				if err != nil {
					tools.Println(err)
					return nil
				} else {
					sendUserID = proxy.UserID
				}
			}
		}
	}

	for _, u := range userList {
		if u.UserID == msg.UserID || u.UserID == msg.RepresentativeID || common.CheckIsBlockUser(ctx, clientID, u.UserID) {
			continue
		}
		_data := ""
		// 处理 撤回 消息
		if msg.Category == mixin.MessageCategoryMessageRecall {
			if recallMsgIDMap[u.UserID] == "" {
				continue
			}
			data, err := json.Marshal(map[string]string{"message_id": recallMsgIDMap[u.UserID]})
			if err != nil {
				return err
			}
			msg.QuoteMessageID = ""
			_data = tools.Base64Encode(data)
		}
		// 处理 PIN 消息
		if msg.Category == "MESSAGE_PIN" {
			if pinMsgIDs[u.UserID] == nil || len(pinMsgIDs[u.UserID]) == 0 {
				continue
			}
			data, _ := json.Marshal(map[string]interface{}{"message_ids": pinMsgIDs[u.UserID], "action": action})
			_data = tools.Base64Encode(data)
		}
		if msg.QuoteMessageID != "" && quoteMessageIDMap[u.UserID] == "" {
			quoteMessageIDMap[u.UserID] = msg.QuoteMessageID
		}

		// 处理 聊天记录 消息
		msgID := mixin.UniqueConversationID(clientID+u.UserID+msg.MessageID, u.UserID+msg.MessageID+clientID)
		if msg.Category == "PLAIN_TRANSCRIPT" ||
			msg.Category == "ENCRYPTED_TRANSCRIPT" {
			t := make([]map[string]interface{}, 0)
			err := json.Unmarshal(tools.Base64Decode(msg.Data), &t)
			if err != nil {
				tools.Println(err)
				return err
			}
			for i := range t {
				t[i]["transcript_id"] = msgID
			}
			byteData, err := json.Marshal(t)
			if err != nil {
				tools.Println(err)
				return err
			}
			_data = tools.Base64Encode(byteData)
		}
		msgs = append(msgs, &models.DistributeMessage{
			ClientID:         clientID,
			UserID:           u.UserID,
			MessageID:        msgID,
			OriginMessageID:  msg.MessageID,
			QuoteMessageID:   quoteMessageIDMap[u.UserID],
			Category:         msg.Category,
			Data:             _data,
			RepresentativeID: sendUserID,
			Level:            u.Priority,
			Status:           models.DistributeMessageStatusPending,
			CreatedAt:        time.Now(),
		})
	}
	if err := common.CreateDistributeMsgToRedis(ctx, msgs); err != nil {
		tools.Println(err)
		return err
	}
	if err := session.Redis(ctx).QSet(ctx, fmt.Sprintf("msg_status:%s", msg.MessageID), strconv.Itoa(models.MessageStatusFinished), time.Hour*24); err != nil {
		return err
	}
	tools.PrintTimeDuration(fmt.Sprintf("%d条消息入库%s", len(msgs), clientID), now)
	return nil
}

func getOriginMsgIDMapAndUpdateMsg(ctx context.Context, clientID string, msg *mixin.MessageView) (map[string]string, error) {
	data := tools.Base64Decode(msg.Data)
	var _msg struct {
		MessageID string `json:"message_id"`
	}
	err := json.Unmarshal(data, &_msg)
	if err != nil {
		return nil, err
	}
	return common.GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, _msg.MessageID)
}
func getPINMsgIDMapAndUpdateMsg(ctx context.Context, msg *mixin.MessageView, clientID string) (map[string][]string, string, error) {
	action, orginMsgIDs := getPinOriginMsgIDs(ctx, msg.Data)
	if len(orginMsgIDs) == 0 {
		return nil, "", nil
	}
	var pinMsgIDMaps map[string][]string
	var err error
	if action == "PIN" {
		pinMsgIDMaps, err = getQuoteMsgIDUserIDsMapsFromRedis(ctx, orginMsgIDs)
	} else if action == "UNPIN" {
		pinMsgIDMaps, err = getUserIDMsgIDMapByOriginMsgIDFromPsql(ctx, orginMsgIDs)
	}
	if err != nil {
		return nil, "", err
	}
	if len(pinMsgIDMaps) == 0 {
		return nil, "", nil
	}
	status := models.MessageStatusPINMsg
	if action == "UNPIN" {
		status = models.MessageStatusFinished
	}
	for _, msgID := range orginMsgIDs {
		if err := updateMessageStatus(ctx, clientID, msgID, status); err != nil {
			tools.Println(err)
		}
	}
	return pinMsgIDMaps, action, nil
}

type MessagePinBody struct {
	Action     string   `json:"action"`
	MessageIDs []string `json:"message_ids"`
}

func getPinOriginMsgIDs(ctx context.Context, msgData string) (string, []string) {
	var msg MessagePinBody
	_ = json.Unmarshal(tools.Base64Decode(msgData), &msg)
	msgIDs := make([]string, 0, len(msg.MessageIDs))
	for _, msgID := range msg.MessageIDs {
		var m *models.DistributeMessage
		var err error
		if msg.Action == "PIN" {
			m, err = common.GetDistributeMsgByMsgIDFromRedis(ctx, msgID)
		} else if msg.Action == "UNPIN" {
			err = session.DB(ctx).Take(&m, "message_id = ?", msgID).Error
		}
		if err != nil {
			tools.Println(err)
		}
		if m != nil && m.OriginMessageID != "" {
			msgIDs = append(msgIDs, m.OriginMessageID)
		}
	}
	return msg.Action, msgIDs
}

func getQuoteMsgIDUserIDsMapsFromRedis(ctx context.Context, originMsgIDs []string) (map[string][]string, error) {
	quoteMsgIDMap := make(map[string][]string)
	for _, originMsgID := range originMsgIDs {
		msgIDMap, err := common.GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
		if err != nil {
			return nil, err
		}
		if len(msgIDMap) == 0 {
			continue
		}
		for userID, msgID := range msgIDMap {
			if quoteMsgIDMap[userID] == nil {
				quoteMsgIDMap[userID] = make([]string, 0)
			}
			quoteMsgIDMap[userID] = append(quoteMsgIDMap[userID], msgID)
		}
	}
	if len(quoteMsgIDMap) == 0 {
		return nil, nil
	}
	return quoteMsgIDMap, nil
}

func getUserIDMsgIDMapByOriginMsgIDFromPsql(ctx context.Context, originMsgIDs []string) (map[string][]string, error) {
	var dms []*models.DistributeMessage
	userIDMsgIDMap := make(map[string][]string)
	err := session.DB(ctx).Where("origin_message_id in ?", originMsgIDs).Find(&dms).Error
	if err != nil {
		return nil, err
	}
	for _, dm := range dms {
		if userIDMsgIDMap[dm.UserID] == nil {
			userIDMsgIDMap[dm.UserID] = make([]string, 0)
		}
		userIDMsgIDMap[dm.UserID] = append(userIDMsgIDMap[dm.UserID], dm.MessageID)
	}

	if _, err = session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, dm := range dms {
			if err := common.BuildOriginMsgAndMsgIndex(ctx, p, dm); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return userIDMsgIDMap, err
}

func createdPINDistributeMsg(ctx context.Context, clientID string, msgIDs []string) {
	// 1. 从 redis 中拿出来
	if len(msgIDs) == 0 {
		return
	}
	result := make([]*redis.StringCmd, 0, len(msgIDs))
	if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, msgID := range msgIDs {
			result = append(result, p.Get(ctx, fmt.Sprintf("msg_origin_idx:%s", msgID)))
		}
		return nil
	}); err != nil {
		tools.Println(err)
	}
	// 2. 存入 psql 中
	dataInserts := make([]*models.DistributeMessage, 0, len(msgIDs))
	for i, v := range result {
		tmp, err := v.Result()
		if err != nil {
			tools.Println(err)
			continue
		}
		msg, err := getOriginMsgFromRedisResult(tmp)
		if err != nil {
			tools.Println(err)
			continue
		}
		dataInserts = append(dataInserts, &models.DistributeMessage{
			ClientID:        clientID,
			UserID:          msg.UserID,
			OriginMessageID: msg.OriginMessageID,
			MessageID:       msgIDs[i],
			Status:          msg.Status,
		})
	}
	if err := session.DB(ctx).Save(&dataInserts); err != nil {
		tools.Println(err)
		return
	}
}

func getOriginMsgFromRedisResult(res string) (*models.DistributeMessage, error) {
	tmp := strings.Split(res, ",")
	if len(tmp) != 3 {
		tools.Println("invalid msg_origin_idx:", res)
		return nil, session.BadDataError(models.Ctx)
	}
	status, err := strconv.Atoi(tmp[2])
	if err != nil {
		return nil, err
	}
	return &models.DistributeMessage{
		OriginMessageID: tmp[0],
		UserID:          tmp[1],
		Status:          status,
	}, nil
}
