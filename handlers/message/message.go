package message

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

func updateMessageStatus(ctx context.Context, clientID, messageID string, status int) error {
	return session.DB(ctx).Model(&models.Message{}).
		Where("client_id = ? AND message_id = ?", clientID, messageID).
		Update("status", status).Error
}
