package message

import (
	"context"

	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

func getMsgByClientIDAndMessageID(ctx context.Context, clientID, msgID string) (*models.Message, error) {
	var m models.Message
	err := session.DB(ctx).Take(&m, "client_id = ? AND message_id = ?", clientID, msgID).Error
	return &m, err
}
