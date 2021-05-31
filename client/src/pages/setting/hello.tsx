import { BackHeader } from "@/components/BackHeader"
import React, { useEffect, useState } from "react"
import styles from "./hello.less"
import { Button, ToastSuccess, ToastWarning } from "@/components/Sub"
import { $get, $set } from "@/stores/localStorage"
import { ApiPutGroupSetting } from "@/apis/group"

let group_id: string = "",
  setting: any
export default () => {
  const [welcome, setWelcome] = useState("")
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const group = $get("group") || {}
    setting = $get("setting")
    group_id = group.group_id
    setWelcome(setting?.welcome || "")
    return () => {
      group_id = ""
      setting = undefined
    }
  }, [])

  const valid = () => {
    if (!group_id) return ToastWarning("没找到社群...")
    if (!welcome) return ToastWarning("欢迎语不能为空")
    return true
  }

  const handleClickSave = async () => {
    setLoading(true)
    if (!valid()) return setLoading(false)
    const res = await ApiPutGroupSetting({ welcome, group_id })
    if (res) {
      $set("setting", { ...setting, welcome })
      ToastSuccess("修改成功")
    } else ToastSuccess("修改失败")
    setLoading(false)
  }

  return (
    <>
      <BackHeader name="入群欢迎语" />
      <div className={styles.container}>
        <textarea
          value={welcome || ""}
          onChange={(e) => {
            console.log(e.target.value)
            setWelcome(e.target.value)
          }}
        />
        <p>欢迎语只有新人入群的成员可以看到，其他群组成员看不到.</p>
      </div>
      <Button
        className={styles.button}
        loading={loading}
        onClick={handleClickSave}
      >
        保存
      </Button>
    </>
  )
}
