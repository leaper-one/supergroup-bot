import { apis } from "./http"
import { getConversationId } from "@/assets/ts/tools"
import { getGroupID } from "@/apis/group"

export interface IUser {
  type?: string
  user_id: string
  identity_number: string
  full_name: string
  avatar_url: string
  authentication_token?: string
  is_manager?: string
}

export interface IParticipantUser extends IUser {
  conversation_id: string
  index: string
  created_at: string
}

export const ApiAuth = (code: string): Promise<IUser> =>
  apis.post(`/auth`, { code, conversation_id: getConversationId() })
export const ApiGetMe = (): Promise<IUser> => apis.get(`/me`)

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
