package models

import (
	"context"
	"database/sql"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Ctx context.Context

func init() {
	db := durable.NewDB()

	Ctx = session.WithDatabase(context.Background(), db)
	Ctx = session.WithRedis(Ctx, durable.NewRedis(context.Background()))

	db.AutoMigrate(
		&Activity{},
		&Airdrop{},
		&ClientAssetLevel{},
		&ClientAssetLpCheck{},
		&Asset{},
		&ExinOtcAsset{},
		&ExinLocalAsset{},
		&Broadcast{},
		&ClientBlockUser{},
		&BlockUser{},
		&ClientMemberAuth{},
		&ClientMenu{},
		&ClientReplay{},
		&ClientUserProxy{},
		&LoginLog{},
		&ClientUser{},
		&ClientWhiteURL{},
		&Client{},
		&DailyData{},
		&Invitation{},
		&InvitationPowerRecord{},
		&LiquidityMining{},
		&LiquidityMiningUser{},
		&LiquidityMiningTx{},
		&LiquidityMiningRecord{},
		&Liquidity{},
		&LiquidityDetail{},
		&LiquidityUser{},
		&LiquiditySnapshot{},
		&LiquidityTx{},
		&Live{},
		&LiveData{},
		&LiveReplay{},
		&LivePlay{},
		&LotteryRecord{},
		&LotterySupply{},
		&LotterySupplyReceived{},
		&Message{},
		&Claim{},
		&Power{},
		&PowerRecord{},
		&PowerExtra{},
		&Property{},
		&Session{},
		&Snapshot{},
		&Transfer{},
		&Swap{},
		&User{},
		&Voucher{},
	)
}

func RunInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return session.DB(ctx).Transaction(fn, &sql.TxOptions{Isolation: sql.LevelSerializable})
}

func CreateIgnoreIfExist(ctx context.Context, v interface{}) error {
	return session.DB(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(v).Error
}
