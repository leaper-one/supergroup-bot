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
	if err := session.DB(ctx).Take(&ld, "live_id=?", liveID).Error; err != nil {
		return err
	}

	l, err := live.GetLiveByID(ctx, ld.LiveID)
	if err != nil {
		return err
	}

	// 1. 获取消息
	var msgViews []*models.Message

	if err := session.DB(ctx).Find(&msgViews, "client_id=? AND created_at>=? AND created_at<=?", l.ClientID, ld.StartAt, ld.EndAt).Error; err != nil {
		return err
	}

	log.Println("msgViews", len(msgViews))
	session.DB(ctx).Delete(&models.LiveReplay{}, "live_id=?", l.LiveID)

	for _, msgView := range msgViews {
		log.Println("msgView", i, msgView.MessageID)
		live.HandleAudioReplay(l.ClientID, msgView)
	}

	session.DB(ctx).Model(&models.LiveReplay{}).Where("client_id=? AND created_at>? AND created_at<?", l.ClientID, ld.StartAt, ld.EndAt).Update("live_id", l.LiveID)

	return nil
}
