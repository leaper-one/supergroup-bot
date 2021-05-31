import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { history } from "umi"
import { getCurrentGroup } from "@/pages/home"
import { ApiGetGroupInfo } from "@/apis/group"
import { $get, $set } from "@/stores/localStorage"

async function getManagerList() {
  const { group_number } = $get("group")
  const [group, groupInfo] = await Promise.all([
    getCurrentGroup(),
    ApiGetGroupInfo(group_number),
  ])
  group!.check = groupInfo.checks!
  group!.invite_status = groupInfo.setting!.invite_status
  $set("group", groupInfo.group)
  $set("setting", groupInfo.setting)

  return [
    [
      {
        icon: "iconshequnxinxi",
        type: "社群信息",
        mount: "",
        route: "/setting/group",
      },
      {
        icon: "iconruqunhuanyingyu",
        type: "入群欢迎语",
        mount: "",
        route: "/setting/hello",
      },
    ],
    [
      {
        icon: "iconchicangbizhong",
        type: "持仓币种",
        mount: `${group?.check.length} 种`,
        route: "/create/coin?from=manager",
      },
      {
        icon: "iconjianchajiange",
        type: "检查间隔",
        mount: `${group?.duration} 小时`,
        route: "/create/check?from=manager",
      },
    ],
    [
      {
        icon: "iconguanliyuan",
        type: "管理员",
        route: "/setting/manager",
      },
      {
        icon: "iconxinqunguanliyuan",
        type: "成员管理",
        mount: `${group?.people} 人`,
        route: "/setting/member",
      },
      {
        icon: "iconheimingdan1",
        type: "黑名单",
        route: "/setting/black",
      },
    ],
    [
      // {
      //   icon: "iconheimingdan1",
      //   type: "邀请奖励",
      //   mount: group!.invite_status === "0" ? "已关闭" : "已开启",
      //   route: "/setting/invite",
      // },
    ],
  ]
}

// const settingLists = [
//   [
//     {
//       icon: "member",
//       type: "成员管理",
//       mount: "3 人",
//     },
//     {
//       icon: "blacklist",
//       type: "黑名单",
//       mount: "0 人",
//     },
//   ],
// ]

interface IItem {
  icon: string
  type: string
  mount?: string
  route?: string
}

const isSuperManager = true

const Item = (props: { list: IItem[] }) => (
  <>
    {props.list.map((item, key) => (
      <div
        key={key}
        className={styles.msg}
        onClick={() => history.push(item.route!)}
      >
        <i className={`iconfont ${item.icon} ${styles.iconUrl}`} />
        <span>{item.type}</span>
        <span className={styles.mount}>{item.mount}</span>
        <i className={`iconfont iconic_arrow ${styles.iconArrow}`} />
      </div>
    ))}
  </>
)

const List = (props: { lists: IItem[][] }) => (
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

  useEffect(() => {
    initPage()
  }, [])
  const initPage = async () => {
    setManagerList(await getManagerList())
  }

  return (
    <>
      {/*{isSuperManager ? */}
      <List lists={managerList} />
      {/*/!* : *!/*/}
      {/*{container(settingLists)}*/}
      {/*}*/}
    </>
  )
}
