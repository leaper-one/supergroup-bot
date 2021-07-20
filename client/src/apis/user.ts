import { apis } from "./http"
import { getConversationId } from "@/assets/ts/tools"
import { getGroupID } from "@/apis/group"

export interface IUser {
  access_token?: string
  client_id?: string
  muted_at?: string
  muted_time?: string
  is_notice_join?: boolean
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
}

export const ApiAuth = (code: string): Promise<IUser> =>
  apis.post(`/auth`, { code, conversation_id: getConversationId() })
export const ApiGetMe = (): Promise<IUser> => apis.get(`/me`)

export const ApiPostChatStatus =
  (is_received: boolean, is_notice_join: boolean): Promise<IUser> => apis.post(`/user/chatStatus`, {
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

