package services

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
)

type ResetLiveService struct{}

var liveID = ""

func (service *ResetLiveService) Run(ctx context.Context) error {
	var lds []common.LiveData
	if err := session.DB(ctx).ConnQuery(ctx, `
SELECT live_id, start_at, end_at FROM live_data WHERE live_id=$1
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var ld common.LiveData
			err := rows.Scan(&ld.LiveID, &ld.StartAt, &ld.EndAt)
			if err != nil {
				return err
			}
			lds = append(lds, ld)
		}
		return nil
	}, liveID); err != nil {
		return err
	}
	log.Println("lds", len(lds))

	for _, ld := range lds {
		live, err := common.GetLiveByID(ctx, ld.LiveID)
		if err != nil {
			return err
		}

		// 1. 获取消息
		var msgViews []*mixin.MessageView
		if err := session.DB(ctx).ConnQuery(ctx, `
SELECT message_id, category, data, created_at FROM messages
WHERE client_id=$1 AND created_at>=$2 AND created_at<=$3
`, func(rows pgx.Rows) error {
			for rows.Next() {
				var msgView mixin.MessageView
				err := rows.Scan(&msgView.MessageID, &msgView.Category, &msgView.Data, &msgView.CreatedAt)
				if err != nil {
					return err
				}
				msgViews = append(msgViews, &msgView)
			}
			return nil
		}, live.ClientID, ld.StartAt, ld.EndAt); err != nil {
			return err
		}
		log.Println("msgViews", len(msgViews))
		session.DB(ctx).Exec(ctx, `DELETE FROM live_replay WHERE live_id=$1`, live.LiveID)
		for _, msgView := range msgViews {
			log.Println("msgView", i, msgView.MessageID)
			common.HandleAudioReplay(live.ClientID, msgView)
		}

		session.DB(ctx).Exec(ctx, `UPDATE live_replay SET live_id=$1 WHERE client_id=$2 AND created_at>$3 AND created_at<$4`, live.LiveID, live.ClientID, ld.StartAt, ld.EndAt)
	}

	return nil
}
