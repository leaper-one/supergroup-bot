import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { history } from "umi"
import { BackHeader } from "@/components/BackHeader";
import { get$t } from "@/locales/tools";
import { useIntl } from "@@/plugin-locale/localeExports";


async function getManagerList($t: any) {
  return [
    [
      {
        icon: "iconic_unselected_5",
        type: $t('manager.base'),
        mount: "",
        route: "/manager/setting/base",
      }
    ],
    [
      {
        icon: "iconruqunhuanyingyu",
        type: $t('broadcast.title'),
        route: "/broadcast"
      }
    ],
    [
      {
        icon: "iconshequnxinxi",
        type: $t('stat.title'),
        route: "/manager/stat"
      }
    ],
  ]
}

interface IItem {
  icon: string
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
        <i className={`iconfont ${item.icon} ${styles.iconUrl}`}/>
        <span>{item.type}</span>
        <span className={styles.mount}>{item.mount}</span>
        <i className={`iconfont iconic_arrow ${styles.iconArrow}`}/>
      </div>
    ))}
  </>
)

export const List = (props: { lists: IItem[][] }) => (
  <>
    {props.lists.map((list, idx) => (
      <div key={idx} className={styles.content}>
        <Item list={list}/>
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
      <BackHeader name="设置"/>
      <List lists={managerList}/>
    </>
  )
}
