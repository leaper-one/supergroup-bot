
export type PowerType = "claim" | "lottery" | "cliam_extra"
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
  power_type?: PowerType // 能量
}

export type RecordByDate = [string, Record[]]