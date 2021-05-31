import React, { useEffect, useState } from "react"
import styles from "./coin.less"
import { BackHeader } from "@/components/BackHeader"
import { IAsset } from "@/apis/asset"
import { Modal, Toast } from "antd-mobile"
import { CoinModal } from "@/components/PopupModal/coinSelect"
import { $get, $set } from "@/stores/localStorage"
import { history } from "umi"
import { Button, ToastSuccess } from "@/components/Sub"
import { getCurrentGroup } from "@/pages/home"
import { ApiGetGroupInfo, ApiPutGroupSetting } from "@/apis/group"

export default () => {
  const [coinList, setCoinList] = useState<IAsset[]>([])
  const [showSelectCoin, setSelectCoin] = useState(false)
  const [avoidCoin, setAvoidCoin] = useState<string[]>([])
  const [isManager, setIsManager] = useState(false)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setIsManager(isFromManager())
    initPage()
  }, [])

  const initPage = async () => {
    if (isFromManager()) {
      const { group_number } = $get("group")
      const group = await ApiGetGroupInfo(group_number)
      setCoinList(group.checks!)
    } else {
      const create = $get("create")
      if (create?.coin) setCoinList(create.coin)
    }
  }

  const select = (asset: IAsset | undefined) => {
    setSelectCoin(false)
    if (!asset) return
    setCoinList([...coinList, asset])
    setAvoidCoin([...avoidCoin, asset.asset_id!])
  }

  const deleteConfirm = (idx: number) => {
    Modal.alert("提示", "确认删除？", [
      { text: "取消", style: "default" },
      {
        text: "确认",
        onPress: () => {
          coinList.splice(idx, 1)
          avoidCoin.slice(idx, 1)
          setCoinList(coinList.slice())
          setAvoidCoin(avoidCoin.slice())
        },
      },
    ])
  }

  const handleClickBtn = async () => {
    setLoading(true)
    if (!isManager) handleClickNext()
    else await handleClickSave()
    setLoading(false)
  }

  const handleClickNext = () => {
    const canReceived = coinList.every((item) => Number(item.amount) > 0.00001)
    if (!canReceived) return Toast.fail("持仓最小为 0.0001", 2)
    const create = $get("create")
    create.coin = coinList
    $set("create", create)
    history.push("/create/check")
  }

  const handleClickSave = async () => {
    const { group_id } = $get("group")
    const res = await ApiPutGroupSetting({
      group_id,
      coin: coinList,
    })
    if (res) {
      const group = await getCurrentGroup()
      if (group) group.check = coinList
      ToastSuccess("修改成功")
    } else {
      ToastSuccess("修改失败")
    }
  }

  return (
    <div className={styles.container}>
      <BackHeader
        name="设置持仓币种"
        action={
          coinList.length > 0 ? (
            <i
              className={styles.addIcon + " iconfont iconic_add"}
              onClick={() => setSelectCoin(true)}
            />
          ) : undefined
        }
      />
      {coinList.length > 0 ? (
        <ul>
          {coinList.map((item, idx) => (
            <li key={idx}>
              <i
                className={styles.icon + " iconfont iconshanchu"}
                onClick={() => deleteConfirm(idx)}
              />
              <div>
                <img src={item.icon_url} alt="" />
                <span>{item.symbol}</span>
              </div>
              <input
                type="number"
                placeholder="最小0.00001"
                value={item.amount}
                onChange={(e) => {
                  item.amount = e.target.value
                  setCoinList(coinList.slice())
                }}
              />
            </li>
          ))}
        </ul>
      ) : (
        <div className={styles.noCoin}>
          <div
            className={styles.addContent}
            onClick={() => setSelectCoin(true)}
          >
            <i className="iconfont icontianjia" />
            <span>添加持仓币种</span>
          </div>
        </div>
      )}
      <footer className={styles.footer}>
        <Button
          disabled={coinList.length === 0}
          onClick={handleClickBtn}
          loading={loading}
        >
          {isManager ? "保存" : "下一步"}
        </Button>
        <p>
          最多对 10
          个币种进行持仓检查，满足任一币种持仓要求即可，不会对管理员进行持仓检查。
        </p>
      </footer>
      <Modal
        popup
        visible={showSelectCoin}
        onClose={() => setSelectCoin(false)}
        animationType="slide-up"
      >
        <CoinModal select={select} avoid={avoidCoin} />
      </Modal>
    </div>
  )
}
export const isFromManager = (): boolean => {
  const { query } = history.location
  return query?.from === "manager"
}
