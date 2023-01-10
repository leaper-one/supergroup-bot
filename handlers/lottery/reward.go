package lottery

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/handlers/common"
	"github.com/MixinNetwork/supergroup/models"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"gorm.io/gorm"
)

func PostLotteryReward(ctx context.Context, u *models.ClientUser, traceID string) (*models.Client, error) {
	if common.CheckIsBlockUser(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	r, err := getLotteryRecordByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	go transferLottery(&r)
	l, err := getLotteryByTrace(ctx, traceID)
	if err != nil {
		return nil, err
	}
	if l.ClientID == "" {
		return nil, nil
	}
	isJoined := checkUserIsJoinedClient(ctx, l.ClientID, u.UserID)
	if !isJoined {
		info, _ := common.GetMixinClientByIDOrHost(ctx, l.ClientID)
		return &info.C, nil
	}
	return nil, nil
}

func getLotteryRecordByTraceID(ctx context.Context, traceID string) (models.LotteryRecord, error) {
	var r models.LotteryRecord
	err := session.DB(ctx).Take(&r, "trace_id=?", traceID).Error
	r.TraceID = traceID
	return r, err
}

func transferLottery(r *models.LotteryRecord) {
	ctx := models.Ctx
	lClient := GetLotteryClient()
	if lClient.ClientID == "11efbb75-e7fe-44d7-a14f-698535289310" {
		r.AssetID = "965e5c6e-434c-3fa9-b780-c50f43cd955c"
	}
	snapshot, err := lClient.Transfer(context.Background(), &mixin.TransferInput{
		AssetID:    r.AssetID,
		Amount:     r.Amount,
		TraceID:    r.TraceID,
		OpponentID: r.UserID,
		Memo:       "lottery",
	}, lClient.PIN)
	if err != nil {
		if strings.Contains(err.Error(), "20117") {
			a, _ := common.GetAssetByID(ctx, nil, r.AssetID)
			tools.SendMonitorGroupMsg(fmt.Sprintf("转账失败！请及时充值！%s", a.Symbol))
			tools.SendMonitorGroupMsg("mixin://transfer/" + lClient.ClientID)
		} else {
			tools.Println(err)
			time.Sleep(time.Second * 5)
			transferLottery(r)
		}
	} else {
		if err := models.RunInTransaction(ctx, func(tx *gorm.DB) error {
			if err := tx.Model(&models.LotterySupplyReceived{}).Where("trace_id=?", r.TraceID).Update("status", 2).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.LotteryRecord{}).Where("trace_id=?", r.TraceID).Updates(map[string]interface{}{
				"is_received": true,
				"snapshot_id": snapshot.SnapshotID,
			}).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			tools.Println(err)
		}
	}
}

type LotteryClient struct {
	*mixin.Client
	PIN string `json:"pin"`
}

var lClient *LotteryClient

func GetLotteryClient() *LotteryClient {
	if lClient == nil {
		var l LotteryClient
		lc := config.Config.Lottery
		if lc.ClientID == "" {
			return nil
		}
		l.Client, _ = mixin.NewFromKeystore(&mixin.Keystore{
			ClientID:   lc.ClientID,
			SessionID:  lc.SessionID,
			PinToken:   lc.PinToken,
			PrivateKey: lc.PrivateKey,
		})
		l.PIN = lc.PIN
		lClient = &l
	}
	return lClient
}

func checkUserIsJoinedClient(ctx context.Context, clientID, userID string) bool {
	_, err := common.GetClientUserByClientIDAndUserID(ctx, clientID, userID)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	return true
}
