package models

import (
	"context"
	"strconv"
	"time"

	"github.com/MixinNetwork/supergroup/config"
	"github.com/MixinNetwork/supergroup/durable"
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
	status       SMALLINT NOT NULL, -- 1 上架 2 手动下架 3 库存为0 自动下架
	inventory		 INT NOT NULL DEFAULT -1,
	created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
`

type LotterySupply struct {
	SupplyID  string          `json:"supply_id"`
	LotteryID string          `json:"lottery_id"`
	AssetID   string          `json:"asset_id"`
	Inventory int             `json:"inventory"`
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
	supply_id    VARCHAR(36),
	user_id 		 VARCHAR(36) NOT NULL,
	trace_id 		 VARCHAR(36) NOT NULL,
	status 		   SMALLINT NOT NULL DEFAULT 1, -- 1 待领取 2 已领取
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

func getLotteryByTrace(ctx context.Context, traceID string) (*config.Lottery, error) {
	supplyID := ""
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT supply_id FROM lottery_supply_received WHERE trace_id=$1
	`, traceID).Scan(&supplyID); durable.CheckNotEmptyError(err) != nil {
		return nil, err
	}
	if supplyID != "" {
		supply, err := getLotterySupplyBySupplyID(ctx, supplyID)
		if err != nil {
			return nil, err
		}
		return &config.Lottery{
			LotteryID: supply.LotteryID,
			AssetID:   supply.AssetID,
			Amount:    supply.Amount,
			IconURL:   supply.IconURL,
			ClientID:  supply.ClientID,
			SupplyID:  supply.SupplyID,
			Inventory: supply.Inventory,
		}, nil
	}
	lotteryID := ""
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT lottery_id FROM lottery_record WHERE trace_id=$1
	`, traceID).Scan(&lotteryID); err != nil {
		return nil, err
	}
	list := getInitListingLottery()
	i, _ := strconv.Atoi(lotteryID)
	return &list[i], nil
}

func getLotteryByID(ctx context.Context, id, userID string) *config.Lottery {
	list := getUserListingLottery(ctx, userID)
	i, _ := strconv.Atoi(id)
	return &list[i]
}

func getUserListingLottery(ctx context.Context, userID string) [16]config.Lottery {
	lotteryList := getInitListingLottery()
	lss, err := getAllListingLottery(ctx)
	if err != nil {
		session.Logger(ctx).Println(err)
		return [16]config.Lottery{}
	}
	lssID := make([]string, 0)
	for id := range lss {
		lssID = append(lssID, id)
	}
	query := `
SELECT supply_id
FROM lottery_supply_received
WHERE user_id=$1
AND supply_id=ANY($2)`
	if err := session.Database(ctx).ConnQuery(ctx, query, func(rows pgx.Rows) error {
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				return err
			}
			delete(lss, id)
		}
		return nil
	}, userID, lssID); err != nil {
		session.Logger(ctx).Println(err)
	}
	for _, ls := range lss {
		id, _ := strconv.Atoi(ls.LotteryID)
		lotteryList[id] = config.Lottery{
			LotteryID: ls.LotteryID,
			AssetID:   ls.AssetID,
			Amount:    ls.Amount,
			IconURL:   ls.IconURL,
			ClientID:  ls.ClientID,
			SupplyID:  ls.SupplyID,
			Inventory: ls.Inventory,
		}
	}
	return lotteryList
}

func getInitListingLottery() [16]config.Lottery {
	test := [16]config.Lottery{}
	for i, v := range config.Config.Lottery.List {
		test[i] = v
	}
	return test
}

func getLotterySupplyBySupplyID(ctx context.Context, supplyID string) (*LotterySupply, error) {
	var ls LotterySupply
	if err := session.Database(ctx).QueryRow(ctx, `
SELECT supply_id, lottery_id, asset_id, inventory, amount, client_id, icon_url, status, created_at
FROM lottery_supply
WHERE supply_id=$1
	`, supplyID).Scan(&ls.SupplyID, &ls.LotteryID, &ls.AssetID, &ls.Inventory, &ls.Amount, &ls.ClientID, &ls.IconURL, &ls.Status, &ls.CreatedAt); err != nil {
		return nil, err
	}
	return &ls, nil
}

func getAllListingLottery(ctx context.Context) (map[string]*LotterySupply, error) {
	lss := make(map[string]*LotterySupply)
	if err := session.Database(ctx).ConnQuery(ctx, `
SELECT supply_id, lottery_id, asset_id, amount, client_id, icon_url, inventory
FROM lottery_supply 
WHERE status=1`,
		func(rows pgx.Rows) error {
			for rows.Next() {
				var ls LotterySupply
				if err := rows.Scan(&ls.SupplyID, &ls.LotteryID, &ls.AssetID, &ls.Amount, &ls.ClientID, &ls.IconURL, &ls.Inventory); err != nil {
					return err
				}
				if ls.Inventory != 0 {
					lss[ls.SupplyID] = &ls
				}
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return lss, nil
}

func createLotterySupplyRecord(ctx context.Context, tx pgx.Tx, supplyID, userID, traceID string) error {
	_, err := tx.Exec(ctx, `INSERT INTO lottery_supply_received(supply_id, user_id, trace_id) VALUES($1, $2, $3)`, supplyID, userID, traceID)
	return err
}
