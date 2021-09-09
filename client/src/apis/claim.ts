import { apis } from './http'


// 获取抽奖页面数据
export const ApiGetClaimPageData = () => apis.get('/claim')

// 点击签到
export const ApiPostClain = () => apis.post('/claim')

// 获取能量记录
export const ApiGetClaimRecord = (page = 1) => apis.get('/power/record', { page })

// 兑换抽奖
export const ApiPostLotteryExchange = () => apis.post('/lottery/exchange')

// 开始抽奖
export const ApiPostLottery = () => apis.post('/lottery')

// 获取抽奖奖励
export const ApiGetLotteryReward = (trace_id: string) => apis.get('/lottery/reward', { trace_id })

// 获取抽奖列表
export const ApiGetLotteryRecord = (page = 1) => apis.get('/lottery/record', { page })