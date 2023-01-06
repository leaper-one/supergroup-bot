package common

import (
	"context"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const client_block_user_DDL = `
CREATE TABLE IF NOT EXISTS client_block_user (
  client_id           VARCHAR(36) NOT NULL,
  user_id             VARCHAR(36) NOT NULL,
	operator_id         VARCHAR(36) NOT NULL,
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
  PRIMARY KEY (client_id,user_id)
);
CREATE INDEX IF NOT EXISTS client_block_user_idx ON client_block_user(client_id);
ALTER TABLE client_block_user ADD COLUMN IF NOT EXISTS operator_id VARCHAR(36) default '';
`

const block_user_DDL = `
CREATE TABLE IF NOT EXISTS block_user (
  user_id             VARCHAR(36) NOT NULL PRIMARY KEY,
	operator_id         VARCHAR(36) NOT NULL DEFAULT '',
	memo								VARCHAR(255) NOT NULL DEFAULT '',
  created_at          TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
`

type ClientBlockUser struct {
	ClientID  string
	UserID    string
	CreatedAt time.Time
}

type BlockUser struct {
	UserID    string
	CreatedAt time.Time
}

var cacheBlockClientUserIDMap = tools.NewMutex()

// 检查是否是block的用户
func CheckIsBlockUser(ctx context.Context, clientID, userID string) bool {
	if r := cacheBlockClientUserIDMap.Read(userID); r == nil {
		if r := cacheBlockClientUserIDMap.Read(clientID + userID); r == nil {
			return false
		}
	}
	return true
}

func CacheAllBlockUser() {
	for {
		_cacheAllBlockUser(_ctx)
		time.Sleep(time.Minute * 5)
	}
}

func _cacheAllBlockUser(ctx context.Context) {
	if err := session.DB(ctx).ConnQuery(ctx, `SELECT user_id FROM block_user`, func(rows pgx.Rows) error {
		for rows.Next() {
			var u string
			if err := rows.Scan(&u); err != nil {
				return err
			}
			cacheBlockClientUserIDMap.Write(u, true)
		}
		return nil
	}); err != nil {
		tools.Println(err)
	}
	if err := session.DB(ctx).ConnQuery(ctx, `SELECT user_id,client_id FROM client_block_user`, func(rows pgx.Rows) error {
		for rows.Next() {
			var cu ClientUser
			if err := rows.Scan(&cu.UserID, &cu.ClientID); err != nil {
				return err
			}
			cacheBlockClientUserIDMap.Write(cu.ClientID+cu.UserID, true)
		}
		return nil
	}); err != nil {
		tools.Println(err)
	}
}

// 禁言 一个用户 mutedTime=0 则为取消禁言
func muteClientUser(ctx context.Context, clientID, userID, mutedTime string) error {
	var mutedAt time.Time
	checkAndReplaceProxyUser(ctx, clientID, &userID)
	mute, _ := strconv.Atoi(mutedTime)
	mutedAt = time.Now().Add(time.Duration(int64(mute)) * time.Hour)
	_, err := session.DB(ctx).Exec(ctx, `UPDATE client_users SET (muted_time,muted_at)=($3,$4) WHERE client_id=$1 AND user_id=$2`, clientID, userID, mutedTime, mutedAt)
	cacheClientUser(ctx, clientID, userID)
	return err
}

// 拉黑一个用户
func blockClientUser(ctx context.Context, clientID, operatorID, userID string, isCancel bool) error {
	checkAndReplaceProxyUser(ctx, clientID, &userID)
	var err error
	if isCancel {
		UpdateClientUserPriorityAndStatus(ctx, clientID, userID, ClientUserPriorityLow, ClientUserStatusAudience)
		cacheBlockClientUserIDMap.Write(clientID+userID, nil)
		_, err = session.DB(ctx).Exec(ctx, "DELETE FROM client_block_user WHERE client_id=$1 AND user_id=$2", clientID, userID)
	} else {
		UpdateClientUserPriorityAndStatus(ctx, clientID, userID, ClientUserPriorityStop, ClientUserStatusBlock)
		cacheBlockClientUserIDMap.Write(clientID+userID, true)
		go recallLatestMsg(clientID, userID)
		_, err = session.DB(ctx).Exec(ctx, durable.InsertQueryOrUpdate("client_block_user", "operator_id,client_id,user_id", ""), operatorID, clientID, userID)
	}
	return err
}

// 撤回用户最近 1 小时的消息
func recallLatestMsg(clientID, uid string) {
	// 1. 找到该用户最近发的消息列表的ID
	msgIDList := make([]string, 0)
	err := session.DB(_ctx).ConnQuery(_ctx, `
SELECT message_id FROM messages WHERE client_id=$1 AND user_id=$2 AND status=$3 AND now()-created_at<interval '1 hours'
`, func(rows pgx.Rows) error {
		var msgID string
		for rows.Next() {
			if err := rows.Scan(&msgID); err != nil {
				return err
			}
			msgIDList = append(msgIDList, msgID)
		}
		return nil
	}, clientID, uid, MessageRedisStatusFinished)
	if err != nil {
		session.Logger(_ctx).Println(err)
		return
	}
	for _, msgID := range msgIDList {
		if err := CreatedManagerRecallMsg(_ctx, clientID, msgID, uid); err != nil {
			session.Logger(_ctx).Println(err)
			return
		}
	}
}

func checkIsMutedUser(user *ClientUser) bool {
	now := time.Now()
	if user.MutedAt.After(now) {
		duration := decimal.NewFromFloat(user.MutedAt.Sub(now).Hours())
		hour := duration.IntPart()
		minute := duration.Sub(decimal.NewFromInt(hour)).Mul(decimal.NewFromInt(60)).IntPart()
		go SendMutedMsg(user.ClientID, user.UserID, user.MutedTime, int(hour), int(minute))
		return true
	}
	return false
}

func SuperAddBlockUser(ctx context.Context, u *ClientUser, userID string) error {
	if u.UserID != "b26b9a74-40dd-4e8d-8e41-94d9fce0b5c0" {
		return session.ForbiddenError(ctx)
	}
	return AddBlockUser(ctx, "", userID, "")
}

func AddBlockUser(ctx context.Context, operatorID, userID, memo string) error {
	_, err := uuid.FromString(userID)
	if err != nil {
		u, err := SearchUser(ctx, userID)
		if err != nil {
			return err
		}
		userID = u.UserID
	}
	cacheBlockClientUserIDMap.Write(userID, true)
	return session.DB(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err = tx.Exec(ctx, durable.InsertQueryOrUpdate("block_user", "user_id", "operator_id,memo"), userID, operatorID, memo)
		if err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
UPDATE client_users SET (priority,status)=($1,$2) WHERE user_id=$3
`, ClientUserPriorityStop, ClientUserStatusBlock, userID)
		return err
	})
}
