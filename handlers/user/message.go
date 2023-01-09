package user

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

func SendWelcomeAndLatestMsg(clientID, userID string) {
	ctx := models.Ctx
	c, err := common.GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}
	common.SendClientUserTextMsg(clientID, userID, c.Welcome, "")
	btns := mixin.AppButtonGroupMessage{
		{Label: config.Text.Home, Action: c.Host, Color: "#5979F0"},
	}
	if c.AssetID != "" {
		btns = append(btns, mixin.AppButtonMessage{Label: config.Text.Transfer, Action: fmt.Sprintf("%s/trade/%s", c.Host, c.AssetID), Color: "#8A64D0"})
	}
	if err := common.SendBtnMsg(ctx, clientID, userID, btns); err != nil {
		tools.Println(err)
	}
	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}
	conversationStatus := common.GetClientConversationStatus(ctx, clientID)
	if conversationStatus == "" ||
		conversationStatus == models.ClientConversationStatusNormal ||
		conversationStatus == models.ClientConversationStatusMute {
		go sendLatestMsgAndPINMsg(client, userID, 20)
	} else if conversationStatus == models.ClientConversationStatusAudioLive {
		go sendLatestLiveMsg(client, userID)
	}
}

func sendLatestMsgAndPINMsg(client *common.MixinClient, userID string, msgCount int) {
	ctx := models.Ctx
	c, err := common.GetClientUserByClientIDAndUserID(ctx, client.ClientID, userID)
	if err != nil {
		tools.Println(err)
		return
	}
	_ = common.UpdateClientUserPart(ctx, client.ClientID, userID, map[string]interface{}{"priority": models.ClientUserPriorityPending})
	sendPendingMsgByCount(ctx, client.ClientID, userID, msgCount)
	_ = common.UpdateClientUserPart(ctx, client.ClientID, userID, map[string]interface{}{"priority": c.Priority})
	common.SendAssetsNotPassMsg(client.ClientID, userID, "", true)
}

func sendLatestLiveMsg(client *common.MixinClient, userID string) {
	ctx := models.Ctx
	c, err := common.GetClientUserByClientIDAndUserID(ctx, client.ClientID, userID)
	if err != nil {
		tools.Println(err)
		return
	}
	_ = common.UpdateClientUserPart(ctx, client.ClientID, userID, map[string]interface{}{"priority": models.ClientUserPriorityPending})
	// 1. 获取直播的开始时间
	var startAt time.Time
	if err := session.DB(ctx).Table("live_data ld").
		Select("ld.start_at").
		Joins("LEFT JOIN lives l ON ld.live_id=l.live_id").
		Where("l.status=1").
		Scan(&startAt).Error; err != nil {
		tools.Println(err)
		return
	}

	sendLeftMsgToNow(ctx, client.ClientID, userID, startAt)
	_ = common.UpdateClientUserPart(ctx, client.ClientID, userID, map[string]interface{}{"priority": c.Priority})
}

func sendPendingMsgByCount(ctx context.Context, clientID, uid string, count int) {
	var msgs []*models.Message
	if err := session.DB(ctx).Raw(`
(SELECT user_id,message_id,category,data,status,created_at
		FROM messages
		WHERE client_id=?
		AND status IN (4,6)
		AND category!='MESSAGE_RECALL'
			AND category!='MESSAGE_PIN'
		AND created_at > CURRENT_DATE-1
		ORDER BY created_at DESC
		LIMIT ?) UNION (SELECT user_id,message_id,category,data,status,created_at
		FROM messages
		WHERE client_id=?
		AND status=10
		AND category!='MESSAGE_PIN'
		ORDER BY created_at DESC) order by created_at desc
`, clientID, count, clientID).Scan(&msgs).Error; err != nil {
		tools.Println(err)
		return
	}
	lastCreatedAt, err := distributeMsg(ctx, msgs, clientID, uid)
	if err != nil {
		tools.Println(err)
		return
	}
	sendLeftMsgToNow(ctx, clientID, uid, lastCreatedAt)
}

func sendLeftMsgToNow(ctx context.Context, clientID, uid string, startTime time.Time) {
	lastCreatedAt := startTime
	var err error
	for {
		if lastCreatedAt.IsZero() {
			break
		}
		lastCreatedAt, err = sendLeftMsg(ctx, clientID, uid, lastCreatedAt)
		if err != nil {
			tools.Println(err)
			return
		}
	}
}

func sendLeftMsg(ctx context.Context, clientID, userID string, leftTime time.Time) (time.Time, error) {
	var msgs []*models.Message
	if err := session.DB(ctx).Raw(`
SELECT user_id,message_id,category,data,status,created_at
FROM messages
WHERE client_id=?
AND created_at>?
AND status IN (4,6)
AND category!='MESSAGE_RECALL'
AND category!='MESSAGE_PIN'
ORDER BY created_at DESC
`, clientID, leftTime).Scan(&msgs).Error; err != nil {
		tools.Println(err)
		return time.Time{}, err
	}
	return distributeMsg(ctx, msgs, clientID, userID)
}

func distributeMsg(ctx context.Context, msgList []*models.Message, clientID, userID string) (time.Time, error) {
	if len(msgList) == 0 {
		return time.Time{}, nil
	}
	msgs := make([]*mixin.MessageRequest, 0)
	conversationID := mixin.UniqueConversationID(clientID, userID)
	for _, message := range msgList {
		if userID == message.UserID {
			continue
		}
		msgID := tools.GetUUID()
		msgs = append([]*mixin.MessageRequest{{
			ConversationID:   conversationID,
			RecipientID:      userID,
			MessageID:        msgID,
			Category:         common.GetPlainCategory(message.Category),
			Data:             message.Data,
			RepresentativeID: message.UserID,
		}}, msgs...)
		if message.Status == models.MessageStatusPINMsg {
			dataByte, _ := json.Marshal(map[string]interface{}{"message_ids": []string{msgID}, "action": "PIN"})
			dataStr := tools.Base64Encode(dataByte)
			msgs = append(msgs, &mixin.MessageRequest{
				ConversationID: conversationID,
				RecipientID:    userID,
				MessageID:      mixin.UniqueConversationID(message.UserID, message.MessageID),
				Category:       "MESSAGE_PIN",
				Data:           dataStr,
			})
		}
	}
	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return time.Time{}, err
	}
	// 存入成功之后再发送
	for _, m := range msgs {
		if err := common.CreateDistributeMsgToRedis(ctx, []*models.DistributeMessage{{
			ClientID:        clientID,
			UserID:          userID,
			OriginMessageID: m.MessageID,
			ConversationID:  m.ConversationID,
			ShardID:         "0",
			MessageID:       m.MessageID,
			QuoteMessageID:  "",
			CreatedAt:       time.Now(),
		}}); err != nil {
			tools.Println(err)
			continue
		}

		_ = common.SendMessage(ctx, client.Client, m, true)
	}
	return msgList[0].CreatedAt, nil
}
