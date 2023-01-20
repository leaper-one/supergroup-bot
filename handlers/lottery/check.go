package lottery

import (
	"context"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

var ignoreDoubleList = make(map[string]bool)

func checkIsIgnoreDoubleClaim(ctx context.Context, clientID string) bool {
	if len(ignoreDoubleList) == 0 {
		ignoreList, err := session.Redis(ctx).QSMembers(ctx, "double_ignore")
		if err != nil {
			tools.Println(err)
			return true
		}
		for _, v := range ignoreList {
			ignoreDoubleList[v] = true
		}
	}
	return ignoreDoubleList[clientID]
}

func CheckUserIsVIP(ctx context.Context, userID string) bool {
	var count int64
	if err := session.DB(ctx).Table("client_users").Where("user_id=? AND status>1", userID).Count(&count).Error; err != nil {
		tools.Println(err)
	}
	return count > 0
}

func checkIsDoubleClaimClient(ctx context.Context, clientID string) bool {
	list, err := getDoubleClaimClientList(ctx)
	if err != nil {
		return false
	}
	for _, v := range list {
		if v.ClientID == clientID {
			return true
		}
	}
	return false
}
