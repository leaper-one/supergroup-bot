import { RecordByDate, LotteryRecord, Record } from "@/types"
import { apis } from "./http"

const transfromRecords = (data: Record[]): RecordByDate[] => {
  return data.reduce((acc, cur) => {
    if (acc.length && acc.some(([time]) => time === cur.date)) {
      return acc.map(([time, records]): RecordByDate => {
        return time === cur.date ? [time, records.concat(cur)] : [time, records]
      })
    }

    return [...acc, [cur.date, [cur]]]
  }, [] as RecordByDate[])
}

// 获取抽奖页面数据
export const ApiGetClaimPageData = () => apis.get("/claim")

// 点击签到
export const ApiPostClain = () => apis.post("/claim")

// 获取能量记录
export const ApiGetClaimRecord = (page = 1): Promise<RecordByDate[]> =>
  apis.get("/power/record", { page }).then(transfromRecords)

// 兑换抽奖
export const ApiPostLotteryExchange = () => apis.post("/lottery/exchange")

// 开始抽奖
export const ApiPostLottery = () => apis.post("/lottery")

// 获取抽奖奖励
export const ApiGetLotteryReward = (trace_id: string) =>
  apis.post("/lottery/reward", { trace_id })

// 获取抽奖列表
export const ApiGetLotteryRecord = (page = 1): Promise<RecordByDate[]> =>
  apis.get("/lottery/record", { page }).then(transfromRecords)
