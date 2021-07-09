package models

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
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
  join_url           VARCHAR DEFAULT '', -- 入群前发送的url

  welcome            TEXT DEFAULT '', -- 入群时发送的内容

  limit_reject       TEXT DEFAULT '', -- 1分钟发言次数超过限额
  muted_reject       TEXT DEFAULT '', -- 被禁言

  category_reject    TEXT DEFAULT '', -- 类型 被拦截消息

  url_reject         TEXT DEFAULT '', -- 链接被拦截消息
  url_admin          TEXT DEFAULT '', -- 转发给管理员的url消息

  balance_reject     TEXT DEFAULT '', -- 不满足持仓，不能发言
  updated_at         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type ClientReplay struct {
	ClientID       string    `json:"client_id,omitempty"`
	JoinMsg        string    `json:"join_msg,omitempty"`
	JoinURL        string    `json:"join_url,omitempty"`
	Welcome        string    `json:"welcome,omitempty"`
	LimitReject    string    `json:"limit_reject,omitempty"`
	MutedReject    string    `json:"muted_reject,omitempty"`
	CategoryReject string    `json:"category_reject,omitempty"`
	URLReject      string    `json:"url_reject,omitempty"`
	URLAdmin       string    `json:"url_admin,omitempty"`
	BalanceReject  string    `json:"balance_reject,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

var _ctx context.Context

func updateClientWelcome(ctx context.Context, clientID, welcome string) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_replay SET welcome=$2 WHERE client_id=$1`, clientID, welcome)
	return err
}

func UpdateClientReplay(ctx context.Context, c *ClientReplay) error {
	query := durable.InsertQueryOrUpdate("client_replay", "client_id", "join_msg,join_url,welcome,limit_reject,category_reject,url_reject,url_admin,balance_reject,muted_reject,updated_at")
	_, err := session.Database(ctx).Exec(ctx, query, c.ClientID, c.JoinMsg, c.JoinURL, c.Welcome, c.LimitReject, c.CategoryReject, c.URLReject, c.URLAdmin, c.BalanceReject, c.MutedReject, time.Now())
	return err
}

var cacheClientReplay = make(map[string]ClientReplay)
var nilClientReplay = ClientReplay{}

func GetClientReplay(clientID string) (ClientReplay, error) {
	if cacheClientReplay[clientID] == nilClientReplay {
		var c ClientReplay
		if err := session.Database(_ctx).QueryRow(_ctx, `
		SELECT client_id,join_msg,join_url,welcome,limit_reject,category_reject,url_reject,url_admin,balance_reject,muted_reject,updated_at
		FROM client_replay WHERE client_id=$1
		`, clientID).Scan(&c.ClientID, &c.JoinMsg, &c.JoinURL, &c.Welcome, &c.LimitReject, &c.CategoryReject, &c.URLReject, &c.URLAdmin, &c.BalanceReject, &c.MutedReject, &c.UpdatedAt); err != nil {
			return ClientReplay{}, err
		}
		cacheClientReplay[clientID] = c
	}
	return cacheClientReplay[clientID], nil
}

func SendJoinMsg(clientID, userID string) {
	client, r, err := GetReplayAndMixinClientByClientID(clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := SendTextMsg(_ctx, client, userID, r.JoinMsg); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := SendBtnMsg(_ctx, client, userID, mixin.AppButtonGroupMessage{
		{config.Config.Text.Join, fmt.Sprintf("%s/auth", client.Host), "#5979F0"},
	}); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendCategoryMsg(clientID, userID, category string) {
	client, r, err := GetReplayAndMixinClientByClientID(clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	msg := strings.ReplaceAll(r.CategoryReject, "{category}", config.Config.Text.Category[category])
	if err := SendTextMsg(_ctx, client, userID, msg); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendWelcomeAndLatestMsg(clientID, userID string) {
	client, r, err := GetReplayAndMixinClientByClientID(clientID)
	if err != nil {
		return
	}
	if err := SendTextMsg(_ctx, client, userID, r.Welcome); err != nil {
		session.Logger(_ctx).Println(err)
	}
	btns := mixin.AppButtonGroupMessage{
		{config.Config.Text.Home, client.Host, "#5979F0"},
	}
	if client.AssetID != "" {
		btns = append(btns, mixin.AppButtonMessage{Label: config.Config.Text.Transfer, Action: fmt.Sprintf("%s/trade/%s", client.Host, client.AssetID), Color: "#8A64D0"})
	}
	if err := SendBtnMsg(_ctx, client, userID, btns); err != nil {
		session.Logger(_ctx).Println(err)
	}
	conversationStatus := getClientConversationStatus(_ctx, clientID)
	if conversationStatus == ClientConversationStatusNormal ||
		conversationStatus == ClientConversationStatusMute {
		go sendLatestMsg(client, userID, 20)
	} else if conversationStatus == ClientConversationStatusAudioLive {
		go sendLatestLiveMsg(client, userID)
	}
}

func sendLatestMsg(client *MixinClient, userID string, msgCount int) {
	ctx := _ctx
	c, err := GetClientUserByClientIDAndUserID(ctx, client.ClientID, userID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	_ = UpdateClientUserPriority(ctx, client.ClientID, userID, ClientUserPriorityPending)
	sendPendingMsg(ctx, client.ClientID, userID, msgCount)
	_ = UpdateClientUserPriority(ctx, client.ClientID, userID, c.Priority)
}

func sendLatestLiveMsg(client *MixinClient, userID string) {
	ctx := _ctx
	c, err := GetClientUserByClientIDAndUserID(ctx, client.ClientID, userID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	_ = UpdateClientUserPriority(ctx, client.ClientID, userID, ClientUserPriorityPending)
	sendPendingLiveMsg(ctx, client.ClientID, userID)
	_ = UpdateClientUserPriority(ctx, client.ClientID, userID, c.Priority)
}

func SendAssetsNotPassMsg(clientID, userID string) {
	client, r, err := GetReplayAndMixinClientByClientID(clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	l, err := GetClientAssetLevel(_ctx, clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	a, err := GetAssetByID(_ctx, client.Client, client.AssetID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	msg := r.BalanceReject
	msg = strings.ReplaceAll(msg, "{amount}", l.Fresh.String())
	msg = strings.ReplaceAll(msg, "{symbol}", a.Symbol)
	if err := SendTextMsg(_ctx, client, userID, msg); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := SendBtnMsg(_ctx, client, userID, mixin.AppButtonGroupMessage{
		{config.Config.Text.Auth, fmt.Sprintf("%s/auth", client.Host), "#5979F0"},
		{config.Config.Text.Transfer, fmt.Sprintf("%s/trade/%s", client.Host, client.AssetID), "#5979F0"},
	}); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendLimitMsg(clientID, userID string, limit int) {
	client, r, err := GetReplayAndMixinClientByClientID(clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	msg := strings.ReplaceAll(r.LimitReject, "{limit}", strconv.Itoa(limit))
	if err := SendTextMsg(_ctx, client, userID, msg); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendURLMsg(clientID, userID string) {
	client, r, err := GetReplayAndMixinClientByClientID(clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := SendTextMsg(_ctx, client, userID, r.URLReject); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendMutedMsg(clientID, userID string, mutedTime string, hour, minuted int) {
	client, r, err := GetReplayAndMixinClientByClientID(clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	msg := strings.ReplaceAll(r.MutedReject, "{muted_time}", mutedTime)
	msg = strings.ReplaceAll(msg, "{hours}", strconv.Itoa(hour))
	msg = strings.ReplaceAll(msg, "{minutes}", strconv.Itoa(minuted))
	if err := SendTextMsg(_ctx, client, userID, msg); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}
func SendClientMuteMsg(clientID, userID string) {
	client := GetMixinClientByID(_ctx, clientID)
	if err := SendTextMsg(_ctx, &client, userID, config.Config.Text.Muting); err != nil {
		session.Logger(_ctx).Println(err)
	}
}

func SendAuthSuccessMsg(clientID, userID string) {
	client := GetMixinClientByID(_ctx, clientID)
	if err := SendTextMsg(_ctx, &client, userID, config.Config.Text.AuthSuccess); err != nil {
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
			Category:         msg.Category,
			Data:             msg.Data,
			RepresentativeID: userID,
			QuoteMessageID:   quoteMsgIDMap[id],
		}
		if id == uid {
			msg.RepresentativeID = ""
		}
		msgList = append(msgList, msg)
	}
	client := GetMixinClientByID(_ctx, clientID)
	if client.ClientID == "" {
		return
	}
	_ = SendMessages(_ctx, client.Client, msgList)
}

// 处理 用户的 链接 或 二维码的消息
func handleURLMsg(clientID string, msg *mixin.MessageView) {
	// 1. 给用户发送 禁止的消息
	go SendURLMsg(clientID, msg.UserID)
	if err := createMessage(_ctx, clientID, msg, MessageStatusLeaveMessage); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	client, r, err := GetReplayAndMixinClientByClientID(clientID)
	if err != nil {
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
		originMsgID := tools.GetUUID()
		conversationID := mixin.UniqueConversationID(clientID, uid)
		oriMsg = append(oriMsg, &mixin.MessageRequest{
			ConversationID:   conversationID,
			RecipientID:      uid,
			MessageID:        originMsgID,
			Category:         msg.Category,
			Data:             msg.Data,
			RepresentativeID: msg.UserID,
		})
		quoteNoticeMsg = append(quoteNoticeMsg, &mixin.MessageRequest{
			ConversationID: conversationID,
			RecipientID:    uid,
			MessageID:      tools.GetUUID(),
			Category:       mixin.MessageCategoryPlainText,
			Data:           tools.Base64Encode([]byte(r.URLAdmin)),
			QuoteMessageID: originMsgID,
		})
		btnMsg = append(btnMsg, &mixin.MessageRequest{
			ConversationID: conversationID,
			RecipientID:    uid,
			MessageID:      tools.GetUUID(),
			Category:       mixin.MessageCategoryAppButtonGroup,
			Data: getBtnMsg([]mixin.AppButtonMessage{
				{Label: config.Config.Text.Forward, Action: fmt.Sprintf("input:---operation,%s,%s", "forward", msg.MessageID), Color: "#5979F0"},
				{Label: config.Config.Text.Mute, Action: fmt.Sprintf("input:---operation,%s,%s", "mute", msg.MessageID), Color: "#5979F0"},
				{Label: config.Config.Text.Block, Action: fmt.Sprintf("input:---operation,%s,%s", "block", msg.MessageID), Color: "#5979F0"},
			}),
		})
	}
	err = SendMessages(_ctx, client.Client, oriMsg)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	//   2.2. 发送 quote 原消息的 提醒消息
	err = SendMessages(_ctx, client.Client, quoteNoticeMsg)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	// 	 2.3. 发送 三个 btn
	err = SendMessages(_ctx, client.Client, btnMsg)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func SendClientTextMsg(clientID, msg, userID string, isJoinMsg bool) {
	if isJoinMsg && checkIsBlockUser(_ctx, clientID, userID) {
		return
	}
	client := GetMixinClientByID(_ctx, clientID).Client
	msgList := make([]*mixin.MessageRequest, 0)
	users, err := GetClientUserByPriority(_ctx, clientID, []int{ClientUserPriorityHigh, ClientUserPriorityLow}, isJoinMsg, false)
	if err != nil {
		session.Logger(_ctx).Println(err)
	}
	if len(users) <= 0 {
		return
	}
	msgBase64 := tools.Base64Encode([]byte(msg))
	var dataInsert [][]interface{}
	originMsgID := tools.GetUUID()
	if isJoinMsg {
		if err := createMessage(_ctx, clientID, &mixin.MessageView{
			ConversationID: mixin.UniqueConversationID(clientID, userID),
			UserID:         userID,
			MessageID:      originMsgID,
			Category:       mixin.MessageCategoryPlainText,
			Data:           msgBase64,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}, MessageStatusFinished); err != nil {
			session.Logger(_ctx).Println(err)
		}
	}

	for _, uid := range users {
		msgID := tools.GetUUID()
		if isJoinMsg {
			dataInsert = append(dataInsert,
				_createDistributeMessage(_ctx, clientID, uid, originMsgID, msgID, "", mixin.MessageCategoryPlainText, msgBase64, "", DistributeMessageLevelHigher, MessageStatusBroadcast, time.Now()))
		}
		msgList = append(msgList, &mixin.MessageRequest{
			ConversationID: mixin.UniqueConversationID(clientID, uid),
			RecipientID:    uid,
			MessageID:      msgID,
			Category:       mixin.MessageCategoryPlainText,
			Data:           msgBase64,
		})
	}
	if isJoinMsg {
		if err := createDistributeMsgList(_ctx, dataInsert); err != nil {
			session.Logger(_ctx).Println(err)
		}
	}

	if err := SendBatchMessages(_ctx, client, msgList); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
}

func GetReplayAndMixinClientByClientID(clientID string) (*MixinClient, *ClientReplay, error) {
	r, err := GetClientReplay(clientID)
	if err != nil {
		return nil, nil, err
	}
	client := GetMixinClientByID(_ctx, clientID)
	return &client, &r, nil
}

func SendTextMsg(ctx context.Context, client *MixinClient, userID, data string) error {
	conversationID := mixin.UniqueConversationID(client.ClientID, userID)
	if err := SendMessage(ctx, client.Client, &mixin.MessageRequest{
		ConversationID: conversationID,
		RecipientID:    userID,
		MessageID:      tools.GetUUID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           tools.Base64Encode([]byte(data)),
	}, false); err != nil {
		return err
	}
	return nil
}

func SendBtnMsg(ctx context.Context, client *MixinClient, userID string, data mixin.AppButtonGroupMessage) error {
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
	client := GetMixinClientByID(_ctx, clientID)
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

func sendPendingMsg(ctx context.Context, clientID, userID string, count int) {
	query := `SELECT user_id,message_id,category,data FROM messages WHERE client_id=$1 ORDER BY created_at DESC`
	if count != 0 {
		query = fmt.Sprintf(`%s LIMIT %d`, query, count)
	}
	lastCreatedAt, err := getLeftMsgAndDistribute(ctx, query, clientID, userID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	if lastCreatedAt.IsZero() {
		return
	}
	for {
		lastCreatedAt, err = sendLeftMsg(ctx, clientID, userID, lastCreatedAt)
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
		if lastCreatedAt.IsZero() {
			break
		}
	}
}
func sendPendingLiveMsg(ctx context.Context, clientID, userID string) {
	// 1. 获取直播的开始时间
	var lastCreatedAt time.Time
	err := session.Database(ctx).QueryRow(ctx, `
SELECT ld.start_at FROM live_data ld
LEFT JOIN lives l ON ld.live_id=l.live_id
WHERE l.status=1
`).Scan(&lastCreatedAt)
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	for {
		lastCreatedAt, err = sendLeftMsg(ctx, clientID, userID, lastCreatedAt)
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
		if lastCreatedAt.IsZero() {
			break
		}
	}

}

func sendLeftMsg(ctx context.Context, clientID, userID string, leftTime time.Time) (time.Time, error) {
	query := fmt.Sprintf(`SELECT user_id,message_id,category,data FROM messages WHERE client_id=$1 AND created_at>'%s' ORDER BY created_at DESC`, leftTime.Format("2006-01-02T15:04:05.000+08:00"))
	return getLeftMsgAndDistribute(ctx, query, clientID, userID)
}

func getLeftMsgAndDistribute(ctx context.Context, query, clientID, userID string) (time.Time, error) {
	msgs := make([]*mixin.MessageRequest, 0)
	dms := make([]*DistributeMessage, 0)
	conversationID := mixin.UniqueConversationID(clientID, userID)

	if err := session.Database(ctx).ConnQuery(ctx, query, func(rows pgx.Rows) error {
		for rows.Next() {
			var originMsgID string
			msgID := tools.GetUUID()
			msg := mixin.MessageRequest{
				ConversationID: conversationID,
				RecipientID:    userID,
				MessageID:      msgID,
			}
			if err := rows.Scan(&msg.RepresentativeID, &originMsgID, &msg.Category, &msg.Data); err != nil {
				return err
			}
			if msg.RecipientID == msg.RepresentativeID {
				continue
			}
			msgs = append([]*mixin.MessageRequest{&msg}, msgs...)
			dms = append([]*DistributeMessage{{
				ClientID:         clientID,
				UserID:           userID,
				ConversationID:   conversationID,
				ShardID:          "0",
				OriginMessageID:  originMsgID,
				MessageID:        msgID,
				Data:             msg.Data,
				Category:         msg.Category,
				RepresentativeID: msg.RepresentativeID,
				CreatedAt:        time.Now(),
			}}, dms...)
		}
		return nil
	}, clientID); err != nil {
		session.Logger(ctx).Println(err)
		return time.Time{}, err
	}
	if len(dms) == 0 {
		return time.Time{}, nil
	}
	client := GetMixinClientByID(ctx, clientID)
	// 存入成功之后再发送
	for i, dm := range dms {
		if err := createFinishedDistributeMsg(ctx, dm); err != nil {
			session.Logger(ctx).Println(err)
		}
		_ = SendMessage(ctx, client.Client, msgs[i], true)
	}

	return dms[len(dms)-1].CreatedAt, nil
}

func init() {
	_ctx = session.WithDatabase(context.Background(), durable.NewDatabase(context.Background()))
	_ctx = session.WithRedis(_ctx, durable.NewRedis(context.Background()))
}
