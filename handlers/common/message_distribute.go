package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
)

// 获取指定的消息
var cacheMessageData = tools.NewMutex()

func GetDistributeMessageIDMapByOriginMsgID(ctx context.Context, clientID, originMsgID string) (map[string]string, string, error) {
	// 2. 用 origin_message_id 和 user_id 找出 对应会话 里的 message_id ，这个 message_id 就是要 quote 的 id
	mapList, err := GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx, originMsgID)
	if err != nil {
		tools.Println(err)
		return nil, "", err
	}
	msg, err := getMsgByClientIDAndMessageID(ctx, clientID, originMsgID)
	if err == nil {
		mapList[msg.UserID] = originMsgID
		return mapList, msg.UserID, nil
	}
	return mapList, "", nil
}

func DeleteDistributeMsgByClientID(ctx context.Context, clientID string) {
	if err := session.DB(ctx).Table("messages").
		Where("client_id=? AND status=?", clientID, MessageStatusPending).
		Update("status", MessageStatusRemoveMsg).Error; err != nil {
		tools.Println(err)
		DeleteDistributeMsgByClientID(ctx, clientID)
		return
	}

	for i := 0; i < int(config.MessageShardSize); i++ {
		_, err := session.Redis(ctx).QPipelined(ctx, func(p redis.Pipeliner) error {
			if err := p.Del(ctx, fmt.Sprintf("s_msg:%s:%d", clientID, i)).Err(); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			tools.Println(err)
		}
	}
	go func() {
		ctx = models.Ctx
		dMsgs, err := session.Redis(ctx).QKeys(ctx, fmt.Sprintf("d_msg:%s:*", clientID))
		if err != nil {
			return
		}
		if len(dMsgs) > 0 {
			if err := session.Redis(ctx).W.Unlink(ctx, dMsgs...).Err(); err != nil {
				return
			}
		}
		oMsgIDs := make(map[string]bool)
		for _, res := range dMsgs {
			msgID := strings.Split(res, ":")[2]
			res, err := session.Redis(ctx).QGet(ctx, "msg_origin_idx:"+msgID).Result()
			if err != nil {
				return
			}
			msg, err := getOriginMsgFromRedisResult(res)
			if err != nil {
				return
			}
			oMsgIDs[msg.OriginMessageID] = true
			if err := session.Redis(ctx).W.Unlink(ctx, "msg_origin_idx:"+msgID).Err(); err != nil {
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
		for msgID := range oMsgIDs {
			if err := session.Redis(ctx).W.Unlink(ctx, "l_msg:"+msgID).Err(); err != nil {
				return
			}
			if err := session.Redis(ctx).W.Unlink(ctx, "origin_msg_idx:"+msgID).Err(); err != nil {
				return
			}
			if err := session.Redis(ctx).W.Unlink(ctx, "msg_status:"+msgID).Err(); err != nil {
				return
			}
			time.Sleep(time.Millisecond * 100)
		}
	}()
}
