package broadcast

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

type Broadcast struct {
	MessageID string `json:"message_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Category  string `json:"category,omitempty"`
	Data      string `json:"data,omitempty"`
	FullName  string `json:"full_name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Status    int    `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	TopAt     string `json:"top_at,omitempty"`
}

func GetBroadcast(ctx context.Context, u *models.ClientUser) ([]*Broadcast, error) {
	broadcasts := make([]*Broadcast, 0)

	if err := session.DB(ctx).Table("messages as m").
		Select("m.message_id, m.user_id, m.category, m.data, u.full_name, u.avatar_url, b.status, b.created_at, b.top_at").
		Joins("LEFT JOIN users as u ON m.user_id=u.user_id").
		Joins("LEFT JOIN broadcast as b ON m.message_id=b.message_id").
		Where("m.message_id IN (SELECT message_id FROM broadcast WHERE client_id=?)", u.ClientID).
		Order("b.created_at DESC").
		Scan(&broadcasts).Error; err != nil {
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
