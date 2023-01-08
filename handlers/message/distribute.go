package message

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
	"github.com/shopspring/decimal"
)

func createDistributeMsgToRedis(ctx context.Context, msgs []*models.DistributeMessage) error {
	if len(msgs) == 0 {
		return nil
	}
	_, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
		for _, msg := range msgs {
			dMsgKey := fmt.Sprintf("d_msg:%s:%s", msg.ClientID, msg.MessageID)
			if err := p.HSet(ctx, dMsgKey,
				map[string]interface{}{
					"user_id":           msg.UserID,
					"origin_message_id": msg.OriginMessageID,
					"message_id":        msg.MessageID,
					"quote_message_id":  msg.QuoteMessageID,
					"data":              msg.Data,
					"representative_id": msg.RepresentativeID,
				},
			).Err(); err != nil {
				tools.Println(err)
				return err
			}
			if err := p.PExpire(ctx, dMsgKey, time.Hour*24).Err(); err != nil {
				return err
			}
			if msg.Status == models.DistributeMessageStatusPending {
				score := msg.CreatedAt.UnixNano()
				if msg.Level == models.ClientUserPriorityHigh {
					score = score / 2
				}
				if err := p.ZAdd(ctx, fmt.Sprintf("s_msg:%s:%s", msg.ClientID, getShardID(msg.ClientID, msg.UserID)), &redis.Z{
					Score:  float64(score),
					Member: msg.MessageID,
				}).Err(); err != nil {
					tools.Println(err)
					return err
				}
			} else {
				if err := p.PExpire(ctx, dMsgKey, config.QuoteMsgSavedTime).Err(); err != nil {
					return err
				}
			}
			if err := common.BuildOriginMsgAndMsgIndex(ctx, p, msg); err != nil {
				return err
			}
		}
		if msgs[0].Status == models.DistributeMessageStatusPending {
			lKey := fmt.Sprintf("l_msg:%s", msgs[0].OriginMessageID)
			cmcKey := fmt.Sprintf("client_msg_count:%s:%s", msgs[0].ClientID, tools.GetMinuteTime(time.Now()))
			if err := p.IncrBy(ctx, lKey, int64(len(msgs))).Err(); err != nil {
				return err
			}
			if err := p.IncrBy(ctx, cmcKey, int64(len(msgs))).Err(); err != nil {
				return err
			}
			if err := p.PExpire(ctx, lKey, time.Hour*24).Err(); err != nil {
				return err
			}
			if err := p.PExpire(ctx, cmcKey, time.Minute*2).Err(); err != nil {
				return err
			}
		}
		return nil
	})
	if msgs[0].Status == models.DistributeMessageStatusPending {
		if err := session.Redis(ctx).QPublish(ctx, "distribute", msgs[0].ClientID); err != nil {
			return err
		}
	}
	return err
}

func getShardID(clientID, userID string) string {
	shardID := ClientShardIDMap[clientID][userID]
	if shardID == "" {
		shardID = strconv.Itoa(rand.Intn(int(config.MessageShardSize)))
	}
	return shardID
}

var ClientShardIDMap = make(map[string]map[string]string)

func InitShardID(ctx context.Context, clientID string) error {
	ClientShardIDMap[clientID] = make(map[string]string)
	// 1. 获取有多少个协程，就分配多少个编号
	count := decimal.NewFromInt(config.MessageShardSize)
	// 2. 获取优先级高/低的所有用户，及高低比例
	high, low, err := GetClientUserReceived(ctx, clientID)
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

func GetClientUserReceived(ctx context.Context, clientID string) ([]string, []string, error) {
	userList, err := common.GetDistributeMsgUser(ctx, clientID, false, false)
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
