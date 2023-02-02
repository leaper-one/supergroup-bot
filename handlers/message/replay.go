package message

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

// 处理 用户的 留言消息
func handleLeaveMsg(clientID, userID, originMsgID string, msg *mixin.MessageView) {
	ctx := models.Ctx
	forwardList, err := getClientManager(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return
	}
	msgList := make([]*mixin.MessageRequest, 0)
	// 组织管理员的消息
	quoteMsgIDMap, uid, err := GetDistributeMessageIDMapByOriginMsgID(ctx, clientID, originMsgID)
	if err != nil {
		tools.Println(err)
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
	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}
	common.SendMessages(client.Client, msgList)
}
func sendClientMuteMsg(clientID, userID string) {
	common.SendClientUserTextMsg(clientID, userID, config.Text.Muting, "")
}

// 处理 用户的 链接 或 二维码的消息
func rejectMsgAndDeliverManagerWithOperationBtns(clientID string, msg *mixin.MessageView, sendToReceiver, sendToManager string) {
	ctx := models.Ctx
	// 1. 给用户发送 禁止的消息
	if sendToReceiver != "" {
		go common.SendClientUserTextMsg(clientID, msg.UserID, sendToReceiver, "")
	}
	if err := common.CreateMessage(ctx, clientID, msg, models.MessageRedisStatusFinished); err != nil {
		tools.Println(err)
		return
	}
	// 2. 给管理员发送检测到的消息
	managers, err := getClientManager(ctx, clientID)
	if err != nil {
		tools.Println(err)
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
	c, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}
	client := c.Client
	err = common.SendMessages(client, oriMsg)
	if err != nil {
		tools.Println(err)
		return
	}
	//   2.2. 发送 quote 原消息的 提醒消息
	err = common.SendMessages(client, quoteNoticeMsg)
	if err != nil {
		tools.Println(err)
		return
	}
	// 	 2.3. 发送 三个 btn
	err = common.SendMessages(client, btnMsg)
	if err != nil {
		tools.Println(err)
		return
	}
}

func getBtnMsg(data mixin.AppButtonGroupMessage) string {
	btnData, err := json.Marshal(data)
	if err != nil {
		tools.Println(err)
		return ""
	}
	return tools.Base64Encode(btnData)
}

func sendForbidMsg(clientID, userID, category string) {
	msg := strings.ReplaceAll(
		config.Text.Forbid,
		"{category}",
		config.Text.Category[category],
	)
	common.SendClientUserTextMsg(clientID, userID, msg, "")
}
func sendAssetsNotPassMsg(clientID, userID, quoteMsgID string, isJoin bool) {
	ctx := models.Ctx
	if isJoin {
		common.SendClientUserTextMsg(clientID, userID, config.Text.JoinMsgInfo, "")
	} else {
		u, err := getClientAdmin(ctx, clientID)
		if err != nil {
			return
		}
		msg := strings.ReplaceAll(config.Text.BalanceReject, "{admin_name}", u.FullName)
		common.SendClientUserTextMsg(clientID, userID, msg, quoteMsgID)
	}
	sendMemberCentreBtn(clientID, userID)
}

func sendMutedMsg(clientID, userID string, mutedTime string, hour, minuted int) {
	msg := strings.ReplaceAll(config.Text.MutedReject, "{muted_time}", mutedTime)
	msg = strings.ReplaceAll(msg, "{hours}", strconv.Itoa(hour))
	msg = strings.ReplaceAll(msg, "{minutes}", strconv.Itoa(minuted))
	common.SendClientUserTextMsg(clientID, userID, msg, "")
}
func sendJoinMsg(clientID, userID string) {
	ctx := models.Ctx
	c, err := common.GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		tools.Println(err)
		return
	}
	common.SendClientUserTextMsg(clientID, userID, c.JoinMsg, "")
	if err := common.SendBtnMsg(ctx, clientID, userID, mixin.AppButtonGroupMessage{
		{Label: config.Text.Join, Action: fmt.Sprintf("%s/auth", c.Host), Color: "#5979F0"},
	}); err != nil {
		tools.Println(err)
		log.Println(clientID, userID)
		return
	}
}

func getClientAdmin(ctx context.Context, clientId string) (*models.User, error) {
	c, err := common.GetClientByIDOrHost(ctx, clientId)
	if err != nil {
		return nil, err
	}
	adminId := c.AdminID
	if adminId == "" {
		adminId = c.OwnerID
	}
	return common.SearchUser(ctx, clientId, adminId)
}

func sendMemberCentreBtn(clientID, userID string) {
	ctx := models.Ctx
	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return
	}
	if err := common.SendBtnMsg(ctx, clientID, userID, mixin.AppButtonGroupMessage{
		{Label: config.Text.MemberCentre, Action: fmt.Sprintf("%s/member", client.C.Host), Color: "#5979F0"},
	}); err != nil {
		tools.Println(err)
		return
	}
}

func getPlainCategory(category string) string {
	if strings.HasPrefix(category, "ENCRYPTED_") {
		category = strings.Replace(category, "ENCRYPTED_", "PLAIN_", 1)
	}
	return category
}

var statusLimitMap = map[int]int64{
	models.ClientUserStatusAudience: 5,
	models.ClientUserStatusFresh:    10,
	models.ClientUserStatusSenior:   15,
	models.ClientUserStatusLarge:    20,
	models.ClientUserStatusAdmin:    30,
	models.ClientUserStatusGuest:    30,
}

func sendLimitMsg(clientID, userID string, limit int64) {
	msg := strings.ReplaceAll(config.Text.LimitReject, "{limit}", strconv.Itoa(int(limit)))
	if limit < statusLimitMap[models.ClientUserStatusGuest] {
		msg += config.Text.MemberTips
	}
	common.SendClientUserTextMsg(clientID, userID, msg, "")
}

func sendCategoryMsg(clientID, userID, category string, status int) {
	if strings.HasPrefix(category, "ENCRYPTED_") {
		category = strings.Replace(category, "ENCRYPTED_", "PLAIN_", 1)
	}
	msg := strings.ReplaceAll(config.Text.CategoryReject, "{category}", config.Text.Category[category])
	isFreshMember := status < models.ClientUserStatusLarge
	if isFreshMember {
		msg += config.Text.MemberTips
	}
	common.SendClientUserTextMsg(clientID, userID, msg, "")
	if isFreshMember {
		sendMemberCentreBtn(clientID, userID)
	}
}
