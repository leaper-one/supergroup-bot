import React, { useEffect, useState } from "react"
import { BackHeader } from "@/components/BackHeader"
import { DatePicker, Switch } from "antd-mobile"
import { PopCoinModal } from "@/pages/red/modal"
import { IAsset } from "@/apis/asset"
import {
  ApiGetGroupAssets,
  ApiPutGroupInviteSetting,
  ApiPutGroupStatus,
} from "@/apis/group"
import { AssetSelect } from "@/pages/red/red"
import redStyle from "@/pages/red/red.less"
import { InstantAmountItem } from "@/pages/red/instant"
import { get$t } from "@/locales/tools"
import { useIntl } from "umi"
import styles from "./invite.less"
import { Button, ToastFailed, ToastSuccess } from "@/components/Sub"
import { $get } from "@/stores/localStorage"

let isInit = false
export default () => {
  const { invite_status } = $get("setting")
  const [openStatus, setOpenStatus] = useState(invite_status)
  const [coinModal, setCoinModal] = useState(false)
  const [activeCoin, setActiveCoin] = useState<IAsset>()
  const [assetList, setAssetList] = useState<IAsset[]>([] as IAsset[])
  const [amount, setAmount] = useState("0.01")
  const [sendTime, setSendTime] = useState(new Date())
  const [duration, setDuration] = useState("7")
  const [times, setTimes] = useState("10")

  const $t = get$t(useIntl())

  const updateInviteSetting = async () => {
    const { group_id } = $get("group")
    const res = await ApiPutGroupInviteSetting({
      group_id,
      amount,
      asset_id: activeCoin?.asset_id,
      send_at: formatTime(sendTime),
      duration: Number(duration),
      times: Number(times),
    })
    if (res === true) {
      ToastSuccess("操作成功")
    }
  }

  useEffect(() => {
    ApiGetGroupAssets().then((item) => {
      setAssetList(item)
      const [activeCoin] = item
      isInit = false
      setActiveCoin(activeCoin)
    })
  }, [])

  useEffect(() => {
    if (!activeCoin) return
    const price = Number(activeCoin.price_usd)
    if (price === 0) return setAmount("10000")
    const amount = Number((1 / price).toFixed(8)).toString()
    if (isInit) updateInviteSetting().then(() => setAmount(amount))
    isInit = true
  }, [activeCoin])

  return (
    <div className={styles.container}>
      <BackHeader name="邀请入群" />

      <div className={styles.content}>
        <ul className={redStyle.list}>
          <li className={styles.formItem}>
            <p>邀请奖励</p>
            <Switch
              color="black"
              checked={openStatus !== "0"}
              onChange={async () => {
                const status = openStatus === "0" ? "1" : "0"
                const res = await ApiPutGroupStatus("invite_status", status)
                if (res) {
                  ToastSuccess("操作成功")
                  setOpenStatus(status)
                } else {
                  ToastFailed("操作失败")
                }
              }}
            />
          </li>
          {openStatus === "1" && (
            <>
              <AssetSelect
                setCoinModal={setCoinModal}
                activeCoin={activeCoin}
              />
              <InstantAmountItem
                amount={amount}
                setAmount={setAmount}
                onBlur={(amount) => updateInviteSetting()}
                $t={$t}
                noRight
                placeholder="数量"
              />

              <DatePicker
                mode="time"
                value={sendTime}
                onChange={(sendTime) => {
                  updateInviteSetting()
                  setSendTime(sendTime)
                }}
              >
                <li className={styles.formItem}>
                  <p>发放时间</p>
                  <span>{formatTime(sendTime)}</span>
                </li>
              </DatePicker>

              <li className={styles.formItem}>
                <input
                  type="number"
                  placeholder="请输入间隔天数"
                  value={duration}
                  onChange={(e) => setDuration(e.target.value)}
                  onBlur={(e) => updateInviteSetting()}
                />
                <span>天</span>
              </li>

              <li className={styles.formItem}>
                <input
                  type="number"
                  placeholder="请输入发放次数"
                  value={times}
                  onChange={(e) => setTimes(e.target.value)}
                  onBlur={(e) => updateInviteSetting()}
                />
                <span>次数</span>
              </li>
            </>
          )}
        </ul>

        {openStatus === "1" && (
          <>
            <p className={styles.desc}>
              被邀请人入群每满 {duration} 天给邀请人发一次奖励，每次奖励{" "}
              {amount} ，每个有效邀请人分 {times} 次历时{" "}
              {(Number(times) * Number(duration)) | 0} 天总共发
              {(Number(amount) * Number(times)) | 0} 邀请奖励。邀请 1000
              人预计花费 {(Number(amount) * Number(times) * 1000) | 0}{" "}
              {activeCoin?.symbol}。
            </p>
            <p className={styles.desc}>
              如果邀请人在 48
              小时内没有领取当次邀请奖励，奖励将自动作废并退回至当前社群资金账户。
            </p>

            <Button className={styles.btn}>充值 {activeCoin?.symbol}</Button>
          </>
        )}
      </div>

      <PopCoinModal
        coinModal={coinModal}
        setCoinModal={setCoinModal}
        activeCoin={activeCoin}
        setActiveCoin={setActiveCoin}
        assetList={assetList}
      />
    </div>
  )
}

function formatTime(date: Date) {
  const hours = date.getHours()
  const minutes = date.getMinutes()
  return [hours, minutes].join(":")
}
