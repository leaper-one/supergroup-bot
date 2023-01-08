package message

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

func GetPendingMessageByClientID(ctx context.Context, clientID string) ([]*models.Message, error) {
	ms := make([]*models.Message, 0)
	err := session.DB(ctx).Order("created_at").Find(&ms, "client_id=? AND status=?", clientID, models.MessageStatusPending).Error
	return ms, err
}
