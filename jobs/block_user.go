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
	_cacheAllBlockUser(models.Ctx)
	go func() {
		for {
			time.Sleep(time.Minute * 5)
			_cacheAllBlockUser(models.Ctx)
		}
	}()
}

func _cacheAllBlockUser(ctx context.Context) {
	var blockUserIDs []string
	if err := session.DB(ctx).Table("block_user").Pluck("user_id", &blockUserIDs).Error; err != nil {
		tools.Println(err)
	}

	var clientBlockUsers []*models.ClientBlockUser
	if err := session.DB(ctx).Table("client_block_user").Find(&clientBlockUsers).Error; err != nil {
		tools.Println(err)
	}
	common.CacheBlockClientUserIDMap.Lock()
	defer common.CacheBlockClientUserIDMap.Unlock()
	for _, u := range blockUserIDs {
		common.CacheBlockClientUserIDMap.V[u] = true
	}
	for _, cu := range clientBlockUsers {
		common.CacheBlockClientUserIDMap.V[cu.ClientID+cu.UserID] = true
	}
}
