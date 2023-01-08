package live

import (
	"time"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

func handleStatistics(l *models.Live, startAt, endAt time.Time) error {
	ctx := models.Ctx
	var readCount, deliverCount, msgCount, userCount int64
	if err := session.DB(ctx).Model(&models.ClientUser{}).
		Where("client_id = ? AND read_at > ?", l.ClientID, startAt).
		Count(&readCount).Error; err != nil {
		return err
	}
	if err := session.DB(ctx).Model(&models.ClientUser{}).
		Where("client_id = ? AND deliver_at > ?", l.ClientID, startAt).
		Count(&deliverCount).Error; err != nil {
		return err
	}
	if err := session.DB(ctx).Model(&models.Message{}).
		Where("client_id = ? AND created_at > ? AND created_at < ?", l.ClientID, startAt, endAt).
		Count(&msgCount).Error; err != nil {
		return err
	}
	if err := session.DB(ctx).Model(&models.Message{}).
		Where("client_id = ? AND created_at > ? AND created_at < ?", l.ClientID, startAt, endAt).
		Select("DISTINCT(user_id)").
		Count(&userCount).Error; err != nil {
		return err
	}
	err := session.DB(ctx).Model(&models.LiveData{}).
		Where("live_id = ?", l.LiveID).
		Updates(map[string]interface{}{
			"read_count":    readCount,
			"deliver_count": deliverCount,
			"msg_count":     msgCount,
			"user_count":    userCount,
			"end_at":        endAt,
		}).Error
	return err
}
