package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
)

const messages_DDL = `
-- 消息
CREATE TABLE IF NOT EXISTS messages (
  client_id           VARCHAR(36) NOT NULL,
  user_id             VARCHAR(36) NOT NULL,
  conversation_id     VARCHAR(36) NOT NULL,
  message_id          VARCHAR(36) NOT NULL,
  quote_message_id    VARCHAR(36) NOT NULL DEFAULT '',
  category            VARCHAR,
  data                TEXT,
  status              SMALLINT NOT NULL, -- 1 pending 2 privilege 3 normal 4 finished
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
  PRIMARY KEY(client_id, message_id)
);
`

type Message struct {
	ClientID       string    `json:"client_id,omitempty"`
	UserID         string    `json:"user_id,omitempty"`
	ConversationID string    `json:"conversation_id,omitempty"`
	MessageID      string    `json:"message_id,omitempty"`
	QuoteMessageID string    `json:"quote_message_id,omitempty"`
	Category       string    `json:"category,omitempty"`
	Data           string    `json:"data,omitempty"`
	Status         int       `json:"status,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`

	FullName  string `json:"full_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`

	TopAt time.Time `json:"top_at,omitempty"`
}

const (
	MessageStatusPending      = 1
	MessageStatusPrivilege    = 2
	MessageStatusNormal       = 3 // 其他的消息
	MessageStatusFinished     = 4
	MessageStatusLeaveMessage = 5
	MessageStatusBroadcast    = 6
	MessageStatusJoinMsg      = 7
	MessageStatusRecallMsg    = 8
	MessageStatusClientMsg    = 9  // 客户端发送的消息
	MessageStatusPINMsg       = 10 // PIN 消息
)

var statusLimitMap = map[int]int{
	ClientUserStatusAudience: 5,
	ClientUserStatusFresh:    10,
	ClientUserStatusSenior:   15,
	ClientUserStatusLarge:    20,
	ClientUserStatusAdmin:    30,
	ClientUserStatusGuest:    30,
}

func getMsgByClientIDAndMessageID(ctx context.Context, clientID, msgID string) (*Message, error) {
	var m Message
	err := session.Database(ctx).QueryRow(ctx, `
SELECT user_id,message_id,category,data,conversation_id,created_at FROM messages
WHERE client_id=$1 AND message_id=$2
`, clientID, msgID).Scan(&m.UserID, &m.MessageID, &m.Category, &m.Data, &m.ConversationID, &m.CreatedAt)
	return &m, err
}

func GetLongestMessageByStatus(ctx context.Context, clientID string, status int) (*Message, error) {
	var m Message
	err := session.Database(ctx).QueryRow(ctx, `
SELECT message_id, category, data, user_id, quote_message_id FROM messages 
WHERE client_id=$1 AND status=$2
ORDER BY created_at ASC LIMIT 1
`, clientID, status).Scan(&m.MessageID, &m.Category, &m.Data, &m.UserID, &m.QuoteMessageID)
	return &m, err
}

// 1. 存入 message 表中
func createMessage(ctx context.Context, clientID string, msg *mixin.MessageView, status int) error {
	query := `INSERT INTO messages(client_id,user_id,conversation_id,message_id,category,data,quote_message_id,status,created_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := session.Database(ctx).Exec(ctx, query,
		clientID, msg.UserID, msg.ConversationID, msg.MessageID, msg.Category, msg.Data, msg.QuoteMessageID, status, msg.CreatedAt)
	if status == MessageStatusPending {
		go session.Redis(_ctx).QPublish(_ctx, "create", clientID)
	}
	return err
}

func updateMessageStatus(ctx context.Context, clientID, messageID string, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE messages SET status=$3 WHERE client_id=$1 AND message_id=$2`, clientID, messageID, status)
	return err
}

func ReceivedMessage(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	now := time.Now().UnixNano()
	if msg.Category == "MESSAGE_PIN" {
		return nil
	}
	// 检查是否是黑名单用户
	if checkIsBlockUser(ctx, clientID, msg.UserID) {
		return nil
	}
	conversationStatus := getClientConversationStatus(ctx, clientID)
	// 检查是红包的话单独处理
	if msg.UserID == config.Config.LuckCoinAppID &&
		checkIsContact(ctx, clientID, msg.ConversationID) {
		if checkCanNotSendLuckyCoin(ctx, clientID, msg.Data, conversationStatus) {
			return nil
		}
		if err := createAndDistributeMessage(ctx, clientID, msg); err != nil {
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
		if err := createAndDistributeMessage(ctx, clientID, msg); err != nil {
			return err
		}
		return nil
	}
	clientUser, err := GetClientUserByClientIDAndUserID(ctx, clientID, msg.UserID)
	// 检测是不是新用户
	if errors.Is(err, pgx.ErrNoRows) || clientUser.Status == ClientUserStatusExit {
		if checkIsSendJoinMsg(msg.UserID) {
			return nil
		}
		go SendJoinMsg(clientID, msg.UserID)
		return nil
	} else if err != nil {
		return err
	}
	// 如果是失活用户, 激活一下
	if clientUser.Priority == ClientUserPriorityStop {
		activeUser(clientUser)
	}
	// 检测一下是不是激活指令
	if (msg.Category == mixin.MessageCategoryPlainText || msg.Category == "ENCRYPTED_TEXT") &&
		string(tools.Base64Decode(msg.Data)) == "/received_message" {
		return nil
	}
	// 更新一下用户最后已读时间
	go UpdateClientUserDeliverTime(_ctx, clientID, msg.MessageID, msg.CreatedAt, "READ")
	// 检查是不是刚入群发的 Hi 你好 消息
	if checkIsJustJoinGroup(clientUser) && checkIsIgnoreLeaveMsg(msg) {
		return nil
	}
	// 检查是不是禁言用户的的消息
	if checkIsMutedUser(clientUser) {
		return nil
	}
	// 查看该群组是否开启了持仓发言
	client, err := GetClientByID(ctx, clientID)
	if err != nil {
		return err
	}
	// 查看该用户是否是管理员或嘉宾
	switch clientUser.Status {
	case ClientUserStatusAudience:
		// 观众
		if client.SpeakStatus == ClientSpeckStatusOpen {
			// 不能发言
			if checkIsIgnoreLeaveMsg(msg) {
				return nil
			}
			go SendAssetsNotPassMsg(clientID, msg.UserID, msg.MessageID, false)
			if checkMessageCountLimit(ctx, clientID, msg.UserID, ClientUserStatusAudience) {
				go SendToClientManager(clientID, msg, true, true)
			}
			return nil
		}
		fallthrough
	// 入门
	case ClientUserStatusFresh:
		fallthrough
	// 资深
	case ClientUserStatusSenior:
		fallthrough
	// 大户
	case ClientUserStatusLarge:

		if checkMsgIsForbid(clientUser, msg) {
			return nil
		}

		// 检查语言是否符合大群
		if checkMsgLanguage(msg) {
			go rejectMsgAndDeliverManagerWithOperationBtns(clientID, msg, config.Text.LanguageReject, config.Text.LanguageAdmin)
			return nil
		}
		// 检查这个社群状态是否是禁言中
		if conversationStatus == ClientConversationStatusMute ||
			conversationStatus == ClientConversationStatusAudioLive {
			// 给用户发一条禁言中...
			go SendClientMuteMsg(clientID, msg.UserID)
			return nil
		}
		// 检测是否含有链接
		if !checkHasClientMemberAuth(ctx, clientID, "url", clientUser.Status) &&
			checkHasURLMsg(ctx, clientID, msg) {
			var rejectMsg string
			if msg.Category == mixin.MessageCategoryPlainText ||
				msg.Category == "ENCRYPTED_TEXT" {
				rejectMsg = config.Text.URLReject
			} else if msg.Category == mixin.MessageCategoryPlainImage || msg.Category == "ENCRYPTED_IMAGE" {
				rejectMsg = config.Text.QrcodeReject
			}
			go rejectMsgAndDeliverManagerWithOperationBtns(
				clientID,
				msg,
				rejectMsg,
				config.Text.URLAdmin,
			)
			return nil
		}
		// 检测最近5s是否发了多个 sticker
		if checkStickerLimit(ctx, clientID, msg) {
			go muteClientUser(ctx, clientID, msg.UserID, "2")
			return nil
		}
		if msg.Category == "MESSAGE_PIN" {
			return nil
		}
		fallthrough
	// 管理员
	case ClientUserStatusAdmin:
		// 1. 如果是管理员的消息，则检查 quote 的消息是否为留言的消息
		if clientUser.Status == ClientUserStatusAdmin {
			if ok, err := checkIsQuoteLeaveMessage(ctx, clientUser, msg); err != nil {
				session.Logger(ctx).Println(err)
			} else if ok {
				return nil
			}
			// 2. 检查 是否是 帮转/禁言/拉黑 的按钮消息
			if isOperation, err := checkIsButtonOperation(ctx, clientID, msg); err != nil {
				session.Logger(ctx).Println(err)
			} else if isOperation {
				return nil
			}
			// 3. 检查是否是 recall/禁言/拉黑/info 别人 的消息
			// 4. 检测是否是 mute open mute close 的消息
			isOperationMsg, err := checkIsOperationMsg(ctx, clientID, msg)
			if err != nil {
				session.Logger(ctx).Println(err)
			}
			if isOperationMsg {
				return nil
			}
		}
		fallthrough
	// 嘉宾
	case ClientUserStatusGuest:
		if !checkMessageCountLimit(ctx, clientID, msg.UserID, clientUser.Status) {
			// 达到限制
			go SendLimitMsg(clientID, msg.UserID, statusLimitMap[clientUser.Status])
			return nil
		}
		// 检测是否是需要忽略的消息类型
		if !checkCategory(ctx, clientID, msg.Category, clientUser.Status) {
			go SendCategoryMsg(clientID, msg.UserID, msg.Category, clientUser.Status)
			return nil
		}
		if conversationStatus == ClientConversationStatusAudioLive {
			go HandleAudioReplay(clientID, msg)
		}
		msg.Data = tools.SafeBase64Encode(msg.Data)
		if config.Config.Encrypted && strings.HasPrefix(msg.Category, "ENCRYPTED_") {
			msg.Data, err = decryptMessageData(msg.Data, &client)
			if err != nil {
				session.Logger(ctx).Println(err)
				return nil
			}
		}
		if err := createMessage(ctx, clientID, msg, MessageStatusPending); err != nil && !durable.CheckIsPKRepeatError(err) {
			session.Logger(ctx).Println(err)
			return err
		}
		if err := createFinishedDistributeMsg(ctx, clientID, msg.UserID, msg.MessageID, msg.ConversationID, "0", msg.MessageID, msg.QuoteMessageID, msg.CreatedAt); err != nil && !durable.CheckIsPKRepeatError(err) {
			session.Logger(ctx).Println(err)
			return err
		}
	}
	tools.PrintTimeDuration(clientID+"ack 消息...", now)
	return nil
}

func GetClientLastMsg(ctx context.Context) (map[string]time.Time, error) {
	clients, err := getAllClient(ctx)
	if err != nil {
		return nil, err
	}
	lms := make(map[string]time.Time)
	for _, client := range clients {
		var lm time.Time
		if err := session.Database(ctx).QueryRow(ctx,
			`SELECT created_at FROM messages WHERE client_id=$1 ORDER BY created_at DESC LIMIT 1`, client).Scan(&lm); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return nil, err
			}
			lm = time.Now()
		}
		lms[client] = lm
	}
	return lms, nil
}
