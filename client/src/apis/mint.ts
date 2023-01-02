import { apis } from './http';

export interface IMint {
  asset_id: string;
  created_at: string;
  daily_amount: string;
  daily_end: string;
  daily_time: string;
  description: string;
  extra_asset_id: string;
  extra_daily_amount: string;
  extra_first_amount: string;
  first_amount: string;
  first_end: string;
  first_time: string;
  mining_id: string;
  reward_asset_id: string;
  status: string;
  symbol: string;
  bg: string;
  title: string;
  faq: string;
  join_tips: string;
  join_url: string;

  reward_symbol: string;
  extra_symbol: string;
  first_desc: string;
  daily_desc: string;
}

export interface IMintRecord {
  record_id?: string;
  status?: number;
  date?: string;
  profit?: string;
  amount?: string;
  symbol?: string;
  items?: IMintRecord[];
}

export const ApiGetMintByID = (id: string): Promise<IMint> => apis.get(`/mint/${id}`);

export const ApiPostMintByID = (record_id: string): Promise<string> => apis.post(`/mint`, { record_id });

export const ApiGetMintRecord = (mint_id: string): Promise<IMintRecord[]> => apis.get(`/mint/record`, { mint_id });

export interface LiquidityResp {
  info: LiquidityInfo;
  list: LiquidityList[];
  yesterday_amount: string;
  is_join: boolean;
  scope: string;
}

interface LiquidityList {
  liquidity_id: string;
  idx: number;
  start_at: string;
  end_at: string;
  asset_id: string;
  amount: string;
  symbol: string;
}

interface LiquidityInfo {
  liquidity_id: string;
  client_id: string;
  title: string;
  description: string;
  lp_desc: string;
  lp_url: string;
  start_at: string;
  end_at: string;
  asset_ids: string;
  min_amount: string;
  created_at: string;
}

export const ApiGetLiquidityByID = (id: string): Promise<LiquidityResp> => apis.get(`/liquidity/${id}`);

export const ApiPostLiquidityJoin = (id: string) => apis.post(`/liquidity/join`, { id });

export interface LiquidityRecordResp {
  duration: string;
  status: string;
  list: LiquidityRecord[];

  is_open: boolean;
}

interface LiquidityRecord {
  date: string;
  lp_symbol: string;
  lp_amount: string;
  usd_value: string;
}

export const ApiGetLiquidityRecordByID = (id: string): Promise<LiquidityRecordResp[]> => apis.get(`/liquidity/record`, { id });
