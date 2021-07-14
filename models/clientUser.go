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
  status             SMALLINT NOT NULL DEFAULT 0, -- 0 未入群 1 观众 2 入门 3 资深 5 大户 8 嘉宾 9 管理
  muted_time         VARCHAR DEFAULT '',
  muted_at           TIMESTAMP WITH TIME ZONE default '1970-01-01 00:00:00+00',
  is_received        BOOLEAN NOT NULL DEFAULT true,
  is_notice_join     BOOLEAN NOT NULL DEFAULT true,
  deliver_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  read_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
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
	Status       int       `json:"status,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	IsReceived   bool      `json:"is_received,omitempty"`
	IsNoticeJoin bool      `json:"is_notice_join,omitempty"`
	MutedTime    string    `json:"muted_time,omitempty"`
	MutedAt      time.Time `json:"muted_at,omitempty"`
	ReadAt       time.Time `json:"read_at,omitempty"`
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
	isNewUser := false
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 第一次入群
			isNewUser = true
			cs := getClientConversationStatus(ctx, user.ClientID)
			if cs != ClientConversationStatusMute &&
				cs != ClientConversationStatusAudioLive {
				tFullName := []rune(fullName)
				if len(tFullName) > 12 {
					fullName = string(tFullName[:12]) + "..."
				}
				go SendClientTextMsg(user.ClientID, strings.ReplaceAll(config.Config.Text.JoinMsg, "{name}", fullName), user.UserID, true)
			}
		}
	}
	go SendAuthSuccessMsg(user.ClientID, user.UserID)
	if u.Status == ClientUserStatusManager || u.Status == ClientUserStatusGuest {
		user.Status = u.Status
		user.Priority = ClientUserPriorityHigh
	}
	query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,status")
	_, err = session.Database(ctx).Exec(ctx, query, user.ClientID, user.UserID, user.AccessToken, user.Priority, user.Status)
	if isNewUser {
		go SendWelcomeAndLatestMsg(user.ClientID, user.UserID)
	}
	return err
}

func CreateOrUpdateClientUser(ctx context.Context, u *ClientUser) error {
	query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,status,deliver_at,read_at")
	_, err := session.Database(ctx).Exec(ctx, query, u.ClientID, u.UserID, u.AccessToken, u.Priority, u.Status, u.DeliverAt, u.ReadAt)
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
SELECT cu.client_id, cu.user_id, cu.access_token, cu.priority, cu.status, c.asset_id, c.speak_status, cu.deliver_at
FROM client_users AS cu
LEFT JOIN client AS c ON c.client_id=cu.client_id
WHERE cu.priority IN (1,2) AND cu.status NOT IN (0,8,9)
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var cu ClientUser
			if err := rows.Scan(&cu.ClientID, &cu.UserID, &cu.AccessToken, &cu.Priority, &cu.Status, &cu.AssetID, &cu.SpeakStatus, &cu.DeliverAt); err != nil {
				return err
			}
			allUser = append(allUser, &cu)
		}
		return nil
	})
	return allUser, err
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

func UpdateClientUserActive(ctx context.Context, clientID, userID string, priority, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3,status=$4,deliver_at=$5,read_at=$5 WHERE client_id=$1 AND user_id=$2`, clientID, userID, priority, status, time.Now())
	return err
}

func UpdateClientUserDeliverTime(ctx context.Context, clientID, msgID string, deliverTime time.Time, status string) error {
	if status != "DELIVERED" && status != "READ" {
		return nil
	}
	dm, err := getDistributeMessageByClientIDAndMessageID(ctx, clientID, msgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	user, err := GetClientUserByClientIDAndUserID(ctx, clientID, dm.UserID)
	if err != nil {
		return err
	}
	if user.Priority == ClientUserPriorityStop {
		activeUser(user)
	}
	if status == "READ" {
		_, _ = session.Database(ctx).Exec(ctx, `UPDATE client_users SET read_at=$3,deliver_at=$3 WHERE client_id=$1 AND user_id=$2`, clientID, dm.UserID, deliverTime)
	} else {
		_, _ = session.Database(ctx).Exec(ctx, `UPDATE client_users SET deliver_at=$3 WHERE client_id=$1 AND user_id=$2`, clientID, dm.UserID, deliverTime)
	}
	return nil
}

func LeaveGroup(ctx context.Context, u *ClientUser) error {
	//_, err := session.Database(ctx).Exec(ctx, `DELETE FROM client_users WHERE client_id=$1 AND user_id=$2`, u.ClientID, u.UserID)
	if err := UpdateClientUserStatus(ctx, u.ClientID, u.UserID, ClientUserStatusExit); err != nil {
		return err
	}
	client := GetMixinClientByID(ctx, u.ClientID)
	go SendTextMsg(_ctx, &client, u.UserID, config.Config.Text.LeaveGroup)
	return nil
}

func UpdateClientUserChatStatusByHost(ctx context.Context, u *ClientUser, isReceived, isNoticeJoin bool) (*ClientUser, error) {
	msg := ""
	if isReceived {
		msg = config.Config.Text.OpenChatStatus
	} else {
		msg = config.Config.Text.CloseChatStatus
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

func SendDistributeMsgAloneList(ctx context.Context, clientID, userID string, priority, curStatus int) {
	startAt, err := getLeftDistributeMsgAndDistribute(ctx, clientID, userID)
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	for {
		if startAt.IsZero() {
			break
		}
		startAt, err = sendLeftMsg(ctx, clientID, userID, startAt)
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
	}
	err = UpdateClientUserPriorityAndStatus(ctx, clientID, userID, priority, curStatus)
	if err != nil {
		session.Logger(ctx).Println(err)
	}
}

func CheckUserIsActive(ctx context.Context, user *ClientUser, lastMsgCreatedAt time.Time) error {
	if lastMsgCreatedAt.Sub(user.DeliverAt).Hours() > config.NotActiveCheckTime {
		// 标记用户为不活跃，停止消息推送
		go SendStopMsg(user.ClientID, user.UserID)
		if err := UpdateClientUserPriority(ctx, user.ClientID, user.UserID, ClientUserPriorityStop); err != nil {
			session.Logger(ctx).Println(err)
			return err
		}
	} else if user.Priority == ClientUserPriorityStop {
		activeUser(user)
	}
	return nil
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

func activeUser(u *ClientUser) {
	if u.Priority != ClientUserPriorityStop {
		return
	}
	curStatus, err := GetClientUserStatusByClientUser(_ctx, u)
	if err != nil {
		session.Logger(_ctx).Println(err)
	}
	priority := ClientUserPriorityLow
	if curStatus != ClientUserStatusAudience {
		priority = ClientUserPriorityHigh
	}
	if err := UpdateClientUserActive(_ctx, u.ClientID, u.UserID, priority, curStatus); err != nil {
		session.Logger(_ctx).Println(err)
	}
}

func GetPendingClientUser(ctx context.Context) ([]*ClientUser, error) {
	cus := make([]*ClientUser, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT client_id,user_id,access_token
FROM client_users
WHERE priority=$1
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var cu ClientUser
			if err := rows.Scan(&cu.ClientID, &cu.UserID, &cu.AccessToken); err != nil {
				return err
			}
			cus = append(cus, &cu)
		}
		return nil
	}, ClientUserPriorityPending)
	return cus, err
}
