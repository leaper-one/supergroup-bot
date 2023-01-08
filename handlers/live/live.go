package live

import (
	"context"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
)

func UpdateLive(ctx context.Context, u *models.ClientUser, l models.Live) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if l.LiveID == "" {
		l.LiveID = tools.GetUUID()
	}
	err := session.DB(ctx).Save(&l).Error
	return err
}

func GetLiveByID(ctx context.Context, liveID string) (*models.Live, error) {
	var l models.Live
	err := session.DB(ctx).Take(&l, "live_id = ?", liveID).Error
	return &l, err
}

func GetLivesByClientID(ctx context.Context, u *models.ClientUser) ([]*models.Live, error) {
	ls := make([]*models.Live, 0)
	err := session.DB(ctx).Find(&ls, "client_id = ?", u.ClientID).Error
	return ls, err
}

func StartLive(ctx context.Context, u *models.ClientUser, liveID, url string) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	l, err := GetLiveByID(ctx, liveID)
	if err != nil {
		return err
	}
	if l.Category == models.LiveCategoryAudioAndImage {
		if err := common.SetClientConversationStatusByIDAndStatus(ctx, l.ClientID, models.ClientConversationStatusAudioLive); err != nil {
			return err
		}
		common.DeleteDistributeMsgByClientID(models.Ctx, u.ClientID)
		go common.SendClientTextMsg(l.ClientID, config.Text.Living, "", false)

	} else {
		go sendLiveCard(u.ClientID, url, liveID)
	}
	return startLive(ctx, l)
}

func StopLive(ctx context.Context, u *models.ClientUser, liveID string) error {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	l, err := GetLiveByID(ctx, liveID)
	if err != nil {
		return err
	}
	if l.Category == models.LiveCategoryAudioAndImage {
		go common.SendClientTextMsg(l.ClientID, config.Text.LiveEnd, "", false)
		if err := common.SetClientConversationStatusByIDAndStatus(ctx, l.ClientID, models.ClientConversationStatusNormal); err != nil {
			return err
		}
	} else {
		go sendLiveCard(u.ClientID, "", liveID)
	}
	return stopLive(ctx, l)
}

func StatLive(ctx context.Context, u *models.ClientUser, liveID string) (*models.LiveData, error) {
	var l models.LiveData
	err := session.DB(ctx).Take(&l, "live_id = ?", liveID).Error
	return &l, err
}

// 视频直播开始
func startLive(ctx context.Context, l *models.Live) error {
	// 直接开始
	if err := session.DB(ctx).Model(&models.LiveData{}).
		Where("live_id=?", l.LiveID).
		Update("start_at", time.Now()).Error; err != nil {
		return err
	}
	return updateLivePart(ctx, l.LiveID, map[string]interface{}{"status": models.LiveStatusLiving})
}

func updateLivePart(ctx context.Context, liveID string, update map[string]interface{}) error {
	return session.DB(ctx).Model(&models.LiveData{}).Where("live_id=?", liveID).Updates(update).Error
}

// 视频直播结束
func stopLive(ctx context.Context, l *models.Live) error {
	var liveData models.LiveData
	if err := session.DB(ctx).Take(&liveData, "live_id = ?", l.LiveID).Error; err != nil {
		return err
	}
	endAt := time.Now()
	if l.Category == models.LiveCategoryAudioAndImage {
		if err := session.DB(ctx).Model(&models.LiveReplay{}).
			Where("client_id = ? AND created_at > ? AND created_at < ?", l.ClientID, liveData.StartAt, endAt).
			Update("live_id", l.LiveID).Error; err != nil {
			return err
		}
	}
	go func() {
		// 统计观看用户。 广播用户。 直播时长。 发言人数。 发言数量
		if err := UpdateClientUserActiveTimeFromRedis(models.Ctx, l.ClientID); err != nil {
			tools.Println(err)
		}
		if err := handleStatistics(l, liveData.StartAt, endAt); err != nil {
			tools.Println(err)
			return
		}
	}()
	return updateLivePart(ctx, l.LiveID, map[string]interface{}{"status": models.LiveStatusFinished})
}

func sendLiveCard(clientID, url, liveID string) error {
	ctx := models.Ctx
	var msg string
	if url == "" {
		msg = config.Text.VideoLiveEnd
		if err := session.DB(ctx).Table("live_replay").
			Where("live_id = ?", liveID).
			Select("data").Scan(&url).Error; err != nil {
			tools.Println(err)
		}
	} else {
		msg = config.Text.VideoLiving
	}
	time.Sleep(2 * time.Minute)
	common.SendClientTextMsg(clientID, msg, "", false)
	client, err := common.GetMixinClientByIDOrHost(ctx, clientID)
	if err != nil {
		return err
	}
	if err := common.SendMessage(ctx, client.Client, &mixin.MessageRequest{
		ConversationID: mixin.UniqueConversationID(clientID, "b523c28b-1946-4b98-a131-e1520780e8af"),
		RecipientID:    "b523c28b-1946-4b98-a131-e1520780e8af",
		MessageID:      tools.GetUUID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           tools.Base64Encode([]byte(url)),
	}, false); err != nil {
		tools.Println(err)
		return err
	}
	if url != "" {
		if err := session.DB(ctx).Save(&models.LiveReplay{
			MessageID: tools.GetUUID(),
			ClientID:  clientID,
			LiveID:    liveID,
			Category:  "",
			Data:      url,
			CreatedAt: time.Now(),
		}).Error; err != nil {
			tools.Println(err)
			return err
		}
	}
	return nil
}

func TopNews(ctx context.Context, u *models.ClientUser, newsID string, isCancel bool) error {
	t := time.Now()
	if isCancel {
		t, _ = time.Parse("2006-1-2", "1970-1-1")
	}
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return session.ForbiddenError(ctx)
	}
	if err := updateLivePart(ctx, newsID, map[string]interface{}{"top_at": t}); err != nil {
		return err
	}
	if err := session.DB(ctx).Model(&models.Broadcast{}).
		Where("message_id = ?", newsID).
		Update("top_at", t).Error; err != nil {
		return err
	}
	return nil
}
