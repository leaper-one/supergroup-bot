package common

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/jackc/pgx/v4"
	"github.com/robfig/cron/v3"
	"github.com/shopspring/decimal"
)

type Liquidity struct {
	LiquidityID string          `json:"liquidity_id"`
	ClientID    string          `json:"client_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	StartAt     time.Time       `json:"start_at"`
	EndAt       time.Time       `json:"end_at"`
	AssetIDs    string          `json:"asset_ids"`
	MinAmount   decimal.Decimal `json:"min_amount"`
	LpDesc      string          `json:"lp_desc"`
	LpURL       string          `json:"lp_url"`
	CreatedAt   time.Time       `json:"created_at"`
}

const liquidity_DDL = `
CREATE TABLE IF NOT EXISTS liquidity (
	liquidity_id VARCHAR(36) NOT NULL PRIMARY KEY,
	client_id VARCHAR(36) NOT NULL,
	title varchar DEFAULT '',
	description varchar DEFAULT '',
	start_at timestamp WITH TIME ZONE NOT NULL,
	end_at timestamp WITH TIME ZONE NOT NULL,
	asset_ids varchar DEFAULT '',
	min_amount varchar DEFAULT '0',
	lp_desc varchar DEFAULT '',
	lp_url varchar DEFAULT '',
	created_at timestamp WITH TIME ZONE DEFAULT NOW()
);
`

type LiquidityDetail struct {
	LiquidityID string          `json:"liquidity_id,omitempty"`
	Idx         int             `json:"idx,omitempty"`
	StartAt     time.Time       `json:"start_at,omitempty"`
	EndAt       time.Time       `json:"end_at,omitempty"`
	AssetID     string          `json:"asset_id,omitempty"`
	Amount      decimal.Decimal `json:"amount,omitempty"`
	Symbol      string          `json:"symbol,omitempty"`
}

const liquidity_detail_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_detail (
	liquidity_id VARCHAR(36),
	idx int NOT NULL,
	start_at timestamp WITH TIME ZONE NOT NULL,
	end_at timestamp WITH TIME ZONE NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	amount varchar DEFAULT '0',
	symbol varchar DEFAULT '',
	created_at timestamp WITH TIME ZONE DEFAULT NOW()
);
`

type LiquidityUser struct {
	LiquidityID string    `json:"liquidity_id"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

const liquidity_user_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_user (
	liquidity_id VARCHAR(36) NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	created_at timestamp WITH TIME ZONE DEFAULT NOW(),
	PRIMARY KEY (liquidity_id, user_id)
);
`

type LiquiditySnapshot struct {
	UserID      string          `json:"user_id,omitempty"`
	LiquidityID string          `json:"liquidity_id,omitempty"`
	Idx         int             `json:"idx,omitempty"`
	Date        string          `json:"date,omitempty"`
	LpSymbol    string          `json:"lp_symbol,omitempty"`
	LpAmount    decimal.Decimal `json:"lp_amount,omitempty"`
	UsdValue    decimal.Decimal `json:"usd_value,omitempty"`
}

const liquidity_snapshot_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_snapshot (
	user_id VARCHAR(36) NOT NULL,
	liquidity_id VARCHAR(36) NOT NULL,
	idx int NOT NULL,
	date date NOT NULL,
	lp_symbol varchar DEFAULT '',
	lp_amount varchar DEFAULT '0',
	usd_value varchar DEFAULT '0'
);
`

type LiquidityTx struct {
	LiquidityID string    `json:"liquidity_id"`
	Month       time.Time `json:"month"`
	Idx         int       `json:"idx"`
	UserID      string    `json:"user_id"`
	AssetID     string    `json:"asset_id"`
	Amount      string    `json:"amount"`
	TraceID     string    `json:"trace_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	LiquidityTxWait    = "W"
	LiquidityTxPending = "P"
	LiquidityTxSuccess = "S"

	LiquidityTxFail = "F"
)

const liquidity_tx_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_tx (
	trace_id VARCHAR(36) NOT NULL PRIMARY KEY,
	month date NOT NULL,
	liquidity_id VARCHAR(36) NOT NULL,
	idx int NOT NULL,
	user_id VARCHAR(36) NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	amount varchar DEFAULT '0',
	status varchar DEFAULT 'W',
	created_at timestamp WITH TIME ZONE DEFAULT NOW()
);
`

type liquidityResp struct {
	Info            *Liquidity         `json:"info"`
	List            []*LiquidityDetail `json:"list"`
	YesterdayAmount decimal.Decimal    `json:"yesterday_amount"`
	IsJoin          bool               `json:"is_join"`
	Scope           string             `json:"scope"`
}

// 获取活动页面详情
func GetLiquidityInfo(ctx context.Context, u *ClientUser, id string) (*liquidityResp, error) {
	info, err := GetLiquidityByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var list []*LiquidityDetail
	if err := session.DB(ctx).ConnQuery(ctx, `
SELECT liquidity_id, idx, start_at, end_at, asset_id, amount, symbol
FROM liquidity_detail WHERE liquidity_id = $1 ORDER BY idx
	`, func(rows pgx.Rows) error {
		for rows.Next() {
			var ld LiquidityDetail
			err := rows.Scan(&ld.LiquidityID, &ld.Idx, &ld.StartAt, &ld.EndAt, &ld.AssetID, &ld.Amount, &ld.Symbol)
			if err != nil {
				return err
			}
			list = append(list, &ld)
		}
		return nil
	}, id); err != nil {
		return nil, err
	}
	var yesterdayAmount decimal.Decimal
	if err := session.DB(ctx).QueryRow(ctx, `
SELECT lp_amount from liquidity_snapshot
WHERE user_id=$1 AND date=CURRENT_DATE-1
	`, u.UserID).Scan(&yesterdayAmount); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
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

func GetLiquidityByID(ctx context.Context, id string) (*Liquidity, error) {
	var info Liquidity
	if err := session.DB(ctx).QueryRow(ctx, `
SELECT liquidity_id, client_id, title, description, asset_ids, start_at, end_at, min_amount, lp_desc, lp_url, created_at
FROM liquidity WHERE liquidity_id = $1
	`, id).Scan(&info.LiquidityID, &info.ClientID, &info.Title, &info.Description, &info.AssetIDs, &info.StartAt, &info.EndAt, &info.MinAmount, &info.LpDesc, &info.LpURL, &info.CreatedAt); err != nil {
		return nil, err
	}
	return &info, nil
}

func checkUserIsJoinLiquidity(ctx context.Context, lid, uid string) bool {
	var join int64
	if err := session.DB(ctx).QueryRow(ctx, `
SELECT COUNT(1) FROM liquidity_user
WHERE liquidity_id=$1 AND user_id=$2
`, lid, uid).Scan(&join); err != nil {
		tools.Println(err)
		return false
	}
	return join == 1
}

// 参与
func PostLiquidity(ctx context.Context, u *ClientUser, id string) (string, error) {
	if time.Now().UTC().Day() != 1 {
		return "miss", nil
	}
	l, err := GetLiquidityByID(ctx, id)
	if err != nil {
		return "", err
	}
	asset, err := GetUserAsset(ctx, u, l.AssetIDs)
	if err != nil {
		return "", err
	}
	if _, err := GetUserSnapshots(ctx, u, "", time.Now(), "", 1); err != nil {
		return "", err
	}
	if checkUserIsJoinLiquidity(ctx, id, u.UserID) {
		return "success", nil
	}
	if _, err := session.DB(ctx).Exec(ctx, durable.InsertQuery("liquidity_user", "liquidity_id,user_id"), id, u.UserID); err != nil {
		return "", err
	}
	if asset.Balance.LessThan(l.MinAmount) {
		return "limit", nil
	}
	return "success", nil
}

type liquidityRecord struct {
	Duration string               `json:"duration"`
	Status   string               `json:"status"`
	List     []*LiquiditySnapshot `json:"list"`
}

func GetLiquiditySnapshots(ctx context.Context, u *ClientUser, id string) ([]*liquidityRecord, error) {
	var lts []*LiquidityTx
	if err := session.DB(ctx).ConnQuery(ctx, `
SELECT liquidity_id, month, amount, status
FROM liquidity_tx WHERE user_id = $1 AND liquidity_id = $2
ORDER BY month DESC
	`, func(rows pgx.Rows) error {
		for rows.Next() {
			var lt LiquidityTx
			if err := rows.Scan(&lt.LiquidityID, &lt.Month, &lt.Amount, &lt.Status); err != nil {
				return err
			}
			lts = append(lts, &lt)
		}
		return nil
	}, u.UserID, id); err != nil {
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
		var lss []*LiquiditySnapshot
		if err := session.DB(ctx).ConnQuery(ctx, `
SELECT to_char(date, 'YYYY-MM-DD'), lp_amount, lp_symbol
FROM liquidity_snapshot WHERE user_id = $1 AND liquidity_id = $2 AND date >= $3 AND date <= $4
ORDER BY date DESC
		`, func(rows pgx.Rows) error {
			for rows.Next() {
				var ls LiquiditySnapshot
				if err := rows.Scan(&ls.Date, &ls.LpAmount, &ls.LpSymbol); err != nil {
					return err
				}
				lss = append(lss, &ls)
			}
			return nil
		}, u.UserID, id, lt.Month, lt.Month.AddDate(0, 1, -1)); err != nil {
			return nil, err
		}
		item.List = lss
		res = append(res, item)
	}
	return res, nil
}

func StartLiquidityDailyJob() {
	c := cron.New(cron.WithLocation(time.UTC))
	_, err := c.AddFunc("1 0 * * *", func() {
		log.Println("start liquidity job")
		StatisticLiquidityDaily(_ctx)
		// StatisticLiquidityMonth(_ctx)
	})
	if err != nil {
		session.Logger(_ctx).Println(err)
		SendMsgToDeveloper(_ctx, "", "定时任务StartLiquidityJob。。。出问题了。。。")
		return
	}
	c.Start()
}

// 统计每日的情况
func StatisticLiquidityDaily(ctx context.Context) error {
	// 1. 获取所有的 liquidity
	var liquidities []*Liquidity
	if err := session.DB(ctx).ConnQuery(ctx, `
SELECT liquidity_id, client_id, asset_ids, min_amount
FROM liquidity
WHERE now() > start_at AND now() <= end_at
`, func(rows pgx.Rows) error {
		for rows.Next() {
			var l Liquidity
			if err := rows.Scan(&l.LiquidityID, &l.ClientID, &l.AssetIDs, &l.MinAmount); err != nil {
				tools.Println(err)
				return err
			}
			liquidities = append(liquidities, &l)
		}
		return nil
	}); err != nil {
		return err
	}
	// 2. 选择一个 liquidity，然后获取所有的用户
	for _, l := range liquidities {
		var users []string
		if err := session.DB(ctx).ConnQuery(ctx, `
SELECT user_id FROM liquidity_user
WHERE liquidity_id=$1
		`, func(rows pgx.Rows) error {
			for rows.Next() {
				var user string
				if err := rows.Scan(&user); err != nil {
					return err
				}
				users = append(users, user)
			}
			return nil
		}, l.LiquidityID); err != nil {
			tools.Println(err)
			continue
		}
		// 10 9 9:00 9:24
		now := time.Now().UTC()
		endAt := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		startAt := endAt.AddDate(0, 0, -1)
		asset, err := GetAssetByID(ctx, nil, l.AssetIDs)
		if err != nil {
			tools.Println(err)
			continue
		}
		for _, uid := range users {
			// 3.0 判断上一次是否达标，如果不达标则跳过
			if time.Now().UTC().Day() != 1 {
				amount, err := getRecentSnapshot(ctx, l.LiquidityID, uid)
				if err != nil {
					tools.Println(err)
					continue
				}
				if !amount.IsZero() && amount.LessThan(l.MinAmount) {
					continue
				}
			}

			// 3.1 获取该用户的资产, 使指定 asset 为初始值
			u, err := GetClientUserByClientIDAndUserID(ctx, l.ClientID, uid)
			if err != nil {
				tools.Println(err)
				continue
			}
			a, err := GetUserAsset(ctx, &u, l.AssetIDs)
			if err != nil {
				tools.Println(err)
				if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "401") {
					a, err := GetAssetByID(ctx, nil, l.AssetIDs)
					if err != nil {
						tools.Println(err)
						continue
					}
					if _, err := session.DB(ctx).Exec(ctx,
						durable.InsertQuery("liquidity_snapshot",
							"user_id,liquidity_id,idx,date,lp_symbol,lp_amount,usd_value"),
						// TODO
						u.UserID, l.LiquidityID, 1, startAt, a.Symbol, "0", "0"); err != nil {
						tools.Println(err)
						continue
					}
					if _, err := session.DB(ctx).Exec(ctx,
						durable.InsertQuery("liquidity_tx",
							"trace_id,month,liquidity_id,idx,user_id,asset_id,status"),
						tools.GetUUID(), time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), l.LiquidityID, 0, u.UserID, "", LiquidityTxFail); err != nil {
						tools.Println(err)
						continue
					}
				}
				continue
			}
			minAmount, err := GetMinAmount(ctx, &u, l.AssetIDs, startAt, endAt, a.Balance, true)
			if err != nil {
				tools.Println(err)
				continue
			}
			// 5. 保存该值，结束
			if _, err := session.DB(ctx).Exec(ctx,
				durable.InsertQuery("liquidity_snapshot",
					"user_id,liquidity_id,idx,date,lp_symbol,lp_amount,usd_value"),
				// TODO
				u.UserID, l.LiquidityID, 1, startAt, a.Symbol, minAmount, asset.PriceUsd.Mul(minAmount)); err != nil {
				tools.Println(err)
				continue
			}

			if minAmount.LessThan(l.MinAmount) {
				if _, err := session.DB(ctx).Exec(ctx,
					durable.InsertQuery("liquidity_tx",
						"trace_id,month,liquidity_id,idx,user_id,asset_id,status"),
					tools.GetUUID(), time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC), l.LiquidityID, 0, u.UserID, "", LiquidityTxFail); err != nil {
					tools.Println(err)
					continue
				}
			}
		}
	}

	return nil
}

func StatisticLiquidityMonth(ctx context.Context) error {
	now := time.Now().UTC()
	if now.Day() != 1 {
		return nil
	}
	lastMonthFirstDay := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -1, 0)
	lastMonthLastDay := lastMonthFirstDay.AddDate(0, 1, 0).Add(-time.Second)
	// 1. 获取上个月的 liquidity_detail
	var lds []*LiquidityDetail
	if err := session.DB(ctx).ConnQuery(ctx, `
SELECT liquidity_id, asset_id, amount FROM liquidity_detail
WHERE start_at>=$1 AND end_at<$2
	`, func(rows pgx.Rows) error {
		for rows.Next() {
			var ld LiquidityDetail
			if err := rows.Scan(&ld.LiquidityID, &ld.AssetID, &ld.Amount); err != nil {
				return err
			}
			lds = append(lds, &ld)
		}
		return nil
	}, lastMonthFirstDay, lastMonthLastDay); err != nil {
		return err
	}

	for _, ld := range lds {
		// 2. 获取该 liquidity 的所有参与者
		var users []string
		if err := session.DB(ctx).ConnQuery(ctx, `
SELECT user_id FROM liquidity_user
WHERE liquidity_id=$1
		`, func(rows pgx.Rows) error {
			for rows.Next() {
				var uid string
				if err := rows.Scan(&uid); err != nil {
					return err
				}
				users = append(users, uid)
			}
			return nil
		}, ld.LiquidityID); err != nil {
			return err
		}

		// 3. 遍历所有参与者，获得 lp_amount_map
		lpUserAmountMap := make(map[string]decimal.Decimal)
		totalAmount := decimal.Zero
		for _, uid := range users {
			// 3.1 判断该用户是否存在 liquidity_tx
			var txStatus string
			if err := session.DB(ctx).QueryRow(ctx, `
SELECT status FROM liquidity_tx
WHERE liquidity_id=$1 AND user_id=$2 AND month=$3
		`, ld.LiquidityID, uid, lastMonthFirstDay).Scan(&txStatus); err != nil {
				if !errors.Is(err, pgx.ErrNoRows) {
					return err
				}
			}
			if txStatus == LiquidityTxFail {
				continue
			}

			// 3.2 获取该用户上月的 lp_amount，
			if err := session.DB(ctx).ConnQuery(ctx, `
SELECT lp_amount FROM liquidity_snapshot
WHERE liquidity_id=$1 AND user_id=$2 AND date>=$3 AND date<$4
		`, func(rows pgx.Rows) error {
				for rows.Next() {
					var amount decimal.Decimal
					if err := rows.Scan(&amount); err != nil {
						return err
					}
					lpUserAmountMap[uid] = lpUserAmountMap[uid].Add(amount)
					totalAmount = totalAmount.Add(amount)
				}
				return nil
			}, ld.LiquidityID, uid, lastMonthFirstDay, lastMonthLastDay); err != nil {
				return err
			}
		}

		for uid, amount := range lpUserAmountMap {
			// 4. 计算每个用户的分成
			share := amount.Div(totalAmount)
			// 5. 计算每个用户的分成金额
			amount = share.Mul(ld.Amount).Truncate(8)
			// 6. 插入 liquidity_tx
			if _, err := session.DB(ctx).Exec(ctx,
				durable.InsertQuery("liquidity_tx",
					"trace_id,month,liquidity_id,idx,user_id,asset_id,status,amount"),
				tools.GetUUID(), lastMonthFirstDay, ld.LiquidityID, 0, uid, ld.AssetID, LiquidityTxSuccess, amount); err != nil {
				return err
			}
		}
	}
	return nil
}

func getRecentSnapshot(ctx context.Context, lid, uid string) (decimal.Decimal, error) {
	var amount decimal.Decimal
	if err := session.DB(ctx).QueryRow(ctx, `
SELECT lp_amount
FROM liquidity_snapshot
WHERE liquidity_id=$1 AND user_id=$2
ORDER BY date DESC
LIMIT 1
`, lid, uid).Scan(&amount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return decimal.Zero, nil
		}
		return decimal.Zero, err
	}
	return amount, nil
}

func GetMinAmount(ctx context.Context, u *ClientUser, assetID string, startAt, endAt time.Time, minAmount decimal.Decimal, isStart bool) (decimal.Decimal, error) {
	// 4. 获取该用户的 snapshot，遍历最近一天的 snapshot，取 asset 的最低值
	ss, err := GetUserSnapshots(ctx, u, assetID, endAt, "DESC", 500)
	if err != nil {
		tools.Println(err)
		return decimal.Zero, err
	}
	if !isStart {
		if len(ss) == 1 {
			return minAmount, nil
		}
		ss = ss[1:]
	}
	for _, s := range ss {
		if s.CreatedAt.Before(startAt) {
			return minAmount, nil
		}
		if s.ClosingBalance.LessThan(minAmount) {
			minAmount = s.ClosingBalance
		}
	}
	return GetMinAmount(ctx, u, assetID, startAt, ss[len(ss)-1].CreatedAt, minAmount, false)
}
