import { apis } from './http'

// 获取邀请数据
export interface IInvitationResp {
  code: string
  count: number
}
export const ApiGetInvitation = (): Promise<IInvitationResp> => apis.get("/invitation")


export const ApiGetInviteList = () => apis.get('')
export interface IInviteItem {
  full_name: string
  avatar_url: string
  identity_number: string
  amount: string
  updated_at: string
}