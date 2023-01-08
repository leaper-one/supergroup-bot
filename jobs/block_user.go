package jobs

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

func CacheAllBlockUser() {
	for {
		_cacheAllBlockUser(models.Ctx)
		time.Sleep(time.Minute * 5)
	}
}

func _cacheAllBlockUser(ctx context.Context) {
	var blockUserIDs []string
	if err := session.DB(ctx).Table("block_user").Pluck("user_id", &blockUserIDs).Error; err != nil {
		tools.Println(err)
	}

	for _, u := range blockUserIDs {
		common.CacheBlockClientUserIDMap.Write(u, true)
	}

	var clientBlockUsers []*models.ClientBlockUser
	if err := session.DB(ctx).Table("client_block_user").Find(&clientBlockUsers).Error; err != nil {
		tools.Println(err)
	}

	for _, cu := range clientBlockUsers {
		common.CacheBlockClientUserIDMap.Write(cu.ClientID+cu.UserID, true)
	}
}
