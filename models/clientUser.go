package models

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
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
  is_received        BOOLEAN NOT NULL DEFAULT true,
  is_notice_join     BOOLEAN NOT NULL DEFAULT true,
  deliver_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  PRIMARY KEY (client_id, user_id)
);
CREATE INDEX client_user_idx ON client_users(client_id);
CREATE INDEX client_user_send_idx ON transfers(client_id, can_send);
`

type ClientUser struct {
	ClientID     string    `json:"client_id,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	Priority     int       `json:"priority,omitempty"`
	IsAsync      bool      `json:"is_async,omitempty"`
	Status       int       `json:"status,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	IsReceived   bool      `json:"is_received,omitempty"`
	IsNoticeJoin bool      `json:"is_notice_join,omitempty"`
	MutedTime    string    `json:"muted_time,omitempty"`
	MutedAt      time.Time `json:"muted_at,omitempty"`
	DeliverAt    time.Time `json:"deliver_at,omitempty"`

	AssetID     string `json:"asset_id,omitempty"`
	SpeakStatus int    `json:"speak_status,omitempty"`
}

const (
	ClientUserPriorityHigh    = 1 // 高优先级
	ClientUserPriorityLow     = 2 // 低优先级
	ClientUserPriorityPending = 3 // 补发中
	ClientUserPriorityStop    = 4 // 暂停发送

	ClientUserStatusExit     = 0 // 退群
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
			go SendWelcomeAndLatestMsg(user.ClientID, user.UserID)

			if !checkClientIsMute(ctx, user.ClientID) {
				go SendClientTextMsg(user.ClientID, strings.ReplaceAll(config.JoinMsg, "{name}", fullName), user.UserID, true)
			}
		}
	}
	go SendAuthSuccessMsg(user.ClientID, user.UserID)
	if u.Status == ClientUserStatusManager || u.Status == ClientUserStatusGuest {
		user.Status = u.Status
		user.Priority = ClientUserPriorityHigh
	}
	query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,is_async,status")
	_, err = session.Database(ctx).Exec(ctx, query, user.ClientID, user.UserID, user.AccessToken, user.Priority, user.IsAsync, user.Status)
	return err
}

func GetClientUserByClientIDAndUserID(ctx context.Context, clientID, userID string) (*ClientUser, error) {
	var b ClientUser
	err := session.Database(ctx).QueryRow(ctx, `
SELECT client_id,user_id,priority,access_token,status,muted_time,muted_at,is_received,is_notice_join,created_at 
FROM client_users 
WHERE client_id=$1 AND user_id=$2
`, clientID, userID).Scan(&b.ClientID, &b.UserID, &b.Priority, &b.AccessToken, &b.Status, &b.MutedTime, &b.MutedAt, &b.IsReceived, &b.IsNoticeJoin, &b.CreatedAt)
	return &b, err
}

func GetClientUserReceived(ctx context.Context, clientID string) ([]string, []string, error) {
	privilegeUserList, err := GetClientUserByPriority(ctx, clientID, []int{ClientUserPriorityHigh}, false, false)
	if err != nil {
		return nil, nil, err
	}
	normalList := []int{ClientUserPriorityLow}
	normalUserList, err := GetClientUserByPriority(ctx, clientID, normalList, false, false)
	if err != nil {
		return nil, nil, err
	}
	return privilegeUserList, normalUserList, nil
}

func GetClientUserByPriority(ctx context.Context, clientID string, priority []int, isJoinMsg, isBroadcast bool) ([]string, error) {
	userList := make([]string, 0)
	addQuery := ""
	if isJoinMsg {
		addQuery = "AND is_notice_join=true"
	}
	if !isBroadcast {
		addQuery = fmt.Sprintf("%s %s", addQuery, "AND is_received=true")
	}
	query := fmt.Sprintf(`
SELECT user_id FROM client_users 
WHERE client_id=$1 AND priority=ANY($2) %s AND status!=$3
ORDER BY created_at
`, addQuery)

	err := session.Database(ctx).ConnQuery(ctx, query, func(rows pgx.Rows) error {
		for rows.Next() {
			var b string
			if err := rows.Scan(&b); err != nil {
				return err
			}
			userList = append(userList, b)
		}
		return nil
	}, clientID, priority, ClientUserStatusExit)
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
WHERE cu.priority IN (1,2) AND cu.status NOT IN (0,8,9)
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
	err := session.Database(ctx).QueryRow(ctx, `SELECT is_async FROM client_users WHERE client_id=$1 AND user_id=$2`,
		clientID, userID,
	).Scan(&isAsync)
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

func UpdateClientUserPriorityAndStatus(ctx context.Context, clientID, userID string, priority, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3,status=$4 WHERE client_id=$1 AND user_id=$2`, clientID, userID, priority, status)
	return err
}

var debounceUserMap = make(map[string]func(func()))

func UpdateClientUserDeliverTime(ctx context.Context, clientID, msgID string, deliverTime time.Time) error {
	dm, err := getDistributeMessageByClientIDAndMessageID(ctx, clientID, msgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if debounceUserMap[dm.UserID] == nil {
		debounceUserMap[dm.UserID] = tools.Debounce(config.DebounceTime)
	}
	debounceUserMap[dm.UserID](func() {
		user, err := GetClientUserByClientIDAndUserID(ctx, clientID, dm.UserID)
		if err != nil {
			return
		}
		if user.Priority == ClientUserPriorityStop {
			status, err := GetClientUserStatusByClientIDAndUserID(ctx, clientID, dm.UserID)
			if err != nil {
				status = ClientUserStatusAudience
			}
			priority := ClientUserPriorityLow
			if status != ClientUserStatusAudience {
				priority = ClientUserPriorityHigh
			}
			_, _ = session.Database(ctx).Exec(ctx, `UPDATE client_users SET deliver_at=$3, status=$4, priority=$5 WHERE client_id=$1 AND user_id=$2`, clientID, dm.UserID, deliverTime, status, priority)
		} else {
			_, _ = session.Database(ctx).Exec(ctx, `UPDATE client_users SET deliver_at=$3 WHERE client_id=$1 AND user_id=$2`, clientID, dm.UserID, deliverTime)
		}
	})

	return nil
}
func UpdateClientUserPriorityAndAsync(ctx context.Context, clientID, userID string, priority int, isAsync bool) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3, is_async=$4 WHERE client_id=$1 AND user_id=$2`, clientID, userID, priority, isAsync)
	return err
}

func LeaveGroup(ctx context.Context, u *ClientUser) error {
	//_, err := session.Database(ctx).Exec(ctx, `DELETE FROM client_users WHERE client_id=$1 AND user_id=$2`, u.ClientID, u.UserID)
	if err := UpdateClientUserStatus(ctx, u.ClientID, u.UserID, ClientUserStatusExit); err != nil {
		return err
	}
	client := GetMixinClientByID(ctx, u.ClientID)
	go SendTextMsg(_ctx, &client, u.UserID, config.LeaveGroup)
	return nil
}

func UpdateClientUserChatStatusByHost(ctx context.Context, u *ClientUser, isReceived, isNoticeJoin bool) (*ClientUser, error) {
	msg := ""
	if isReceived {
		msg = config.OpenChatStatus
	} else {
		msg = config.CloseChatStatus
		isNoticeJoin = false
	}
	if err := UpdateClientUserChatStatus(ctx, u.ClientID, u.UserID, isReceived, isNoticeJoin); err != nil {
		return nil, err
	}
	client := GetMixinClientByID(ctx, u.ClientID)
	if u.IsReceived != isReceived {
		go SendTextMsg(_ctx, &client, u.UserID, msg)
	}
	return GetClientUserByClientIDAndUserID(ctx, u.ClientID, u.UserID)
}

func UpdateClientUserChatStatus(ctx context.Context, clientID, userID string, isReceived, isNoticeJoin bool) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET is_received=$3,is_notice_join=$4 WHERE client_id=$1 AND user_id=$2`, clientID, userID, isReceived, isNoticeJoin)
	return err
}

func UpdateClientUserAndMessageToPending(ctx context.Context, clientID, userID string) error {
	// TODO 低状态变高状态需要 搞成一个单独的队列来操作
	// 现在先简单的直接变成高状态消息
	//	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
	//		_, err := tx.Exec(ctx, `
	//UPDATE client_users
	//SET priority=$5, is_async=$6
	//WHERE client_id=$1 AND user_id=$2 AND status IN ($3,$4)`,
	//			clientID, userID, MessageStatusPending, MessageStatusNormal, ClientUserPriorityPending, false)
	//		if err != nil {
	//			return err
	//		}
	//		_, err = tx.Exec(ctx, `UPDATE distribute_messages SET level=$3 WHERE client_id=$1 AND user_id=$2 AND status=1`, clientID, userID, DistributeMessageLevelAlone)
	//		return err
	//	})
	_, err := session.Database(ctx).Exec(ctx, `UPDATE distribute_messages SET level=1 WHERE client_id=$1 AND user_id=$2 AND status=2`, clientID, userID)
	return err
}

func CheckUserIsActive(ctx context.Context, users []*ClientUser) {
	for _, user := range users {
		lastMsg, err := getLastMsgByClientID(ctx, user.ClientID)
		if err != nil {
			session.Logger(ctx).Println(err)
			continue
		}
		if lastMsg.CreatedAt.Sub(user.DeliverAt).Hours() > config.NotActiveCheckTime {
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

func getClientPeopleCount(ctx context.Context, clientID string) (decimal.Decimal, decimal.Decimal, error) {
	queryAll := `SELECT COUNT(1) FROM client_users WHERE client_id=$1 AND status!=$2`
	queryWeek := queryAll + " AND NOW() - created_at < interval '7 days'"
	var all, week decimal.Decimal
	if err := session.Database(ctx).QueryRow(ctx, queryAll, clientID, ClientUserStatusExit).Scan(&all); err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	if err := session.Database(ctx).QueryRow(ctx, queryWeek, clientID, ClientUserStatusExit).Scan(&week); err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	return all, week, nil

}
