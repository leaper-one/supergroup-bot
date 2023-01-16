package live

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

func GetLiveReplayByLiveID(ctx context.Context, u *models.ClientUser, liveID, addr string) ([]*models.LiveReplay, error) {
	lrs := make([]*models.LiveReplay, 0)
	if err := session.DB(ctx).Order("created_at").Find(&lrs, "live_id=?", liveID).Error; err != nil {
		return nil, err
	}
	if err := session.DB(ctx).Create(&models.LivePlay{
		LiveID: liveID,
		UserID: u.UserID,
		Addr:   addr,
	}).Error; err != nil {
		tools.Println(err)
	}
	return lrs, nil
}
