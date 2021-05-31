import { apis } from "@/apis/http"

export interface IInviteItem {
  avatar_url: string
  created_at: string
  full_name: string
  identity_number: string
  status: string
  user_id: string
}

interface IStat {
  people: number
  amount?: string
  symbol?: string
  price_usd?: string
}

export const ApiGetInviteList = (
  id: string,
  page: number,
): Promise<IInviteItem[]> => apis.get(`/invite/${id}/${page}`)

export const ApiGetInviteCount = (id: string): Promise<IStat> =>
  apis.get(`/invite/count/${id}`)
