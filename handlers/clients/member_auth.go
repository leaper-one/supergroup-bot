package clients

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

func InitClientMemberAuth(ctx context.Context, clientID string) error {
	states := []int{1, 2, 5}
	for _, state := range states {
		if err := models.CreateIgnoreIfExist(ctx, &models.ClientMemberAuth{
			ClientID:        clientID,
			UserStatus:      state,
			PlainText:       true,
			PlainSticker:    true,
			PlainImage:      true,
			PlainVideo:      true,
			PlainPost:       true,
			PlainData:       true,
			PlainLive:       true,
			PlainContact:    true,
			PlainTranscript: true,
			AppCard:         true,
			URL:             true,
			LuckyCoin:       true,
			UpdatedAt:       time.Now(),
		}); err != nil {
			return err
		}
	}
	return nil
}

func GetClientMemberAuth(ctx context.Context, clientID string) (map[int]models.ClientMemberAuth, error) {
	cmas := make([]*models.ClientMemberAuth, 0)
	if err := session.DB(ctx).Find(&cmas, "client_id=?", clientID).Error; err != nil {
		return nil, err
	}
	cmaMap := make(map[int]models.ClientMemberAuth)
	for _, cma := range cmas {
		cmaMap[cma.UserStatus] = *cma
	}
	return cmaMap, nil
}

func UpdateClientMemberAuth(ctx context.Context, u *models.ClientUser, auth models.ClientMemberAuth) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if !checkUserStatusIsValid(auth.UserStatus) {
		return session.BadDataError(ctx)
	}
	auth.PlainText = true
	auth.UpdatedAt = time.Now()
	return session.DB(ctx).Save(&auth).Error
}

func checkUserStatusIsValid(userStatus int) bool {
	return userStatus == models.ClientUserStatusFresh ||
		userStatus == models.ClientUserStatusAudience ||
		userStatus == models.ClientUserStatusLarge
}
