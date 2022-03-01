package models

import (
	"context"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

var swapBot = []string{
	"6a4a121d-9673-4a7e-a93e-cb9ca4bb83a2",
	"a753e0eb-3010-4c4a-a7b2-a7bda4063f62",
	"",
}

const trading_competition_DDL = `
CREATE TABLE IF NOT EXISTS trading_competition (
	competition_id VARCHAR(36) NOT NULL PRIMARY KEY,
	client_id VARCHAR(36) NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	amount VARCHAR NOT NULL,
	start_at DATE NOT NULL,
	end_at DATE NOT NULL,
	title VARCHAR(255) NOT NULL,
	tips VARCHAR(255) NOT NULL,
	rules VARCHAR(255) NOT NULL,
	reward VARCHAR NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const user_snapshots_DDL = `
CREATE TABLE IF NOT EXISTS user_snapshots (
	snapshot_id VARCHAR(36) NOT NULL PRIMARY KEY,
	user_id VARCHAR(36) NOT NULL,
	opponent_id VARCHAR(36) NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	amount VARCHAR NOT NULL,
	opening_balance VARCHAR NOT NULL,
	closing_balance VARCHAR NOT NULL,
	source VARCHAR NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
`

const trading_rank_DDL = `
CREATE TABLE IF NOT EXISTS trading_rank (
	competition_id VARCHAR(36) NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	amount VARCHAR NOT NULL,
	updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
	PRIMARY KEY (competition_id,user_id)
);
`

type TradingCompetition struct {
	CompetitionID string          `json:"competition_id"`
	ClientID      string          `json:"client_id"`
	AssetID       string          `json:"asset_id"`
	Amount        decimal.Decimal `json:"amount"`
	Title         string          `json:"title"`
	Tips          string          `json:"tips"`
	Rules         string          `json:"rules"`
	Reward        string          `json:"reward"`
	StartAt       time.Time       `json:"start_at"`
	EndAt         time.Time       `json:"end_at"`
	CreatedAt     time.Time       `json:"created_at"`
}

type UserSnapshot struct {
	SnapshotID     string          `json:"snapshot_id"`
	ClientID       string          `json:"client_id"`
	UserID         string          `json:"user_id"`
	OpponentID     string          `json:"opponent_id"`
	AssetID        string          `json:"asset_id"`
	OpeningBalance decimal.Decimal `json:"opening_balance,omitempty"`
	ClosingBalance decimal.Decimal `json:"closing_balance,omitempty"`
	Amount         decimal.Decimal `json:"amount"`
	Source         string          `json:"source"`
	CreatedAt      time.Time       `json:"created_at"`
}

type TradingRank struct {
	CompetitionID string          `json:"competition_id,omitempty"`
	AssetID       string          `json:"asset_id,omitempty"`
	UserID        string          `json:"user_id"`
	Amount        decimal.Decimal `json:"amount"`
	UpdatedAt     time.Time       `json:"-"`

	FullName       string `json:"full_name,omitempty"`
	Avatar         string `json:"avatar,omitempty"`
	IdentityNumber string `json:"identity_number,omitempty"`
}

type TradingCompetitionResp struct {
	TradingCompetition *TradingCompetition `json:"trading_competition"`
	Asset              *Asset              `json:"asset"`
	Status             string              `json:"status"` // 1 待授权 2 已授权
}

func GetTradingCompetetionByID(ctx context.Context, u *ClientUser, id string) (*TradingCompetitionResp, error) {
	tc, err := getTradingCompetetionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	_, err = mixin.ReadSnapshots(ctx, u.AccessToken, "", time.Now(), "DESC", 1)
	var status string
	if err == nil {
		status = "2"
	} else if strings.Contains(err.Error(), "[202/403] Forbidden") {
		status = "1"
	} else {
		return nil, err
	}
	if status == "2" && time.Now().Before(tc.EndAt) {
		go runTradingCheck(_ctx, tc, u, nil)
	}
	asset, _ := GetAssetByID(ctx, nil, tc.AssetID)
	return &TradingCompetitionResp{
		TradingCompetition: tc,
		Asset:              &asset,
		Status:             status,
	}, nil
}

type TradingRankResp struct {
	List   []*TradingRank  `json:"list"`
	Amount decimal.Decimal `json:"amount"`
	Symbol string          `json:"symbol"`
}

func GetRandingCompetetionRankByID(ctx context.Context, u *ClientUser, id string) (*TradingRankResp, error) {
	ranks := make([]*TradingRank, 0)
	tc, err := getTradingCompetetionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	asset, err := GetAssetByID(ctx, nil, tc.AssetID)
	if err != nil {
		return nil, err
	}
	err = session.Database(ctx).ConnQuery(ctx, `
SELECT tr.user_id,tr.amount,u.full_name,u.avatar_url,u.identity_number
FROM trading_rank tr 
LEFT JOIN users u ON u.user_id = tr.user_id
WHERE competition_id = $1
ORDER BY amount::NUMERIC DESC
LIMIT 10
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var r TradingRank
			if err := rows.Scan(&r.UserID, &r.Amount, &r.FullName, &r.Avatar, &r.IdentityNumber); err != nil {
				return err
			}
			r.Amount = decimal.NewFromInt(r.Amount.IntPart())
			id := string(r.IdentityNumber[0])
			id += "****"
			id += string(r.IdentityNumber[len(r.IdentityNumber)-1])
			r.IdentityNumber = id
			ranks = append(ranks, &r)
		}
		return nil
	}, id)
	if err != nil {
		return nil, err
	}

	var amount decimal.Decimal
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT amount FROM trading_rank
WHERE user_id=$1 AND competition_id=$2
`, u.UserID, tc.CompetitionID).Scan(&amount); err != nil {
		return nil, err
	}

	return &TradingRankResp{
		List:   ranks,
		Symbol: asset.Symbol,
		Amount: amount,
	}, nil
}

func getTradingCompetetionByID(ctx context.Context, id string) (*TradingCompetition, error) {
	var tc TradingCompetition
	err := session.Database(ctx).QueryRow(ctx, `
SELECT competition_id,client_id,asset_id,amount,title,tips,rules,reward,start_at,end_at+1 as end_at FROM trading_competition
WHERE competition_id=$1
`, id).Scan(&tc.CompetitionID, &tc.ClientID, &tc.AssetID, &tc.Amount, &tc.Title, &tc.Tips, &tc.Rules, &tc.Reward, &tc.StartAt, &tc.EndAt)
	return &tc, err
}

func runTradingCheck(ctx context.Context, tc *TradingCompetition, u *ClientUser, specTime *time.Time) error {
	if err := StatisticUserSnapshots(ctx, u, tc.StartAt); err != nil {
		if strings.Contains(err.Error(), "maybe invalid token") ||
			strings.Contains(err.Error(), "Forbidden") {
			if _, err := session.Database(ctx).Exec(ctx, `
				INSERT INTO trading_rank (competition_id,asset_id,user_id,amount,updated_at)
				VALUES ($1,$2,$3,$4,NOW())
				ON CONFLICT (competition_id,user_id) DO UPDATE SET amount=$4,updated_at=NOW()
					`, tc.CompetitionID, tc.AssetID, u.UserID, decimal.Zero); err != nil {
				session.Logger(ctx).Println(err)
				return err
			}
			return nil
		}
		session.Logger(ctx).Println(err)
		return err
	}
	amount, err := getTransferAmount(ctx, tc, u)
	if err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	if amount.GreaterThan(decimal.Zero) {
		diff, err := getBlanceChangeAmount(ctx, u.UserID, tc.AssetID, tc.StartAt, tc.EndAt)
		if err != nil {
			session.Logger(ctx).Println(err)
			return err
		}
		if diff.LessThan(amount) {
			amount = diff
		}
	}

	lpAmount, err := getTradingCompetitionLpAssetAmount(ctx, tc, u, specTime)
	if err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	amount = amount.Add(lpAmount)
	if _, err := session.Database(ctx).Exec(ctx, `
INSERT INTO trading_rank (competition_id,asset_id,user_id,amount,updated_at)
VALUES ($1,$2,$3,$4,NOW())
ON CONFLICT (competition_id,user_id) DO UPDATE SET amount=$4,updated_at=NOW()
	`, tc.CompetitionID, tc.AssetID, u.UserID, amount); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	return nil
}

func StatisticUserSnapshots(ctx context.Context, u *ClientUser, startAt time.Time) error {
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT created_at FROM user_snapshots
WHERE user_id=$1
ORDER BY created_at DESC LIMIT 1
	`, u.UserID).Scan(&startAt); durable.CheckNotEmptyError(err) != nil {
		return err
	}
	ss, err := mixin.ReadSnapshots(ctx, u.AccessToken, "", startAt, "ASC", 500)
	if err != nil {
		return err
	}
	for _, s := range ss {
		if _, err := session.Database(ctx).Exec(ctx, `
INSERT INTO user_snapshots (snapshot_id,user_id,opponent_id,asset_id,amount,opening_balance,closing_balance ,source,created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
`, s.SnapshotID, u.UserID, s.OpponentID, s.AssetID, s.Amount, s.OpeningBalance, s.ClosingBalance, s.Type, s.CreatedAt); err != nil {
			if !durable.CheckIsPKRepeatError(err) {
				return err
			}
		}
	}
	if len(ss) < 500 {
		return nil
	}
	return StatisticUserSnapshots(ctx, u, startAt)
}

func getTransferAmount(ctx context.Context, tc *TradingCompetition, u *ClientUser) (decimal.Decimal, error) {
	var amount decimal.Decimal
	err := session.Database(ctx).QueryRow(ctx, `
SELECT coalesce(SUM(amount::NUMERIC),0) FROM user_snapshots
WHERE user_id=$1
AND asset_id=$2
AND opponent_id=ANY($3)
AND created_at BETWEEN $4 AND $5
AND source IN ('transfer', 'raw')
`, u.UserID, tc.AssetID, swapBot, tc.StartAt, tc.EndAt).Scan(&amount)
	return amount, err
}

func getBlanceChangeAmount(ctx context.Context, userID, assetID string, startAt, endAt time.Time) (decimal.Decimal, error) {
	var start decimal.Decimal
	var end decimal.Decimal
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT opening_balance FROM user_snapshots
WHERE user_id=$1
AND asset_id=$2
AND created_at BETWEEN $3 AND $4
ORDER BY created_at ASC LIMIT 1
`, userID, assetID, startAt, endAt).Scan(&start); durable.CheckNotEmptyError(err) != nil {
		session.Logger(ctx).Println(err)
		return decimal.Zero, err
	}
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT closing_balance FROM user_snapshots
WHERE user_id=$1
AND asset_id=$2
AND created_at BETWEEN $3 AND $4
ORDER BY created_at DESC LIMIT 1
`, userID, assetID, startAt, endAt).Scan(&end); durable.CheckNotEmptyError(err) != nil {
		session.Logger(ctx).Println(err)
		return decimal.Zero, err
	}
	diff := end.Sub(start)
	return diff, nil
}

func getTradingCompetitionLpAssetAmount(ctx context.Context, tc *TradingCompetition, u *ClientUser, specTime *time.Time) (decimal.Decimal, error) {
	a := decimal.Zero
	lpList, err := GetClientAssetLPCheckMapByID(ctx, u.ClientID)
	if err != nil {
		return a, err
	}
	if len(lpList) == 0 {
		return a, nil
	}
	asset, err := GetAssetByID(ctx, nil, tc.AssetID)
	if err != nil {
		return a, err
	}

	if specTime != nil {
		// 如果指定了时间，则只计算指定时间的资产价格
		for assetId := range lpList {
			ticker, err := mixin.ReadTicker(ctx, assetId, *specTime)
			if err != nil {
				return a, err
			}
			lpList[assetId] = ticker.PriceUSD
		}
		ticker, err := mixin.ReadTicker(ctx, tc.AssetID, *specTime)
		if err != nil {
			return a, err
		}
		asset.PriceUsd = ticker.PriceUSD
	}

	for assetID, price := range lpList {
		amount, err := getBlanceChangeAmount(ctx, u.UserID, assetID, tc.StartAt, tc.EndAt)
		if err != nil {
			session.Logger(ctx).Println(err)
			return decimal.Zero, err
		}
		valAmount := amount.Mul(price).Div(decimal.NewFromInt(2)).Div(asset.PriceUsd)
		a = a.Add(valAmount)
	}
	return a, nil
}

func autoDrawlTradingJob() {
	for {
		tcs := make([]*TradingCompetition, 0)
		session.Database(_ctx).ConnQuery(_ctx, `
SELECT competition_id,client_id,asset_id,amount,title,tips,rules,reward,start_at,end_at+1 as end_at FROM trading_competition
WHERE start_at<NOW() AND end_at::date+1>NOW()
	`, func(rows pgx.Rows) error {
			for rows.Next() {
				var tc TradingCompetition
				if err := rows.Scan(&tc.CompetitionID, &tc.ClientID, &tc.AssetID, &tc.Amount, &tc.Title, &tc.Tips, &tc.Rules, &tc.Reward, &tc.StartAt, &tc.EndAt); err != nil {
					return err
				}
				tcs = append(tcs, &tc)
			}
			return nil
		})
		for _, tc := range tcs {
			if err := DrawlTradingJob(tc, nil); err != nil {
				session.Logger(_ctx).Println(err)
			}
		}
		time.Sleep(time.Minute)
	}
}

func DrawlTradingJob(tc *TradingCompetition, specTime *time.Time) error {
	users := make([]string, 0)
	if err := session.Database(_ctx).ConnQuery(_ctx, `
SELECT user_id FROM trading_rank WHERE competition_id=$1
	`, func(rows pgx.Rows) error {
		for rows.Next() {
			var userID string
			if err := rows.Scan(&userID); err != nil {
				return err
			}
			users = append(users, userID)
		}
		return nil
	}, tc.CompetitionID); err != nil {
		return err
	}

	for _, userID := range users {
		u, err := GetClientUserByClientIDAndUserID(_ctx, tc.ClientID, userID)
		if err != nil {
			return err
		}
		if err := runTradingCheck(_ctx, tc, u, specTime); err != nil {
			return err
		}
	}
	return nil
}

func DrawlTradingJobWithSpecTime(ctx context.Context, competitionID string, specTime time.Time) error {
	tc, err := getTradingCompetetionByID(ctx, competitionID)
	if err != nil {
		return err
	}
	if err := DrawlTradingJob(tc, &specTime); err != nil {
		return err
	}
	return nil
}
