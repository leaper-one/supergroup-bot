export interface Lottery {
  amount: string
  asset_id: string
  client_id: string
  icon_url: string
  lottery_id: string
}
export interface Prize {
  amount: string
  client_id: string
  asset_id: string
  created_at: string
  icon_url: string
  is_received: boolean
  lottery_id: string
  price_usd: string
  symbol: string
  trace_id: string
  user_id: string
  description?: string
}
