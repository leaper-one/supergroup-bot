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
	"github.com/go-redis/redis/v8"
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
  pay_status         SMALLINT NOT NULL DEFAULT 1,
  pay_expired_at     TIMESTAMP WITH TIME ZONE default '1970-01-01 00:00:00+00',
  is_received        BOOLEAN NOT NULL DEFAULT true,
  is_notice_join     BOOLEAN NOT NULL DEFAULT true,
  deliver_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  read_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  created_at         TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
  PRIMARY KEY (client_id, user_id)
);
CREATE INDEX IF NOT EXISTS client_user_idx ON client_users(client_id);
`

type ClientUser struct {
	ClientID     string    `json:"client_id,omitempty" redis:"client_id"`
	UserID       string    `json:"user_id,omitempty" redis:"user_id"`
	AccessToken  string    `json:"access_token,omitempty" redist:"access_token"`
	Priority     int       `json:"priority,omitempty" redist:"priority"`
	Status       int       `json:"status,omitempty" redist:"status"`
	PayStatus    int       `json:"pay_status,omitempty" redist:"pay_status"`
	CreatedAt    time.Time `json:"created_at,omitempty" redist:"created_at"`
	IsReceived   bool      `json:"is_received,omitempty" redist:"is_received"`
	IsNoticeJoin bool      `json:"is_notice_join,omitempty" redist:"is_notice_join"`
	MutedTime    string    `json:"muted_time,omitempty" redist:"muted_time"`
	MutedAt      time.Time `json:"muted_at,omitempty" redist:"muted_at"`
	PayExpiredAt time.Time `json:"pay_expired_at,omitempty" redist:"pay_expired_at"`
	ReadAt       time.Time `json:"read_at,omitempty" redist:"read_at"`
	DeliverAt    time.Time `json:"deliver_at,omitempty" redist:"deliver_at"`

	MuteAtInt       int64 `json:"-" redis:"mute_at_int"`
	PayExpiredAtInt int64 `json:"-" redis:"pay_expired_at_int"`
	ReadAtInt       int64 `json:"-" redis:"read_at_int"`
	DeliverAtInt    int64 `json:"-" redis:"deliver_at_int"`

	AssetID     string `json:"asset_id,omitempty" redis:"asset_id"`
	SpeakStatus int    `json:"speak_status,omitempty" redis:"speak_status"`
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
	ClientUserStatusBlock    = 4 // 拉黑
	ClientUserStatusLarge    = 5 // 大户
	ClientUserStatusGuest    = 8 // 嘉宾
	ClientUserStatusAdmin    = 9 // 管理员
)

func UpdateClientUser(ctx context.Context, user *ClientUser, fullName string) (bool, error) {
	u, err := GetClientUserByClientIDAndUserID(ctx, user.ClientID, user.UserID)
	isNewUser := false
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 第一次入群
			isNewUser = true
		}
	}
	if user.AccessToken != "" {
		go SendAuthSuccessMsg(user.ClientID, user.UserID)
		var msg string
		if user.Status == ClientUserStatusLarge {
			if u.Status < ClientUserStatusLarge {
				msg = config.Text.AuthForLarge
			}
		} else if user.Status != ClientUserStatusAudience {
			if u.Status < user.Status {
				msg = config.Text.AuthForFresh
			}
		}
		go SendClientUserTextMsg(_ctx, user.ClientID, user.UserID, msg, "")
	}
	if u.Status == ClientUserStatusAdmin || u.Status == ClientUserStatusGuest {
		user.Status = u.Status
		user.Priority = ClientUserPriorityHigh
	} else if u.PayStatus > ClientUserStatusAudience {
		user.Status = u.PayStatus
		user.Priority = ClientUserPriorityHigh
	}
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", user.ClientID, user.UserID))
	if user.PayExpiredAt.IsZero() {
		query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,status")
		_, err = session.Database(ctx).Exec(ctx, query, user.ClientID, user.UserID, user.AccessToken, user.Priority, user.Status)
	} else {
		query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,status,pay_status,pay_expired_at")
		_, err = session.Database(ctx).Exec(ctx, query, user.ClientID, user.UserID, user.AccessToken, user.Priority, user.Status, ClientUserStatusLarge, user.PayExpiredAt)
	}
	if isNewUser {
		cs := getClientConversationStatus(ctx, user.ClientID)
		// conversation 状态为普通的时候入群通知是打开的，就通知用户入群。
		if cs == ClientConversationStatusNormal &&
			getClientNewMemberNotice(ctx, user.ClientID) == ClientNewMemberNoticeOn {
			go SendClientTextMsg(user.ClientID, strings.ReplaceAll(config.Text.JoinMsg, "{name}", tools.SplitString(fullName, 12)), user.UserID, true)
		}
		go SendWelcomeAndLatestMsg(user.ClientID, user.UserID)
	}
	return isNewUser, err
}

// 用户导入时
func CreateOrUpdateClientUser(ctx context.Context, u *ClientUser) error {
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", u.ClientID, u.UserID))
	query := durable.InsertQueryOrUpdate("client_users", "client_id,user_id", "access_token,priority,status,deliver_at,read_at")
	_, err := session.Database(ctx).Exec(ctx, query, u.ClientID, u.UserID, u.AccessToken, u.Priority, u.Status, u.DeliverAt, u.ReadAt)
	return err
}

func GetClientUserByClientIDAndUserID(ctx context.Context, clientID, userID string) (ClientUser, error) {
	key := fmt.Sprintf("client_user:%s:%s", clientID, userID)
	var u ClientUser
	if err := session.Redis(ctx).StructScan(ctx, key, &u); err != nil {
		if errors.Is(err, redis.Nil) {
			return cacheClientUser(ctx, clientID, userID)
		}
		if !errors.Is(err, context.Canceled) {
			session.Logger(ctx).Println(err)
		}
		return ClientUser{}, err
	}
	return u, nil
}

func cacheClientUser(ctx context.Context, clientID, userID string) (ClientUser, error) {
	key := fmt.Sprintf("client_user:%s:%s", clientID, userID)
	var b ClientUser
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT cu.client_id,cu.user_id,cu.priority,cu.access_token,cu.status,cu.muted_time,cu.muted_at,cu.is_received,cu.is_notice_join,cu.pay_status,cu.pay_expired_at,cu.deliver_at,cu.read_at,cu.created_at,
c.asset_id,c.speak_status
FROM client_users cu
LEFT JOIN client c ON cu.client_id=c.client_id
WHERE cu.client_id=$1 AND cu.user_id=$2
`, clientID, userID).Scan(&b.ClientID, &b.UserID, &b.Priority, &b.AccessToken, &b.Status, &b.MutedTime, &b.MutedAt, &b.IsReceived, &b.IsNoticeJoin, &b.PayStatus, &b.PayExpiredAt, &b.DeliverAt, &b.ReadAt, &b.CreatedAt, &b.AssetID, &b.SpeakStatus); err != nil {
		return ClientUser{}, err
	}
	go func(key string, b ClientUser) {
		if err := session.Redis(_ctx).StructSet(_ctx, key, b); err != nil {
			session.Logger(_ctx).Println(err)
		}
	}(key, b)
	return b, nil
}

func cacheAllClientUser() {
	for {
		lastTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local)
		count := -1
		total := 0
		for {
			count, lastTime = _cacheAllClientUser(_ctx, lastTime)
			total += count
			if count == 0 {
				break
			}
		}
		time.Sleep(5 * time.Minute)
	}
}

func _cacheAllClientUser(ctx context.Context, lastTime time.Time) (int, time.Time) {
	cus := make([]ClientUser, 0, 1000)
	session.Database(ctx).ConnQuery(ctx, `
SELECT cu.client_id,cu.user_id,cu.priority,cu.access_token,cu.status,cu.muted_time,cu.muted_at,cu.is_received,cu.is_notice_join,cu.pay_status,cu.pay_expired_at,cu.deliver_at,cu.read_at,cu.created_at,
c.asset_id,c.speak_status
FROM client_users cu
LEFT JOIN client c ON cu.client_id=c.client_id
WHERE cu.created_at>$1
ORDER BY cu.created_at ASC LIMIT 1000
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var b ClientUser
			if err := rows.Scan(&b.ClientID, &b.UserID, &b.Priority, &b.AccessToken, &b.Status, &b.MutedTime, &b.MutedAt, &b.IsReceived, &b.IsNoticeJoin, &b.PayStatus, &b.PayExpiredAt, &b.DeliverAt, &b.ReadAt, &b.CreatedAt, &b.AssetID, &b.SpeakStatus); err != nil {
				return err
			}
			cus = append(cus, b)
		}
		return nil
	}, lastTime)
	if len(cus) == 0 {
		return 0, lastTime
	}
	if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, u := range cus {
			key := fmt.Sprintf("client_user:%s:%s", u.ClientID, u.UserID)
			uStr, err := json.Marshal(u)
			if err != nil {
				session.Logger(ctx).Println(err)
			}
			if err := p.Set(ctx, key, uStr, 15*time.Minute).Err(); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		session.Logger(ctx).Println(err)
	}
	return len(cus), cus[len(cus)-1].CreatedAt
}

func GetClientUserReceived(ctx context.Context, clientID string) ([]string, []string, error) {
	userList, err := GetDistributeMsgUser(ctx, clientID, false, false)
	if err != nil {
		return nil, nil, err
	}
	privilegeUserList := make([]string, 0)
	normalUserList := make([]string, 0)
	for _, u := range userList {
		if u.Priority == ClientUserPriorityHigh {
			privilegeUserList = append(privilegeUserList, u.UserID)
		} else if u.Priority == ClientUserPriorityLow {
			normalUserList = append(normalUserList, u.UserID)
		}
	}
	return privilegeUserList, normalUserList, nil
}

func GetDistributeMsgUser(ctx context.Context, clientID string, isJoinMsg, isBroadcast bool) ([]*ClientUser, error) {
	userList := make([]*ClientUser, 0)
	addQuery := ""
	if isJoinMsg {
		addQuery = "AND is_notice_join=true"
	}
	if !isBroadcast {
		addQuery = fmt.Sprintf("%s %s", addQuery, "AND is_received=true")
	}
	query := fmt.Sprintf(`
SELECT user_id, priority FROM client_users 
WHERE client_id=$1 AND priority IN (1,2) %s AND status IN (1,2,3,5,8,9)
ORDER BY created_at
`, addQuery)
	err := session.Database(ctx).ConnQuery(ctx, query, func(rows pgx.Rows) error {
		for rows.Next() {
			var u ClientUser
			if err := rows.Scan(&u.UserID, &u.Priority); err != nil {
				return err
			}
			userList = append(userList, &u)
		}
		return nil
	}, clientID)
	if err != nil {
		return nil, err
	}
	return userList, nil
}

func GetAllClientNeedAssetsCheckUser(ctx context.Context, hasPayedUser bool) ([]*ClientUser, error) {
	allUser := make([]*ClientUser, 0)
	query := `
SELECT cu.client_id, cu.user_id, cu.access_token, cu.priority, cu.status, coalesce(c.asset_id, '') as asset_id, c.speak_status, cu.deliver_at
FROM client_users AS cu
LEFT JOIN client AS c ON c.client_id=cu.client_id
WHERE cu.priority IN (1,2)
AND cu.status IN (1,2,3,4,5)
`
	if !hasPayedUser {
		query += `AND cu.pay_expired_at<NOW()`
	}
	err := session.Database(ctx).ConnQuery(ctx, query, func(rows pgx.Rows) error {
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

func updateClientUserStatus(ctx context.Context, clientID, userID string, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET status=$3 WHERE client_id=$1 AND user_id=$2`, clientID, userID, status)
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", clientID, userID))
	return err
}

func UpdateClientUserPriority(ctx context.Context, clientID, userID string, priority int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3 WHERE client_id=$1 AND user_id=$2`, clientID, userID, priority)
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", clientID, userID))
	return err
}

func UpdateClientUserPriorityAndStatus(ctx context.Context, clientID, userID string, priority, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3,status=$4 WHERE client_id=$1 AND user_id=$2`, clientID, userID, priority, status)
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", clientID, userID))
	return err
}

func UpdateClientUserActive(ctx context.Context, clientID, userID string, priority, status int) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET priority=$3,status=$4,deliver_at=$5,read_at=$5 WHERE client_id=$1 AND user_id=$2`, clientID, userID, priority, status, time.Now())
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", clientID, userID))
	return err
}

func UpdateClientUserPayStatus(ctx context.Context, clientID, userID string, status int, expiredAt time.Time) error {
	_, err := session.Database(ctx).Exec(ctx, `UPDATE client_users SET status=$3,priority=1,pay_status=$3,pay_expired_at=$4 WHERE client_id=$1 AND user_id=$2`, clientID, userID, status, expiredAt)
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", clientID, userID))
	return err
}

func UpdateClientUserActiveTimeToRedis(ctx context.Context, clientID, msgID string, deliverTime time.Time, status string) error {
	if status != "DELIVERED" && status != "READ" {
		return nil
	}
	dm, err := getDistributeMsgByMsgIDFromRedis(ctx, msgID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}
		session.Logger(ctx).Println(err)
		return err
	}
	user, err := GetClientUserByClientIDAndUserID(ctx, clientID, dm.UserID)
	if err != nil {
		return err
	}
	go activeUser(&user)
	if status == "READ" {
		if err := session.Redis(ctx).QSet(ctx, fmt.Sprintf("ack_msg:read:%s:%s", clientID, user.UserID), deliverTime, time.Hour*2); err != nil {
			return err
		}
	} else {
		if err := session.Redis(ctx).QSet(ctx, fmt.Sprintf("ack_msg:deliver:%s:%s", clientID, user.UserID), deliverTime, time.Hour*2); err != nil {
			return err
		}
	}
	return nil
}

func taskUpdateActiveUserToPsql() {
	ctx := _ctx
	list, err := GetClientList(ctx)
	if err != nil {
		session.Logger(ctx).Println(err)
	}
	for {
		time.Sleep(time.Hour)
		for _, client := range list {
			UpdateClientUserActiveTimeFromRedis(ctx, client.ClientID)
		}
	}
}

func UpdateClientUserActiveTimeFromRedis(ctx context.Context, clientID string) error {
	if err := UpdateClientUserActiveTime(ctx, clientID, "deliver"); err != nil {
		return err
	}
	if err := UpdateClientUserActiveTime(ctx, clientID, "read"); err != nil {
		return err
	}
	return nil
}

func UpdateClientUserActiveTime(ctx context.Context, clientID, status string) error {
	allUser, err := getClientUserByClientID(ctx, clientID, 0)
	if err != nil {
		return err
	}
	keys := make([]string, len(allUser))
	for _, userID := range allUser {
		keys = append(keys, fmt.Sprintf("ack_msg:%s:%s:%s", status, clientID, userID))
	}
	for {
		if len(keys) == 0 {
			break
		}
		var currentKeys []string
		if len(keys) > 500 {
			currentKeys = keys[:500]
			keys = keys[500:]
		} else {
			currentKeys = keys
			keys = nil
		}
		results := make([]*redis.StringCmd, 0, len(currentKeys))
		if _, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
			for _, key := range currentKeys {
				results = append(results, p.Get(ctx, key))
			}
			return nil
		}); err != nil {
			if !errors.Is(err, redis.Nil) {
				session.Logger(ctx).Println(err)
			}
		}

		if err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
			for _, v := range results {
				t, err := v.Result()
				if err != nil {
					if !errors.Is(err, redis.Nil) {
						session.Logger(ctx).Println(err)
					}
					continue
				}
				key := v.Args()[1].(string)
				clientID := strings.Split(key, ":")[2]
				userID := strings.Split(key, ":")[3]
				_, err = tx.Exec(ctx,
					fmt.Sprintf(`UPDATE client_users SET %s_at=$3 WHERE client_id=$1 AND user_id=$2`, status),
					clientID, userID, t)
				if err != nil {
					session.Logger(ctx).Println(err)
				}
			}
			return nil
		}); err != nil {
			session.Logger(ctx).Println(err)
			return err
		}
		time.Sleep(time.Second)
	}
	return nil
}

func LeaveGroup(ctx context.Context, u *ClientUser) error {
	if err := updateClientUserStatus(ctx, u.ClientID, u.UserID, ClientUserStatusExit); err != nil {
		return err
	}
	go SendClientUserTextMsg(_ctx, u.ClientID, u.UserID, config.Text.LeaveGroup, "")
	return nil
}

func UpdateClientUserChatStatus(ctx context.Context, u *ClientUser, isReceived, isNoticeJoin bool) (ClientUser, error) {
	msg := ""
	if isReceived {
		msg = config.Text.OpenChatStatus
	} else {
		msg = config.Text.CloseChatStatus
		isNoticeJoin = false
	}

	_, err := session.Database(ctx).Exec(ctx, `
UPDATE client_users 
SET is_received=$3,is_notice_join=$4 
WHERE client_id=$1 AND user_id=$2
`, u.ClientID, u.UserID, isReceived, isNoticeJoin)
	if err != nil {
		return ClientUser{}, err
	}
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", u.ClientID, u.UserID))
	if u.IsReceived != isReceived {
		go SendClientUserTextMsg(_ctx, u.ClientID, u.UserID, msg, "")
	}
	return GetClientUserByClientIDAndUserID(ctx, u.ClientID, u.UserID)
}

func CheckUserIsActive(ctx context.Context, user *ClientUser, lastMsgCreatedAt time.Time) error {
	if lastMsgCreatedAt.IsZero() {
		return nil
	}
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

func getClientManager(ctx context.Context, clientID string) ([]string, error) {
	users, err := getClientUserByClientID(ctx, clientID, ClientUserStatusAdmin)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func getClientUserByClientID(ctx context.Context, clientID string, status int) ([]string, error) {
	users := make([]string, 0)
	query := "SELECT user_id FROM client_users WHERE client_id=$1"
	if status != 0 {
		query += " AND status=" + strconv.Itoa(status)
	}
	err := session.Database(ctx).ConnQuery(ctx, query, func(rows pgx.Rows) error {
		for rows.Next() {
			var user string
			if err := rows.Scan(&user); err != nil {
				return err
			}
			users = append(users, user)
		}
		return nil
	}, clientID)
	return users, err
}

var queryAll = `SELECT COUNT(1) FROM client_users WHERE client_id=$1 AND status!=$2`
var queryWeek = queryAll + " AND NOW() - created_at < interval '7 days'"

func getClientPeopleCount(ctx context.Context, clientID string) (decimal.Decimal, decimal.Decimal, error) {
	var all, week decimal.Decimal
	allString, err := session.Redis(ctx).QGet(ctx, "people_count_all:"+clientID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := session.Database(ctx).QueryRow(ctx, queryAll, clientID, ClientUserStatusExit).Scan(&all); err != nil {
				return decimal.Zero, decimal.Zero, err
			}
			if err := session.Redis(ctx).QSet(ctx, "people_count_all:"+clientID, all.String(), time.Minute); err != nil {
				session.Logger(ctx).Println(err)
			}
		} else {
			return decimal.Zero, decimal.Zero, err
		}
	} else {
		all, _ = decimal.NewFromString(allString)
	}
	weekString, err := session.Redis(ctx).QGet(ctx, "people_count_week:"+clientID).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := session.Database(ctx).QueryRow(ctx, queryWeek, clientID, ClientUserStatusExit).Scan(&week); err != nil {
				return decimal.Zero, decimal.Zero, err
			}
			if err := session.Redis(ctx).QSet(ctx, "people_count_week:"+clientID, week.String(), time.Minute); err != nil {
				session.Logger(ctx).Println(err)
			}
		} else {
			return decimal.Zero, decimal.Zero, err
		}
	} else {
		week, _ = decimal.NewFromString(weekString)
	}
	return all, week, nil
}

func activeUser(u *ClientUser) {
	if u.Priority != ClientUserPriorityStop {
		return
	}
	var err error
	status := ClientUserStatusAudience
	priority := ClientUserPriorityLow
	if u.PayExpiredAt.After(time.Now()) {
		status = u.PayStatus
	} else {
		status, err = GetClientUserStatusByClientUser(_ctx, u)
		if err != nil {
			session.Logger(_ctx).Println(err)
		}
	}
	if status != ClientUserStatusAudience {
		priority = ClientUserPriorityHigh
	}
	if err := UpdateClientUserActive(_ctx, u.ClientID, u.UserID, priority, status); err != nil {
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

type clientUserView struct {
	UserID         string    `json:"user_id,omitempty"`
	AvatarURL      string    `json:"avatar_url,omitempty"`
	FullName       string    `json:"full_name,omitempty"`
	IdentityNumber string    `json:"identity_number,omitempty"`
	Status         int       `json:"status,omitempty"`
	ActiveAt       time.Time `json:"active_at,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
}

var clientUserViewPrefix = `SELECT u.user_id,avatar_url,full_name,identity_number,status,deliver_at,cu.created_at
FROM client_users cu
LEFT JOIN users u ON cu.user_id=u.user_id 
WHERE client_id=$1 `

// {state}: all 全部 mute 禁言 block 拉黑 guest 嘉宾 admin 管理员
func GetClientUserList(ctx context.Context, u *ClientUser, page int, status string) ([]*clientUserView, error) {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	cs, err := getMuteOrBlockClientUserList(ctx, u, status)
	if err != nil {
		return nil, err
	}
	if cs == nil {
		cs, err = getAllOrGuestOrAdminClientUserList(ctx, u, page, status)
	}
	return cs, err
}

func getMuteOrBlockClientUserList(ctx context.Context, u *ClientUser, status string) ([]*clientUserView, error) {
	if status == "mute" {
		// 获取禁言用户列表
		return getClientUserView(ctx, clientUserViewPrefix+`
AND muted_at>NOW()
`, u.ClientID)
	}
	if status == "block" {
		// 获取拉黑用户列表
		return getClientBlockUserView(ctx, u.ClientID)
	}
	return nil, nil
}

func getClientBlockUserView(ctx context.Context, clientID string) ([]*clientUserView, error) {
	cus := make([]*clientUserView, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT cbu.user_id,avatar_url,full_name,identity_number,cu.status,deliver_at,cu.created_at
FROM client_block_user cbu
LEFT JOIN client_users cu ON cbu.user_id=cu.user_id AND cbu.client_id=cu.client_id
LEFT JOIN users u ON cu.user_id=u.user_id
WHERE cbu.client_id=$1
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var u clientUserView
			if err := rows.Scan(&u.UserID, &u.AvatarURL, &u.FullName, &u.IdentityNumber, &u.Status, &u.ActiveAt, &u.CreatedAt); err != nil {
				return err
			}
			cus = append(cus, &u)
		}
		return nil
	}, clientID)
	return cus, err
}

var clientUserStatusMap = map[string][]int{
	"all": {
		ClientUserStatusAudience,
		ClientUserStatusFresh,
		ClientUserStatusSenior,
		ClientUserStatusLarge,
	},
	"guest": {ClientUserStatusGuest},
	"admin": {ClientUserStatusAdmin},
}

func getAllOrGuestOrAdminClientUserList(ctx context.Context, u *ClientUser, page int, status string) ([]*clientUserView, error) {
	if status == "mute" || status == "block" {
		session.Logger(ctx).Println("status::", status)
		return nil, nil
	}
	statusList := clientUserStatusMap[status]
	if statusList == nil {
		session.Logger(ctx).Println("status::", status)
		return nil, nil
	}
	cs, err := getClientUserView(ctx, clientUserViewPrefix+`
AND status=ANY($3)
ORDER BY created_at ASC OFFSET $2 LIMIT 20`, u.ClientID, (page-1)*20, statusList)
	if err != nil {
		return nil, err
	}
	if page == 1 && status == "all" {
		if users, err := getClientUserView(ctx, clientUserViewPrefix+`
AND status IN (8,9)
ORDER BY status DESC 
`, u.ClientID); err != nil {
			return nil, err
		} else {
			c, _ := GetClientByIDOrHost(ctx, u.ClientID)
			for i, v := range users {
				if v.UserID == c.OwnerID {
					users[0], users[i] = users[i], users[0]
					break
				}
			}
			cs = append(users, cs...)
		}
	}
	return cs, nil
}

func GetAdminAndGuestUserList(ctx context.Context, u *ClientUser) ([]*clientUserView, error) {
	return getClientUserView(ctx, clientUserViewPrefix+`
AND status IN (8,9)
ORDER BY status DESC 
`, u.ClientID)
}

// 获取 全部用户数量/禁言用户数量/拉黑用户数量/嘉宾数量/管理员数量
func GetClientUserStat(ctx context.Context, u *ClientUser) (map[string]int, error) {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	var allUserCount, muteUserCount, blockUserCount, guestUserCount, adminUserCount int
	// res := make(map[string]int)
	if err := session.Database(ctx).QueryRow(ctx, `SELECT count(1) FROM client_users WHERE client_id=$1`, u.ClientID).Scan(&allUserCount); err != nil {
		return nil, err
	}
	if err := session.Database(ctx).QueryRow(ctx, `SELECT count(1) FROM client_block_user WHERE client_id=$1`, u.ClientID).Scan(&blockUserCount); err != nil {
		return nil, err
	}
	if err := session.Database(ctx).QueryRow(ctx, `SELECT count(1) FROM client_users WHERE client_id=$1 AND muted_at>NOW()`, u.ClientID).Scan(&muteUserCount); err != nil {
		return nil, err
	}
	if err := session.Database(ctx).QueryRow(ctx, `SELECT count(1) FROM client_users WHERE client_id=$1 AND status=8`, u.ClientID).Scan(&guestUserCount); err != nil {
		return nil, err
	}
	if err := session.Database(ctx).QueryRow(ctx, `SELECT count(1) FROM client_users WHERE client_id=$1 AND status=9`, u.ClientID).Scan(&adminUserCount); err != nil {
		return nil, err
	}
	return map[string]int{
		"all":   allUserCount,
		"mute":  muteUserCount,
		"block": blockUserCount,
		"guest": guestUserCount,
		"admin": adminUserCount,
	}, nil
}

func GetClientUserByIDOrName(ctx context.Context, u *ClientUser, key string) ([]*clientUserView, error) {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	return getClientUserView(ctx, clientUserViewPrefix+`
AND (
	(u.identity_number ILIKE '%' || $2 || '%') OR 
	(u.full_name ILIKE '%' || $2 || '%')
)
LIMIT 20
`, u.ClientID, key)
}

func getClientUserView(ctx context.Context, query string, p ...interface{}) ([]*clientUserView, error) {
	cs := make([]*clientUserView, 0)
	err := session.Database(ctx).ConnQuery(ctx, query, func(rows pgx.Rows) error {
		for rows.Next() {
			var u clientUserView
			if err := rows.Scan(&u.UserID, &u.AvatarURL, &u.FullName, &u.IdentityNumber, &u.Status, &u.ActiveAt, &u.CreatedAt); err != nil {
				return err
			}
			cs = append(cs, &u)
		}
		return nil
	}, p...)
	return cs, err
}

func UpdateClientUserStatus(ctx context.Context, u *ClientUser, userID string, status int, isCancel bool) error {
	if status == ClientUserStatusAdmin {
		if !checkIsOwner(ctx, u.ClientID, u.UserID) {
			return session.ForbiddenError(ctx)
		}
	} else {
		if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
			return session.ForbiddenError(ctx)
		}
	}
	var s string
	var msg string
	if status == ClientUserStatusAdmin {
		s = config.Text.StatusAdmin
	} else if status == ClientUserStatusGuest {
		s = config.Text.StatusGuest
	}
	if isCancel {
		msg = config.Text.StatusCancel
		status = ClientUserStatusLarge
	} else {
		msg = config.Text.StatusSet
	}

	if _, err := session.Database(ctx).Exec(ctx, `
	UPDATE client_users SET status=$3 WHERE client_id=$1 AND user_id=$2
	`, u.ClientID, userID, status); err != nil {
		return err
	}
	session.Redis(ctx).QDel(ctx, fmt.Sprintf("client_user:%s:%s", u.ClientID, userID))
	user, err := getUserByID(ctx, userID)
	if err != nil {
		session.Logger(ctx).Println("设置用户状态的时候没找到用户...", err)
		return err
	}
	msg = strings.ReplaceAll(msg, "{full_name}", user.FullName)
	msg = strings.ReplaceAll(msg, "{identity_number}", user.IdentityNumber)
	msg = strings.ReplaceAll(msg, "{status}", s)
	if !isCancel && status == ClientUserStatusGuest {
		go SendClientUserTextMsg(_ctx, u.ClientID, userID, msg, "")
	}
	go SendToClientManager(u.ClientID, &mixin.MessageView{
		ConversationID: mixin.UniqueConversationID(u.ClientID, userID),
		UserID:         userID,
		MessageID:      tools.GetUUID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           tools.Base64Encode([]byte(msg)),
		CreatedAt:      time.Now(),
	}, false, false)
	return nil
}

func MuteUserByID(ctx context.Context, u *ClientUser, userID, muteTime string) error {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	return muteClientUser(ctx, u.ClientID, userID, muteTime)
}

func BlockUserByID(ctx context.Context, u *ClientUser, userID string, isCancel bool) error {
	if !checkIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	return blockClientUser(ctx, u.ClientID, userID, isCancel)
}

func checkUserIsVIP(ctx context.Context, userID string) bool {
	var count int
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT count(1) FROM client_users WHERE user_id=$1 AND status>1
`, userID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}

func checkUserIsInSystem(ctx context.Context, userID string) bool {
	var count int
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT count(1) FROM client_users WHERE user_id=$1
`, userID).Scan(&count); err != nil {
		return false
	}
	return count > 0
}
