package services

import (
	"context"
	"log"

	"github.com/MixinNetwork/supergroup/handlers/live"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
)

type ResetLiveService struct{}

var liveID = ""

func (service *ResetLiveService) Run(ctx context.Context) error {
	var ld models.LiveData
	// 	if err := session.DB(ctx).ConnQuery(ctx, `
	// SELECT live_id, start_at, end_at FROM live_data WHERE live_id=$1
	// `, func(rows pgx.Rows) error {
	// 		for rows.Next() {
	// 			var ld models.LiveData
	// 			err := rows.Scan(&ld.LiveID, &ld.StartAt, &ld.EndAt)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			lds = append(lds, ld)
	// 		}
	// 		return nil
	// 	}, liveID); err != nil {
	// 		return err
	// 	}

	if err := session.DB(ctx).Take(&ld, "live_id=?", liveID).Error; err != nil {
		return err
	}

	l, err := live.GetLiveByID(ctx, ld.LiveID)
	if err != nil {
		return err
	}

	// 1. 获取消息
	var msgViews []*models.Message
	// 	if err := session.DB(ctx).ConnQuery(ctx, `
	// SELECT message_id, category, data, created_at FROM messages
	// WHERE client_id=$1 AND created_at>=$2 AND created_at<=$3
	// `, func(rows pgx.Rows) error {
	// 		for rows.Next() {
	// 			var msgView mixin.MessageView
	// 			err := rows.Scan(&msgView.MessageID, &msgView.Category, &msgView.Data, &msgView.CreatedAt)
	// 			if err != nil {
	// 				return err
	// 			}
	// 			msgViews = append(msgViews, &msgView)
	// 		}
	// 		return nil
	// 	}, l.ClientID, ld.StartAt, ld.EndAt); err != nil {
	// 		return err
	// 	}
	if err := session.DB(ctx).Find(&msgViews, "client_id=? AND created_at>=? AND created_at<=?", l.ClientID, ld.StartAt, ld.EndAt).Error; err != nil {
		return err
	}

	log.Println("msgViews", len(msgViews))
	// session.DB(ctx).Exec(ctx, `DELETE FROM live_replay WHERE live_id=$1`, l.LiveID)

	session.DB(ctx).Delete(&models.LiveReplay{}, "live_id=?", l.LiveID)

	for _, msgView := range msgViews {
		log.Println("msgView", i, msgView.MessageID)
		live.HandleAudioReplay(l.ClientID, msgView)
	}

	session.DB(ctx).Model(&models.LiveReplay{}).Where("client_id=? AND created_at>? AND created_at<?", l.ClientID, ld.StartAt, ld.EndAt).Update("live_id", l.LiveID)

	return nil
}
