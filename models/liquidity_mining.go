package models

import (
	"context"
	"errors"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/MixinNetwork/supergroup/tools"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const liquidity_mining_DDL = `
CREATE TABLE IF NOT EXISTS liquidity_mining (
	mining_id VARCHAR(36) NOT NULL PRIMARY KEY,

	title VARCHAR NOT NULL,
	description VARCHAR NOT NULL,
	faq VARCHAR NOT NULL,
	asset_id VARCHAR(36) NOT NULL,
	client_id VARCHAR(36) NOT NULL,
	first_time timestamp NOT NULL DEFAULT NOW(),
	first_end  timestamp NOT NULL DEFAULT NOW(),
	daily_time timestamp NOT NULL DEFAULT NOW(),
	daily_end  timestamp NOT NULL DEFAULT NOW(),
	
	reward_asset_id VARCHAR(36) NOT NULL,
	first_amount varchar NOT NULL DEFAULT '0',
	daily_amount varchar NOT NULL DEFAULT '0',

	extra_asset_id varchar NOT NULL DEFAULT '',
	extra_first_amount varchar NOT NULL DEFAULT '0',
	extra_daily_amount varchar NOT NULL DEFAULT '0',
	created_at timestamp NOT NULL DEFAULT NOW()
);
`

type LiquidityMining struct {
	MiningID         string          `json:"mining_id"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Faq              string          `json:"faq"`
	AssetID          string          `json:"asset_id"`
	ClientID         string          `json:"client_id"`
	FirstTime        time.Time       `json:"first_time"`
	FirstEnd         time.Time       `json:"first_end"`
	DailyTime        time.Time       `json:"daily_time"`
	DailyEnd         time.Time       `json:"daily_end"`
	RewardAssetID    string          `json:"reward_asset_id"`
	FirstAmount      decimal.Decimal `json:"first_amount"`
	DailyAmount      decimal.Decimal `json:"daily_amount"`
	ExtraAssetID     string          `json:"extra_asset_id"`
	ExtraFirstAmount decimal.Decimal `json:"extra_first_amount"`
	ExtraDailyAmount decimal.Decimal `json:"extra_daily_amount"`
	CreatedAt        time.Time       `json:"created_at"`
}

const (
	LiquidityMiningFirst = 1 // 头矿挖矿
	LiquidityMiningDaily = 2 // 日矿挖矿
)

func CreateLiquidityMining(ctx context.Context, m *LiquidityMining) error {
	query := durable.InsertQuery("liquidity_mining", "mining_id, title, description, asset_id, first_time, first_end, daily_time, daily_end, reward_asset_id, first_amount, daily_amount, extra_asset_id, extra_first_amount, extra_daily_amount")
	_, err := session.Database(ctx).Exec(ctx, query, m.MiningID, m.Title, m.Description, m.AssetID, m.FirstTime, m.FirstEnd, m.DailyTime, m.DailyEnd, m.RewardAssetID, m.FirstAmount, m.DailyAmount, m.ExtraAssetID, m.ExtraFirstAmount, m.ExtraDailyAmount)
	return err
}

func GetLiquidityMiningByID(ctx context.Context, id string) (*LiquidityMining, error) {
	var m LiquidityMining
	err := session.Database(ctx).QueryRow(ctx, `
SELECT mining_id, asset_id, first_time, first_end, daily_time, daily_end, reward_asset_id, first_amount, daily_amount, extra_asset_id, extra_first_amount, extra_daily_amount
FROM liquidity_mining WHERE mining_id=$1`, id).
		Scan(&m.MiningID, &m.AssetID, &m.FirstTime, &m.FirstEnd, &m.DailyTime, &m.DailyEnd, &m.RewardAssetID, &m.FirstAmount, &m.DailyAmount, &m.ExtraAssetID, &m.ExtraFirstAmount, &m.ExtraDailyAmount)
	return &m, err
}

func GetLiquidtityMiningListByID(ctx context.Context) ([]*LiquidityMining, error) {
	ms := make([]*LiquidityMining, 0)
	err := session.Database(ctx).ConnQuery(ctx, `
SELECT mining_id, asset_id, first_time, first_end, daily_time, daily_end, reward_asset_id, first_amount, daily_amount, extra_asset_id, extra_first_amount, extra_daily_amount
FROM liquidity_mining`, func(rows pgx.Rows) error {
		for rows.Next() {
			var m LiquidityMining
			if err := rows.Scan(&m.MiningID, &m.AssetID, &m.FirstTime, &m.FirstEnd, &m.DailyTime, &m.DailyEnd, &m.RewardAssetID, &m.FirstAmount, &m.DailyAmount, &m.ExtraAssetID, &m.ExtraFirstAmount, &m.ExtraDailyAmount); err != nil {
				return err
			}
			ms = append(ms, &m)
		}
		return nil
	})
	return ms, err
}

func HandleStatictis(ctx context.Context) {
	ms, err := GetLiquidtityMiningListByID(ctx)
	if err != nil {
		session.Logger(ctx).Println(err)
		return
	}
	for _, m := range ms {
		// 1. 还没到 first_time 结束
		if m.FirstTime.After(time.Now()) {
			continue
		}
		// 2. 到了first_time，没到 first_end，这个时候是头矿奖励
		if m.FirstEnd.After(time.Now()) {
			// 处理头矿奖励
			continue
		}
		// 3. 到了first_end，没到 daily_time 结束
		if m.DailyTime.After(time.Now()) {
			continue
		}
		// 4. 到了daily_time，没到 daily_end，这个时候是每日奖励
		if m.DailyEnd.After(time.Now()) {
			// 处理挖矿奖励
			continue
		}
		// 5. 到了daily_end，结束
	}
}

func handleStatisticsAssets(ctx context.Context, m *LiquidityMining, mintStatus int) error {
	// 获取参与活动的用户
	users, err := GetLiquidityMiningUsersByID(ctx, m.MiningID)
	if err != nil {
		return err
	}
	// 获取流动性资产
	lpAssets, err := GetClientAssetLPCheckMapByID(ctx, m.ClientID)
	if err != nil {
		return err
	}

	assetReward := m.FirstAmount
	extraReward := m.ExtraFirstAmount

	if mintStatus == LiquidityMiningDaily {
		assetReward = m.DailyAmount
		extraReward = m.ExtraDailyAmount
	}

	totalAmount, usersAmount := statisticsUsersPartAndTotalAmount(ctx, m.MiningID, users, lpAssets)
	if err != nil {
		return err
	}
	if totalAmount.IsZero() {
		session.Logger(ctx).Println(m.MiningID + "... totalAmount is zero...")
		return nil
	}
	for userID, v := range usersAmount {
		if v.IsZero() {
			continue
		}
		// 份额
		part := v.Div(totalAmount)
		assetRewardAmount := assetReward.Mul(part).Truncate(8)
		extraAssetRewardAmount := extraReward.Mul(part).Truncate(8)
		if err := session.Database(ctx).RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
			if err := CreateLiquidityMiningRecordWithTx(ctx, tx, &LiquidityMiningRecord{
				MiningID: m.MiningID,
				AssetID:  m.AssetID,
				UserID:   userID,
				Amount:   assetRewardAmount,
				TraceID:  tools.GetUUID(),
				Status:   LiquidityMiningRecordStatusPending,
			}); err != nil {
				return err
			}
			if err := CreateLiquidityMiningRecordWithTx(ctx, tx, &LiquidityMiningRecord{
				MiningID: m.MiningID,
				AssetID:  m.ExtraAssetID,
				UserID:   userID,
				Amount:   extraAssetRewardAmount,
				TraceID:  tools.GetUUID(),
				Status:   LiquidityMiningRecordStatusPending,
			}); err != nil {
				return err
			}
			return nil
		}); err != nil {
			session.Logger(ctx).Println(err)
			continue
		}
	}
	return nil
}

func statisticsUsersPartAndTotalAmount(ctx context.Context, mintID string, users []*User, lpAssets map[string]decimal.Decimal) (decimal.Decimal, map[string]decimal.Decimal) {
	// 统计每个用户的流动性资产
	totalAmount := decimal.Zero
	usersAmount := make(map[string]decimal.Decimal, 0)
	for _, u := range users {
		userAssets, err := GetUserAssets(ctx, u.AccessToken)
		if err != nil {
			if errors.Is(err, session.ForbiddenError(ctx)) {
				// 取消授权的用户，添加一条未参与的记录
				if err := CreateLiquidityMiningRecord(ctx, &LiquidityMiningRecord{
					MiningID: mintID,
					UserID:   u.UserID,
					AssetID:  "",
					Amount:   decimal.Zero,
					Status:   LiquidityMiningRecordStatusFailed,
					TraceID:  "",
				}); err != nil {
					session.Logger(ctx).Println(err)
				}
				continue
			}
			session.Logger(ctx).Println(err)
			continue
		}
		// 检查流动性资产
		for _, a := range userAssets {
			if price, ok := lpAssets[a.AssetID]; ok {
				if price.GreaterThan(decimal.Zero) {
					// 用户的分数 和 总分数加
					if _, ok := usersAmount[u.UserID]; !ok {
						usersAmount[u.UserID] = decimal.Zero
					}
					addPart := a.Balance.Mul(price)
					usersAmount[u.UserID] = usersAmount[u.UserID].Add(addPart)
					totalAmount = totalAmount.Add(addPart)
				}
				break
			}
		}
	}
	return totalAmount, usersAmount
}
