import React, { useEffect, useState } from "react"
import styles from "./broadcast.less"
import { SwipeAction } from "antd-mobile"
import { history, useIntl } from "umi"
import { ApiGetBroadcastList, ApiGetBroadcastRecall, IBroadcast, } from "@/apis/broadcast"
import moment from "moment"
import { Confirm, ToastSuccess } from "@/components/Sub"
import { get$t } from "@/locales/tools"
import { BackHeader } from "@/components/BackHeader"

// const statusMap = {
//   "0": "发送中",
//   "1": "已发布",
//   "2": "撤回中",
//   "3": "已撤回",
// }

let tmpData: IBroadcast[] = []
export const Broadcast = () => {
  const [broadcastList, setBroadcastList] = useState<IBroadcast[]>([])
  const [searchKey, setSearchKey] = useState("")
  const $t = get$t(useIntl())
  useEffect(() => {
    initPage()
  }, [])
  useEffect(() => {
    if (!searchKey) return setBroadcastList(tmpData)
    setBroadcastList(
      tmpData.filter((data) =>
        data.data.toLowerCase().includes(searchKey.toLowerCase()),
      ),
    )
  }, [searchKey])
  const initPage = async () => {
    tmpData = await ApiGetBroadcastList()
    setBroadcastList(tmpData)
  }
  return (
    <div className={styles.container}>
      <BackHeader
        name={$t("broadcast.a")}
        action={
          <i
            className={styles.addIcon + " iconfont iconic_add"}
            onClick={() => history.push(`/broadcast/send`)}
          />
        }
      />
      <div className={styles.input}>
        <img src={require("@/assets/img/svg/search.svg")} alt="" />
        <input
          value={searchKey}
          onChange={(e) => setSearchKey(e.target.value)}
          type="text"
          placeholder="Keyword"
        />
      </div>
      <div className={styles.list}>
        {broadcastList.map((item, index) => (
          <SwipeAction
            key={index}
            className={styles.item}
            right={[
              {
                text: $t("broadcast.recall"),
                style: { backgroundColor: "#FA596D", color: "#fff" },
                onPress: async () => {
                  const isConfirm = await Confirm(
                    $t("action.tips"),
                    $t("broadcast.confirmRecall"),
                  )
                  if (isConfirm) {
                    const res = await ApiGetBroadcastRecall(item.message_id)
                    if (res) ToastSuccess($t("broadcast.recallSuccess"))
                    initPage()
                  }
                },
              },
            ]}
          >
            <li className={styles.bItem}>
              <img src={item.avatar_url} alt="" />
              <p>
                {item.data.length > 100
                  ? item.data.slice(0, 100) + "..."
                  : item.data}
              </p>
              <span className={styles[`status` + (item.status || '0')]}>
                {$t('broadcast.status' + (item.status || "0"))}
              </span>
              <span className={styles.date}>
                {moment(item.created_at).format("YYYY/MM/DD HH:mm:ss")}
              </span>
            </li>
          </SwipeAction>
        ))}
      </div>
    </div>
  )
}

export default Broadcast
