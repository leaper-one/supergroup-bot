import { apis } from '@/apis/http';

export interface IAppointResp {
  status: number;
  symbol?: string;
  amount?: string;
}

export const ApiGetAirdrop = (airdropID: string): Promise<IAppointResp> => apis.get(`/airdrop/${airdropID}`);

export const ApiAirdropReceived = (airdropID: string): Promise<IAppointResp> => apis.post(`/airdrop/${airdropID}`);
