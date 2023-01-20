package common

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/go-redis/redis/v8"
)

func DeleteDistributeMsgByClientID(ctx context.Context, clientID string) {
	if err := session.DB(ctx).Table("messages").
		Where("client_id=? AND status=?", clientID, models.MessageStatusPending).
		Update("status", models.MessageStatusRemoveMsg).Error; err != nil {
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
			msg, err := GetOriginMsgFromRedisResult(res)
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

func GetOriginMsgFromRedisResult(res string) (*models.DistributeMessage, error) {
	tmp := strings.Split(res, ",")
	if len(tmp) != 3 {
		tools.Println("invalid msg_origin_idx:", res)
		return nil, session.BadDataError(models.Ctx)
	}
	status, err := strconv.Atoi(tmp[2])
	if err != nil {
		return nil, err
	}
	return &models.DistributeMessage{
		OriginMessageID: tmp[0],
		UserID:          tmp[1],
		Status:          status,
	}, nil
}
