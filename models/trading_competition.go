package models

import (
	"context"
	"log"
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
	if status == "2" {
		go runTradingCheck(_ctx, tc, u.UserID, u.AccessToken)
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
SELECT competition_id,client_id,asset_id,amount,title,tips,rules,reward,start_at,end_at FROM trading_competition 
WHERE competition_id=$1
`, id).Scan(&tc.CompetitionID, &tc.ClientID, &tc.AssetID, &tc.Amount, &tc.Title, &tc.Tips, &tc.Rules, &tc.Reward, &tc.StartAt, &tc.EndAt)
	return &tc, err
}

func runTradingCheck(ctx context.Context, tc *TradingCompetition, userID, token string) error {
	if err := StatisticUserSnapshots(ctx, userID, token, tc.AssetID, tc.StartAt); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	var amount decimal.Decimal
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT coalesce(SUM(amount::NUMERIC),0) FROM user_snapshots
WHERE user_id=$1 
AND asset_id=$2 
AND opponent_id=ANY($3) 
AND created_at BETWEEN $4 AND $5
`, userID, tc.AssetID, swapBot, tc.StartAt, tc.EndAt).Scan(&amount); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	if amount.GreaterThan(decimal.Zero) {
		var start decimal.Decimal
		var end decimal.Decimal
		if err := session.Database(ctx).QueryRow(ctx, `
SELECT opening_balance FROM user_snapshots
WHERE user_id=$1 
AND asset_id=$2 
AND opponent_id=ANY($3)
AND created_at BETWEEN $4 AND $5
ORDER BY created_at ASC LIMIT 1
`, userID, tc.AssetID, swapBot, tc.StartAt, tc.EndAt).Scan(&start); err != nil {
			session.Logger(ctx).Println(err)
			return err
		}
		if err := session.Database(ctx).QueryRow(ctx, `
SELECT closing_balance FROM user_snapshots
WHERE user_id=$1
AND asset_id=$2 
AND opponent_id=ANY($3)
AND created_at BETWEEN $4 AND $5
ORDER BY created_at DESC LIMIT 1
`, userID, tc.AssetID, swapBot, tc.StartAt, tc.EndAt).Scan(&end); err != nil {
			session.Logger(ctx).Println(err)
			return err
		}
		diff := end.Sub(start)
		log.Println("start", start)
		log.Println("end", end)
		if diff.LessThanOrEqual(decimal.Zero) {
			diff = decimal.Zero
		}
		if diff.LessThan(amount) {
			amount = diff
		}
	}
	if _, err := session.Database(ctx).Exec(ctx, `
INSERT INTO trading_rank (competition_id,asset_id,user_id,amount,updated_at)
VALUES ($1,$2,$3,$4,NOW())
ON CONFLICT (competition_id,user_id) DO UPDATE SET amount=$4,updated_at=NOW()
	`, tc.CompetitionID, tc.AssetID, userID, amount); err != nil {
		session.Logger(ctx).Println(err)
		return err
	}
	return nil
}

func StatisticUserSnapshots(ctx context.Context, userID, token, assetID string, startAt time.Time) error {
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT created_at FROM user_snapshots 
WHERE user_id=$1
ORDER BY created_at DESC LIMIT 1
	`, userID).Scan(&startAt); durable.CheckNotEmptyError(err) != nil {
		return err
	}
	ss, err := mixin.ReadSnapshots(ctx, token, assetID, startAt, "ASC", 500)
	if err != nil {
		return err
	}
	for _, s := range ss {
		if _, err := session.Database(ctx).Exec(ctx, `
INSERT INTO user_snapshots (snapshot_id,user_id,opponent_id,asset_id,amount,opening_balance,closing_balance ,source,created_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
`, s.SnapshotID, userID, s.OpponentID, s.AssetID, s.Amount, s.OpeningBalance, s.ClosingBalance, s.Source, s.CreatedAt); err != nil {
			if !durable.CheckIsPKRepeatError(err) {
				return err
			}
		}
	}

	if len(ss) < 500 {
		return nil
	}
	return StatisticUserSnapshots(ctx, userID, token, assetID, startAt)
}
