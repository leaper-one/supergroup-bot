package message

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"sort"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"gorm.io/gorm"
)

const (
	UserCategoryPlain     = "PLAIN"
	UserCategoryEncrypted = "ENCRYPTED"
)

func InitCacheSession(ctx context.Context) {
	sessions := make([]*models.Session, 0)
	err := session.DB(ctx).Find(&sessions, "client_id=''").Error
	if err != nil {
		tools.Println(err)
		return
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
	sessionCache.Lock()
	defer sessionCache.Unlock()
	for uid, su := range set {
		sessionCache.V[uid] = su
	}
}

func SyncSession(ctx context.Context, sessions []*models.Session) error {
	if len(sessions) < 1 {
		return nil
	}
	var userIDs []string
	for _, s := range sessions {
		userIDs = append(userIDs, s.UserID)
		sessionCache.Delete(s.UserID)
	}
	return models.RunInTransaction(ctx, func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Session{}, "client_id = '' AND user_id IN ?", userIDs).Error; err != nil {
			return err
		}
		dataInsert := make([]*models.Session, 0)
		repeatIds := make(map[string]bool)
		for _, s := range sessions {
			if repeatIds[s.UserID+s.SessionID] {
				continue
			}
			repeatIds[s.UserID+s.SessionID] = true
			dataInsert = append(dataInsert, s)
		}
		return tx.Create(&dataInsert).Error
	})
}

type SimpleUser struct {
	Category string
	Sessions []*models.Session
}

var sessionCache = tools.NewMutex()

func ReadSessionSetByUsers(ctx context.Context, _userIDs []string) (map[string]*SimpleUser, error) {
	set := make(map[string]*SimpleUser)
	userIDs := make([]string, 0)
	for _, uid := range _userIDs {
		su := sessionCache.Read(uid)
		if su != nil {
			set[uid] = su.(*SimpleUser)
		} else {
			userIDs = append(userIDs, uid)
		}
	}

	if len(userIDs) < 1 {
		return set, nil
	}

	sessions := make([]*models.Session, 0)
	err := session.DB(ctx).Where("client_id = '' AND user_id IN ?", userIDs).Find(&sessions).Error
	if err != nil {
		return nil, session.TransactionError(ctx, err)
	}

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

	sessionCache.Lock()
	defer sessionCache.Unlock()
	for _, uid := range userIDs {
		if set[uid] != nil {
			sessionCache.V[uid] = set[uid]
		}
	}
	return set, nil
}

func GenerateUserChecksum(sessions []*models.Session) string {
	if len(sessions) < 1 {
		return ""
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].SessionID < sessions[j].SessionID
	})
	h := md5.New()
	for _, s := range sessions {
		io.WriteString(h, s.SessionID)
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:])
}
