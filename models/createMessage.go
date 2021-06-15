package models

import (
	"context"
	"encoding/json"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"time"
)

// 创建消息 和 分发消息列表
func createAndDistributeMessage(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	// 1. 创建消息
	err := createMessage(ctx, clientID, msg, MessageStatusNormal)
	if err != nil && !durable.CheckIsPKRepeatError(err) {
		session.Logger(ctx).Println(err)
		return err
	}
	// 2. 创建分发消息列表
	return CreateDistributeMsgAndMarkStatus(ctx, clientID, msg, []int{ClientUserPriorityHigh})
}

// 创建分发消息 标记 并把消息标记
func CreateDistributeMsgAndMarkStatus(ctx context.Context, clientID string, msg *mixin.MessageView, priorityList []int) error {
	userList, err := GetClientUserByPriority(ctx, clientID, priorityList, false, false)
	if err != nil {
		return err
	}
	recallMsgIDMap := make(map[string]string)
	level := priorityList[0]
	var status int
	if level == ClientUserPriorityHigh {
		status = MessageStatusPrivilege
	} else if level == ClientUserPriorityLow {
		status = MessageStatusFinished
	}
	if msg.Category == mixin.MessageCategoryMessageRecall {
		recallMsgIDMap, err = getOriginMsgIDMap(ctx, clientID, msg)
		if err != nil {
			return err
		}
		if recallMsgIDMap == nil {
			if err := updateMessageStatus(ctx, clientID, msg.MessageID, status); err != nil {
				session.Logger(ctx).Println(err)
				return err
			}
			return nil
		}
	}
	// 创建消息
	var dataToInsert [][]interface{}
	quoteMessageIDMap := make(map[string]string)
	if msg.QuoteMessageID != "" {
		originMsgID, err := getOriginDistributeMsgID(ctx, clientID, msg.QuoteMessageID)
		if err != nil {
			session.Logger(ctx).Println(err)
		}
		if originMsgID != "" {
			quoteMessageIDMap, err = getDistributeMessageIDMapByOriginMsgID(ctx, clientID, originMsgID)
			if err != nil {
				session.Logger(ctx).Println(err)
			}
		}
	}
	for _, s := range userList {
		if s == msg.UserID || s == msg.RepresentativeID || checkIsBlockUser(ctx, clientID, s) {
			continue
		}
		if msg.Category == mixin.MessageCategoryMessageRecall {
			data, err := json.Marshal(map[string]string{"message_id": recallMsgIDMap[s]})
			if err != nil {
				return err
			}
			msg.QuoteMessageID = ""
			msg.UserID = ""
			msg.Data = tools.Base64Encode(data)
		}
		row := _createDistributeMessage(ctx, clientID, s, msg.MessageID, tools.GetUUID(), quoteMessageIDMap[s], msg.Category, msg.Data, msg.UserID, level, DistributeMessageStatusPending, time.Now())
		dataToInsert = append(dataToInsert, row)
	}
	if err := createDistributeMsgList(ctx, dataToInsert); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	// 3. 标记消息为 privilege
	if err := updateMessageStatus(ctx, clientID, msg.MessageID, status); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	return nil
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

func createDistributeMsgList(ctx context.Context, insert [][]interface{}) error {
	var ident = pgx.Identifier{"distribute_messages"}
	var cols = []string{"client_id", "user_id", "shard_id", "conversation_id", "origin_message_id", "message_id", "quote_message_id", "category", "data", "representative_id", "level", "status", "created_at"}
	_, err := session.Database(ctx).CopyFrom(ctx, ident, cols, pgx.CopyFromRows(insert))
	if err != nil {
		session.Logger(ctx).Println(err)
	}
	return nil
}

func getOriginMsgIDMap(ctx context.Context, clientID string, msg *mixin.MessageView) (map[string]string, error) {
	recallMsgIDMap := make(map[string]string)
	originMsgID := getRecallOriginMsgID(ctx, msg.Data)
	var count int
	if err := session.Database(ctx).ConnQuery(ctx, `
SELECT message_id, user_id
FROM distribute_messages
WHERE client_id=$1 AND origin_message_id=$2`, func(rows pgx.Rows) error {
		for rows.Next() {
			var msgID, userID string
			if err := rows.Scan(&msgID, &userID); err != nil {
				return err
			}
			recallMsgIDMap[userID] = msgID
			count++
		}
		return nil
	}, clientID, originMsgID); err != nil {
		return nil, err
	}
	if count == 0 {
		// 消息已经被删除...
		return nil, nil
	}
	return recallMsgIDMap, nil

}

func _createDistributeMessage(ctx context.Context, clientID, userID, originMsgID, messageID, quoteMsgID, category, data, representativeID string, level, status int, createdAt time.Time) []interface{} {
	conversationID := mixin.UniqueConversationID(clientID, userID)
	shardID := tools.ShardId(conversationID, userID)
	var row []interface{}
	row = append(row, clientID)
	row = append(row, userID)
	row = append(row, shardID)
	row = append(row, conversationID)
	row = append(row, originMsgID)
	row = append(row, messageID)
	row = append(row, quoteMsgID)
	row = append(row, category)
	row = append(row, data)
	row = append(row, representativeID)
	row = append(row, level)
	row = append(row, status)
	row = append(row, createdAt)
	return row
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
