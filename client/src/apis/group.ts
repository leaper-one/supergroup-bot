import { apis } from './http';
import { GlobalData } from '@/stores/store';
import { $get } from '@/stores/localStorage';

export interface IGroup {
  client_id: string;
  name: string;
  description: string;
  icon_url: string;
  total_people?: string;
  week_people?: string;

  button?: string;
  asset_id?: string;
  created_at?: string;
  welcome?: string;
  identity_number?: string;
}

export interface IGroupSetting {
  welcome?: string;
  description?: string;
}

export interface IGroupItem extends IGroup {
  welcome: string;
  amount?: string;
  symbol?: string;
  host?: string;
}

export interface IGroupStat {
  today: {
    messages: number;
    users: number;
  };
  list: {
    date: string;
    messages: number;
    users: number;
  }[];
}

export interface IGroupMenu {
  name_zh: string;
  name_en: string;
  name_ja: string;
  icon: string;
  url: string;
}

export interface IGroupInfo {
  asset_id: string;
  change_usd: string;
  client_id: string;
  created_at: string;
  description: string;
  icon_url: string;
  information_url: string;
  name: string;
  price_usd: string;
  symbol: string;
  total_people: string;
  week_people: string;
  speak_status: number;
  activity: IActivity[];
  has_reward: boolean;
  amount?: string;
  large_amount?: string;
  menus?: IGroupMenu[];
}

export interface IActivity {
  activity_index: number;
  action: string;
  expire_at: string;
  expire_img_url: string;
  img_url: string;
  start_at: string;
  isExpire?: boolean;
}

export interface IAuth {
  app_card: boolean;
  limit: number;
  lucky_coin: boolean;
  plain_contact: boolean;
  plain_data: boolean;
  plain_image: boolean;
  plain_live: boolean;
  plain_post: boolean;
  plain_sticker: boolean;
  plain_text: boolean;
  plain_transcript: boolean;
  plain_video: boolean;
  url: boolean;
}

export interface IVipAmount {
  level: {
    client_id: string;
    fresh_amount: string;
    large_amount: string;
  };
  auth: {
    1: IAuth;
    2: IAuth;
    5: IAuth;
  };
}

export const ApiGetGroup = (): Promise<IGroupInfo> => apis.get(`/group`);

export const ApiGetGroupVipAmount = (): Promise<IVipAmount> => apis.get(`/group/vip`);

export const ApiDeleteGroup = () => apis.delete(`/group`);

export const ApiGetGroupList = async (): Promise<IGroupItem[]> => {
  if (!GlobalData.groupList) {
    GlobalData.groupList = await apis.get(`/groupList`);
    let locale = $get('umi_locale');
    locale = locale === 'en' ? 'en' : 'zh';
    GlobalData.groupList = GlobalData.groupList.sort((a: IGroupItem, b: IGroupItem) => Number(b.total_people)! - Number(a.total_people)!);
  }
  return GlobalData.groupList;
};

export const ApiGetGroupStat = async (): Promise<IGroupStat> => {
  if (!GlobalData.GroupStat) GlobalData.GroupStat = await apis.get(`/group/stat`);
  return GlobalData.GroupStat;
};

export const ApiPutGroupSetting = (groupInfo: IGroupSetting) => apis.put(`/group/setting`, groupInfo);

// 0 普通模式 1 禁言模式 2 图文直播模式
export const ApiGetGroupStatus = (): Promise<string> => apis.get(`/group/status`);

export const getGroupID = () => $get('group').group_id;
