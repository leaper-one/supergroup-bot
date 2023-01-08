package broadcast

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

func GetBroadcast(ctx context.Context, u *models.ClientUser) ([]*models.Message, error) {
	broadcasts := make([]*models.Message, 0)

	if err := session.DB(ctx).Table("messages as m").
		Select("m.message_id, m.user_id, m.category, m.data, u.full_name, u.avatar_url, b.status, b.created_at, b.top_at").
		Joins("LEFT JOIN users as u ON m.user_id=u.user_id").
		Joins("LEFT JOIN broadcast as b ON m.message_id=b.message_id").
		Where("m.message_id IN (SELECT message_id FROM broadcast WHERE client_id=?)", u.ClientID).
		Order("b.created_at DESC").
		Find(&broadcasts).Error; err != nil {
		return nil, err
	}
	for _, b := range broadcasts {
		b.Data = string(tools.Base64Decode(b.Data))
	}
	return broadcasts, nil
}

func UpdateBroadcast(ctx context.Context, clientID, msgID string, status int) error {
	return session.DB(ctx).Model(&models.Broadcast{}).
		Where("client_id=? AND message_id=?", clientID, msgID).
		Update("status", status).Error
}
