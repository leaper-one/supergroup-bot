package services

import (
	"context"

	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	msgList, err := session.Redis(ctx).ZRangeWithScores(ctx, "s_msg:d419d2b0-5c20-4dd7-9a5c-177375c249b8:0", 0, -1).Result()
	if err != nil {
		return err
	}
	tools.PrintJson(msgList)
	idList := make([]interface{}, 0, len(msgList))
	for _, z := range msgList {
		if s, ok := z.Member.(string); ok {
			msg, err := session.Redis(ctx).HGetAll(ctx, "d_msg:d419d2b0-5c20-4dd7-9a5c-177375c249b8:"+s).Result()
			if err != nil {
				return err
			}
			idList = append(idList, s)
			tools.PrintJson(msg)
		}
	}
	// session.Redis(ctx).ZRem(ctx, "s_msg:d419d2b0-5c20-4dd7-9a5c-177375c249b8:0", idList...)
	return nil
}
