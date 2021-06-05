package models

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"mvdan.cc/xurls"
	"time"
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
}

const (
	MessageStatusPending      = 1
	MessageStatusPrivilege    = 2
	MessageStatusNormal       = 3
	MessageStatusFinished     = 4
	MessageStatusLeaveMessage = 5
)

var statusMsgCategoryMap = map[int]map[string]bool{
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
		mixin.MessageCategoryPlainPost:    true,
		mixin.MessageCategoryPlainLive:    true,
		mixin.MessageCategoryAppCard:      true,
		mixin.MessageCategoryPlainImage:   true,
		mixin.MessageCategoryPlainVideo:   true,
	},
	ClientUserStatusLarge: {
		mixin.MessageCategoryPlainText:    true,
		mixin.MessageCategoryPlainSticker: true,
		mixin.MessageCategoryPlainPost:    true,
		mixin.MessageCategoryPlainLive:    true,
		mixin.MessageCategoryAppCard:      true,
		mixin.MessageCategoryPlainImage:   true,
		mixin.MessageCategoryPlainVideo:   true,
	},
}

var statusLimitMap = map[int]int{
	ClientUserStatusAudience: 3,
	ClientUserStatusFresh:    10,
	ClientUserStatusSenior:   15,
	ClientUserStatusLarge:    20,
	ClientUserStatusManager:  30,
	ClientUserStatusGuest:    30,
}

func getMsgByClientIDAndMessageID(ctx context.Context, clientID, msgID string) (*Message, error) {
	var m Message
	err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT user_id, message_id,category,data,user_id,conversation_id FROM messages
WHERE client_id=$1 AND message_id=$2
`, func(row pgx.Row) error {
		return row.Scan(&m.UserID, &m.MessageID, &m.Category, &m.Data, &m.UserID, &m.ConversationID)
	}, clientID, msgID)
	return &m, err
}

func GetLongestMessageByStatus(ctx context.Context, clientID string, status int) (*Message, error) {
	var m Message
	err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT message_id, category, data, user_id, quote_message_id FROM messages 
WHERE client_id=$1 AND status=$2
ORDER BY created_at ASC LIMIT 1
`, func(row pgx.Row) error {
		return row.Scan(&m.MessageID, &m.Category, &m.Data, &m.UserID, &m.QuoteMessageID)
	}, clientID, status)
	return &m, err
}

// 1. 存入 message 表中, 标记为 `pending`
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

func ReceivedMessage(ctx context.Context, clientID string, _msg mixin.MessageView) error {
	now := time.Now().UnixNano()
	msg := &_msg
	if isBlock := checkIsBlockUser(ctx, clientID, msg.UserID); isBlock {
		return nil
	}
	clientUser, err := GetClientUserByClientIDAndUserID(ctx, clientID, msg.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			go SendJoinMsg(clientID, msg.UserID)
			return nil
		}
		return err
	}
	if checkIsLuckCoin(msg) {
		if err := createAndDistributeMessage(ctx, clientID, clientUser.Status, msg); err != nil {
			return err
		}
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
			if ignore := checkIsIgnoreLeaveMsg(msg); !ignore {
				go SendAssetsNotPassMsg(clientID, msg.UserID)
				// 将留言消息发给管理员
				go SendToClientManager(clientID, msg)
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
		hasURL := checkHasURLMsg(ctx, clientID, msg)
		if hasURL {
			go handleURLMsg(clientID, msg)
			return nil
		}
		fallthrough
	// 管理员
	case ClientUserStatusManager:
		// 1. 如果是管理员的消息，则检查 quote 的消息是否为留言的消息
		if clientUser.Status == ClientUserStatusManager {
			if ok, err := checkIsQuoteLeaveMessage(ctx, clientUser, msg); err != nil {
				session.Logger(ctx).Println(err)
			} else if ok {
				return nil
			}
			// 2. 检查 是否是 帮转/禁言/拉黑 的消息
			if isOperation, err := checkIsOperation(ctx, clientID, msg); err != nil {
				session.Logger(ctx).Println(err)
			} else if isOperation {
				return nil
			}
			// 3. 检查是否是 recall 别人 的消息
			notThrough, err := checkIsRecallMsg(ctx, clientID, msg)
			if err != nil {
				session.Logger(ctx).Println(err)
			}
			if notThrough {
				return nil
			}
		}
		fallthrough
	// 嘉宾
	case ClientUserStatusGuest:
		if client.SpeakStatus == ClientSpeckStatusOpen {
			canSpeak, err := checkCanSpeak(ctx, clientID, msg.UserID, clientUser.Status)
			if err != nil {
				return err
			}
			if !canSpeak {
				// 达到限制
				go SendLimitMsg(clientID, msg.UserID, statusLimitMap[clientUser.Status])
				return nil
			}
			canCategory := checkCategory(msg.Category, clientUser.Status)
			if !canCategory {
				// 消息类型
				go SendCategoryMsg(clientID, msg.UserID, msg.Category)
				return nil
			}
		}

		err := createMessage(ctx, clientID, msg, MessageStatusPending)
		if err != nil && !durable.CheckIsPKRepeatError(err) {
			session.Logger(ctx).Println(err)
			return err
		}
	}
	tools.PrintTimeDuration(fmt.Sprintf("处理消息 %s ", msg.MessageID), now)
	return nil
}

func checkCanSpeak(ctx context.Context, clientID, userID string, status int) (bool, error) {
	limit := statusLimitMap[status]
	count := 0
	err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT count(1) FROM messages 
WHERE client_id=$1 
AND user_id=$2 
AND now()-created_at<interval '1 minutes'
`, func(row pgx.Row) error {
		return row.Scan(&count)
	}, clientID, userID)
	return count < limit, err
}

func checkCategory(category string, status int) bool {
	if category == mixin.MessageCategoryMessageRecall ||
		status == ClientUserStatusManager ||
		status == ClientUserStatusGuest {
		return true
	}
	return statusMsgCategoryMap[status][category]
}

var cacheClientIDLastMsgMap = make(map[string]Message)
var nilClientMsgMap = Message{}

func getLastMsgByClientID(ctx context.Context, clientID string) (Message, error) {
	if cacheClientIDLastMsgMap[clientID] == nilClientMsgMap {
		var msg Message
		if err := session.Database(ctx).ConnQueryRow(ctx,
			`SELECT created_at FROM messages WHERE client_id=$1 ORDER BY created_at DESC LIMIT 1`,
			func(row pgx.Row) error {
				return row.Scan(&msg.CreatedAt)
			}, clientID); err != nil {
			return Message{}, err
		}
		cacheClientIDLastMsgMap[clientID] = msg
	}
	return cacheClientIDLastMsgMap[clientID], nil
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

func checkIsLuckCoin(msg *mixin.MessageView) bool {
	if msg.Category == mixin.MessageCategoryAppCard {
		dataByte := tools.Base64Decode(msg.Data)
		var card mixin.AppCardMessage
		if err := json.Unmarshal(dataByte, &card); err != nil {
			return false
		}
		if card.AppID == config.LuckCoinAppID {
			return true
		}
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
