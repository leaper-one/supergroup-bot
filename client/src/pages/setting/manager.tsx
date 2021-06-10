import React, { useEffect, useState } from "react"
import styles from "./manager.less"
import { history } from "umi"
import { BackHeader } from "@/components/BackHeader";

// import { getCurrentGroup } from "@/pages/home"

async function getManagerList() {
  return [
    [
      {
        icon: "iconic_unselected_5",
        type: "基本设置",
        mount: "",
        route: "/manager/setting/base",
      },
      // {
      //   icon: "iconruqunhuanyingyu",
      //   type: "入群欢迎语",
      //   mount: "",
      //   route: "/setting/hello",
      // },
    ],
    [
      // {
      //   icon: "iconguanliyuan",
      //   type: "管理员",
      //   route: "/setting/manager",
      // },
      // {
      //   icon: "iconxinqunguanliyuan",
      //   type: "成员管理",
      //   route: "/setting/member",
      // },
      // {
      //   icon: "iconheimingdan1",
      //   type: "黑名单",
      //   route: "/setting/black",
      // },
    ],
    [
      {
        icon: "iconruqunhuanyingyu",
        type: "公告管理",
        route: "/broadcast"
      }
    ],
    [
      // {
      //   icon: "iconshequnxinxi",
      //   type: "高级管理",
      //   route: "/broadcast"
      // }
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

  useEffect(() => {
    initPage()
  }, [])
  const initPage = async () => {
    setManagerList(await getManagerList())
  }

  return (
    <>
      <BackHeader name="设置"/>
      <List lists={managerList}/>
    </>
  )
}
