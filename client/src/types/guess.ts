export interface GuessResponse {
  client_id: string
  guess_id: string
  symbol: string
  price_usd: string
  rules: string // 分号分割
  explain: string // 分号分割
  start_time: string
  end_time: string
  start_at: string
  end_at: string
  created_at: string
}

export interface Guess {
  client_id: string
  guess_id: string
  symbol: string
  price_usd: string
  rules: string[] // 分号分割
  explain: string[] // 分号分割
  start_time: string
  end_time: string
  start_at: string
  end_at: string
  created_at: string
}

export type GuessPageInitData<T = Guess> = {
  today_guess: GuessType
} & T

export enum GuessType {
  Up = 1,
  Down = 2,
  Flat = 3,
}

// 没有result 未参加
export enum GuessResult {
  NotStart = -1,
  Pending = 0,
  Win = 1,
  lose = 2,
}

export type GuessTypeKeys = keyof typeof GuessType

export interface GuessRecord {
  guess_id: string
  user_id: string
  guess_type: GuessType
  result?: GuessResult
  date: string
}
