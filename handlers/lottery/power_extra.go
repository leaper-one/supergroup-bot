package lottery

import (
	"context"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
)

func getDoubleClaimClientList(ctx context.Context) ([]*models.Client, error) {
	clientList := make([]*models.Client, 0)
	if err := session.DB(ctx).Table("power_extra as pe").
		Select("c.name, c.description, c.icon_url, c.identity_number, c.client_id, c.created_at, pe.description as welcome").
		Joins("LEFT JOIN client c ON pe.client_id = c.client_id").
		Where("pe.start_at <= CURRENT_DATE AND pe.end_at >= CURRENT_DATE").
		Scan(&clientList).Error; err != nil {
		tools.Println(err)
		return nil, err
	}
	return clientList, nil
}

func needAddExtraPower(ctx context.Context, userID string) bool {
	passDays := int(time.Now().Weekday())
	if config.Config.Lang == "zh" {
		if passDays == 0 {
			passDays = 7
		}
		if passDays < 5 {
			return false
		}

	} else {
		if passDays < 4 {
			return false
		}
		passDays++
	}
	var count int64
	if err := session.DB(ctx).Table("claim").
		Where(fmt.Sprintf("user_id = ? AND date >= CURRENT_DATE - %d", passDays), userID).
		Count(&count).Error; err != nil {
		tools.Println(err)
		return false
	}
	return count == 4
}
