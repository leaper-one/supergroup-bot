package jobs

func StartWithHttpServiceJob() {
	// 每天更新社群信息
	go StartDailyDataJob()
	// 挖矿任务
	go StartMintJob()
	// 更新 exinlocal 的广告
	go UpdateExinLocalAD()
	// 处理转账
	go HandleTransfer()
	// 每日更新白名单
	go DailyUpdateClientWhiteURL()
	// 每日更新猜奖
	// go common.timeToUpdateGuessResult()
	// 每日统计抽奖消息
	go DailyStatisticMsg()
	// cache + 定期更新用户
	go CacheAllClientUser()
	// cache + 定期更新黑名单用户
	go CacheAllBlockUser()
	// 更新用户的活跃时间
	go TaskUpdateActiveUserToPsql()
	// 每日更新代理用户信息
	go DailyUpdateProxyUserProfile()
	// 处理 liquidity 任务
	go StartLiquidityDailyJob()
}
