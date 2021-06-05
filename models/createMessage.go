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
func createAndDistributeMessage(ctx context.Context, clientID string, clientUserStatus int, msg *mixin.MessageView) error {
	// 1. 创建消息
	err := createMessage(ctx, clientID, msg, MessageStatusPending)
	if err != nil && !durable.CheckIsPKRepeatError(err) {
		session.Logger(ctx).Println(err)
		return err
	}
	// 2. 创建分发消息列表
	return CreateDistributeMsgAndMarkPrivilege(ctx, clientID, clientUserStatus, msg)
}

// 创建分发消息 标记 并把消息标记完成
func CreateDistributeMsgAndMarkPrivilege(ctx context.Context, clientID string, clientUserStatus int, msg *mixin.MessageView) error {
	if err := createDistributeMessageList(ctx, clientID, msg, clientUserStatus == ClientUserStatusManager); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	// 3. 标记消息为 privilege
	if err := updateMessageStatus(ctx, clientID, msg.MessageID, MessageStatusFinished); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	return nil
}

func createDistributeMessageList(ctx context.Context, clientID string, msg *mixin.MessageView, isManager bool) error {
	high, low, err := GetClientUserReceived(ctx, clientID, isManager)
	if err != nil {
		return err
	}
	recallMsgIDMap := make(map[string]string)
	if msg.Category == mixin.MessageCategoryMessageRecall {
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
			return err
		}
		if count == 0 {
			// 消息已经被删除...
			return nil
		}
	}

	for _, s := range high {
		if err := createDistributeMessageByLevel(ctx, clientID, s, msg, recallMsgIDMap, DistributeMessageLevelHigher); err != nil {
			session.Logger(ctx).Println(err)
		}
	}
	for _, s := range low {
		if err := createDistributeMessageByLevel(ctx, clientID, s, msg, recallMsgIDMap, DistributeMessageLevelLower); err != nil {
			session.Logger(ctx).Println(err)
		}
	}
	return nil
}

func createDistributeMessageByLevel(ctx context.Context, clientID, userID string, msg *mixin.MessageView, recallMsgIDMap map[string]string, status int) error {
	if userID == msg.UserID || userID == msg.RepresentativeID {
		return nil
	}
	if msg.Category == mixin.MessageCategoryMessageRecall {
		if recallMsgIDMap == nil {
			recallMsgIDMap = make(map[string]string)
			originMsgID := getRecallOriginMsgID(ctx, msg.Data)
			if originMsgID == "" {
				return nil
			}
			msgID := ""
			if err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT message_id
FROM distribute_messages
WHERE client_id=$1 AND origin_message_id=$2 AND user_id=$3
`, func(row pgx.Row) error {
				return row.Scan(&msgID)
			}, clientID, originMsgID, userID); err != nil {
				return err
			}
			recallMsgIDMap[userID] = msgID
		}
		if recallMsgIDMap[userID] == "" {
			return nil
		}
		data, err := json.Marshal(map[string]string{"message_id": recallMsgIDMap[userID]})
		if err != nil {
			return err
		}
		msg.QuoteMessageID = ""
		msg.UserID = ""
		msg.Data = tools.Base64Encode(data)
	}

	if err := createDistributeMessage(ctx, clientID, userID, msg, status); err == nil {
		return nil
	} else {
		if !durable.CheckIsPKRepeatError(err) {
			session.Logger(ctx).Println(err)
			return nil
		}
	}
	return nil
}

func createDistributeMessage(ctx context.Context, clientID, userID string, msg *mixin.MessageView, level int) error {
	if msg.QuoteMessageID == "" {
		return _createDistributeMessage(ctx, clientID, userID, msg.MessageID, tools.GetUUID(), "", msg.Category, msg.Data, msg.UserID, level, DistributeMessageStatusPending, time.Now())
	}
	// 1. 获取 originMessageID
	originMsgID, err := getOriginDistributeMsgID(ctx, clientID, msg.QuoteMessageID)
	if err != nil {
		session.Logger(ctx).Println(err)
	}
	var quoteMessageID string
	if originMsgID != "" {
		quoteMessageID, err = getDistributeMessageIDByOriginMsgID(ctx, clientID, userID, originMsgID)
		if err != nil {
			session.Logger(ctx).Println(err)
		}
	}
	return _createDistributeMessage(ctx, clientID, userID, msg.MessageID, tools.GetUUID(), quoteMessageID, msg.Category, msg.Data, msg.UserID, level, DistributeMessageStatusPending, time.Now())
}

var insertDistributeMessageQuery = durable.InsertQuery("distribute_messages", "client_id,user_id,shard_id,conversation_id,origin_message_id,message_id,quote_message_id,category,data,representative_id,level,status,created_at")

func _createDistributeMessage(ctx context.Context, clientID, userID, originMsgID, messageID, quoteMsgID, category, data, representativeID string, level, status int, createdAt time.Time) error {
	conversationID := mixin.UniqueConversationID(clientID, userID)
	shardID := tools.ShardId(conversationID, userID)
	_, err := session.Database(ctx).Exec(ctx, insertDistributeMessageQuery,
		clientID, userID, shardID, conversationID, originMsgID, messageID, quoteMsgID, category, data, representativeID, level, status, createdAt)
	return err
}

func getRecallOriginMsgID(ctx context.Context, msgData string) string {
	data := tools.Base64Decode(msgData)
	var msg struct{ MessageID string `json:"message_id"` }
	err := json.Unmarshal(data, &msg)
	if err != nil {
		session.Logger(ctx).Println(err)
		return ""
	}
	return msg.MessageID
}
