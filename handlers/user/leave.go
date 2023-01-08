package user

import (
	"context"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
)

func LeaveGroup(ctx context.Context, u *models.ClientUser) error {
	if err := common.UpdateClientUserPart(ctx, u.ClientID, u.UserID, map[string]interface{}{"status": models.ClientUserStatusExit}); err != nil {
		return err
	}
	go common.SendClientUserTextMsg(u.ClientID, u.UserID, config.Text.LeaveGroup, "")
	return nil
}
