import { apis } from "./http"
import { IAppointResp } from "@/apis/airdrop"
import { IGroup, IGroupSetting } from "@/apis/group"
import { getConversationId } from "@/assets/ts/tools"

export interface IConversationCheck {
  is_owner: boolean
  is_manager: boolean
  group: IGroup
  setting?: IGroupSetting
}

export const ApiCheckGroup = (id: string): Promise<IConversationCheck> =>
  apis.get(`/checkGroup/${id}`)

export interface ICheckTransfer {
  can_create: boolean
}

export const ApiCheckTransfer = (): Promise<ICheckTransfer> =>
  apis.get(`/checkTransfer`)

export interface IGenerateTransfer {
  trace: string
}

export const ApiGenerateTransfer = (): Promise<IGenerateTransfer> =>
  apis.get(`/generateTransfer`)

export interface ICheck {
  asset_id: string
  name: string
  amount: string
  symbol: string
}

export interface ICanJoin {
  code_id?: string
  can_join: boolean
  reject_reason: string
  check_list?: Array<ICheck>
}

export const ApiGetCheckCanJoin = (group_id: string): Promise<ICanJoin> =>
  apis.get(`/checkCanJoin/${group_id}`, {
    conversation_id: getConversationId(),
  })
