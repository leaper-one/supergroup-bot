package invitation

import (
	"context"
	"errors"
	"log"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type InvitationListResp struct {
	UserID         string          `json:"user_id"`
	AvatarURL      string          `json:"avatar_url"`
	FullName       string          `json:"full_name"`
	IdentityNumber string          `json:"identity_number"`
	Amount         decimal.Decimal `json:"amount"`
	CreatedAt      string          `json:"created_at"`
}

func GetInvitationListByUserID(ctx context.Context, u *models.ClientUser, page int) ([]*InvitationListResp, error) {
	list := make([]*InvitationListResp, 0)
	if page == 0 {
		page = 1
	}
	// TODO...
	if err := session.DB(ctx).Raw(`
SELECT a.invitee_id,a.amount,u.full_name,u.identity_number,u.avatar_url,to_char(i.created_at, 'YYYY/MM/DD') FROM
	(SELECT invitee_id, COALESCE(SUM(amount::int),0) as amount FROM invitation_power_record
	WHERE inviter_id = ?
	GROUP BY invitee_id) as a
LEFT JOIN users u ON u.user_id=a.invitee_id
LEFT JOIN invitation i ON i.invitee_id=a.invitee_id
ORDER BY i.created_at DESC
	`, u.UserID).Scan(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

type InviteDataResp struct {
	Code  string `json:"code"`
	Count int64  `json:"count"`
	Power int64  `json:"power"`
}

func GetInviteDataByUserID(ctx context.Context, userID string) (*InviteDataResp, error) {
	i := InviteDataResp{}
	err := session.DB(ctx).Table("invitation").Select("invite_code").Where("invitee_id=?", userID).Scan(&i.Code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			i.Code, err = common.CreateInvitation(ctx, userID, "", "")
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	i.Count = common.GetInviteCountByUserID(ctx, userID)
	if err := session.DB(ctx).Table("invitation_power_record").
		Where("inviter_id=?", userID).
		Select("COALESCE(SUM(amount::int),0)").
		Scan(&i.Power).Error; err != nil {
		return nil, err
	}
	return &i, nil
}
func GetLiquidityMiningRespByID(ctx context.Context, u *models.ClientUser, id string) (*models.LiquidityMining, error) {
	var m models.LiquidityMining
	err := session.DB(ctx).Where("mining_id=?", id).First(&m).Error
	if err != nil {
		return nil, err
	}
	a, err := common.GetAssetByID(ctx, nil, m.AssetID)
	if err != nil {
		return nil, err
	}
	m.Symbol = a.Symbol
	// 如果没有token则跳授权页
	m.Status = models.LiquidityMiningStatusAuth

	rewardAsset, err := common.GetAssetByID(ctx, nil, m.RewardAssetID)
	if err != nil {
		return nil, err
	}
	m.RewardSymbol = rewardAsset.Symbol
	extraAsset, err := common.GetAssetByID(ctx, nil, m.ExtraAssetID)
	if err != nil {
		return nil, err
	}
	m.ExtraSymbol = extraAsset.Symbol
	// 检查token是否有资产权限
	assets, err := common.GetUserAssets(ctx, u)
	if err == nil && len(assets) > 0 {
		// 有授权资产则跳已参与活动页面
		m.Status = models.LiquidityMiningStatusPending
		lpAssets, err := common.GetClientAssetLPCheckMapByID(ctx, u.ClientID)
		if err != nil {
			return nil, err
		}
		for _, a := range assets {
			if _, ok := lpAssets[a.AssetID]; ok {
				if a.Balance.GreaterThan(decimal.Zero) {
					m.Status = models.LiquidityMiningStatusDone
					break
				}
			}
		}
		if m.Status == models.LiquidityMiningStatusDone {
			// 添加到已参与活动用户
			var u models.LiquidityMiningUser
			err := session.DB(ctx).Where("mining_id=? AND user_id=?", m.MiningID, u.UserID).First(&u).Error
			if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
				if err := session.DB(ctx).Create(&models.LiquidityMiningUser{
					MiningID: m.MiningID,
					UserID:   u.UserID,
				}).Error; err != nil {
					log.Println(err)
					return nil, err
				}
			}
		}
	}
	return &m, nil
}
