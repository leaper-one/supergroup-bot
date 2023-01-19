package message

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"gorm.io/gorm"
)

func ReceivedMessage(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	now := time.Now()
	// 检查是否是黑名单用户
	if common.CheckIsBlockUser(ctx, clientID, msg.UserID) {
		return nil
	}
	conversationStatus := common.GetClientConversationStatus(ctx, clientID)
	// 查看该群组是否开启了持仓发言
	client, err := common.GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		return err
	}
	msg.Data = tools.SafeBase64Encode(msg.Data)
	if config.Config.Encrypted && strings.HasPrefix(msg.Category, "ENCRYPTED_") {
		msg.Data, err = decryptMessageData(msg.Data, &client)
		if err != nil {
			tools.Println(err)
			return nil
		}
	}
	if msg.UserID == config.Config.LuckCoinAppID &&
		checkIsContact(ctx, clientID, msg.ConversationID) {
		if checkCanNotSendLuckyCoin(ctx, clientID, msg.Data, conversationStatus) {
			return nil
		}
		if err := createPendingMessage(ctx, clientID, msg); err != nil {
			return err
		}
		return nil
	}
	// 检查是直播卡片消息单独处理
	if msg.UserID == "b523c28b-1946-4b98-a131-e1520780e8af" &&
		(msg.Category == mixin.MessageCategoryPlainLive ||
			msg.Category == "ENCRYPTED_LIVE") &&
		checkIsContact(ctx, clientID, msg.ConversationID) {
		msg.UserID = clientID
		if err := createPendingMessage(ctx, clientID, msg); err != nil {
			return err
		}
		return nil
	}
	clientUser, err := common.GetClientUserByClientIDAndUserID(ctx, clientID, msg.UserID)
	// 检测是不是新用户
	if errors.Is(err, gorm.ErrRecordNotFound) || clientUser.Status == models.ClientUserStatusExit {
		if checkIsSendJoinMsg(msg.UserID) {
			return nil
		}
		go sendJoinMsg(clientID, msg.UserID)
		return nil
	} else if err != nil {
		return err
	}
	// 如果是失活用户, 激活一下
	if clientUser.Priority == models.ClientUserPriorityStop {
		ActiveUser(&clientUser)
	}
	// 检测一下是不是激活指令
	if (msg.Category == mixin.MessageCategoryPlainText || msg.Category == "ENCRYPTED_TEXT") &&
		string(tools.Base64Decode(msg.Data)) == "/received_message" {
		return nil
	}
	// 更新一下用户最后已读时间
	go UpdateClientUserActiveTimeToRedis(clientID, msg.MessageID, msg.CreatedAt, "READ")
	// 检查是不是刚入群发的 Hi 你好 消息
	if checkIsJustJoinGroup(&clientUser) && checkIsIgnoreLeaveMsg(msg) {
		return nil
	}
	// 检查是不是禁言用户的的消息
	if checkIsMutedUser(&clientUser) {
		return nil
	}
	// 查看该用户是否是管理员或嘉宾
	switch clientUser.Status {
	case models.ClientUserStatusAudience:
		// 观众
		if client.SpeakStatus == models.ClientSpeckStatusOpen {
			// 不能发言
			if checkIsIgnoreLeaveMsg(msg) {
				return nil
			}
			go sendAssetsNotPassMsg(clientID, msg.UserID, msg.MessageID, false)
			if checkMessageCountLimit(ctx, clientID, msg.UserID, models.ClientUserStatusAudience) {
				go common.SendToClientManager(clientID, msg, true, true)
			}
			return nil
		}
		fallthrough
	// 入门
	case models.ClientUserStatusFresh:
		fallthrough
	// 资深
	case models.ClientUserStatusSenior:
		fallthrough
	// 大户
	case models.ClientUserStatusLarge:
		if checkMsgIsForbid(&clientUser, msg) {
			return nil
		}
		// 检查语言是否符合大群
		if checkMsgLanguage(msg, clientID) {
			go rejectMsgAndDeliverManagerWithOperationBtns(clientID, msg, config.Text.LanguageReject, config.Text.LanguageAdmin)
			return nil
		}
		// 检查这个社群状态是否是禁言中
		if conversationStatus == models.ClientConversationStatusMute ||
			conversationStatus == models.ClientConversationStatusAudioLive {
			// 给用户发一条禁言中...
			go sendClientMuteMsg(clientID, msg.UserID)
			return nil
		}
		// 检测是否含有链接
		if !common.CheckHasClientMemberAuth(ctx, clientID, "url", clientUser.Status) &&
			checkHasURLMsg(ctx, clientID, msg) {
			var rejectMsg string
			admin, err := getClientAdmin(ctx, clientID)
			if err != nil {
				return err
			}
			if msg.Category == mixin.MessageCategoryPlainText ||
				msg.Category == "ENCRYPTED_TEXT" {
				client, err := common.GetClientByIDOrHost(ctx, clientID)
				if err != nil {
					return err
				}
				rejectMsg = strings.ReplaceAll(config.Text.URLReject, "{group_name}", client.Name)
				rejectMsg = strings.ReplaceAll(rejectMsg, "{admin_name}", admin.FullName)
			} else if msg.Category == mixin.MessageCategoryPlainImage || msg.Category == "ENCRYPTED_IMAGE" {
				rejectMsg = strings.ReplaceAll(config.Text.QrcodeReject, "{group_name}", client.Name)
				rejectMsg = strings.ReplaceAll(rejectMsg, "{admin_name}", admin.FullName)
			}
			go rejectMsgAndDeliverManagerWithOperationBtns(clientID, msg, rejectMsg, config.Text.URLAdmin)
			return nil
		}
		// 检测最近5s是否发了多个 sticker
		if checkStickerLimit(ctx, clientID, msg) {
			go common.MuteClientUser(ctx, clientID, msg.UserID, "2")
			return nil
		}
		if msg.Category == "MESSAGE_PIN" {
			return nil
		}
		fallthrough
	// 管理员
	case models.ClientUserStatusAdmin:
		// 1. 如果是管理员的消息，则检查 quote 的消息是否为留言的消息
		if clientUser.Status == models.ClientUserStatusAdmin {
			if ok, err := checkIsQuoteLeaveMessage(ctx, &clientUser, msg); err != nil {
				tools.Println(err)
			} else if ok {
				return nil
			}
			// 2. 检查 是否是 帮转/禁言/拉黑 的按钮消息
			isOperation, err := checkIsButtonOperation(ctx, clientID, msg)
			if err != nil {
				tools.Println(err)
			}
			if isOperation {
				return nil
			}
			// 3. 检查是否是 recall/禁言/拉黑/info 别人 的消息
			// 4. 检测是否是 mute open mute close 的消息
			isOperationMsg, err := checkIsOperationMsg(ctx, &clientUser, msg)
			if err != nil {
				tools.Println(err)
			}
			if isOperationMsg {
				return nil
			}
		}
		fallthrough
	// 嘉宾
	case models.ClientUserStatusGuest:
		if !checkMessageCountLimit(ctx, clientID, msg.UserID, clientUser.Status) {
			// 达到限制
			go sendLimitMsg(clientID, msg.UserID, statusLimitMap[clientUser.Status])
			return nil
		}
		// 检测是否是需要忽略的消息类型
		if !checkCategory(ctx, clientID, msg.Category, clientUser.Status) {
			go sendCategoryMsg(clientID, msg.UserID, msg.Category, clientUser.Status)
			return nil
		}
		if conversationStatus == models.ClientConversationStatusAudioLive {
			go HandleAudioReplay(clientID, msg)
		}
		if err := createPendingMessage(ctx, clientID, msg); err != nil {
			return err
		}
	}
	tools.PrintTimeDuration(clientID+"ack 消息...", now)
	return nil
}
