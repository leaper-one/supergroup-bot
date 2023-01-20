package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/MixinNetwork/supergroup/durable"
	"github.com/MixinNetwork/supergroup/session"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var Ctx context.Context

type Report struct {
	ReporterID string    `json:"reporter_id" gorm:"type:varchar(36);not null;"`
	ReportedID string    `json:"reported_id" gorm:"type:varchar(36);not null"`
	Category   string    `json:"category" gorm:"type:varchar(4);not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"type:timestamp with time zone;default: now()"`
}

func (Report) Table() string {
	return "reports"
}
func init() {
	db := durable.NewDB()

	Ctx = session.WithDatabase(context.Background(), db)
	Ctx = session.WithRedis(Ctx, durable.NewRedis(context.Background()))

}
func AutoMigrate() {
	durable.NewDB().AutoMigrate(
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
		&Report{},
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
