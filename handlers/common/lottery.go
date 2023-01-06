package common

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const lottery_record_DDL = `
CREATE TABLE IF NOT EXISTS lottery_record (
	lottery_id		 VARCHAR(36) NOT NULL,
	user_id    VARCHAR(36) NOT NULL,
	asset_id   VARCHAR(36) NOT NULL,
	trace_id  VARCHAR(36) NOT NULL,
	snapshot_id VARCHAR(36) NOT NULL DEFAULT '',
	is_received BOOLEAN NOT NULL DEFAULT false,
	amount 	   VARCHAR NOT NULL DEFAULT '0',
	
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);`

type LotteryRecord struct {
	LotteryID   string          `json:"lottery_id"`
	UserID      string          `json:"user_id"`
	AssetID     string          `json:"asset_id"`
	TraceID     string          `json:"trace_id"`
	IsReceived  bool            `json:"is_received"`
	Amount      decimal.Decimal `json:"amount"`
	CreatedAt   time.Time       `json:"created_at"`
	IconURL     string          `json:"icon_url,omitempty"`
	Symbol      string          `json:"symbol,omitempty"`
	FullName    string          `json:"full_name,omitempty"`
	PriceUsd    decimal.Decimal `json:"price_usd,omitempty"`
	ClientID    string          `json:"client_id,omitempty"`
	Date        string          `json:"date,omitempty"`
	Description string          `json:"description,omitempty"`
}

type LotteryList struct {
	config.Lottery
	Description string          `json:"description"`
	Symbol      string          `json:"symbol"`
	PriceUSD    decimal.Decimal `json:"price_usd"`
}

// 获取抽奖列表
func getLotteryList(ctx context.Context, u *ClientUser) []LotteryList {
	ls := make([]LotteryList, 0)
	list := getUserListingLottery(ctx, u.UserID)
	for _, lottery := range list {
		var l LotteryList
		l.Lottery = lottery
		if lottery.ClientID != "" {
			client, _ := GetClientByIDOrHost(ctx, lottery.ClientID)
			l.Description = client.Description
		}
		if lottery.AssetID != "" {
			asset, _ := GetAssetByID(ctx, nil, lottery.AssetID)
			l.Symbol = asset.Symbol
			l.PriceUSD = asset.PriceUsd
		}
		if l.Inventory != 0 {
			l.Inventory = 0
		}
		ls = append(ls, l)
	}
	return ls
}

// 点击抽奖
func PostLottery(ctx context.Context, u *ClientUser) (string, error) {
	if CheckIsBlockUser(ctx, u.ClientID, u.UserID) {
		return "", session.ForbiddenError(ctx)
	}
	lotteryID := ""
	err := session.DB(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. 检查是否有足够的能量
		pow := getPowerWithTx(ctx, tx, u.UserID)
		if pow.LotteryTimes < 1 {
			return session.ForbiddenError(ctx)
		}
		// 2. 根据概率获取 lottery
		lottery := GetRandomLottery(ctx, u)
		// 3. 保存当前的 power
		updatePower(ctx, tx, u.UserID, pow.Balance.String(), pow.LotteryTimes-1)
		// 6. 保存 lottery 记录
		traceID := tools.GetUUID()
		if err := createLotteryRecord(ctx, tx, lottery, u.UserID, traceID); err != nil {
			return err
		}

		if lottery.SupplyID != "" {
			// 记录抽到了项目方的奖品
			if err := createLotterySupplyRecord(ctx, tx, lottery.SupplyID, u.UserID, traceID); err != nil {
				return err
			}
			if lottery.Inventory == 0 {
				return session.ForbiddenError(ctx)
			}
			if lottery.Inventory == 1 {
				if _, err := tx.Exec(ctx, `UPDATE lottery_supply SET inventory=$2,status=3 WHERE supply_id=$1`, lottery.SupplyID, lottery.Inventory-1); err != nil {
					return err
				}
			} else if lottery.Inventory > 1 {
				if _, err := tx.Exec(ctx, `UPDATE lottery_supply SET inventory=$2 WHERE supply_id=$1`, lottery.SupplyID, lottery.Inventory-1); err != nil {
					return err
				}
			}
		}
		lotteryID = lottery.LotteryID
		return nil
	})
	if lotteryID == "" {
		return "", session.ForbiddenError(ctx)
	}
	return lotteryID, err
}

// 获取抽奖奖励
// 如果 string 有值则表示要弹框加入社群
func PostLotteryReward(ctx context.Context, u *ClientUser, traceID string) (*Client, error) {
	if CheckIsBlockUser(ctx, u.ClientID, u.UserID) {
		return nil, session.ForbiddenError(ctx)
	}
	r, err := getLotteryRecordByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	go transferLottery(_ctx, &r)
	l, err := getLotteryByTrace(ctx, traceID)
	if err != nil {
		return nil, err
	}
	if l.ClientID == "" {
		return nil, nil
	}
	isJoined := checkUserIsJoinedClient(ctx, l.ClientID, u.UserID)
	if !isJoined {
		info, _ := GetClientInfoByHostOrID(ctx, l.ClientID)
		return info.Client, nil
	}
	return nil, nil
}

func transferLottery(ctx context.Context, r *LotteryRecord) {
	lClient := getLotteryClient()
	if lClient.ClientID == "11efbb75-e7fe-44d7-a14f-698535289310" {
		r.AssetID = "965e5c6e-434c-3fa9-b780-c50f43cd955c"
	}
	snapshot, err := lClient.Transfer(ctx, &mixin.TransferInput{
		AssetID:    r.AssetID,
		Amount:     r.Amount,
		TraceID:    r.TraceID,
		OpponentID: r.UserID,
		Memo:       "lottery",
	}, lClient.PIN)
	if err != nil {
		if strings.Contains(err.Error(), "20117") {
			a, _ := GetAssetByID(ctx, nil, r.AssetID)
			SendMonitorGroupMsg(ctx, fmt.Sprintf("转账失败！请及时充值！%s", a.Symbol))
			SendMonitorGroupMsg(ctx, "mixin://transfer/"+lClient.ClientID)
		} else {
			tools.Println(err)
			time.Sleep(time.Second * 5)
			transferLottery(ctx, r)
		}
	} else {
		if err := session.DB(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE lottery_supply_received SET status=2 WHERE trace_id=$1`, r.TraceID)
			if err != nil {
				return err
			}
			_, err = tx.Exec(ctx, "UPDATE lottery_record SET is_received=true,snapshot_id=$1 WHERE trace_id=$2", snapshot.SnapshotID, r.TraceID)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			tools.Println(err)
		}
	}
}

// 获取抽奖列表
func GetLotteryRecordList(ctx context.Context, u *ClientUser, page int) ([]LotteryRecord, error) {
	if page < 1 {
		page = 1
	}
	var list []LotteryRecord
	if err := session.DB(ctx).ConnQuery(ctx, `
SELECT asset_id, amount, to_char(created_at, 'YYYY-MM-DD') AS date 
FROM lottery_record 
WHERE user_id = $1 
ORDER BY created_at DESC 
OFFSET $2 LIMIT 20
`,
		func(rows pgx.Rows) error {
			for rows.Next() {
				var r LotteryRecord
				if err := rows.Scan(&r.AssetID, &r.Amount, &r.Date); err != nil {
					return err
				}
				a, _ := GetAssetByID(ctx, nil, r.AssetID)
				r.Symbol = a.Symbol
				r.IconURL = a.IconUrl
				list = append(list, r)
			}
			return nil
		}, u.UserID, (page-1)*20); err != nil {
		return nil, err
	}
	return list, nil
}

func getLotteryRecordByTraceID(ctx context.Context, traceID string) (LotteryRecord, error) {
	var r LotteryRecord
	err := session.DB(ctx).
		QueryRow(ctx, "SELECT lottery_id,asset_id, amount, user_id FROM lottery_record WHERE trace_id = $1", traceID).
		Scan(&r.LotteryID, &r.AssetID, &r.Amount, &r.UserID)
	r.TraceID = traceID
	return r, err
}

func createLotteryRecord(ctx context.Context, tx pgx.Tx, l *config.Lottery, userID, traceID string) error {
	query := durable.InsertQuery("lottery_record", "lottery_id,user_id,asset_id,amount,trace_id,snapshot_id")
	_, err := tx.Exec(ctx, query, l.LotteryID, userID, l.AssetID, l.Amount.String(), traceID, "")
	return err
}

// 获取随机的抽奖奖励
func GetRandomLottery(ctx context.Context, u *ClientUser) *config.Lottery {
	// 通过转账获取一个随机数
	rand.Seed(time.Now().UnixNano())
	random := decimal.NewFromInt(int64(rand.Intn(10000)))
	for lotteryID, rate := range config.Config.Lottery.Rate {
		if random.LessThanOrEqual(rate) {
			return getLotteryByID(ctx, lotteryID, u.UserID)
		} else {
			random = random.Sub(rate)
		}
	}
	tools.Println("get random lottery error")
	return &config.Config.Lottery.List[0]
}

type LotteryClient struct {
	*mixin.Client
	PIN string `json:"pin"`
}

var lClient *LotteryClient

func getLotteryClient() *LotteryClient {
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
	_, err := GetClientUserByClientIDAndUserID(ctx, clientID, userID)
	if err != nil && errors.Is(err, pgx.ErrNoRows) {
		return false
	}
	return true
}

func getLastLottery(ctx context.Context) []LotteryRecord {
	list := make([]LotteryRecord, 0)
	session.DB(ctx).ConnQuery(ctx, `
SELECT lr.asset_id, lr.amount, u.full_name
FROM lottery_record lr 
LEFT JOIN users u ON u.user_id = lr.user_id
ORDER BY lr.created_at DESC LIMIT 5`,
		func(rows pgx.Rows) error {
			for rows.Next() {
				var r LotteryRecord
				if err := rows.Scan(&r.AssetID, &r.Amount, &r.FullName); err != nil {
					return err
				}
				a, _ := GetAssetByID(ctx, nil, r.AssetID)
				r.Symbol = a.Symbol
				r.IconURL = a.IconUrl
				r.PriceUsd = a.PriceUsd.Mul(r.Amount).Round(2)
				list = append(list, r)
			}
			return nil
		})
	return list
}

func getReceivingLottery(ctx context.Context, userID string) *LotteryRecord {
	var r LotteryRecord
	if err := session.DB(ctx).QueryRow(ctx, `
SELECT lottery_id, asset_id, amount, trace_id
FROM lottery_record
WHERE is_received = false AND user_id = $1
ORDER BY created_at ASC LIMIT 1`, userID).
		Scan(&r.LotteryID, &r.AssetID, &r.Amount, &r.TraceID); durable.CheckNotEmptyError(err) != nil {
		tools.Println("get client id error", err)
		return nil
	}
	a, _ := GetAssetByID(ctx, nil, r.AssetID)
	r.IconURL = a.IconUrl
	r.Symbol = a.Symbol
	r.PriceUsd = a.PriceUsd
	clientID := ""
	lottery, err := getLotteryByTrace(ctx, r.TraceID)
	if err != nil {
		return nil
	}
	clientID = lottery.ClientID
	if clientID != "" {
		c, err := GetClientByIDOrHost(ctx, clientID)
		if err != nil {
			tools.Println("get client error", err)
			return nil
		}
		r.Description = c.Description
	}
	return &r
}
