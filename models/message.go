package models

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"mvdan.cc/xurls"
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
)

var openStatusMsgCategoryMap = map[int]map[string]bool{
	ClientUserStatusAudience: {
		mixin.MessageCategoryPlainText: true,
	},
	ClientUserStatusFresh: {
		mixin.MessageCategoryPlainText:    true,
		mixin.MessageCategoryPlainSticker: true,
	},
	ClientUserStatusSenior: {
		mixin.MessageCategoryPlainText:    true,
		mixin.MessageCategoryPlainSticker: true,
		mixin.MessageCategoryPlainImage:   true,
		mixin.MessageCategoryPlainVideo:   true,
		mixin.MessageCategoryPlainPost:    true,
		mixin.MessageCategoryPlainData:    true,
		mixin.MessageCategoryPlainLive:    true,
		"PLAIN_TRANSCRIPT":                true,
	},
	ClientUserStatusLarge: {
		mixin.MessageCategoryPlainText:    true,
		mixin.MessageCategoryPlainSticker: true,
		mixin.MessageCategoryPlainImage:   true,
		mixin.MessageCategoryPlainVideo:   true,
		mixin.MessageCategoryPlainPost:    true,
		mixin.MessageCategoryPlainData:    true,
		mixin.MessageCategoryPlainLive:    true,
		"PLAIN_TRANSCRIPT":                true,
	},
}

var ignoreCategoryMsg = map[string]bool{
	mixin.MessageCategoryPlainContact: true,
	mixin.MessageCategoryPlainAudio:   true,
}

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
	return err
}

func updateMessageStatus(ctx context.Context, clientID, messageID string, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE messages SET status=$3 WHERE client_id=$1 AND message_id=$2`, clientID, messageID, status)
	return err
}

var cacheSendJoinMsg = make(map[string]time.Time)

func checkIsSendJoinMsg(userID string) bool {
	if cacheSendJoinMsg[userID].IsZero() {
		cacheSendJoinMsg[userID] = time.Now()
		return false
	}
	if cacheSendJoinMsg[userID].Add(time.Minute * 5).Before(time.Now()) {
		cacheSendJoinMsg[userID] = time.Now()
		return false
	}
	return true
}

func checkIsJustJoinGroup(u *ClientUser) bool {
	return u.CreatedAt.Add(time.Minute * 5).After(time.Now())
}

func ReceivedMessage(ctx context.Context, clientID string, _msg mixin.MessageView) error {
	now := time.Now().UnixNano()
	msg := &_msg
	conversationStatus := getClientConversationStatus(ctx, clientID)
	if checkIsBlockUser(ctx, clientID, msg.UserID) {
		return nil
	}
	if msg.UserID == config.Config.LuckCoinAppID &&
		checkIsContact(ctx, clientID, msg.ConversationID) {
		if checkLuckCoinIsBlockUserOrMutedAndNotManager(ctx, clientID, msg.Data, conversationStatus) {
			return nil
		}
		if err := createAndDistributeMessage(ctx, clientID, msg); err != nil {
			return err
		}
		return nil
	}
	if msg.UserID == "b523c28b-1946-4b98-a131-e1520780e8af" &&
		msg.Category == mixin.MessageCategoryPlainLive &&
		checkIsContact(ctx, clientID, msg.ConversationID) {
		msg.UserID = clientID
		if err := createAndDistributeMessage(ctx, clientID, msg); err != nil {
			return err
		}
		return nil
	}

	clientUser, err := GetClientUserByClientIDAndUserID(ctx, clientID, msg.UserID)
	if errors.Is(err, pgx.ErrNoRows) || clientUser.Status == ClientUserStatusExit {
		if checkIsSendJoinMsg(msg.UserID) {
			return nil
		}
		go SendJoinMsg(clientID, msg.UserID)
		return nil
	} else if err != nil {
		return err
	}
	if clientUser.Priority == ClientUserPriorityStop {
		activeUser(clientUser)
		return nil
	}
	if msg.Category == mixin.MessageCategoryPlainText &&
		string(tools.Base64Decode(msg.Data)) == "/received_message" {
		return nil
	}
	go UpdateClientUserDeliverTime(_ctx, clientID, msg.MessageID, msg.CreatedAt, "READ")
	if checkIsLuckCoin(msg) {
		if err := createAndDistributeMessage(ctx, clientID, msg); err != nil {
			return err
		}
		return nil
	}
	if checkIsJustJoinGroup(clientUser) && checkIsIgnoreLeaveMsg(msg) {
		return nil
	}
	if checkIsMutedUser(clientUser) {
		return nil
	}

	// 查看该群组是否开启了持仓发言
	client, err := GetClientByID(ctx, clientID)
	if err != nil {
		return err
	}

	// 1. 查看该用户是否是管理员或嘉宾
	// 1. 是管理员或者是嘉宾
	switch clientUser.Status {
	case ClientUserStatusAudience:
		// 观众
		if client.SpeakStatus == ClientSpeckStatusOpen {
			// 不能发言
			if checkIsIgnoreLeaveMsg(msg) {
				return nil
			}
			go SendAssetsNotPassMsg(clientID, msg.UserID)
			if checkCanSpeak(ctx, clientID, msg.UserID, ClientUserStatusAudience, true) {
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
		if ignoreCategoryMsg[msg.Category] {
			go SendCategoryMsg(clientID, msg.UserID, msg.Category)
			return nil
		}
		if conversationStatus == ClientConversationStatusMute ||
			conversationStatus == ClientConversationStatusAudioLive {
			// 1. 给用户发一条禁言中...
			go SendClientMuteMsg(clientID, msg.UserID)
			return nil
		}
		if checkHasURLMsg(ctx, clientID, msg) {
			go handleURLMsg(clientID, msg, true)
			return nil
		}
		if checkStickerLimit(ctx, clientID, msg) {
			go muteClientUser(ctx, clientID, msg.UserID, "2")
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
			// 3. 检查是否是 recall/禁言/拉黑 别人 的消息
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
		isOpen := client.SpeakStatus == ClientSpeckStatusOpen
		if !checkCanSpeak(ctx, clientID, msg.UserID, clientUser.Status, isOpen) {
			// 达到限制
			go SendLimitMsg(clientID, msg.UserID, statusLimitMap[clientUser.Status])
			return nil
		}
		if !checkCategory(msg.Category, clientUser.Status, isOpen) {
			// 消息类型
			if isOpen {
				go SendCategoryMsg(clientID, msg.UserID, msg.Category)
			} else if msg.Category == mixin.MessageCategoryPlainImage ||
				msg.Category == mixin.MessageCategoryPlainVideo ||
				msg.Category == mixin.MessageCategoryPlainPost {
				// 转发给管理员
				go handleURLMsg(clientID, msg, false)
			}
			return nil
		}
		if conversationStatus == ClientConversationStatusAudioLive {
			go HandleAudioReplay(clientID, msg)
		}
		if err := createMessage(ctx, clientID, msg, MessageStatusPending); err != nil && !durable.CheckIsPKRepeatError(err) {
			session.Logger(ctx).Println(err)
			return err
		}
		if err := createFinishedDistributeMsg(ctx, &DistributeMessage{
			ClientID:         clientID,
			UserID:           msg.UserID,
			ConversationID:   msg.ConversationID,
			ShardID:          "0",
			OriginMessageID:  msg.MessageID,
			MessageID:        msg.MessageID,
			QuoteMessageID:   msg.QuoteMessageID,
			Data:             msg.Data,
			Category:         msg.Category,
			RepresentativeID: msg.RepresentativeID,
			CreatedAt:        msg.CreatedAt,
		}); err != nil && !durable.CheckIsPKRepeatError(err) {
			session.Logger(ctx).Println(err)
			return err
		}
	}
	tools.PrintTimeDuration(clientID+"ack 消息...", now)
	return nil
}

func checkCanSpeak(ctx context.Context, clientID, userID string, status int, isOpen bool) bool {
	count := 0
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT count(1) FROM messages 
WHERE client_id=$1 
AND user_id=$2 
AND now()-created_at<interval '1 minutes'
`, clientID, userID).Scan(&count); err != nil {
		return false
	}
	if isOpen {
		limit := statusLimitMap[status]
		return count < limit
	}
	var limit int
	if status == ClientUserStatusAdmin ||
		status == ClientUserStatusGuest {
		limit = statusLimitMap[status]
	} else {
		limit = config.NotOpenAssetsCheckMsgLimit
	}
	return count < limit
}

func checkCategory(category string, status int, isOpen bool) bool {
	if category == mixin.MessageCategoryMessageRecall ||
		status == ClientUserStatusAdmin ||
		status == ClientUserStatusGuest {
		return true
	}
	if isOpen {
		return openStatusMsgCategoryMap[status][category]
	}
	if category == mixin.MessageCategoryPlainText ||
		category == mixin.MessageCategoryPlainSticker {
		return true
	}
	return false
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

func checkHasURLMsg(ctx context.Context, clientID string, msg *mixin.MessageView) bool {
	hasURL := false
	if msg.Category == mixin.MessageCategoryPlainImage {
		if _hasURL, err := tools.MessageQRFilter(ctx, GetMixinClientByID(ctx, clientID).Client, msg); err == nil {
			hasURL = _hasURL
		} else {
			session.Logger(ctx).Println(err)
		}
	} else if msg.Category == mixin.MessageCategoryPlainText {
		data := tools.Base64Decode(msg.Data)
		if xurls.Relaxed.Match(data) {
			hasURL = true
		}
	}
	return hasURL
}

func checkStickerLimit(ctx context.Context, clientID string, msg *mixin.MessageView) bool {
	count := 0
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT count(1) FROM messages 
WHERE client_id=$1 AND user_id=$2 AND category=$3
AND now()-created_at<interval '5 seconds'
`, clientID, msg.UserID, mixin.MessageCategoryPlainSticker).Scan(&count); err != nil {
		session.Logger(ctx).Println(err)
		return false
	}
	if count == 2 {
		go SendStickerLimitMsg(clientID, msg.UserID)
	}
	return count >= 5
}

func checkIsLuckCoin(msg *mixin.MessageView) bool {
	if msg.Category == mixin.MessageCategoryAppCard {
		dataByte := tools.Base64Decode(msg.Data)
		var card mixin.AppCardMessage
		if err := json.Unmarshal(dataByte, &card); err != nil {
			return false
		}
		if card.AppID == config.Config.LuckCoinAppID {
			return true
		}
	}
	return false
}

func checkIsContact(ctx context.Context, clientID, conversationID string) bool {
	c, err := GetMixinClientByID(ctx, clientID).ReadConversation(ctx, conversationID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return false
	}
	if c.Category == mixin.ConversationCategoryContact {
		return true
	}
	return false
}

func checkLuckCoinIsBlockUserOrMutedAndNotManager(ctx context.Context, clientID, data, status string) bool {
	var m mixin.AppCardMessage
	err := json.Unmarshal(tools.Base64Decode(data), &m)
	if err != nil {
		session.Logger(ctx).Println(err)
		return true
	}
	u, err := url.Parse(m.Action)
	if err != nil {
		session.Logger(ctx).Println(err)
		return true
	}
	query, _ := url.ParseQuery(u.RawQuery)
	if len(query["uid"]) == 0 {
		return true
	}
	uid := query["uid"][0]
	if checkIsBlockUser(ctx, clientID, uid) {
		return true
	}

	if (status == ClientConversationStatusMute ||
		status == ClientConversationStatusAudioLive) &&
		!checkIsAdmin(ctx, clientID, uid) {
		return true
	}

	return false
}

var ignoreMsgList = []string{"Hi", "你好"}

func checkIsIgnoreLeaveMsg(msg *mixin.MessageView) bool {
	data := string(tools.Base64Decode(msg.Data))
	for _, s := range ignoreMsgList {
		if data == s {
			return true
		}
	}
	return false
}
