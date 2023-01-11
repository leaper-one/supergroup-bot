package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func HandleTransfer() {
	for {
		handleTransfer(models.Ctx)
		time.Sleep(5 * time.Second)
	}
}

type Transfer struct {
	ClientID   string          `json:"client_id"`
	TraceID    string          `json:"trace_id"`
	AssetID    string          `json:"asset_id"`
	OpponentID string          `json:"opponent_id"`
	Amount     decimal.Decimal `json:"amount"`
	Memo       string          `json:"memo"`
}

func handleTransfer(ctx context.Context) {
	ts := make([]*Transfer, 0)
	if err := session.DB(ctx).Model(&models.Transfer{}).
		Select("client_id,trace_id,asset_id,opponent_id,amount,memo").
		Where("status = ?", models.TransferStatusPending).
		Scan(&ts).Error; err != nil {
		tools.Println("select transfer_pending error", err)
		return
	}
	for _, t := range ts {
		client, err := common.GetMixinClientByIDOrHost(models.Ctx, t.ClientID)
		if err != nil {
			continue
		}
		c, err := common.GetClientByIDOrHost(models.Ctx, t.ClientID)
		if err != nil || c.Pin == "" {
			tools.Println("get pin error", err)
			continue
		}
		s, err := client.Transfer(models.Ctx, &mixin.TransferInput{
			AssetID:    t.AssetID,
			OpponentID: t.OpponentID,
			Amount:     t.Amount,
			TraceID:    t.TraceID,
			Memo:       t.Memo,
		}, c.Pin)
		if err != nil {
			tools.Println("transfer error", err)
			if strings.Contains(err.Error(), "20117") {
				a, _ := common.GetAssetByID(ctx, nil, t.AssetID)
				tools.SendMonitorGroupMsg(fmt.Sprintf("转账失败！请及时充值！%s (%s)\n\n 5分钟后重启转账队列...", a.Symbol, t.Memo))
				tools.SendMonitorGroupMsg("mixin://transfer/" + client.ClientID)
				time.Sleep(5 * time.Minute)
			}
			continue
		}
		models.RunInTransaction(models.Ctx, func(tx *gorm.DB) error {
			// 1. 添加转账记录
			if err := tx.Save(&models.Snapshot{
				SnapshotID: s.SnapshotID,
				ClientID:   t.ClientID,
				TraceID:    s.TraceID,
				UserID:     s.UserID,
				AssetID:    s.AssetID,
				Amount:     s.Amount,
				Memo:       s.Memo,
				CreatedAt:  s.CreatedAt,
			}).Error; err != nil {
				return err
			}
			// 2. 更新transfer_pending
			if err := tx.Model(&models.Transfer{}).
				Where("trace_id = ?", t.TraceID).
				Update("status", models.TransferStatusSucceed).
				Error; err != nil {
				return err
			}
			return nil
		})
	}
}
