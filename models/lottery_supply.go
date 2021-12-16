package models

import (
	"context"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/session"
	"github.com/jackc/pgx/v4"
	"github.com/shopspring/decimal"
)

const lottery_supply_DDL = `
CREATE TABLE IF NOT EXISTS lottery_supply (
	supply_id    VARCHAR(36) PRIMARY KEY,
	lottery_id   VARCHAR NOT NULL,
	asset_id     VARCHAR NOT NULL,
	amount       VARCHAR NOT NULL,
	client_id    VARCHAR NOT NULL,
	icon_url     VARCHAR NOT NULL,
	status       SMALLINT NOT NULL,
	created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

type LotterySupply struct {
	SupplyID  string          `json:"supply_id"`
	LotteryID string          `json:"lottery_id"`
	AssetID   string          `json:"asset_id"`
	Amount    decimal.Decimal `json:"amount"`
	ClientID  string          `json:"client_id"`
	IconURL   string          `json:"icon_url"`
	Status    int             `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

const (
	LotterySupplyStatusListing = 1
	LotterySupplyStatusEnd     = 2
)

const lottery_supply_received_DDL = `
CREATE TABLE IF NOT EXISTS lottery_supply_received (
	supply_id    VARCHAR(36) PRIMARY KEY,
	user_id 		 VARCHAR(36) NOT NULL,
	trace_id 		 VARCHAR(36) NOT NULL,
	created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
	PRIMARY KEY (supply_id, user_id)
);
`

type LotterySupplyReceived struct {
	SupplyID  string    `json:"supply_id"`
	UserID    string    `json:"user_id"`
	TraceID   string    `json:"trace_id"`
	CreatedAt time.Time `json:"created_at"`
}

func getUserListingLottery(ctx context.Context, u *ClientUser) []config.Lottery {
	lotteryList := config.Config.Lottery.List
	lss, err := getAllListingLottery(ctx)
	if err != nil {
		session.Logger(ctx).Println(err)
		return nil
	}
	lssID := make([]string, 0)
	for id := range lss {
		lssID = append(lssID, id)
	}
	session.Database(ctx).ConnQuery(ctx, `
SELECT supply_id 
FROM lottery_supply_received 
WHERE user_id=$1 
AND supply_id=ANY($2)`, func(rows pgx.Rows) error {
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			delete(lss, id)
		}
		return nil
	}, u.UserID, lssID)

	for _, ls := range lss {
		id, _ := strconv.Atoi(ls.LotteryID)
		lotteryList[id] = config.Lottery{
			LotteryID: ls.LotteryID,
			AssetID:   ls.AssetID,
			Amount:    ls.Amount,
			IconURL:   ls.IconURL,
			ClientID:  ls.ClientID,
			SupplyID:  ls.SupplyID,
		}
	}

	return lotteryList
}

func getAllListingLottery(ctx context.Context) (map[string]*LotterySupply, error) {
	lss := make(map[string]*LotterySupply)
	if err := session.Database(ctx).ConnQuery(ctx,
		`SELECT supply_id, lottery_id, asset_id, amount, client_id, icon_url FROM lottery_supply WHERE status = 1`,
		func(rows pgx.Rows) error {
			for rows.Next() {
				var ls LotterySupply
				if err := rows.Scan(&ls.SupplyID, &ls.LotteryID, &ls.AssetID, &ls.Amount, &ls.ClientID, &ls.IconURL); err != nil {
					return err
				}
				lss[ls.SupplyID] = &ls
			}
			return nil
		}); err != nil {
		return nil, err
	}

	return nil, nil
}
