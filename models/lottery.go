package models

import (
	"context"
	"errors"
	"math/rand"
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
	LotteryID  string          `json:"lottery_id"`
	UserID     string          `json:"user_id"`
	AssetID    string          `json:"asset_id"`
	TraceID    string          `json:"trace_id"`
	IsReceived bool            `json:"is_received"`
	Amount     decimal.Decimal `json:"amount"`
	CreatedAt  time.Time       `json:"created_at"`

	IconURL     string          `json:"icon_url,omitempty"`
	Symbol      string          `json:"symbol,omitempty"`
	FullName    string          `json:"full_name,omitempty"`
	PriceUsd    decimal.Decimal `json:"price_usd,omitempty"`
	ClientID    string          `json:"client_id,omitempty"`
	Date        string          `json:"date,omitempty"`
	Description string          `json:"description,omitempty"`
}

type Lottery struct {
	LotteryID string          `json:"lottery_id"`
	AssetID   string          `json:"asset_id"`
	Amount    decimal.Decimal `json:"amount"`
	IconURL   string          `json:"icon_url"`
	ClientID  string          `json:"client_id"`
}

var lotteryList = []Lottery{
	{LotteryID: "SAT50", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(50), IconURL: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", ClientID: ""},
	{LotteryID: "SAT100", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(100), IconURL: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", ClientID: ""},
	{LotteryID: "SAT200", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(200), IconURL: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", ClientID: ""},
	{LotteryID: "SAT500", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(500), IconURL: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", ClientID: ""},
	{LotteryID: "SAT99999", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(99999), IconURL: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", ClientID: ""},
	{LotteryID: "AKITA5000", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(5000), IconURL: "https://mixin-images.zeromesh.net/JSxN4FxhH3LNDowo22bEV3fGMdrGmKrYzGyNqGbYe72GFEitLVFfwmxrjEE8ZDzqAc14LWUcuHtHiO8l7ODyExmnLwM3aPdx8D0Z=s128", ClientID: ""},
	{LotteryID: "AKITA10000", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(10000), IconURL: "https://mixin-images.zeromesh.net/JSxN4FxhH3LNDowo22bEV3fGMdrGmKrYzGyNqGbYe72GFEitLVFfwmxrjEE8ZDzqAc14LWUcuHtHiO8l7ODyExmnLwM3aPdx8D0Z=s128", ClientID: ""},
	{LotteryID: "AKITA50000", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(50000), IconURL: "https://mixin-images.zeromesh.net/JSxN4FxhH3LNDowo22bEV3fGMdrGmKrYzGyNqGbYe72GFEitLVFfwmxrjEE8ZDzqAc14LWUcuHtHiO8l7ODyExmnLwM3aPdx8D0Z=s128", ClientID: ""},
	{LotteryID: "EPC100", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(100), IconURL: "https://mixin-images.zeromesh.net/HMXlpSt6KF9i-jp_ZQix9wFcMD27DrYox5kDrju6KkjvlQjQPZ2zimKKFYBJwecRTw5YAaMt4fpHXd1W0mwIxQ=s128", ClientID: ""},
	{LotteryID: "EPC50", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(50), IconURL: "https://mixin-images.zeromesh.net/HMXlpSt6KF9i-jp_ZQix9wFcMD27DrYox5kDrju6KkjvlQjQPZ2zimKKFYBJwecRTw5YAaMt4fpHXd1W0mwIxQ=s128", ClientID: ""},
	{LotteryID: "SHIB1000", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(1000), IconURL: "https://mixin-images.zeromesh.net/fgSEd6CY07BiZP76--7JA9P-rKIWRoXD8Eis8RUL6mP85_QPsbMoyJtWJ6MjE9jWFEjabNF0AKb8i2QOfdbCS6BJMntySps-8GfvJQ=s128", ClientID: ""},
	{LotteryID: "SHIB5000", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(5000), IconURL: "https://mixin-images.zeromesh.net/fgSEd6CY07BiZP76--7JA9P-rKIWRoXD8Eis8RUL6mP85_QPsbMoyJtWJ6MjE9jWFEjabNF0AKb8i2QOfdbCS6BJMntySps-8GfvJQ=s128", ClientID: ""},
	{LotteryID: "SHIB10000", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(10000), IconURL: "https://mixin-images.zeromesh.net/fgSEd6CY07BiZP76--7JA9P-rKIWRoXD8Eis8RUL6mP85_QPsbMoyJtWJ6MjE9jWFEjabNF0AKb8i2QOfdbCS6BJMntySps-8GfvJQ=s128", ClientID: ""},
	{LotteryID: "DOGE1", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(1), IconURL: "https://mixin-images.zeromesh.net/gtz8ocdxuC4N2rgEDKGc4Q6sZzWWCIGDWYBT6mHmtRubLqpE-xafvlABX6cvZ74VXL4HjyIocnX-H_Vxrz3En9tMcIKED0c-2MhH=s128", ClientID: ""}, {LotteryID: "MOB0.1", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(0.1), IconURL: "https://mixin-images.zeromesh.net/eckqDQi50ZUCoye5mR7y6BvlbXX6CBzkP89BfGNNH6TMNuyXYcCUd7knuIDpV_0W7nT1q3Oo9ooVnMDGjl8-oiENuA5UVREheUu2=s128", ClientID: ""},
	{LotteryID: "XIN0.1", AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c", Amount: decimal.NewFromFloat(0.1), IconURL: "https://mixin-images.zeromesh.net/UasWtBZO0TZyLTLCFQjvE_UYekjC7eHCuT_9_52ZpzmCC-X-NPioVegng7Hfx0XmIUavZgz5UL-HIgPCBECc-Ws=s128", ClientID: ""},
}

var lotterRate = map[string]decimal.Decimal{
	"SAT50":      decimal.NewFromInt(2200),
	"SAT100":     decimal.NewFromInt(900),
	"SAT200":     decimal.NewFromInt(600),
	"SAT500":     decimal.NewFromInt(300),
	"SAT99999":   decimal.NewFromInt(1),
	"AKITA5000":  decimal.NewFromInt(2100),
	"AKITA10000": decimal.NewFromInt(600),
	"AKITA50000": decimal.NewFromInt(300),
	"EPC100":     decimal.NewFromInt(30),
	"EPC50":      decimal.NewFromInt(100),
	"SHIB1000":   decimal.NewFromInt(1900),
	"SHIB5000":   decimal.NewFromInt(600),
	"SHIB10000":  decimal.NewFromInt(300),
	"DOGE1":      decimal.NewFromInt(50),
	"MOB0.1":     decimal.NewFromInt(18),
	"XIN0.1":     decimal.NewFromInt(1),
}

// 获取抽奖列表
func getLotteryList(ctx context.Context) []Lottery {
	return lotteryList[:16]
}

// 点击抽奖
func PostLottery(ctx context.Context, u *ClientUser) (string, error) {
	lotteryID := ""
	session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// 1. 检查是否有足够的能量
		pow := getPower(ctx, u.UserID)
		if pow.LotteryTimes < 1 {
			return session.ForbiddenError(ctx)
		}
		// 2. 根据概率获取 lottery
		snapshotID, lottery := getRandomLottery(ctx)
		// 3. 保存当前的 power
		updatePower(ctx, tx, u.UserID, pow.Balance.String(), pow.LotteryTimes-1)
		// 6. 保存 lottery 记录
		if err := createLotteryRecord(ctx, tx, lottery, u.UserID, snapshotID); err != nil {
			return err
		}
		lotteryID = lottery.LotteryID
		return nil
	})
	if lotteryID == "" {
		return "", session.ForbiddenError(ctx)
	}
	return lotteryID, nil
}

// 获取抽奖奖励
// 如果 string 有值则表示要弹框加入社群
func PostLotteryReward(ctx context.Context, u *ClientUser, traceID string) (*Client, error) {
	r, err := getLotteryRecordByTraceID(ctx, traceID)
	if err != nil {
		return nil, err
	}
	go transferLottery(_ctx, &r)
	l := getLotteryByID(ctx, r.LotteryID)
	if l.ClientID == "" {
		return nil, nil
	}

	isJoined := checkUserIsJoinedClient(ctx, l.ClientID, u.UserID)
	if !isJoined {
		info, _ := GetClientInfoByHostOrID(ctx, "", l.ClientID)
		return info.Client, nil
	}
	return nil, nil
}

func transferToGenerateRand(ctx context.Context) string {
	lClient := getLotteryClient()
	s, err := lClient.Transfer(ctx, &mixin.TransferInput{
		AssetID:    "965e5c6e-434c-3fa9-b780-c50f43cd955c",
		Amount:     decimal.NewFromFloat(0.1),
		TraceID:    tools.GetUUID(),
		OpponentID: "11efbb75-e7fe-44d7-a14f-698535289310",
	}, lClient.PIN)
	if err != nil {
		session.Logger(ctx).Println(err)
		time.Sleep(time.Second)
		return transferToGenerateRand(ctx)
	}

	return s.SnapshotID
}

func transferLottery(ctx context.Context, r *LotteryRecord) {
	lClient := getLotteryClient()
	_, err := lClient.Transfer(ctx, &mixin.TransferInput{
		AssetID:    r.AssetID,
		Amount:     r.Amount,
		TraceID:    r.TraceID,
		OpponentID: r.UserID,
	}, lClient.PIN)
	if err != nil {
		session.Logger(ctx).Println(err)
		time.Sleep(time.Second * 5)
		transferLottery(ctx, r)
	} else {
		_, err = session.Database(ctx).Exec(ctx, "UPDATE lottery_record SET is_received = true WHERE trace_id = $1", r.TraceID)
		if err != nil {
			session.Logger(ctx).Println(err)
			return
		}
	}
}

// 获取抽奖列表
func GetLotteryRecordList(ctx context.Context, u *ClientUser, page int) ([]LotteryRecord, error) {
	if page < 1 {
		page = 1
	}
	var list []LotteryRecord
	if err := session.Database(ctx).ConnQuery(ctx, `
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
	err := session.Database(ctx).
		QueryRow(ctx, "SELECT lottery_id,asset_id, amount, user_id FROM lottery_record WHERE trace_id = $1", traceID).
		Scan(&r.LotteryID, &r.AssetID, &r.Amount, &r.UserID)
	r.TraceID = traceID
	return r, err
}

func createLotteryRecord(ctx context.Context, tx pgx.Tx, l *Lottery, userID, snapshotID string) error {
	query := durable.InsertQuery("lottery_record", "lottery_id,user_id,asset_id,amount,trace_id,snapshot_id")
	_, err := tx.Exec(ctx, query, l.LotteryID, userID, l.AssetID, l.Amount.String(), tools.GetUUID(), snapshotID)
	return err
}

// 获取随机的抽奖奖励
func getRandomLottery(ctx context.Context) (string, *Lottery) {
	// 通过转账获取一个随机数
	rand.Seed(time.Now().UnixNano())
	random := decimal.NewFromInt(int64(rand.Intn(10000)))
	for lotteryID, rate := range lotterRate {
		if random.LessThanOrEqual(rate) {
			return "", getLotteryByID(ctx, lotteryID)
		} else {
			random = random.Sub(rate)
		}
	}
	session.Logger(ctx).Println("get random lottery error")
	return "", &lotteryList[0]
}

// 获取随机的抽奖奖励
// func getRandomLottery(ctx context.Context) (string, *Lottery) {
// 	// 通过转账获取一个随机数
// 	snapshotID := transferToGenerateRand(ctx)
// 	s := sha512.Sum512([]byte(snapshotID))
// 	var sum int64
// 	for _, v := range s {
// 		sum += int64(v)
// 	}
// 	random := decimal.NewFromInt(sum % 10000)
// 	for lotteryID, rate := range lotterRate {
// 		if random.LessThanOrEqual(rate) {
// 			return snapshotID, getLotteryByID(ctx, lotteryID)
// 		} else {
// 			random = random.Sub(rate)
// 		}
// 	}
// 	session.Logger(ctx).Println("get random lottery error")
// 	return snapshotID, &lotteryList[0]
// }

type LotteryClient struct {
	*mixin.Client
	PIN string `json:"pin"`
}

var lClient *LotteryClient

func getLotteryClient() *LotteryClient {
	if lClient == nil {
		var l LotteryClient
		lc := config.Config.Lottery
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

func getLotteryByID(ctx context.Context, id string) *Lottery {
	for _, l := range lotteryList {
		if l.LotteryID == id {
			return &l
		}
	}
	return nil
}

func getLastLottery(ctx context.Context) []LotteryRecord {
	list := make([]LotteryRecord, 0)
	session.Database(ctx).ConnQuery(ctx, `
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
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT lottery_id, asset_id, amount, trace_id
FROM lottery_record
WHERE is_received = false AND user_id = $1
ORDER BY created_at ASC LIMIT 1`, userID).
		Scan(&r.LotteryID, &r.AssetID, &r.Amount, &r.TraceID); err != nil {
		return nil
	}
	a, _ := GetAssetByID(ctx, nil, r.AssetID)
	r.IconURL = a.IconUrl
	r.Symbol = a.Symbol
	r.PriceUsd = a.PriceUsd.Mul(r.Amount).Round(2)
	clientID := getLotteryByID(ctx, r.LotteryID).ClientID
	if clientID != "" {
		c, err := GetClientByID(ctx, clientID)
		if err != nil {
			session.Logger(ctx).Println("get client error", err)
			return nil
		}
		r.Description = c.Description
	}
	return &r
}
