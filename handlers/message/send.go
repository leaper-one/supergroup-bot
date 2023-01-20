package message

import (
	"strings"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
)

func SendForbidMsg(clientID, userID, category string) {
	msg := strings.ReplaceAll(config.Text.Forbid, "{category}", config.Text.Category[category])
	go common.SendClientUserTextMsg(clientID, userID, msg, "")
}
