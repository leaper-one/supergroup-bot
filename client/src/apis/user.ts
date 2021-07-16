import { apis } from "./http"
import { getConversationId } from "@/assets/ts/tools"
import { getGroupID } from "@/apis/group"

export interface IUser {
  access_token: string
  client_id: string
  created_at: string
  deliver_at: string
  muted_at: string
  muted_time: string
  user_id: string
  is_notice_join: boolean
  is_received: boolean
  priority: number
  authentication_token: string
  status: number
  is_new: boolean
}

export const ApiAuth = (code: string): Promise<IUser> =>
  apis.post(`/auth`, { code, conversation_id: getConversationId() })
export const ApiGetMe = (): Promise<IUser> => apis.get(`/me`)

export const ApiPostChatStatus =
  (is_received: boolean, is_notice_join: boolean): Promise<IUser> => apis.post(`/user/chatStatus`, {
    is_received,
    is_notice_join
  })

export const ApiGetUserList = (search: string): Promise<IUser[]> =>
  apis.get(`/users/${getGroupID()}`, { search })

export const ApiGetGroupUsers = (status = "", search = "") =>
  apis.get(`/groupUsers/${getGroupID()}`, { status, search })

export interface IUpdateParticipant {
  user_id: string
  conversation_id: string
  status: string
}

export const ApiPutGroupUsers = (user: IUpdateParticipant) =>
  apis.put(`/groupUsers/${getGroupID()}`, user)
