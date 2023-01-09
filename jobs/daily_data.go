package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/clients"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/robfig/cron/v3"
)

func StartDailyDataJob() {
	c := cron.New(cron.WithLocation(time.UTC))
	ctx := models.Ctx
	_, err := c.AddFunc("16 0 * * *", func() {
		cs, err := clients.GetAllClient(ctx)
		if err != nil {
			tools.Println(err)
			return
		}
		now, _ := time.Parse("2006-1-2", time.Now().Format("2006-1-2"))
		yesterday := now.Add(-24 * time.Hour)
		for _, clientID := range cs {
			dd, err := statisticsGroupDailyData(ctx, clientID, yesterday)
			if err != nil {
				tools.Println(err)
				continue
			}
			if err := session.DB(ctx).Save(&dd); err != nil {
				tools.Println(err)
			}
		}
	})
	if err != nil {
		tools.Println(err)
		tools.SendMsgToDeveloper("定时任务。。。出问题了。。。")
		return
	}
	c.Start()
}
func statisticsGroupDailyData(ctx context.Context, clientID string, startAt time.Time) (models.DailyData, error) {
	endAt := startAt.Add(time.Hour * 24)
	var users, messages, activeUsers int64
	if err := session.DB(ctx).Table("client_users").
		Where(`client_id = ? 
AND created_at > ? AND created_at < ? 
AND priority IN (1,2) 
AND status IN (1,2,3,5,8,9)`, clientID, startAt, endAt).
		Count(&users).Error; err != nil {
		return models.DailyData{}, err
	}

	if err := session.DB(ctx).Table("messages").
		Where(`client_id = ? AND created_at > ? AND created_at < ?`, clientID, startAt, endAt).
		Count(&messages).Error; err != nil {
		return models.DailyData{}, err
	}
	if err := session.DB(ctx).Table("client_users").
		Where(fmt.Sprintf(`client_id = ?
AND $1-deliver_at<interval '%f %s'
AND created_at<$1`, config.NotActiveCheckTime, "hours"), clientID, endAt).
		Count(&activeUsers).Error; err != nil {
		return models.DailyData{}, err
	}

	return models.DailyData{
		ClientID:    clientID,
		Users:       users,
		Messages:    messages,
		ActiveUsers: activeUsers,
		Date:        startAt.Format("2006-1-2"),
	}, nil
}
