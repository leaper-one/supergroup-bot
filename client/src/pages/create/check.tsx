import React, { useEffect, useState } from "react"
import styles from "./check.less"
import { BackHeader } from "@/components/BackHeader"
import { $get, $remove } from "@/stores/localStorage"
import { history } from "umi"
import { Loading } from "@/components/Loading"
import { Button, Confirm, ToastFailed, ToastSuccess } from "@/components/Sub"
import { ApiPostGroup, ApiPutGroupSetting } from "@/apis/group"
import { isFromManager } from "@/pages/create/coin"
import { getCurrentGroup } from "@/pages/home"

export default () => {
  const [duration, setDuration] = useState("48")
  const [showLoading, setShowLoading] = useState(false)
  const [isManager, setIsManager] = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (isFromManager()) {
      setIsManager(true)
      initPage()
    }
  }, [])
  const initPage = async () => {
    const data = await getCurrentGroup()
    setDuration(data?.duration ? String(data.duration) : "48")
  }

  const handleClickBtn = async () => {
    setLoading(true)
    if (!isManager) await handleClickCreate()
    else await handleClickSave()
    setLoading(false)
  }

  const handleClickSave = async () => {
    const { group_id } = $get("group")
    const res = await ApiPutGroupSetting({
      group_id,
      duration: Number(duration),
    })
    if (res) {
      const data = await getCurrentGroup()
      data!.duration = Number(duration)
      ToastSuccess("修改成功")
    } else {
      ToastFailed("修改失败")
    }
  }

  const handleClickCreate = async () => {
    const create = $get("create")
    create.duration = Number(duration)
    setShowLoading(true)
    const data = await ApiPostGroup(create)
    setShowLoading(false)
    let msg = ""
    if (data?.group_id) {
      $remove("create")
      msg =
        "恭喜您，社群已经成功创建。现在可以退出机器人，由分群直接进去机器人就可以管理您创建的社群啦..."
    } else msg = "请关闭机器人重新进入..."
    await Confirm("提示", msg)
    history.push("/")
  }

  return (
    <div className={styles.container}>
      <BackHeader name="设置检查间隔" />
      <div className={styles.inputBox}>
        <input
          type="number"
          value={duration}
          onChange={(e) => setDuration(e.target.value)}
        />
        <i>小时</i>
        <p>每 {duration} 小时进行一次持仓检查</p>
      </div>
      <footer>
        <Button
          disabled={duration === ""}
          className="btn"
          onClick={handleClickBtn}
          loading={loading}
        >
          {isManager ? "保存" : "创建"}
        </Button>
        <p>
          每间隔指定时间就进行一次持仓检查，满足任一币种持仓要求即可，不符合持仓要求自动移出群。
        </p>
      </footer>
      {showLoading && <Loading noCancel />}
    </div>
  )
}
