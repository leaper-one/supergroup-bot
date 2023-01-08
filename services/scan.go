package services

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup/handlers/live"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
)

type ScanService struct{}

func (service *ScanService) Run(ctx context.Context) error {
	var lds []models.LiveData
	if err := session.DB(ctx).ConnQuery(ctx, `
SELECT live_id, start_at, end_at FROM live_data WHERE live_id='ca54cd47-e545-4738-af98-7cc5fdadd9be'
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var ld models.LiveData
			err := rows.Scan(&ld.LiveID, &ld.StartAt, &ld.EndAt)
			if err != nil {
				return err
			}
			lds = append(lds, ld)
		}
		return nil
	}); err != nil {
		return err
	}
	log.Println("lds", len(lds))

	for _, ld := range lds {
		l, err := live.GetLiveByID(ctx, ld.LiveID)
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
		}, l.ClientID, ld.StartAt, ld.EndAt); err != nil {
			return err
		}
		log.Println("msgViews", len(msgViews))
		session.DB(ctx).Exec(ctx, `DELETE FROM live_replay WHERE live_id=$1`, l.LiveID)
		for _, msgView := range msgViews {
			log.Println("msgView", i, msgView.MessageID)
			live.HandleAudioReplay(l.ClientID, msgView)
		}

		session.DB(ctx).Exec(ctx, `UPDATE live_replay SET live_id=$1 WHERE client_id=$2 AND created_at>$3 AND created_at<$4`, l.LiveID, l.ClientID, ld.StartAt, ld.EndAt)
	}

	return nil
}
