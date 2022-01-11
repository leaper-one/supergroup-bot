import { IGroup } from './group'
import { apis } from "./http"


export interface Record {
  amount: string
  asset_id: string
  created_at: string
  date: string
  icon_url: string
  is_received: boolean
  lottery_id: string
  price_usd: string
  symbol: string
  trace_id: string
  power_type?: "claim" | "lottery" | "cliam_extra" | "invitation" // 能量
}

export type RecordByDate = [string, Record[]]
interface RecordByDateResponse {
  hasMore: boolean
  list: RecordByDate[]
}

const transfromRecords = (defaultData: RecordByDate[] = []) => (data: Record[] | null): RecordByDateResponse => {
  if (!data || !data.length) return { hasMore: false, list: [] }

  const list = data.reduce((acc, cur, idx) => {
    if (acc.length && acc.some(([time]) => time === cur.date)) {
      return acc.map(([time, records]): RecordByDate => {
        return time === cur.date ? [time, records.concat(cur)] : [time, records]
      })
    }

    return [...acc, [cur.date, [cur]]] as RecordByDate[]
  }, defaultData)

  return {
    hasMore: data.length >= 20,
    list
  }
}


export interface LotteryRecord {
  lottery_id: string
  user_id: string
  asset_id: string
  trace_id: string
  is_received: boolean
  amount: string
  created_at: string
  icon_url?: string
  symbol?: string
  full_name?: string
  price_usd?: string
  client_id?: string
  date?: string
  description?: string
  power_type?: "claim" | "lottery" | "claim_extra"
}

export interface ClaimData {
  count: number
  invite_count: number
  is_claim: boolean
  last_lottery: LotteryRecord[]
  lottery_list: LotteryRecord[]
  power: {
    balance: string
    lottery_times: number
  }
  receiving?: LotteryRecord
  double_claim_list: IGroup[]
}

// 获取抽奖页面数据
export const ApiGetClaimPageData = (): Promise<ClaimData> => apis.get("/claim")

// 点击签到
export const ApiPostClaim = () => apis.post("/claim")

// 获取能量记录
export const ApiGetClaimRecord = (page = 1, defaultData: RecordByDate[]): Promise<RecordByDateResponse> =>
  apis.get("/power/record", { page }).then(transfromRecords(defaultData))

// 兑换抽奖
export const ApiPostLotteryExchange = () => apis.post("/lottery/exchange")

// 开始抽奖
export const ApiPostLottery = () => apis.post("/lottery")

// 获取抽奖奖励
export const ApiGetLotteryReward = (trace_id: string) =>
  apis.post("/lottery/reward", { trace_id })

// 获取抽奖列表
export const ApiGetLotteryRecord = (page = 1, defaultData: RecordByDate[]): Promise<RecordByDateResponse> =>
  apis.get("/lottery/record", { page }).then(transfromRecords(defaultData))
