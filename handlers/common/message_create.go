package common

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
)

type transcript map[string]interface{}

type MessagePinBody struct {
	Action     string   `json:"action"`
	MessageIDs []string `json:"message_ids"`
}

// 创建消息 和 分发消息列表
func CreateAndDistributeMessage(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	// 1. 创建消息
	err := CreateMessage(ctx, clientID, msg, MessageStatusNormal)
	if err != nil && !durable.CheckIsPKRepeatError(err) {
		tools.Println(err)
		return err
	}
	// 2. 创建分发消息列表
	return CreateDistributeMsgAndMarkStatus(ctx, clientID, msg)
}

// 创建分发消息 标记 并把消息标记
func CreateDistributeMsgAndMarkStatus(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	userList, err := GetDistributeMsgUser(ctx, clientID, false, false)
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
			if err := updateMessageStatus(ctx, clientID, msg.MessageID, MessageStatusFinished); err != nil {
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
			if err := updateMessageStatus(ctx, clientID, msg.MessageID, MessageStatusFinished); err != nil {
				tools.Println(err)
				return err
			}
			go SendClientUserTextMsg(_ctx, clientID, msg.UserID, config.Text.PINMessageError, "")
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
				go removePINDistributeMsg(_ctx, msgIDs)
			} else if action == "PIN" {
				go createdPINDistributeMsg(_ctx, clientID, msgIDs)
			}
		}()
	}

	// 处理 quote 消息
	quoteMessageIDMap := make(map[string]string)
	if msg.QuoteMessageID != "" {
		originMsg, _ := GetDistributeMsgByMsgIDFromRedis(ctx, msg.QuoteMessageID)
		if originMsg != nil && originMsg.OriginMessageID != "" {
			quoteMessageIDMap, _, err = GetDistributeMessageIDMapByOriginMsgID(ctx, clientID, originMsg.OriginMessageID)
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
		if GetClientProxy(ctx, clientID) == ClientProxyStatusOn {
			u, err := GetClientUserByClientIDAndUserID(ctx, clientID, sendUserID)
			if err != nil {
				tools.Println(err)
				return nil
			}
			if u.Status != models.ClientUserStatusAdmin &&
				u.Status != models.ClientUserStatusGuest {
				proxy, err := getClientUserProxyByProxyID(ctx, clientID, sendUserID)
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
		if u.UserID == msg.UserID || u.UserID == msg.RepresentativeID || CheckIsBlockUser(ctx, clientID, u.UserID) {
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
			t := make([]transcript, 0)
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
	if err := createDistributeMsgToRedis(ctx, msgs); err != nil {
		tools.Println(err)
		return err
	}
	if err := session.Redis(ctx).QSet(ctx, fmt.Sprintf("msg_status:%s", msg.MessageID), strconv.Itoa(MessageStatusFinished), time.Hour*24); err != nil {
		return err
	}
	tools.PrintTimeDuration(fmt.Sprintf("%d条消息入库%s", len(msgs), clientID), now)
	return nil
}

func getOriginMsgIDMapAndUpdateMsg(ctx context.Context, clientID string, msg *mixin.MessageView) (map[string]string, error) {
	originMsgID := getRecallOriginMsgID(ctx, msg.Data)
	return GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
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
	status := MessageStatusPINMsg
	if action == "UNPIN" {
		status = MessageStatusFinished
	}
	for _, msgID := range orginMsgIDs {
		if err := updateMessageStatus(ctx, clientID, msgID, status); err != nil {
			tools.Println(err)
		}
	}
	return pinMsgIDMaps, action, nil
}

func GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx context.Context, originMsgID string) (map[string]string, error) {
	recallMsgIDMap := make(map[string]string)
	resList, err := session.Redis(ctx).QSMembers(ctx, "origin_msg_idx:"+originMsgID)
	if err != nil {
		return nil, err
	}
	for _, res := range resList {
		msg, err := getMsgOriginFromRedisResult(res)
		if err != nil {
			continue
		}
		recallMsgIDMap[msg.UserID] = msg.MessageID
	}
	return recallMsgIDMap, nil
}

func getQuoteMsgIDUserIDsMapsFromRedis(ctx context.Context, originMsgIDs []string) (map[string][]string, error) {
	quoteMsgIDMap := make(map[string][]string)
	for _, originMsgID := range originMsgIDs {
		msgIDMap, err := GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
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
	err := session.DB(ctx).ConnQuery(ctx, `
SELECT user_id, message_id, status, origin_message_id
FROM distribute_messages
WHERE origin_message_id=ANY($1)
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var dm models.DistributeMessage
			if err := rows.Scan(&dm.UserID, &dm.MessageID, &dm.Status, &dm.OriginMessageID); err != nil {
				return err
			}
			dms = append(dms, &dm)
			if userIDMsgIDMap[dm.UserID] == nil {
				userIDMsgIDMap[dm.UserID] = make([]string, 0)
			}
			userIDMsgIDMap[dm.UserID] = append(userIDMsgIDMap[dm.UserID], dm.MessageID)
		}
		return nil
	}, originMsgIDs)
	if err != nil {
		return nil, err
	}
	if _, err = session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, dm := range dms {
			if err := BuildOriginMsgAndMsgIndex(ctx, p, dm); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return userIDMsgIDMap, err
}

func CreatedManagerRecallMsg(ctx context.Context, clientID string, msgID, uid string) error {
	dataByte, _ := json.Marshal(map[string]string{"message_id": msgID})

	if err := CreateAndDistributeMessage(ctx, clientID, &mixin.MessageView{
		UserID:    uid,
		MessageID: tools.GetUUID(),
		Category:  mixin.MessageCategoryMessageRecall,
		Data:      tools.Base64Encode(dataByte),
	}); err != nil {
		tools.Println(err)
	}

	return nil
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
	dataInserts := make([][]interface{}, 0, len(msgIDs))
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
		dataInserts = append(dataInserts, []interface{}{clientID, msg.UserID, msg.OriginMessageID, msgIDs[i], msg.Status})
	}
	if err := createDistributeMsgList(ctx, dataInserts); err != nil {
		tools.Println(err)
		return
	}
}

func removePINDistributeMsg(ctx context.Context, msgIDs []string) {
	// 1. 从 psql 中标记为已成功，定时任务会 24 小时清理
	if _, err := session.DB(ctx).Exec(ctx, `UPDATE distribute_messages SET status=2 WHERE message_id=ANY($1)`, msgIDs); err != nil {
		tools.Println(err)
		return
	}
}

var distributeCols = []string{"client_id", "user_id", "origin_message_id", "message_id", "status"}

func createDistributeMsgList(ctx context.Context, insert [][]interface{}) error {
	var ident = pgx.Identifier{"distribute_messages"}
	if len(insert) == 0 {
		return nil
	}
	_, err := session.DB(ctx).CopyFrom(ctx, ident, distributeCols, pgx.CopyFromRows(insert))
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key") {
			tools.Println(err)
		}
	}
	return nil
}

func getShardID(clientID, userID string) string {
	shardID := ClientShardIDMap[clientID][userID]
	if shardID == "" {
		shardID = strconv.Itoa(rand.Intn(int(config.MessageShardSize)))
	}
	return shardID
}

func getRecallOriginMsgID(ctx context.Context, msgData string) string {
	data := tools.Base64Decode(msgData)
	var msg struct {
		MessageID string `json:"message_id"`
	}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		tools.Println(err)
		return ""
	}
	return msg.MessageID
}

func getPinOriginMsgIDs(ctx context.Context, msgData string) (string, []string) {
	var msg MessagePinBody
	_ = json.Unmarshal(tools.Base64Decode(msgData), &msg)
	msgIDs := make([]string, 0, len(msg.MessageIDs))
	for _, msgID := range msg.MessageIDs {
		var m *models.DistributeMessage
		var err error
		if msg.Action == "PIN" {
			m, err = GetDistributeMsgByMsgIDFromRedis(ctx, msgID)
		} else if msg.Action == "UNPIN" {
			m, err = getDistributeMsgByMsgIDFromPsql(ctx, msgID)
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

var ClientShardIDMap = make(map[string]map[string]string)
