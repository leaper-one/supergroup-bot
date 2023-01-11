package common

import (
	"context"
	"math/rand"
	"strconv"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/shopspring/decimal"
)

var ClientShardIDMap = make(map[string]map[string]string)

func getShardID(clientID, userID string) string {
	shardID := ClientShardIDMap[clientID][userID]
	if shardID == "" {
		shardID = strconv.Itoa(rand.Intn(int(config.MessageShardSize)))
	}
	return shardID
}

func InitShardID(ctx context.Context, clientID string) error {
	ClientShardIDMap[clientID] = make(map[string]string)
	// 1. 获取有多少个协程，就分配多少个编号
	count := decimal.NewFromInt(config.MessageShardSize)
	// 2. 获取优先级高/低的所有用户，及高低比例
	high, low, err := getClientUserReceived(ctx, clientID)
	if err != nil {
		return err
	}
	// 每个分组的平均人数
	highCount := int(decimal.NewFromInt(int64(len(high))).Div(count).Ceil().IntPart())
	lowCount := int(decimal.NewFromInt(int64(len(low))).Div(count).Ceil().IntPart())
	// 3. 给这个大群里 每个用户进行 编号
	if highCount < 100 {
		highCount = 100
	}
	if lowCount < 100 {
		lowCount = 100
	}
	for shardID := 0; shardID < int(config.MessageShardSize); shardID++ {
		strShardID := strconv.Itoa(shardID)
		cutCount := 0
		hC := len(high)
		for i := 0; i < highCount; i++ {
			if i == hC {
				break
			}
			cutCount++
			ClientShardIDMap[clientID][high[i]] = strShardID
		}
		if cutCount > 0 {
			high = high[cutCount:]
		}

		cutCount = 0
		lC := len(low)
		for i := 0; i < lowCount; i++ {
			if i == lC {
				break
			}
			cutCount++
			ClientShardIDMap[clientID][low[i]] = strShardID
		}
		if cutCount > 0 {
			low = low[cutCount:]
		}
	}
	return nil
}

func getClientUserReceived(ctx context.Context, clientID string) ([]string, []string, error) {
	userList, err := GetDistributeMsgUser(ctx, clientID, false, false)
	if err != nil {
		return nil, nil, err
	}
	privilegeUserList := make([]string, 0)
	normalUserList := make([]string, 0)
	for _, u := range userList {
		if u.Priority == models.ClientUserPriorityHigh {
			privilegeUserList = append(privilegeUserList, u.UserID)
		} else if u.Priority == models.ClientUserPriorityLow {
			normalUserList = append(normalUserList, u.UserID)
		}
	}
	return privilegeUserList, normalUserList, nil
}
