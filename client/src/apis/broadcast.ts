import { apis } from '@/apis/http';

type TypeStatus = '0' | '1' | '2';

export interface IBroadcast {
  avatar_url: string;
  message_id: string;
  category: string;
  created_at: string;
  data: string;
  full_name: string;
  status: TypeStatus;
  user_id: string;
  top_at: string;
  is_top: boolean;
}

export const ApiGetBroadcastList = (): Promise<IBroadcast[]> => apis.get(`/broadcast`);

export const ApiPostBroadcast = (data: string): Promise<boolean> => apis.post(`/broadcast`, { data });

export const ApiGetBroadcastRecall = (broadcast_id: string): Promise<boolean> => apis.delete(`/broadcast/${broadcast_id}`);
