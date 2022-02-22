package models

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
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

type transcript map[string]interface{}

type MessagePinBody struct {
	Action     string   `json:"action"`
	MessageIDs []string `json:"message_ids"`
}

// 创建消息 和 分发消息列表
func createAndDistributeMessage(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	// 1. 创建消息
	err := createMessage(ctx, clientID, msg, MessageStatusNormal)
	if err != nil && !durable.CheckIsPKRepeatError(err) {
		session.Logger(ctx).Println(err)
		return err
	}
	// 2. 创建分发消息列表
	return CreateDistributeMsgAndMarkStatus(ctx, clientID, msg, []int{ClientUserPriorityHigh, ClientUserPriorityLow})
}

// 创建分发消息 标记 并把消息标记
func CreateDistributeMsgAndMarkStatus(ctx context.Context, clientID string, msg *mixin.MessageView, priorityList []int) error {
	userList, err := GetClientUserByPriority(ctx, clientID, priorityList, false, false)
	if err != nil {
		return err
	}
	level := priorityList[0]
	var status int
	if level == ClientUserPriorityHigh {
		status = MessageStatusPrivilege
	} else if level == ClientUserPriorityLow {
		status = MessageStatusFinished
	}
	// 处理 撤回 消息
	recallMsgIDMap := make(map[string]string)
	if msg.Category == mixin.MessageCategoryMessageRecall {
		recallMsgIDMap, err = getOriginMsgIDMapAndUpdateMsg(ctx, clientID, msg)
		if err != nil {
			return err
		}
		if recallMsgIDMap == nil {
			if err := updateMessageStatus(ctx, clientID, msg.MessageID, MessageStatusFinished); err != nil {
				session.Logger(ctx).Println(err)
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
			session.Logger(ctx).Println(err)
			return err
		}
		if pinMsgIDs == nil {
			// 没有 pin 消息（可能被删除了）
			if status == MessageStatusFinished {
				go SendTextMsg(_ctx, clientID, msg.UserID, config.Text.PINMessageErorr)
			}
			if err := updateMessageStatus(ctx, clientID, msg.MessageID, MessageStatusFinished); err != nil {
				session.Logger(ctx).Println(err)
				return err
			}
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
	// 创建消息
	msgs := make([]*DistributeMessage, 0, len(userList))
	quoteMessageIDMap := make(map[string]string)
	if msg.QuoteMessageID != "" {
		originMsg, _ := getDistributeMsgByMsgIDFromRedis(ctx, msg.QuoteMessageID)
		if originMsg != nil && originMsg.OriginMessageID != "" {
			quoteMessageIDMap, _, err = getDistributeMessageIDMapByOriginMsgID(ctx, clientID, originMsg.OriginMessageID)
			if err != nil {
				session.Logger(ctx).Println(err)
			}
		}
	}
	sendUserID := msg.UserID
	if sendUserID != config.Config.LuckCoinAppID &&
		sendUserID != "b523c28b-1946-4b98-a131-e1520780e8af" {
		if GetClientProxy(ctx, clientID) == ClientProxyStatusOn {
			u, err := GetClientUserByClientIDAndUserID(ctx, clientID, sendUserID)
			if err != nil {
				session.Logger(ctx).Println(err)
				return nil
			}
			if u.Status != ClientUserStatusAdmin {
				proxy, err := getClientUserProxyByProxyID(ctx, clientID, sendUserID)
				if err != nil {
					session.Logger(ctx).Println(err)
					return nil
				} else {
					sendUserID = proxy.UserID
				}
			}
		}
	}

	now := time.Now()
	for _, userID := range userList {
		if userID == msg.UserID || userID == msg.RepresentativeID || checkIsBlockUser(ctx, clientID, userID) {
			continue
		}
		_data := ""
		// 处理 撤回 消息
		if msg.Category == mixin.MessageCategoryMessageRecall {
			if recallMsgIDMap[userID] == "" {
				continue
			}
			data, err := json.Marshal(map[string]string{"message_id": recallMsgIDMap[userID]})
			if err != nil {
				return err
			}
			msg.QuoteMessageID = ""
			_data = tools.Base64Encode(data)
		}
		// 处理 PIN 消息
		if msg.Category == "MESSAGE_PIN" {
			if pinMsgIDs[userID] == nil || len(pinMsgIDs[userID]) == 0 {
				continue
			}
			data, _ := json.Marshal(map[string]interface{}{"message_ids": pinMsgIDs[userID], "action": action})
			_data = tools.Base64Encode(data)
		}
		if msg.QuoteMessageID != "" && quoteMessageIDMap[userID] == "" {
			quoteMessageIDMap[userID] = msg.QuoteMessageID
		}

		// 处理 聊天记录 消息
		msgID := mixin.UniqueConversationID(clientID+userID+msg.MessageID, userID+msg.MessageID+clientID)
		if msg.Category == "PLAIN_TRANSCRIPT" ||
			msg.Category == "ENCRYPTED_TRANSCRIPT" {
			t := make([]transcript, 0)
			err := json.Unmarshal(tools.Base64Decode(msg.Data), &t)
			if err != nil {
				session.Logger(ctx).Println(err)
				return err
			}
			for i := range t {
				t[i]["transcript_id"] = msgID
			}
			byteData, err := json.Marshal(t)
			if err != nil {
				session.Logger(ctx).Println(err)
				return err
			}
			_data = tools.Base64Encode(byteData)
		}
		msgs = append(msgs, &DistributeMessage{
			ClientID:         clientID,
			UserID:           userID,
			MessageID:        msgID,
			OriginMessageID:  msg.MessageID,
			QuoteMessageID:   quoteMessageIDMap[userID],
			Category:         msg.Category,
			Data:             _data,
			RepresentativeID: sendUserID,
			Level:            level,
			Status:           DistributeMessageStatusPending,
			CreatedAt:        time.Now(),
		})
	}
	if err := createDistributeMsgToRedis(ctx, msgs); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	if err := session.Redis(ctx).Set(ctx, fmt.Sprintf("msg_status:%s", msg.MessageID), strconv.Itoa(status), -1).Err(); err != nil {
		return err
	}
	tools.PrintTimeDuration(fmt.Sprintf("%d条消息入库%s", len(msgs), clientID), now)
	return nil
}

func getOriginMsgIDMapAndUpdateMsg(ctx context.Context, clientID string, msg *mixin.MessageView) (map[string]string, error) {
	originMsgID := getRecallOriginMsgID(ctx, msg.Data)
	recallMsgIDMap, err := getQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
	if err != nil {
		return nil, err
	}
	if len(recallMsgIDMap) == 0 {
		return nil, nil
	}
	return recallMsgIDMap, nil
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
			session.Logger(ctx).Println(err)
		}
	}
	return pinMsgIDMaps, action, nil
}

func getQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx context.Context, originMsgID string) (map[string]string, error) {
	recallMsgIDMap := make(map[string]string)
	resList, err := session.Redis(ctx).SMembers(ctx, "origin_msg_idx:"+originMsgID).Result()
	if err != nil {
		return nil, err
	}
	for _, res := range resList {
		resList := strings.Split(res, ",")
		if len(resList) != 2 {
			continue
		}
		msgID := resList[0]
		userID := resList[1]
		recallMsgIDMap[userID] = msgID
	}
	if len(recallMsgIDMap) == 0 {
		// 消息已经被删除...
		return nil, nil
	}
	return recallMsgIDMap, nil
}

func getQuoteMsgIDUserIDsMapsFromRedis(ctx context.Context, originMsgIDs []string) (map[string][]string, error) {
	quoteMsgIDMap := make(map[string][]string)
	for _, originMsgID := range originMsgIDs {
		msgIDMap, err := getQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
		if err != nil {
			return nil, err
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
	var dms []*DistributeMessage
	userIDMsgIDMap := make(map[string][]string)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT user_id, message_id, status, origin_message_id
FROM distribute_messages
WHERE origin_message_id=ANY($1)
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var dm DistributeMessage
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
	if _, err = session.Redis(ctx).Pipelined(ctx, func(p redis.Pipeliner) error {
		for _, dm := range dms {
			if err := buildOriginMsgAndMsgIndex(ctx, p, dm); err != nil {
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

	if err := createAndDistributeMessage(ctx, clientID, &mixin.MessageView{
		UserID:    uid,
		MessageID: tools.GetUUID(),
		Category:  mixin.MessageCategoryMessageRecall,
		Data:      tools.Base64Encode(dataByte),
	}); err != nil {
		session.Logger(ctx).Println(err)
	}

	return nil
}

func createdPINDistributeMsg(ctx context.Context, clientID string, msgIDs []string) {
	// 1. 从 redis 中拿出来
	if len(msgIDs) == 0 {
		return
	}
	result := make([]*redis.StringCmd, 0, len(msgIDs))
	if _, err := session.Redis(ctx).Pipelined(ctx, func(p redis.Pipeliner) error {
		for _, msgID := range msgIDs {
			result = append(result, p.Get(ctx, fmt.Sprintf("msg_origin_idx:%s", msgID)))
		}
		return nil
	}); err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	// 2. 存入 psql 中
	dataInserts := make([][]interface{}, 0, len(msgIDs))
	for i, v := range result {
		tmp, err := v.Result()
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
		msg, err := getOriginMsgFromRedisResult(tmp)
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
		dataInserts = append(dataInserts, []interface{}{clientID, msg.UserID, msg.OriginMessageID, msgIDs[i], msg.Status})
	}
	if err := createDistributeMsgList(ctx, dataInserts); err != nil {
		session.Logger(ctx).Println(err)
		return
	}
}

func removePINDistributeMsg(ctx context.Context, msgIDs []string) {
	// 1. 从 psql 中标记为已成功，定时任务会 24 小时清理
	if _, err := session.Database(ctx).Exec(ctx, `UPDATE distribute_messages SET status=2 WHERE message_id=ANY($1)`, msgIDs); err != nil {
		session.Logger(ctx).Println(err)
		return
	}
}

var distributeCols = []string{"client_id", "user_id", "origin_message_id", "message_id", "status"}

func createDistributeMsgList(ctx context.Context, insert [][]interface{}) error {
	var ident = pgx.Identifier{"distribute_messages"}
	if len(insert) == 0 {
		return nil
	}
	_, err := session.Database(ctx).CopyFrom(ctx, ident, distributeCols, pgx.CopyFromRows(insert))
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key") {
			session.Logger(ctx).Println(err)
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
		session.Logger(ctx).Println(err)
		return ""
	}
	return msg.MessageID
}

func getPinOriginMsgIDs(ctx context.Context, msgData string) (string, []string) {
	var msg MessagePinBody
	_ = json.Unmarshal(tools.Base64Decode(msgData), &msg)
	msgIDs := make([]string, 0, len(msg.MessageIDs))
	for _, msgID := range msg.MessageIDs {
		var m *DistributeMessage
		var err error
		if msg.Action == "PIN" {
			m, err = getDistributeMsgByMsgIDFromRedis(ctx, msgID)
		} else if msg.Action == "UNPIN" {
			m, err = getDistributeMsgByMsgIDFromPsql(ctx, msgID)
		}
		if err != nil {
			session.Logger(ctx).Println(err)
		}
		if m != nil && m.OriginMessageID != "" {
			msgIDs = append(msgIDs, m.OriginMessageID)
		}
	}
	return msg.Action, msgIDs
}

var ClientShardIDMap = make(map[string]map[string]string)

func InitShardID(ctx context.Context, clientID string) error {
	ClientShardIDMap[clientID] = make(map[string]string)
	// 1. 获取有多少个协程，就分配多少个编号
	count := decimal.NewFromInt(config.MessageShardSize)
	// 2. 获取优先级高/低的所有用户，及高低比例
	high, low, err := GetClientUserReceived(ctx, clientID)
	if err != nil {
		return err
	}
	// 每个分组的平均人数
	highCount := int(decimal.NewFromInt(int64(len(high))).Div(count).Ceil().IntPart())
	lowCount := int(decimal.NewFromInt(int64(len(low))).Div(count).Ceil().IntPart())
	// 3. 给这个大群里 每个用户进行 编号
	for shardID := 0; shardID < int(config.MessageShardSize); shardID++ {
		strShardID := strconv.Itoa(shardID)
		cutCount := 0
		hC := len(high)
		for i := 0; i < highCount; i++ {
			if i == hC {
				break
			}
			cutCount++
			ClientShardIDMap[clientID][high[i]] = strShardID
		}
		if cutCount > 0 {
			high = high[cutCount:]
		}

		cutCount = 0
		lC := len(low)
		for i := 0; i < lowCount; i++ {
			if i == lC {
				break
			}
			cutCount++
			ClientShardIDMap[clientID][low[i]] = strShardID
		}
		if cutCount > 0 {
			low = low[cutCount:]
		}
	}
	return nil
}
