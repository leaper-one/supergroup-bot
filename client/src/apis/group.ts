import { apis } from "./http"
import { ApiGetAssetByID, IAsset } from "@/apis/asset"
import { GlobalData } from "@/stores/store"
import { $get } from "@/stores/localStorage"
import { IAppointResp } from "@/apis/airdrop"

export interface IGroup {
  client_id:string
  name: string
  description: string
  icon_url: string
  total_people?: number
  week_people?: string

  button?: string
  asset_id?: string
  created_at?: string
}

export interface IGroupSetting {
  welcome?: string
  description?: string
}

export interface IGroupItem extends IGroup {
  welcome: string
  amount?: string
  symbol?: string
  host?: string
}

export interface IGroupInfo {
  group: IGroup
  airdrop?: IAppointResp
  setting?: IGroupSetting
  checks?: IAsset[]
}

export interface IGroupId {
  group_id: string
}

export interface IGroupStat {
  members: number
  broadcasts: number
  conversations: number
  list: {
    date: string
    count: number
  }[]
}

export interface IGroupInviteSetting {
  group_id: string
  asset_id?: string
  amount?: string
  duration?: number
  times?: number
  send_at?: string
}

export interface IGroupInfo1 {
  asset_id: string
  change_usd: string
  client_id: string
  created_at: string
  description: string
  icon_url: string
  information_url: string
  name: string
  price_usd: string
  symbol: string
  total_people: string
  week_people: string
  speak_status: number
  activity: IActivity[]
}

export interface IActivity {
  activity_index: number
  action: string
  expire_at: string
  expire_img_url: string
  img_url: string
  start_at: string
  isExpire?: boolean
}

export const ApiGetGroup = (): Promise<IGroupInfo1> => apis.get(`/group`)

export const ApiPostGroup = (groupInfo: IGroup): Promise<IGroupId> =>
  apis.post(`/group`, groupInfo)

export const ApiGetGroupInfo = (): Promise<IGroupInfo> =>
  apis.get(`/group`)

export const ApiDeleteGroup = () => apis.delete(`/group`)

export const ApiGetGroupList = async (): Promise<IGroupItem[]> => {
  if (!GlobalData.groupList) GlobalData.groupList = await apis.get(`/groupList`)
  let locale = $get("umi_locale")
  locale = locale === "en-US" ? "en" : "zh"
  return GlobalData.groupList
    .sort((a: IGroupItem, b: IGroupItem) => b.total_people! - a.total_people!)
}

export const ApiGetGroupStat = async (): Promise<IGroupStat> => {
  if (!GlobalData.GroupStat)
    GlobalData.GroupStat = await apis.get(`/group/stat/${getGroupID()}`)
  return GlobalData.GroupStat
}

export const ApiPutGroup = (groupInfo: IGroup) => apis.put(`/group`, groupInfo)

export const ApiPutGroupSetting = (groupInfo: IGroupSetting) =>
  apis.put(`/group/setting`, groupInfo)

export const ApiGetGroupManager = () =>
  apis.get(`/group/manager/${getGroupID()}`)

export const ApiPostGroupManager = (users: string[]) =>
  apis.post(`/group/manager/${getGroupID()}`, { users })

export const ApiDeleteGroupManager = (user_id: string) =>
  apis.delete(`/group/manager/${getGroupID()}/${user_id}`)

export const ApiGetGroupAssets = (): Promise<IAsset[]> =>
  apis.get(`/group/manager/assets/${getGroupID()}`)

type snapshotOrigin = "deposit" | "packet_refund" | "packet_send" | "airdrop"

export interface ISnapshotItem {
  snapshot_id: string
  amount: string
  created_at: string
  memo: string
  origin: snapshotOrigin
}

export const ApiGetGroupSnapshots = (
  asset_id: string,
): Promise<ISnapshotItem[]> =>
  apis.get(`/group/manager/snapshots/${getGroupID()}/${asset_id}`)

export const ApiGetWithdrawalAssets = (
  asset_id: string,
  amount: string,
): Promise<boolean> =>
  apis.post(`/group/manager/assets/withdrawal`, {
    group_id: getGroupID(),
    asset_id,
    amount,
  })

export const ApiGetBtcPrice = async (): Promise<string> => {
  if (!GlobalData.btcPrice)
    GlobalData.btcPrice = (
      await ApiGetAssetByID(`c6d0c728-2624-429b-8e0d-d9d19b6592fa`)
    ).price_usd!
  return GlobalData.btcPrice
}

export const ApiPutGroupStatus = async (
  status_name: string,
  status: string,
): Promise<boolean> =>
  apis.put(`/group/manager/groupStatus`, {
    group_id: getGroupID(),
    status_name,
    status,
  })

export const ApiPutGroupInviteSetting = async (
  data: IGroupInviteSetting,
): Promise<boolean> => apis.put(`/group/manager/invite`, data)

export const getGroupID = () => $get("group").group_id
