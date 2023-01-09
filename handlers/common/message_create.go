package common

import (
	"context"
	"encoding/json"
	"math/rand"
	"strconv"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

func CreatedManagerRecallMsg(ctx context.Context, clientID string, msgID, uid string) error {
	dataByte, _ := json.Marshal(map[string]string{"message_id": msgID})

	if err := CreateMessage(ctx, clientID, &mixin.MessageView{
		UserID:    uid,
		MessageID: tools.GetUUID(),
		Category:  mixin.MessageCategoryMessageRecall,
		Data:      tools.Base64Encode(dataByte),
	}, models.MessageStatusPending); err != nil {
		tools.Println(err)
	}

	return nil
}

func getShardID(clientID, userID string) string {
	shardID := ClientShardIDMap[clientID][userID]
	if shardID == "" {
		shardID = strconv.Itoa(rand.Intn(int(config.MessageShardSize)))
	}
	return shardID
}

var ClientShardIDMap = make(map[string]map[string]string)

func GetQuoteMsgIDUserIDMapByOriginMsgIDFromRedis(ctx context.Context, originMsgID string) (map[string]string, error) {
	recallMsgIDMap := make(map[string]string)
	resList, err := session.Redis(ctx).QSMembers(ctx, "origin_msg_idx:"+originMsgID)
	if err != nil {
		return nil, err
	}
	for _, res := range resList {
		msg, err := getMsgOriginFromRedisResult(res)
		if err != nil {
			continue
		}
		recallMsgIDMap[msg.UserID] = msg.MessageID
	}
	return recallMsgIDMap, nil
}
