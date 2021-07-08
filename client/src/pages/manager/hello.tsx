import { BackHeader } from "@/components/BackHeader"
import React, { useState } from "react"
import styles from "./hello.less"
import { Button, ToastSuccess, ToastWarning } from "@/components/Sub"
import { $get, $set } from "@/stores/localStorage"
import { ApiPutGroupSetting } from "@/apis/group"
import { getURLParams } from "@/assets/ts/tools"
import { get$t } from "@/locales/tools";
import { useIntl } from "@@/plugin-locale/localeExports";

export default () => {
  const { status } = getURLParams()
  const $t = get$t(useIntl())
  const group = $get('group')
  const [content, setContent] = useState(group[status])
  const [loading, setLoading] = useState(false)

  const valid = () => {
    if (!content) return ToastWarning($t(`manager.${status}`) + $t(`error.empty`))
    return true
  }

  const handleClickSave = async () => {
    if (!valid()) return
    setLoading(true)
    const params = { [status]: content }
    const res = await ApiPutGroupSetting({ [status]: content })
    if (res === 'success') {
      $set('group', { ...group, ...params })
      ToastSuccess("修改成功")
    } else ToastSuccess("修改失败")
    setLoading(false)
  }

  return (
    <>
      <BackHeader name={$t(`manager.${status}`)}/>
      <div className={styles.container}>
        <textarea
          value={content || ""}
          onChange={(e) => {
            console.log(e.target.value)
            setContent(e.target.value)
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
