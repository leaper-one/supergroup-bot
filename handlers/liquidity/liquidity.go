package liquidity

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type liquidityResp struct {
	Info            *models.Liquidity         `json:"info"`
	List            []*models.LiquidityDetail `json:"list"`
	YesterdayAmount decimal.Decimal           `json:"yesterday_amount"`
	IsJoin          bool                      `json:"is_join"`
	Scope           string                    `json:"scope"`
}

// 获取活动页面详情
func GetLiquidityInfo(ctx context.Context, u *models.ClientUser, id string) (*liquidityResp, error) {
	info, err := GetLiquidityByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var list []*models.LiquidityDetail
	if err := session.DB(ctx).Order("idx").Find(&list, "liquidity_id = ?", id).Error; err != nil {
		return nil, err
	}
	var yesterdayAmount decimal.Decimal
	if err := session.DB(ctx).Table("liquidity_snapshot").
		Select("lp_amount").
		Where("user_id = ? AND date = CURRENT_DATE-1", u.UserID).
		Scan(&yesterdayAmount).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	return &liquidityResp{
		Info:            info,
		List:            list,
		YesterdayAmount: yesterdayAmount,
		IsJoin:          checkUserIsJoinLiquidity(ctx, id, u.UserID),
		Scope:           u.Scope,
	}, nil
}

func GetLiquidityByID(ctx context.Context, id string) (*models.Liquidity, error) {
	var info models.Liquidity
	if err := session.DB(ctx).Take(&info, "liquidity_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &info, nil
}

func checkUserIsJoinLiquidity(ctx context.Context, lid, uid string) bool {
	var join int64
	if err := session.DB(ctx).Table("liquidity_user").
		Where("liquidity_id = ? AND user_id = ?", lid, uid).
		Count(&join).Error; err != nil {
		tools.Println(err)
		return false
	}
	return join == 1
}

// 参与
func PostLiquidity(ctx context.Context, u *models.ClientUser, id string) (string, error) {
	if time.Now().UTC().Day() != 1 {
		return "miss", nil
	}
	l, err := GetLiquidityByID(ctx, id)
	if err != nil {
		return "", err
	}
	asset, err := common.GetUserAsset(ctx, u, l.AssetIDs)
	if err != nil {
		return "", err
	}
	if _, err := common.GetUserSnapshots(ctx, u, "", time.Now(), "", 1); err != nil {
		return "", err
	}
	if checkUserIsJoinLiquidity(ctx, id, u.UserID) {
		return "success", nil
	}
	if err := session.DB(ctx).Create(&models.LiquidityUser{
		LiquidityID: id,
		UserID:      u.UserID,
	}).Error; err != nil {
		return "", err
	}

	if asset.Balance.LessThan(l.MinAmount) {
		return "limit", nil
	}
	return "success", nil
}

type liquidityRecord struct {
	Duration string                      `json:"duration"`
	Status   string                      `json:"status"`
	List     []*models.LiquiditySnapshot `json:"list"`
}

func GetLiquiditySnapshots(ctx context.Context, u *models.ClientUser, id string) ([]*liquidityRecord, error) {
	var lts []*models.LiquidityTx
	if err := session.DB(ctx).Order("month DESC").
		Find(&lts, "user_id = ? AND liquidity_id = ?", u.UserID, id).Error; err != nil {
		return nil, err
	}
	res := make([]*liquidityRecord, 0)
	for _, lt := range lts {
		startAt := lt.Month.Format("2006.01.02")
		endAt := lt.Month.AddDate(0, 1, -1).Format("2006.01.02")
		item := &liquidityRecord{
			Duration: fmt.Sprintf("%s-%s", startAt, endAt),
			Status:   lt.Status,
		}
		var lss []*models.LiquiditySnapshot
		if err := session.DB(ctx).Table("liquidity_snapshot").
			Select("to_char(date, 'YYYY-MM-DD') date,lp_amount,lp_symbol").
			Order("date DESC").
			Find(&lss, "user_id = ? AND liquidity_id = ? AND date >= ? AND date <= ?",
				u.UserID, id, lt.Month, lt.Month.AddDate(0, 1, -1)).Error; err != nil {
			return nil, err
		}
		item.List = lss
		res = append(res, item)
	}
	return res, nil
}
