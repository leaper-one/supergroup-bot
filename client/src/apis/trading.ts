import { IAsset } from './asset';
import { apis } from './http';

export interface ITradingCompetition {
  amount: string;
  asset_id: string;
  client_id: string;
  competition_id: string;
  created_at: string;
  start_at: string;
  end_at: string;
  rules: string;
  tips: string;
  title: string;
  reward: string;
}
export interface ITradingCompetitionResp {
  asset: IAsset;
  status: string;
  trading_competition: ITradingCompetition;
}

export const ApiGetTradingByID = (id: string): Promise<ITradingCompetitionResp> => apis.get(`/trading_competetion/${id}`);

export interface ITradingRank {
  full_name: string;
  identity_number: string;
  user_id: string;
  avatar: string;
  amount: string;
}

export interface ITradingRankResp {
  list: ITradingRank[];
  symbol: string;
  amount: string;
}
export const ApiGetTradingRankByID = (id: string): Promise<ITradingRankResp> => apis.get(`/trading_competetion/${id}/rank`);
