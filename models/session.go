package models

import (
	"context"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
)

const session_DDL = `
CREATE TABLE IF NOT EXISTS session (
	client_id          VARCHAR(36) NOT NULL,
  user_id            VARCHAR(36) NOT NULL,
  session_id         VARCHAR(36) NOT NULL,
  public_key         VARCHAR(128) NOT NULL,
  PRIMARY KEY(client_id, user_id, session_id)
);
`

const (
	UserCategoryPlain     = "PLAIN"
	UserCategoryEncrypted = "ENCRYPTED"
)

type Session struct {
	ClientID  string
	UserID    string
	SessionID string
	PublicKey string
}

var sessionCols = []string{"client_id", "user_id", "session_id", "public_key"}

func SyncSession(ctx context.Context, clientID string, sessions []*Session) error {
	if len(sessions) < 1 {
		return nil
	}
	var userIDs []string
	for _, s := range sessions {
		userIDs = append(userIDs, s.UserID)
	}

	err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "DELETE FROM session WHERE client_id=$1 AND user_id=ANY($2)", clientID, userIDs)
		if err != nil {
			return err
		}
		dataInsert := make([][]interface{}, 0)
		repeatIds := make(map[string]bool)
		for _, s := range sessions {
			if repeatIds[s.ClientID+s.UserID+s.SessionID] {
				continue
			}
			rows := []interface{}{clientID, s.UserID, s.SessionID, s.PublicKey}
			dataInsert = append(dataInsert, rows)
			if err != nil {
				return err
			}
			repeatIds[s.ClientID+s.UserID+s.SessionID] = true
		}
		var ident = pgx.Identifier{"session"}
		_, err = session.Database(ctx).CopyFrom(ctx, ident, sessionCols, pgx.CopyFromRows(dataInsert))
		return err
	})
	if err != nil {
		return session.TransactionError(ctx, err)
	}
	return nil
}

type SimpleUser struct {
	Category string
	Sessions []*Session
}

func ReadSessionSetByUsers(ctx context.Context, clientID string, userIDs []string) (map[string]*SimpleUser, error) {
	set := make(map[string]*SimpleUser)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT user_id,session_id,public_key 
FROM session 
WHERE client_id=$1 
AND user_id=ANY($2)
	`, func(rows pgx.Rows) error {
		for rows.Next() {
			var s Session
			if err := rows.Scan(&s.UserID, &s.SessionID, &s.PublicKey); err != nil {
				return session.TransactionError(ctx, err)
			}
			if set[s.UserID] == nil {
				su := &SimpleUser{
					Category: UserCategoryEncrypted,
					Sessions: []*Session{&s},
				}
				if s.PublicKey == "" {
					su.Category = UserCategoryPlain
				}
				set[s.UserID] = su
				continue
			}
			if s.PublicKey == "" {
				set[s.UserID].Category = UserCategoryPlain
			}
			set[s.UserID].Sessions = append(set[s.UserID].Sessions, &s)
		}
		return nil
	}, clientID, userIDs)
	return set, err
}

func GenerateUserChecksum(sessions []*Session) string {
	if len(sessions) < 1 {
		return ""
	}
	ss := make([]*bot.Session, len(sessions))
	for i, s := range sessions {
		ss[i] = &bot.Session{
			UserID:    s.UserID,
			SessionID: s.SessionID,
			PublicKey: s.PublicKey,
		}
	}
	return bot.GenerateUserChecksum(ss)
}
