import React, { useEffect, useState } from "react"
import styles from "../index.less"
import { history } from "umi"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { Icon } from "@/components/Icon"
import { ApiGetGroupMemberAuth } from '@/apis/user'
import { $set } from '@/stores/localStorage'
import { IItem, Manager } from '..'


function getManagerList($t: any): Array<[Manager]> {
  return [
    [
      {
        icon: "ic_unselected_5",
        type: $t("advance.member.1"),
        route: "/manager/advance/authDesc?s=1",
      },
    ],
    [
      {
        icon: "chengyuanguanli1",
        type: $t("advance.member.2"),
        route: "/manager/advance/authDesc?s=2",
      },
    ],
    [
      {
        icon: "ruqunhuanyingyu",
        type: $t("advance.member.5"),
        route: "/manager/advance/authDesc?s=5",
      },
    ],
  ]
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
    setManagerList(getManagerList($t))
    const auth = await ApiGetGroupMemberAuth()
    $set("auth", auth)
  }

  return (
    <>
      <BackHeader name={$t("advance.msgAuth")} />
      <List lists={managerList} />
    </>
  )
}
