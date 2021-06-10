import { BackHeader } from "@/components/BackHeader"
import React, { useEffect, useState } from "react"
import styles from "./group.less"
import { Modal } from "antd-mobile"
import { CoinModal } from "@/components/PopupModal/coinSelect"
import { IAsset } from "@/apis/asset"
import { $get, $set } from "@/stores/localStorage"
import { ApiPutGroup, IGroup } from "@/apis/group"
import { Button, ToastSuccess, ToastWarning } from "@/components/Sub"

export default () => {
  const [showSelectAsset, setShowSelectAsset] = useState(false)
  const [activeAsset, setActiveAsset] = useState<IAsset>()
  const [groupInfo, setGroupInfo] = useState<IGroup>({} as IGroup)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    const group = $get("group")
    if (group) {
      setGroupInfo(group)
      setActiveAsset({ icon_url: group.icon_url, asset_id: group.asset_id })
    }
  }, [])

  const validForm = (): boolean => {
    if (!groupInfo || !activeAsset) return ToastWarning("信息有误...")
    if (!groupInfo.name) return ToastWarning("群名称不能为空")
    if (!groupInfo.description) return ToastWarning("群描述不能为空")
    if (!activeAsset.asset_id || !activeAsset.icon_url)
      return ToastWarning("选择的资产有问题")
    return true
  }

  const handleClickSave = async () => {
    setLoading(true)
    if (!validForm()) return setLoading(false)
    const res = await ApiPutGroup({
      group_id: groupInfo.group_id,
      asset_id: activeAsset!.asset_id,
      icon_url: activeAsset!.icon_url!,
      name: groupInfo.name,
      description: groupInfo.description,
    })
    if (res) {
      $set("group", groupInfo)
      ToastSuccess("修改成功")
    } else {
      ToastWarning("修改失败")
    }
    setLoading(false)
  }

  return (
    <div className={styles.container}>
      <BackHeader name="社群信息" />
      <div className={styles.content}>
        <div className={styles.icon}>
          <img className={styles.logo} src={activeAsset?.icon_url} alt="" />
          <div className={styles.edit} onClick={() => setShowSelectAsset(true)}>
            <i className="iconfont iconbianji" />
          </div>
        </div>
        {/*<p className={styles.title}>群徽章</p>*/}
        <input
          type="text"
          value={groupInfo.name || ""}
          onChange={(e) => setGroupInfo({ ...groupInfo, name: e.target.value })}
        />
        <p className={styles.tips}>
          名称后缀不要添加“群”字,自动建群时会自动在后面加上.
        </p>
        <textarea
          value={groupInfo.description || ""}
          onChange={(e) =>
            setGroupInfo({ ...groupInfo, description: e.target.value })
          }
        />
        <Button className="btn" onClick={handleClickSave} loading={loading}>
          保存
        </Button>
      </div>

      <Modal
        popup
        visible={showSelectAsset}
        onClose={() => setShowSelectAsset(false)}
        animationType="slide-up"
      >
        <CoinModal
          select={(asset) => {
            setActiveAsset(asset)
            setShowSelectAsset(false)
          }}
          active={activeAsset}
        />
      </Modal>
    </div>
  )
}
