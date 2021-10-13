import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { history } from "umi"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { AppIcons, Icon } from "@/components/Icon"

export interface Manager {
  icon: AppIcons
  type: string
  mount?: string
  route: string
}

async function getManagerList($t: any): Promise<Array<[Manager]>> {
  return [
    [
      {
        icon: "ic_unselected_5",
        type: $t("manager.base"),
        mount: "",
        route: "/manager/setting/base",
      },
    ],
    [
      {
        icon: "ruqunhuanyingyu",
        type: $t("broadcast.title"),
        route: "/broadcast",
      },
    ],
    [
      {
        icon: "shequnxinxi",
        type: $t("stat.title"),
        route: "/manager/stat",
      },
    ],
    [
      {
        icon: "chengyuanguanli1",
        type: $t("member.title"),
        route: "/manager/member",
      },
    ],
    [
      {
        icon: "ruqunhuanyingyu",
        type: $t("advance.title"),
        route: "/manager/advance",
      },
    ],
  ]
}

export interface IItem {
  icon: AppIcons
  type: string
  mount?: string
  route?: string
}

const Item = (props: { list: IItem[] }) => (
  <>
    {props.list.map((item, key) => (
      <div
        key={key}
        className={styles.msg}
        onClick={() => history.push(item.route!)}
      >
        <Icon i={item.icon} className={styles.iconUrl} />
        <span>{item.type}</span>
        <span className={styles.mount}>{item.mount}</span>
        <Icon i="ic_arrow" className={styles.iconArrow} />
      </div>
    ))}
  </>
)

export const List = (props: { lists: IItem[][] }) => (
  <>
    {props.lists.map((list, idx) => (
      <div key={idx} className={styles.content}>
        <Item list={list} />
      </div>
    ))}
  </>
)

export default () => {
  const [managerList, setManagerList] = useState<any[]>([])
  const $t = get$t(useIntl())

  useEffect(() => {
    initPage()
  }, [])
  const initPage = async () => {
    setManagerList(await getManagerList($t))
  }

  return (
    <>
      <BackHeader name={$t("setting.title")} />
      <List lists={managerList} />
    </>
  )
}
