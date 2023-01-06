package models

import (
	"context"

	"github.com/MixinNetwork/supergroup/durable"
)

var Ctx context.Context

func init() {
	db := durable.NewDB()
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
		&Guess{},
		&GuessRecord{},
		&GuessResult{},
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
		&TradingCompetition{},
		&UserSnapshot{},
		&TradingRank{},
		&User{},
		&Voucher{},
	)
}
