package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

const client_replay_DDL = `
-- 自动回复
CREATE TABLE IF NOT EXISTS client_replay (
  client_id          VARCHAR(36) NOT NULL PRIMARY KEY,
  join_msg           TEXT DEFAULT '', -- 入群前的内容
  welcome            TEXT DEFAULT '', -- 入群时发送的内容
  updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type ClientReplay struct {
	ClientID  string    `json:"client_id,omitempty"`
	JoinMsg   string    `json:"join_msg,omitempty"`
	Welcome   string    `json:"welcome,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func updateClientWelcome(ctx context.Context, clientID, welcome string) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_replay SET welcome=$2 WHERE client_id=$1`, clientID, welcome)
	return err
}

func UpdateClientReplay(ctx context.Context, c *ClientReplay) error {
	query := durable.InsertQueryOrUpdate("client_replay", "client_id", "join_msg,welcome,updated_at")
	_, err := session.Database(ctx).Exec(ctx, query, c.ClientID, c.JoinMsg, c.Welcome, time.Now())
	return err
}

func SendJoinMsg(clientID, userID string) {
	c, err := GetClientByIDOrHost(_ctx, clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := SendClientUserTextMsg(_ctx, clientID, userID, c.JoinMsg, ""); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := SendBtnMsg(_ctx, clientID, userID, mixin.AppButtonGroupMessage{
		{Label: config.Text.Join, Action: fmt.Sprintf("%s/auth", c.Host), Color: "#5979F0"},
	}); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendStickerLimitMsg(clientID, userID string) {
	if err := SendClientUserTextMsg(_ctx, clientID, userID, config.Text.StickerWarning, ""); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendCategoryMsg(clientID, userID, category string, status int) {
	if strings.HasPrefix(category, "ENCRYPTED_") {
		category = strings.Replace(category, "ENCRYPTED_", "PLAIN_", 1)
	}
	msg := strings.ReplaceAll(config.Text.CategoryReject, "{category}", config.Text.Category[category])
	isFreshMember := status < ClientUserStatusLarge
	if isFreshMember {
		msg += config.Text.MemberTips
	}
	if err := SendClientUserTextMsg(_ctx, clientID, userID, msg, ""); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if isFreshMember {
		sendMemberCentreBtn(clientID, userID)
	}
}

func SendAssetsNotPassMsg(clientID, userID, quoteMsgID string, isJoin bool) {
	if isJoin {
		if err := SendClientUserTextMsg(_ctx, clientID, userID, config.Text.JoinMsgInfo, ""); err != nil {
			session.Logger(_ctx).Println(err)
			return
		}
	} else {
		u, err := getClientAdmin(_ctx, clientID)
		if err != nil {
			return
		}
		msg := strings.ReplaceAll(config.Text.BalanceReject, "{admin_name}", u.FullName)
		if err := SendClientUserTextMsg(_ctx, clientID, userID, msg, quoteMsgID); err != nil {
			session.Logger(_ctx).Println(err)
			return
		}
	}
	sendMemberCentreBtn(clientID, userID)
}

func SendLimitMsg(clientID, userID string, limit int) {
	msg := strings.ReplaceAll(config.Text.LimitReject, "{limit}", strconv.Itoa(limit))
	if limit < statusLimitMap[ClientUserStatusGuest] {
		msg += config.Text.MemberTips
	}
	if err := SendClientUserTextMsg(_ctx, clientID, userID, msg, ""); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendStopMsg(clientID, userID string) {
	client, err := GetMixinClientByIDOrHost(_ctx, clientID)
	if err != nil {
		return
	}
	if err := SendClientUserTextMsg(_ctx, clientID, userID, config.Text.StopMessage, ""); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := SendBtnMsg(_ctx, clientID, userID, mixin.AppButtonGroupMessage{
		{Label: config.Text.StopClose, Action: "input:/received_message", Color: "#5979F0"},
		{Label: config.Text.StopBroadcast, Action: fmt.Sprintf("%s/news", client.C.Host), Color: "#5979F0"},
	}); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendMutedMsg(clientID, userID string, mutedTime string, hour, minuted int) {
	msg := strings.ReplaceAll(config.Text.MutedReject, "{muted_time}", mutedTime)
	msg = strings.ReplaceAll(msg, "{hours}", strconv.Itoa(hour))
	msg = strings.ReplaceAll(msg, "{minutes}", strconv.Itoa(minuted))
	if err := SendClientUserTextMsg(_ctx, clientID, userID, msg, ""); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendClientMuteMsg(clientID, userID string) {
	if err := SendClientUserTextMsg(_ctx, clientID, userID, config.Text.Muting, ""); err != nil {
		session.Logger(_ctx).Println(err)
	}
}

func SendAuthSuccessMsg(clientID, userID string) {
	if err := SendClientUserTextMsg(_ctx, clientID, userID, config.Text.AuthSuccess, ""); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendForbidMsg(clientID, userID, category string) {
	msg := strings.ReplaceAll(
		config.Text.Forbid,
		"{category}",
		config.Text.Category[category],
	)
	SendClientUserTextMsg(_ctx, clientID, userID, msg, "")
}

func sendMemberCentreBtn(clientID, userID string) {
	client, err := GetMixinClientByIDOrHost(_ctx, clientID)
	if err != nil {
		return
	}
	if err := SendBtnMsg(_ctx, clientID, userID, mixin.AppButtonGroupMessage{
		{Label: config.Text.MemberCentre, Action: fmt.Sprintf("%s/member", client.C.Host), Color: "#5979F0"},
	}); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

// 处理 用户的 留言消息
func handleLeaveMsg(clientID, userID, originMsgID string, msg *mixin.MessageView) {
	forwardList, err := getClientManager(_ctx, clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	msgList := make([]*mixin.MessageRequest, 0)
	// 组织管理员的消息
	quoteMsgIDMap, uid, err := getDistributeMessageIDMapByOriginMsgID(_ctx, clientID, originMsgID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if uid != "" {
		forwardList = append(forwardList, uid)
	}
	for _, id := range forwardList {
		if id == userID || id == "" {
			continue
		}
		msg := &mixin.MessageRequest{
			ConversationID:   mixin.UniqueConversationID(clientID, id),
			RecipientID:      id,
			MessageID:        tools.GetUUID(),
			Category:         getPlainCategory(msg.Category),
			Data:             msg.Data,
			RepresentativeID: userID,
			QuoteMessageID:   quoteMsgIDMap[id],
		}
		if id == uid {
			msg.RepresentativeID = ""
		}
		msgList = append(msgList, msg)
	}
	client, err := GetMixinClientByIDOrHost(_ctx, clientID)
	if err != nil {
		return
	}
	_ = SendMessages(_ctx, client.Client, msgList)
}

// 处理 用户的 链接 或 二维码的消息
func rejectMsgAndDeliverManagerWithOperationBtns(clientID string, msg *mixin.MessageView, sendToReceiver, sendToManager string) {
	// 1. 给用户发送 禁止的消息
	if sendToReceiver != "" {
		go SendClientUserTextMsg(_ctx, clientID, msg.UserID, sendToReceiver, "")
	}
	if err := createMessage(_ctx, clientID, msg, MessageStatusLeaveMessage); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	// 2. 给管理员发送检测到的消息
	managers, err := getClientManager(_ctx, clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	oriMsg := make([]*mixin.MessageRequest, 0)
	quoteNoticeMsg := make([]*mixin.MessageRequest, 0)
	btnMsg := make([]*mixin.MessageRequest, 0)

	//   2.1. 发送原消息
	for _, uid := range managers {
		originMsgID := mixin.UniqueConversationID(msg.MessageID, uid)
		conversationID := mixin.UniqueConversationID(clientID, uid)
		category := msg.Category
		if strings.HasPrefix(category, "ENCRYPTED_") {
			category = strings.Replace(category, "ENCRYPTED_", "PLAIN_", 1)
		}
		oriMsg = append(oriMsg, &mixin.MessageRequest{
			ConversationID:   conversationID,
			RecipientID:      uid,
			MessageID:        originMsgID,
			Category:         category,
			Data:             tools.SafeBase64Encode(msg.Data),
			RepresentativeID: msg.UserID,
		})
		if sendToManager != "" {
			quoteNoticeMsg = append(quoteNoticeMsg, &mixin.MessageRequest{
				ConversationID: conversationID,
				RecipientID:    uid,
				MessageID:      tools.GetUUID(),
				Category:       mixin.MessageCategoryPlainText,
				Data:           tools.Base64Encode([]byte(sendToManager)),
				QuoteMessageID: originMsgID,
			})
		}
		btnMsg = append(btnMsg, &mixin.MessageRequest{
			ConversationID: conversationID,
			RecipientID:    uid,
			MessageID:      tools.GetUUID(),
			Category:       mixin.MessageCategoryAppButtonGroup,
			Data: getBtnMsg([]mixin.AppButtonMessage{
				{Label: config.Text.Forward, Action: fmt.Sprintf("input:---operation,%s,%s", "forward", msg.MessageID), Color: "#5979F0"},
				{Label: config.Text.Mute, Action: fmt.Sprintf("input:---operation,%s,%s", "mute", msg.MessageID), Color: "#5979F0"},
				{Label: config.Text.Block, Action: fmt.Sprintf("input:---operation,%s,%s", "block", msg.MessageID), Color: "#5979F0"},
			}),
		})
	}
	c, err := GetMixinClientByIDOrHost(_ctx, clientID)
	if err != nil {
		return
	}
	client := c.Client
	err = SendMessages(_ctx, client, oriMsg)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	//   2.2. 发送 quote 原消息的 提醒消息
	err = SendMessages(_ctx, client, quoteNoticeMsg)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	// 	 2.3. 发送 三个 btn
	err = SendMessages(_ctx, client, btnMsg)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

// 给客户端的每一个人发送一条消息，userID表示代表发送的用户，可以为空。
func SendClientTextMsg(clientID, msg, userID string, isJoinMsg bool) {
	if isJoinMsg && CheckIsBlockUser(_ctx, clientID, userID) {
		return
	}
	c, err := GetMixinClientByIDOrHost(_ctx, clientID)
	if err != nil {
		return
	}
	msgList := make([]*mixin.MessageRequest, 0)
	users, err := GetDistributeMsgUser(_ctx, clientID, isJoinMsg, false)
	if err != nil {
		session.Logger(_ctx).Println(err)
	}
	if len(users) <= 0 {
		return
	}
	msgBase64 := tools.Base64Encode([]byte(msg))
	originMsgID := tools.GetUUID()
	if isJoinMsg {
		if err := createMessage(_ctx, clientID, &mixin.MessageView{
			ConversationID: mixin.UniqueConversationID(clientID, userID),
			UserID:         userID,
			MessageID:      originMsgID,
			Category:       mixin.MessageCategoryPlainText,
			Data:           msgBase64,
		}, MessageStatusJoinMsg); err != nil {
			session.Logger(_ctx).Println(err)
		}
	}
	dms := make([]*DistributeMessage, 0, len(users))
	for _, u := range users {
		msgID := tools.GetUUID()
		if isJoinMsg {
			dms = append(dms, &DistributeMessage{
				ClientID:        clientID,
				UserID:          u.UserID,
				OriginMessageID: originMsgID,
				MessageID:       msgID,
				Category:        mixin.MessageCategoryPlainText,
				Level:           DistributeMessageLevelHigher,
				Status:          DistributeMessageStatusFinished,
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
		if err := createDistributeMsgToRedis(_ctx, dms); err != nil {
			session.Logger(_ctx).Println(err)
		}
	}

	if err := SendBatchMessages(_ctx, c.Client, msgList); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

// 给社群里的每个人发送一条普通消息
func SendClientMsg(clientID, category, data string) {
	c, err := GetMixinClientByIDOrHost(_ctx, clientID)
	if err != nil {
		return
	}
	msgList := make([]*mixin.MessageRequest, 0)
	users, err := GetDistributeMsgUser(_ctx, clientID, false, false)
	if err != nil {
		session.Logger(_ctx).Println(err)
	}
	if len(users) <= 0 {
		return
	}
	originMsgID := tools.GetUUID()
	if err := createMessage(_ctx, clientID, &mixin.MessageView{
		MessageID: originMsgID,
		Category:  category,
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, MessageStatusClientMsg); err != nil {
		session.Logger(_ctx).Println(err)
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

	if err := SendBatchMessages(_ctx, c.Client, msgList); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

// 指定大群给指定用户发送一条文本消息
func SendClientUserTextMsg(ctx context.Context, clientID, userID, data, quoteMsgID string) error {
	if data == "" {
		return nil
	}
	client, err := GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return errors.New("client is nil")
	}

	representativeID := ""

	if data != config.Text.AuthSuccess {
		admin, err := getClientAdmin(ctx, clientID)
		if err != nil {
			return err
		}

		if representativeID != userID {
			representativeID = admin.UserID
		}
	}

	conversationID := mixin.UniqueConversationID(client.ClientID, userID)
	if err := SendMessage(ctx, client.Client, &mixin.MessageRequest{
		ConversationID:   conversationID,
		RecipientID:      userID,
		MessageID:        tools.GetUUID(),
		Category:         mixin.MessageCategoryPlainText,
		Data:             tools.Base64Encode([]byte(data)),
		QuoteMessageID:   quoteMsgID,
		RepresentativeID: representativeID,
	}, false); err != nil {
		return err
	}
	return nil
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
		session.Logger(_ctx).Println(err)
		return ""
	}
	return tools.Base64Encode(btnData)
}

func SendRecallMsg(clientID string, msg *mixin.MessageView) {
	client, err := GetMixinClientByIDOrHost(_ctx, clientID)
	if err != nil {
		return
	}
	data, _ := json.Marshal(map[string]string{"message_id": msg.QuoteMessageID})

	if err := SendMessage(_ctx, client.Client, &mixin.MessageRequest{
		ConversationID: msg.ConversationID,
		RecipientID:    msg.UserID,
		MessageID:      tools.GetUUID(),
		Category:       mixin.MessageCategoryMessageRecall,
		Data:           tools.Base64Encode(data),
	}, false); err != nil {
		session.Logger(_ctx).Println(err)
	}
}

func getPlainCategory(category string) string {
	if strings.HasPrefix(category, "ENCRYPTED_") {
		category = strings.Replace(category, "ENCRYPTED_", "PLAIN_", 1)
	}
	return category
}
