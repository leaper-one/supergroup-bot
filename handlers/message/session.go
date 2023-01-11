package message

import (
	"context"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"gorm.io/gorm"
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

func SyncSession(ctx context.Context, clientID string, sessions []*models.Session) error {
	if len(sessions) < 1 {
		return nil
	}
	var userIDs []string
	for _, s := range sessions {
		userIDs = append(userIDs, s.UserID)
	}
	return models.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Session{}, "client_id = ? AND user_id IN ?", clientID, userIDs).Error; err != nil {
			return err
		}
		dataInsert := make([]*models.Session, 0)
		repeatIds := make(map[string]bool)
		for _, s := range sessions {
			if repeatIds[s.ClientID+s.UserID+s.SessionID] {
				continue
			}
			repeatIds[s.ClientID+s.UserID+s.SessionID] = true
			dataInsert = append(dataInsert, s)
		}
		return tx.Create(&dataInsert).Error
	})
}

type SimpleUser struct {
	Category string
	Sessions []*models.Session
}

func ReadSessionSetByUsers(ctx context.Context, clientID string, userIDs []string) (map[string]*SimpleUser, error) {
	sessions := make([]*models.Session, 0)
	err := session.DB(ctx).Where("client_id = ? AND user_id IN ?", clientID, userIDs).Find(&sessions).Error
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}

	set := make(map[string]*SimpleUser)
	for _, s := range sessions {
		if set[s.UserID] == nil {
			su := &SimpleUser{
				Category: UserCategoryEncrypted,
				Sessions: []*models.Session{s},
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
		set[s.UserID].Sessions = append(set[s.UserID].Sessions, s)
	}
	return set, err
}

func GenerateUserChecksum(sessions []*models.Session) string {
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
