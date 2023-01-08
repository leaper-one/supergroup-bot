package statistic

import (
	"context"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

type DailyDataResp struct {
	List     []*models.DailyData `json:"list"`
	Today    *models.DailyData   `json:"today"`
	HighUser int64               `json:"high_user"`
}

func GetDailyDataByClientID(ctx context.Context, u *models.ClientUser) (*DailyDataResp, error) {
	if !common.CheckIsAdmin(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	res := DailyDataResp{
		List:     make([]*models.DailyData, 0),
		Today:    nil,
		HighUser: 0,
	}
	if err := session.DB(ctx).Table("daily_data").
		Select("to_char(date, 'YYYY-MM-DD') date, users, messages, active_users").
		Where("client_id = ?", u.ClientID).
		Order("date").
		Scan(&res.List).Error; err != nil {
		return nil, err
	}

	var d models.DailyData
	startAt, _ := time.Parse("2006-1-2", time.Now().Format("2006-1-2"))
	if err := session.DB(ctx).Table("client_users").
		Where("client_id = ? AND created_at > ?", u.ClientID, startAt).
		Count(&d.Users).Error; err != nil {
		tools.Println(err)
		return nil, err
	}

	if err := session.DB(ctx).Table("messages").
		Where("client_id = ? AND created_at > ?", u.ClientID, startAt).
		Count(&d.Messages).Error; err != nil {
		tools.Println(err)
		return nil, err
	}

	if err := session.DB(ctx).Table("client_users").
		Where("client_id = ? AND priority = 1", u.ClientID).
		Count(&res.HighUser).Error; err != nil {
		tools.Println(err)
		return nil, err
	}

	if err := session.DB(ctx).Table("client_users").
		Where(fmt.Sprintf("client_id = ? AND ?-deliver_at<interval '%f %s'", config.NotActiveCheckTime, "hours"), u.ClientID, startAt).
		Count(&d.ActiveUsers).Error; err != nil {
		tools.Println(err)
		return nil, err
	}

	d.Date = time.Now().Format("2006-01-02")
	res.Today = &d
	return &res, nil
}
