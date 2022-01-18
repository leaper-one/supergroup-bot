import { apis } from "./http"
import { getConversationId } from "@/assets/ts/tools"
import { getGroupID } from "@/apis/group"
import { GlobalData } from '@/stores/store'

export interface IUser {
  access_token?: string
  client_id?: string
  muted_at?: string
  muted_time?: string
  is_notice_join?: boolean
  is_proxy?: boolean
  is_received?: boolean
  priority?: number
  authentication_token?: string
  is_new?: boolean

  user_id?: string
  avatar_url?: string
  full_name?: string
  identity_number?: string
  status?: number
  created_at?: string
  active_at?: string

  pay_status?: number
  pay_expired_at?: string | Date

  is_claim?: boolean
  is_block?: boolean
}

export interface IAdvanceSetting {
  conversation_status: string
  new_member_notice: string
  proxy_status: string
}

export const ApiAuth = (code: string, inviteCode: string): Promise<IUser> =>
  apis.post(`/auth`, { code, c: inviteCode })
export const ApiGetMe = (): Promise<IUser> => apis.get(`/me`)

export const ApiPostChatStatus =
  (is_received: boolean, is_notice_join?: boolean): Promise<IUser> => apis.post(`/user/chatStatus`, {
    is_received,
    is_notice_join
  })

export const ApiGetGroupUsers = (status = "", search = "") =>
  apis.get(`/groupUsers/${getGroupID()}`, { status, search })

export interface IUpdateParticipant {
  user_id: string
  conversation_id: string
  status: string
}

export interface IClientUserStat {
  all: number
  mute: number
  block: number
  guest: number
  admin: number
}

export const ApiPutGroupUsers = (user: IUpdateParticipant) =>
  apis.put(`/groupUsers/${getGroupID()}`, user)


export const ApiGetUserList = (page: number, status = "all"): Promise<IUser[]> =>
  apis.get(`/user/list`, { page, status })

export const ApiGetUserStat = (): Promise<IClientUserStat> =>
  apis.get(`/user/stat`)

export const ApiPostSearchUserList = (key: string): Promise<IUser[]> =>
  apis.get(`/user/search`, { key })

export const ApiPutUserStatus = (user_id: string, status: number, is_cancel: boolean): Promise<string> =>
  apis.put(`/user/status`, { user_id, status, is_cancel })


export const ApiPutUserMute = (user_id: string, mute_time: string): Promise<string> =>
  apis.put(`/user/mute`, { user_id, mute_time })

export const ApiPutUserBlock = (user_id: string, is_cancel: boolean): Promise<string> =>
  apis.put(`/user/block`, { user_id, is_cancel })

export const ApiPutUserProxy = (full_name: string, is_proxy: boolean): Promise<string> =>
  apis.put(`/user/proxy`, { full_name, is_proxy })


export const ApiGetAdminAndGuest = async (): Promise<IUser[]> => {
  if (!GlobalData.adminAndGuests) GlobalData.adminAndGuests = await apis.get(`/user/adminAndGuest`)
  return GlobalData.adminAndGuests
}

// 获取 / 修改 全体禁言 / 入群提醒
export const ApiGetGroupAdvanceSetting = (): Promise<IAdvanceSetting> => apis.get(`/group/advance/setting`)
export const ApiPutGroupAdvanceSetting = (setting: IAdvanceSetting) => apis.put(`/group/advance/setting`, setting)


interface IGroupMemberAuth {
  client_id: string
  limit: number
  lucky_coin: boolean
  plain_contact: boolean
  plain_data: boolean
  plain_image: boolean
  plain_live: boolean
  plain_post: boolean
  plain_sticker: boolean
  plain_text: boolean
  plain_transcript: boolean
  plain_video: boolean
  url: boolean
}

interface IGroupMemberAuthResp {
  1: IGroupMemberAuth,
  2: IGroupMemberAuth,
  5: IGroupMemberAuth
}

// 获取 / 修改 会员权限
export const ApiGetGroupMemberAuth = (): Promise<IGroupMemberAuthResp> =>
  apis.get(`/group/member/auth`)
export const ApiPutGroupMemberAuth = (auth: IGroupMemberAuth) => apis.put(`/group/member/auth`, auth)

