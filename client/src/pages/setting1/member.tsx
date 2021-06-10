import React, { useEffect, useState } from "react"
import { BackHeader } from "@/components/BackHeader"
import {
  ApiGetGroupUsers,
  ApiPutGroupUsers,
  IParticipantUser,
} from "@/apis/user"
import timeStyle from "@/pages/red/timingList.less"
import styles from "@/pages/setting/manager.less"
import memberStyle from "./member.less"
import { SwipeAction } from "antd-mobile"
import { Confirm, ToastFailed, ToastSuccess } from "@/components/Sub"
import moment from "moment"
import { history } from "umi"

export default () => {
  const [userList, setUserList] = useState<IParticipantUser[]>([])
  const blackPage = history.location.pathname.includes("black")
  useEffect(() => {
    initPage()
  }, [])
  const initPage = async () => {
    const users = await ApiGetGroupUsers(blackPage ? "9" : "0")
    setUserList(users)
  }

  return (
    <div>
      <BackHeader name={blackPage ? "黑名单" : "成员管理"} />
      <ul className={`${timeStyle.list} ${styles.list}`}>
        {userList.map((user) => (
          <SwipeAction
            key={user.user_id}
            autoClose
            right={
              blackPage
                ? [
                    getRightAction(
                      user,
                      "取消拉黑",
                      "#FA596D",
                      `是否将用户 ${user.full_name}(${user.identity_number}) 取消黑名单？取消黑名单后可以重新加入社群。`,
                      "7",
                      initPage,
                    ),
                  ]
                : [
                    getRightAction(
                      user,
                      "踢人",
                      "#FFA41A",
                      `是否将用户 ${user.full_name}(${user.identity_number}) 移除群聊？移除后该用户 24 小时内无法加入本群。`,
                      "8",
                      initPage,
                    ),
                    getRightAction(
                      user,
                      "拉黑",
                      "#FA596D",
                      `是否将用户 ${user.full_name}(${user.identity_number}) 拉入黑名单？拉入黑名单后用户无法再加入本群。`,
                      "9",
                      initPage,
                    ),
                  ]
            }
          >
            <li
              className={`${timeStyle.item} ${styles.item} ${memberStyle.item}`}
            >
              <img src={user.avatar_url} alt="" />
              <p className={styles.name}>{user.full_name}</p>
              <p className={`${styles.name} ${memberStyle.name}`}>
                {user.index} 群
              </p>
              <span className={styles.status}>{user.identity_number}</span>
              <span className={`${styles.status} ${memberStyle.status}`}>
                {moment(user.created_at).format("MM/DD")}
              </span>
            </li>
          </SwipeAction>
        ))}
      </ul>
    </div>
  )
}

const getRightAction = (
  user: IParticipantUser,
  title: string,
  color: string,
  msg: string,
  status: string,
  initPage: () => void,
) => ({
  text: title,
  style: {
    backgroundColor: color,
    color: "white",
    padding: "0 12px",
    height: "50px",
  },
  onPress: async () => {
    const isConfirm = await Confirm("提示", msg)
    if (!isConfirm) return
    const { user_id, conversation_id } = user
    const t = await ApiPutGroupUsers({ user_id, conversation_id, status })
    if (t) {
      ToastSuccess("操作成功")
      initPage()
    } else {
      ToastFailed("操作失败")
    }
  },
})
