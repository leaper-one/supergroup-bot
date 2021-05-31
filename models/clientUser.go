package models

import (
	"context"
	"errors"
	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"log"
	"time"
)

const bot_user_DDL = `
-- 机器人用户信息表
-- 机器人用户信息表
CREATE TABLE IF NOT EXISTS client_users (
  client_id          VARCHAR(36),
  user_id            VARCHAR(36),
  access_token       VARCHAR(512),
  priority           SMALLINT NOT NULL DEFAULT 2, -- 1 优先级高 2 优先级低 3 补发中 4 超时没接受消息 暂停发送
  is_async           BOOLEAN NOT NULL DEFAULT true,
  status             SMALLINT NOT NULL DEFAULT 0, -- 0 未入群 1 观众 2 入门 3 资深 5 大户 8 嘉宾 9 管理
  muted_time         VARCHAR DEFAULT '',
  muted_at           TIMESTAMP WITH TIME ZONE default '1970-01-01 00:00:00+00',
  deliver_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  PRIMARY KEY (client_id, user_id)
);
CREATE INDEX client_user_idx ON client_users(client_id);
CREATE INDEX client_user_send_idx ON transfers(client_id, can_send);
`

type ClientUser struct {
	ClientID    string    `json:"client_id,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	AccessToken string    `json:"access_token,omitempty"`
	Priority    int       `json:"priority,omitempty"`
	IsAsync     bool      `json:"is_async,omitempty"`
	Status      int       `json:"status,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	MutedTime   string    `json:"muted_time,omitempty"`
	MutedAt     time.Time `json:"muted_at,omitempty"`
	DeliverAt   time.Time `json:"deliver_at,omitempty"`

	AssetID     string `json:"asset_id,omitempty"`
	SpeakStatus int    `json:"speak_status,omitempty"`
}

const (
	ClientUserPriorityHigh    = 1 // 高优先级
	ClientUserPriorityLow     = 2 // 低优先级
	ClientUserPriorityPending = 3 // 补发中
	ClientUserPriorityStop    = 4 // 暂停发送

	//ClientUserStatusNoAuth   = 0 // 未入群
	ClientUserStatusAudience = 1 // 观众
	ClientUserStatusFresh    = 2 // 入门
	ClientUserStatusSenior   = 3 // 资深
	ClientUserStatusLarge    = 5 // 大户
	ClientUserStatusGuest    = 8 // 嘉宾
	ClientUserStatusManager  = 9 // 管理员
)

func UpdateClientUser(ctx context.Context, user *ClientUser, fullName string) error {
	u, err := GetClientUserByClientIDAndUserID(ctx, user.ClientID, user.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 第一次入群
			go SendSomePeopleJoinGroupMsg(user.ClientID, user.UserID, fullName)
			go SendWelcomeMsg(user.ClientID, user.UserID)
		}
	}
	go SendAuthSuccessMsg(user.ClientID, user.UserID)
	if u.Status == ClientUserStatusManager {
		user.Status = ClientUserStatusManager
	}
	query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,is_async,status")
	res, err := session.Database(ctx).Exec(ctx, query, user.ClientID, user.UserID, user.AccessToken, user.Priority, user.IsAsync, user.Status)
	log.Println(res)
	return err
}

func UpdateClientUserIsAsync(ctx context.Context, clientID, userID string, isAsync bool) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET is_async=$3 WHERE client_id=$1 AND user_id=$2`, clientID, userID, isAsync)
	return err
}

func GetClientUserByClientIDAndUserID(ctx context.Context, clientID, userID string) (*ClientUser, error) {
	var b ClientUser
	err := session.Database(ctx).ConnQueryRow(ctx, `
SELECT client_id,user_id,priority,access_token,status,created_at FROM client_users WHERE client_id=$1 AND user_id=$2
`, func(row pgx.Row) error {
		return row.Scan(&b.ClientID, &b.UserID, &b.Priority, &b.AccessToken, &b.Status, &b.CreatedAt)
	}, clientID, userID)
	return &b, err
}

func GetClientUser(ctx context.Context, clientID string, isManager bool) ([]string, []string, error) {
	privilegeUserList, err := GetClientUserByPriority(ctx, clientID, []int{ClientUserPriorityHigh})
	if err != nil {
		return nil, nil, err
	}
	normalList := []int{ClientUserPriorityLow}
	if isManager {
		normalList = append(normalList, ClientUserPriorityStop)
	}
	normalUserList, err := GetClientUserByPriority(ctx, clientID, normalList)
	if err != nil {
		return nil, nil, err
	}
	return privilegeUserList, normalUserList, nil
}

func GetClientUserByPriority(ctx context.Context, clientID string, priority []int) ([]string, error) {
	userList := make([]string, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT user_id FROM client_users WHERE client_id=$1 AND priority=ANY($2)`, func(rows pgx.Rows) error {
		for rows.Next() {
			var b string
			if err := rows.Scan(&b); err != nil {
				return err
			}
			userList = append(userList, b)
		}
		return nil
	}, clientID, priority)
	if err != nil {
		return nil, err
	}
	return userList, nil
}

func GetAllClientUser(ctx context.Context) ([]*ClientUser, error) {
	allUser := make([]*ClientUser, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT cu.client_id, cu.user_id, cu.access_token, cu.priority, cu.is_async, cu.status, c.asset_id, c.speak_status, cu.deliver_at
FROM client_users AS cu
LEFT JOIN client AS c ON c.client_id=cu.client_id
WHERE cu.priority IN (1,2) AND cu.status!=9
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var cu ClientUser
			if err := rows.Scan(&cu.ClientID, &cu.UserID, &cu.AccessToken, &cu.Priority, &cu.IsAsync, &cu.Status, &cu.AssetID, &cu.SpeakStatus, &cu.DeliverAt); err != nil {
				return err
			}
			allUser = append(allUser, &cu)
		}
		return nil
	})
	return allUser, err
}

func GetClientUserIsAsync(ctx context.Context, clientID, userID string) (bool, error) {
	var isAsync bool
	err := session.Database(ctx).ConnQueryRow(ctx, `SELECT is_async FROM client_users WHERE client_id=$1 AND user_id=$2`,
		func(row pgx.Row) error {
			return row.Scan(&isAsync)
		}, clientID, userID,
	)
	return isAsync, err
}

func UpdateClientUserStatus(ctx context.Context, clientID, userID string, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET status=$3 WHERE client_id=$1 AND user_id=$2`, clientID, userID, status)
	return err
}

func UpdateClientUserPriority(ctx context.Context, clientID, userID string, priority int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3 WHERE client_id=$1 AND user_id=$2`, clientID, userID, priority)
	return err
}

var debounceUserMap = make(map[string]func(func()))

func UpdateClientUserDeliverTime(ctx context.Context, clientID, msgID string, deliverTime time.Time) error {
	if userID, err := getUserByDistributeMessageID(ctx, msgID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	} else {
		if debounceUserMap[userID] == nil {
			debounceUserMap[userID] = tools.Debounce(time.Minute * 5)
		}
		debounceUserMap[userID](func() {
			user, err := GetClientUserByClientIDAndUserID(ctx, clientID, userID)
			if err != nil {
				return
			}
			if user.Priority == ClientUserPriorityStop {
				status, err := getClientUserStatusByClientIDAndUserID(ctx, clientID, userID)
				if err != nil {
					return
				}
				priority := ClientUserPriorityLow
				if status != ClientUserStatusAudience {
					priority = ClientUserPriorityHigh
				}
				_, _ = session.Database(ctx).Exec(ctx, `UPDATE client_users SET deliver_at=$3, status=$4, priority=$5 WHERE client_id=$1 AND user_id=$2`, clientID, userID, deliverTime, status, priority)
			} else {
				_, _ = session.Database(ctx).Exec(ctx, `UPDATE client_users SET deliver_at=$3 WHERE client_id=$1 AND user_id=$2`, clientID, userID, deliverTime)
			}
		})
	}
	return nil
}
func UpdateClientUserPriorityAndAsync(ctx context.Context, clientID, userID string, priority int, isAsync bool) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3, is_async=$4 WHERE client_id=$1 AND user_id=$2`, clientID, userID, priority, isAsync)
	return err
}

func UpdateClientUserAndMessageToPending(ctx context.Context, clientID, userID string) error {
	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
UPDATE client_users 
SET priority=$5, is_async=$6
WHERE client_id=$1 AND user_id=$2 AND status IN ($3,$4)`,
			clientID, userID, MessageStatusPending, MessageStatusNormal, ClientUserPriorityPending, false)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE distribute_messages SET level=$3 WHERE client_id=$1 AND user_id=$2 AND status=1`, clientID, userID, DistributeMessageLevelAlone)
		return err
	})
	return err
}

func CheckUserIsActive(ctx context.Context, users []*ClientUser) {
	for _, user := range users {
		lastMsg, err := getLastMsgByClientID(ctx, user.ClientID)
		if err != nil {
			session.Logger(ctx).Println(err)
			continue
		}
		if lastMsg.CreatedAt.Sub(user.DeliverAt).Hours() > 24*7 {
			// 标记用户为不活跃，停止消息推送
			if err := UpdateClientUserPriority(ctx, user.ClientID, user.UserID, ClientUserPriorityStop); err != nil {
				session.Logger(ctx).Println(err)
				continue
			}
		}
	}
}

var cacheManagerMap = make(map[string][]string)

func getClientManager(ctx context.Context, clientID string) ([]string, error) {
	if cacheManagerMap[clientID] == nil {
		users, err := getClientUserByClientIDAndStatus(ctx, clientID, ClientUserStatusManager)
		if err != nil {
			return nil, err
		}
		cacheManagerMap[clientID] = users
	}
	return cacheManagerMap[clientID], nil
}

func getClientUserByClientIDAndStatus(ctx context.Context, clientID string, status int) ([]string, error) {
	users := make([]string, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT user_id FROM client_users WHERE client_id=$1 AND status=$2
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var user string
			if err := rows.Scan(&user); err != nil {
				return err
			}
			users = append(users, user)
		}
		return nil
	}, clientID, status)
	return users, err
}

func SendToClientManager(clientID string, msg *mixin.MessageView) {
	//if msg.Category != mixin.MessageCategoryPlainImage && msg.Category != mixin.MessageCategoryPlainText {
	if msg.Category != mixin.MessageCategoryPlainText {
		return
	}
	users, err := getClientManager(_ctx, clientID)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if len(users) <= 0 {
		log.Println("该社群没有管理员", clientID)
		return
	}
	client := GetMixinClientByID(_ctx, clientID)
	msgList := make([]*mixin.MessageRequest, 0)
	data := config.PrefixLeaveMsg + string(tools.Base64Decode(msg.Data))

	for _, userID := range users {
		conversationID := mixin.UniqueConversationID(clientID, userID)
		msgList = append(msgList, &mixin.MessageRequest{
			ConversationID:   conversationID,
			RecipientID:      userID,
			MessageID:        tools.GetUUID(),
			Category:         msg.Category,
			Data:             tools.Base64Encode([]byte(data)),
			RepresentativeID: msg.UserID,
			QuoteMessageID:   msg.QuoteMessageID,
		})
	}
	if err := SendMessages(_ctx, client.Client, msgList); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	if err := createMessage(_ctx, clientID, msg, MessageStatusLeaveMessage); err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	for _, _msg := range msgList {
		if err := _createDistributeMessage(_ctx, clientID, _msg.RecipientID, msg.MessageID, _msg.MessageID, _msg.QuoteMessageID, ClientUserPriorityHigh, DistributeMessageStatusLeaveMessage, msg.CreatedAt); err != nil {
			session.Logger(_ctx).Println(err)
			continue
		}
	}
}
