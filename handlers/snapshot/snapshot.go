package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type snapshot struct {
	Type   string `json:"type,omitempty"`
	Reward string `json:"reward,omitempty"`
	ID     string `json:"id,omitempty"`
}

func ReceivedSnapshot(ctx context.Context, clientID string, msg *mixin.MessageView) error {
	var s mixin.Snapshot
	if err := json.Unmarshal(tools.Base64Decode(msg.Data), &s); err != nil {
		tools.Println(err)
		tools.PrintJson(msg)
		return nil
	}
	if s.Amount.LessThanOrEqual(decimal.Zero) {
		return nil
	}
	var r snapshot
	memo, err := url.QueryUnescape(s.Memo)
	if err != nil {
		tools.Println(err)
		tools.PrintJson(msg)
		return nil
	} else {
		s.Memo = memo
	}
	if err := json.Unmarshal([]byte(s.Memo), &r); err != nil {
		tools.Println(err)
		tools.PrintJson(msg)
		return nil
	}
	switch r.Type {
	case "":
		fallthrough
	case models.SnapshotTypeReward:
		if err := handelRewardSnapshot(ctx, clientID, &s, r.Reward); err != nil {
			tools.Println(err)
		}
	case models.SnapshotTypeJoin:
		if err := handelJoinSnapshot(ctx, clientID, &s); err != nil {
			tools.Println(err)
		}
	case models.SnapshotTypeVip:
		if err := handelVipSnapshot(ctx, clientID, &s); err != nil {
			tools.Println(err)
		}
	case models.SnapshotTypeAirdrop:
		if err := handelAirdropSnapshot(ctx, clientID, &s, r.ID); err != nil {
			tools.Println(err)
		}
	case models.SnapshotTypeMint:
		if err := session.DB(ctx).Model(&models.LiquidityMiningTx{}).
			Where("trace_id = ?", s.TraceID).
			Update("status", models.LiquidityMiningRecordStatusSuccess).Error; err != nil {
			tools.Println(err)
		}
	}
	return nil
}

func handelRewardSnapshot(ctx context.Context, clientID string, s *mixin.Snapshot, reward string) error {
	if reward == "" {
		tools.Println("reward is empty")
		tools.PrintJson(s)
		return nil
	}
	msg := config.Text.Reward
	from, err := common.SearchUser(ctx, clientID, s.OpponentID)
	if err != nil {
		return err
	}
	to, err := common.SearchUser(ctx, clientID, reward)
	if err != nil {
		return err
	}
	asset, err := common.GetAssetByID(ctx, nil, s.AssetID)
	if err != nil {
		return err
	}
	client, err := common.GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		return err
	}
	msg = strings.ReplaceAll(msg, "{send_name}", from.FullName)
	msg = strings.ReplaceAll(msg, "{reward_name}", to.FullName)
	msg = strings.ReplaceAll(msg, "{amount}", s.Amount.String())
	msg = strings.ReplaceAll(msg, "{symbol}", asset.Symbol)

	msg = tools.SplitString(msg, 36)
	byteMsg, err := json.Marshal([]mixin.AppButtonMessage{
		{Label: msg, Action: fmt.Sprintf("%s/reward?uid=%s", client.Host, to.IdentityNumber), Color: tools.RandomColor()},
	})
	if err != nil {
		return err
	}

	go common.SendClientMsg(clientID, mixin.MessageCategoryAppButtonGroup, tools.Base64Encode(byteMsg))
	go handleReward(clientID, s, from, to)
	return nil
}

func handelJoinSnapshot(ctx context.Context, clientID string, s *mixin.Snapshot) error {
	client, err := common.GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		return err
	}
	if client.PayStatus == models.ClientPayStatusOpen {
		a, err := decimal.NewFromString(client.PayAmount)
		if err != nil {
			return err
		}
		if s.AssetID == client.AssetID && s.Amount.Equal(a) {
			// 这是 一次 付费入群... 成功!
			common.UpdateClientUserPart(ctx, clientID, s.OpponentID, map[string]interface{}{
				"status":         models.ClientUserStatusLarge,
				"pay_expired_at": s.CreatedAt.Add(time.Hour * 24 * 365 * 99),
				"priority":       models.ClientUserPriorityHigh,
			})
		}
		go common.SendClientUserTextMsg(clientID, s.OpponentID, strings.Replace(config.Text.PayForLarge, "{year}", "99", 1), "")
		return nil
	} else {
		tools.Println("error join snapshots...")
		tools.PrintJson(s)
	}
	return nil
}

const (
	USDTAssetID = "4d8c508b-91c5-375b-92b0-ee702ed2dac5" // erc20
)

func handelVipSnapshot(ctx context.Context, clientID string, s *mixin.Snapshot) error {
	c, err := common.GetClientByIDOrHost(ctx, clientID)
	if err != nil {
		return err
	}

	var freshAmount, largeAmount decimal.Decimal
	if c.AssetID == "" {
		c.AssetID = USDTAssetID
		freshAmount = decimal.NewFromInt(1)
		largeAmount = decimal.NewFromInt(10)
	} else {
		cl, err := common.GetClientAssetLevel(ctx, clientID)
		if err != nil {
			tools.Println(err)
			return nil
		}
		freshAmount = cl.FreshAmount
		largeAmount = cl.LargeAmount
	}
	if c.AssetID != s.AssetID {
		log.Println("发现异常的 VIP 转账....")
		tools.PrintJson(s)
		return nil
	}
	var status int
	var msg string
	if s.Amount.Equal(freshAmount) {
		status = models.ClientUserStatusFresh
		msg = config.Text.PayForFresh
	} else if s.Amount.Equal(largeAmount) {
		status = models.ClientUserStatusLarge
		msg = strings.Replace(config.Text.PayForLarge, "{year}", "1", 1)
	} else {
		tools.Println("member to vip amount error...")
		tools.PrintJson(s)
		return nil
	}
	expTime := s.CreatedAt.Add(time.Hour * 24 * 365)
	if err := common.UpdateClientUserPart(ctx, clientID, s.OpponentID, map[string]interface{}{
		"status":         status,
		"priority":       models.ClientUserPriorityHigh,
		"pay_status":     status,
		"pay_expired_at": expTime,
	}); err != nil {
		return err
	}
	go common.SendClientUserTextMsg(clientID, s.OpponentID, msg, "")
	return nil
}

func handelAirdropSnapshot(ctx context.Context, clientID string, s *mixin.Snapshot, airdropID string) error {
	return session.DB(ctx).Model(&models.Airdrop{}).
		Where("trace_id = ?", s.TraceID).
		Update("status", models.AirdropStatusSuccess).
		Error
}

// 处理 reward 的转账添加
func handleReward(clientID string, s *mixin.Snapshot, from, to *models.User) error {
	models.RunInTransaction(models.Ctx, func(tx *gorm.DB) error {
		// 1. 保存转账记录
		if err := tx.Save(&models.Snapshot{
			SnapshotID: s.SnapshotID,
			ClientID:   clientID,
			TraceID:    s.TraceID,
			UserID:     s.UserID,
			AssetID:    s.AssetID,
			Amount:     s.Amount,
			Memo:       s.Memo,
			CreatedAt:  s.CreatedAt,
		}).Error; err != nil {
			return err
		}

		// 2. 添加transfer_pending
		traceID := mixin.UniqueConversationID(s.SnapshotID, s.TraceID)
		memo := strings.ReplaceAll(config.Text.From, "{identity_number}", from.IdentityNumber)
		if err := tx.Save(&models.Transfer{
			ClientID:   clientID,
			TraceID:    traceID,
			AssetID:    s.AssetID,
			OpponentID: to.UserID,
			Memo:       memo,
			Amount:     s.Amount,
			Status:     models.TransferStatusPending,
			CreatedAt:  time.Now(),
		}).Error; err != nil {
			return err
		}
		return nil
	})
	return nil
}
