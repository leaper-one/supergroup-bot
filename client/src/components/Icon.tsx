import React, { FC } from "react"

export type AppIcons =
  | "ic_file_text"
  | "ic_return"
  | "check"
  | "guanbi"
  | "guanbi2"
  | "a-huiyuankaitongshibai1"
  | "shenqingxuzhi"
  | "ic_down"
  | "landun"
  | "yishiming"
  | "inhangka"
  | "zhifubao"
  | "weixinzhifu"
  | "ic_share"
  | "fuzhiruqunlianjie"
  | "ic_load"
  | "loding"
  | "tianjia"
  | "ic_notice"
  | "google"
  | "baidu1"
  | "search"
  | "ic_arrow"
  | "ic_unselected_5"
  | "ruqunhuanyingyu"
  | "shequnxinxi"
  | "chengyuanguanli1"
  | "a-chicangbizhong2"
  | "a-ruqunhuanyingyu1"
  | "jibenshezhi1"
  | "gaojihuiyuan"
  | "chujihuiyuan"
  | "feihuiyuan"
  | "ruquntixing"
  | "quantijinyan"
  | "xiaoxiquanxian"
  | "gaojishezhi"
  | "shujutongji"
  | "gonggaoguanli"
  | "ic_music_open"
  | "ic_music_close"

interface IconProps extends React.HTMLProps<HTMLElement> {
  i: AppIcons
}

export const Icon: FC<IconProps> = ({ className, i, ...rest }) => (
  <i className={`iconfont icon${i} ${className}`} {...rest} />
)
