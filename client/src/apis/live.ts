import { apis } from "@/apis/http";

export interface ILive {
  live_id?: string
  client_id?: string
  img_url: string
  category: number
  title: string
  description: string
  status?: number
  created_at?: string
  top_at: string
  is_top?: boolean
}

export const ApiGetLiveList = (): Promise<ILive[]> => apis.get(`/live`)

export const ApiPostLive = (data: ILive) => apis.post(`/live`, data)

export const ApiGetLiveInfo = (id: string) => apis.get(`/live/${id}`)

export const ApiGetStartLive = (liveID: string, url = "") => apis.get(`/live/${liveID}/start${url ? `?url=${url}` : ''}`)

export const ApiGetStopLive = (liveID: string) => apis.get(`/live/${liveID}/stop`)

export const ApiGetTopNews = (id: string) => apis.get(`/news/${id}/top`)

export const ApiGetCancelTopNews = (id: string) => apis.get(`/news/${id}/cancelTop`)


export interface IReplay {
  category: string
  created_at: string
  data: string
}

export const ApiGetLiveReplayList = (id: string): Promise<IReplay[]> => apis.get(`/live/${id}/replay`)

export interface ILiveData {
  live_id?: string
  end_at: string
  start_at: any
  deliver_count: number
  msg_count: number
  read_count: number
  user_count: number

  duration?: number
}

export const ApiGetLiveStat = (liveID: string): Promise<ILiveData> => apis.get(`/live/${liveID}/stat`)