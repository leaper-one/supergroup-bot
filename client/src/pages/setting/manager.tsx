import React, { useEffect, useState } from "react"
import { BackHeader } from "@/components/BackHeader"
import styles from "./manager.less"
import { Modal, SwipeAction } from "antd-mobile"
import { Button, Confirm, ToastSuccess, ToastWarning } from "@/components/Sub"
import timeStyle from "@/pages/red/timingList.less"
import coinStyle from "@/components/CoinSelect/index.less"
import { ApiGetUserList, IUser } from "@/apis/user"
import {
  ApiDeleteGroupManager,
  ApiGetGroupManager,
  ApiPostGroupManager,
} from "@/apis/group"

export default () => {
  const [selectManager, setSelectManager] = useState(false)

  const [managerList, setManagerList] = useState<IUser[]>([])

  useEffect(() => {
    initPage()
  }, [])

  const initPage = async () => {
    const managerList = await ApiGetGroupManager()
    console.log(managerList)
    setManagerList(managerList)
  }

  return (
    <div className={styles.container}>
      <BackHeader
        name="管理员设置"
        action={
          <i
            className={styles.addIcon + " iconfont iconic_add"}
            onClick={() => setSelectManager(true)}
          />
        }
      />
      <ul className={`${timeStyle.list} ${styles.list}`}>
        {managerList.map((user) => (
          <SwipeAction
            key={user.user_id}
            autoClose
            right={[
              {
                text: "取消",
                style: {
                  backgroundColor: "#FA596D",
                  color: "white",
                  width: "80px",
                  height: "50px",
                },
                onPress: async () => {
                  const isConfirm = await Confirm(
                    "提示",
                    `是否删除管理员 ${user.full_name} ？`,
                  )
                  if (isConfirm) {
                    const res = await ApiDeleteGroupManager(user.user_id)
                    if (res) {
                      ToastSuccess("删除成功")
                      initPage()
                    }
                  }
                },
              },
            ]}
          >
            <li className={`${timeStyle.item} ${styles.item}`}>
              <img src={user.avatar_url} alt="" />
              <p className={styles.name}>{user.full_name}</p>
              <span className={styles.status}>{user.identity_number}</span>
            </li>
          </SwipeAction>
        ))}
      </ul>

      <Modal
        popup
        visible={selectManager}
        onClose={() => setSelectManager(false)}
        animationType="slide-up"
      >
        <ManagerModal
          reload={() => initPage()}
          selected={managerList}
          close={() => setSelectManager(false)}
        />
      </Modal>
    </div>
  )
}

interface Props {
  reload: () => void
  selected: IUser[]
  close: () => void
}

const ManagerModal = (props: Props) => {
  const { reload, selected, close } = props
  const [userList, setUserList] = useState<IUser[]>([])
  const [search, setSearch] = useState("")
  const [selectList, setSelectList] = useState<IUser[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    initPage(search)
  }, [search])
  const initPage = async (search = "") => {
    const t = await ApiGetUserList(search)
    setUserList(t)
  }

  const selectUser = (user: IUser) => {
    if (selected.some((u) => u.user_id === user.user_id)) return
    const idx = selectList.findIndex((u) => u.user_id === user.user_id)
    if (idx === -1) {
      setSelectList([...selectList, user])
    } else {
      selectList.splice(idx, 1)
      setSelectList([...selectList])
    }
  }

  const handleClickSave = async () => {
    if (selectList.length === 0) return ToastWarning("请先选择用户")
    setLoading(true)
    const res = await ApiPostGroupManager(selectList.map((u) => u.user_id))
    if (res) {
      ToastSuccess("新增成功")
      close()
      reload()
    } else {
      ToastSuccess("新增失败")
    }
    setLoading(false)
  }

  return (
    <div className={`${coinStyle.container} ${styles.modalContainer}`}>
      <div className={coinStyle.search + " " + "flex"}>
        <img src={require("@/assets/img/svg/search.svg")} alt="" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Name, Symbol"
          type="text"
        />
        <span onClick={() => close()}>取消</span>
      </div>
      <ul className={coinStyle.list}>
        {userList.map((user, idx) => (
          <li
            className={styles.modalItem}
            key={user.user_id}
            onClick={() => selectUser(user)}
          >
            <i
              className={`iconfont iconxuanzhong1
${styles.select}
${selected.some((u) => u.user_id === user.user_id) ? styles.disable : ""}
${selectList.some((u) => u.user_id === user.user_id) ? styles.active : ""}
`}
            />
            <img src={user.avatar_url} alt="" />
            <p>{user.full_name}</p>
            <i>{user.identity_number}</i>
          </li>
        ))}
      </ul>
      <Button
        loading={loading}
        className={styles.saveBtn}
        onClick={handleClickSave}
      >
        保存
      </Button>
    </div>
  )
}
