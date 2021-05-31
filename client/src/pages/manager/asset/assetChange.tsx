import React, { useEffect, useState } from "react"
import { history, useIntl } from "umi"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import redStyle from "@/pages/red/red.less"
import style from "./assetChange.less"
import { InstantAmountItem } from "@/pages/red/instant"
import { ApiCheckIsPaid, ApiGetMyAssets, IAsset } from "@/apis/asset"
import { PopCoinModal } from "@/pages/red/modal"
import { AssetSelect, MemoEdit } from "@/pages/red/red"
import { Button, ToastSuccess } from "@/components/Sub"
import { payUrl } from "@/apis/http"
import {
  ApiGetGroupAssets,
  ApiGetWithdrawalAssets,
  getGroupID,
} from "@/apis/group"
import { Loading } from "@/components/Loading"
import { delay, getUUID } from "@/assets/ts/tools"
import { $get } from "@/stores/localStorage"

export default () => {
  const path = history.location.pathname.split("/")
  if (!["deposit", "withdrawal"].includes(path[2])) return history.goBack()
  const $t = get$t(useIntl())
  const [coinModal, setCoinModal] = useState(false)
  const [activeCoin, setActiveCoin] = useState<IAsset>()
  const [amount, setAmount] = useState("")
  const [memo, setMemo] = useState("")
  const [loading, setLoading] = useState(false)
  const [assetList, setAssetList] = useState<IAsset[]>()

  useEffect(() => {
    if (path[2] === "withdrawal") {
      // 提现逻辑
      ApiGetGroupAssets().then((item) => {
        setAssetList(item)
        setActiveCoin(item[0])
      })
    } else {
      // 充值逻辑
      ApiGetMyAssets().then((item) => setActiveCoin(item[0]))
    }
  }, [])

  const handleClickBtn = async () => {
    if (path[2] === "deposit") {
      // 充值逻辑
      let trace = getUUID()
      window.location.href = payUrl({
        trace,
        asset: activeCoin!.asset_id!,
        amount,
        memo: JSON.stringify({
          G: getGroupID(),
          M: memo,
          O: "deposit",
        }),
      })
      setLoading(true)
      checkPaid(
        amount,
        activeCoin!.asset_id!,
        $get("user").user_id,
        trace,
        setLoading,
        $t,
      )
    } else {
      // 提现逻辑
      const data = await ApiGetWithdrawalAssets(activeCoin!.asset_id!, amount)
      if (data) {
        ToastSuccess("提现申请成功")
        setTimeout(() => history.goBack(), 500)
      }
    }
  }

  return (
    <div>
      <BackHeader name={$t(`manager.asset.${path[2]}`)} />
      <ul className={redStyle.list}>
        <AssetSelect setCoinModal={setCoinModal} activeCoin={activeCoin} />
        <InstantAmountItem
          amount={amount}
          setAmount={setAmount}
          $t={$t}
          noRight
          placeholder="数量"
        />
        <MemoEdit
          memo={memo}
          setMemo={setMemo}
          $t={$t}
          noRight
          placeholder="备注"
        />
      </ul>

      <Button
        className={style.btn}
        disabled={!activeCoin || !activeCoin.asset_id || !amount}
        onClick={handleClickBtn}
      >
        {$t(`manager.asset.action.${path[2]}`)}
      </Button>

      <PopCoinModal
        coinModal={coinModal}
        setCoinModal={setCoinModal}
        activeCoin={activeCoin}
        setActiveCoin={setActiveCoin}
        assetList={assetList}
      />
      {loading && <Loading />}
    </div>
  )
}

export const checkPaid = async (
  amount: string,
  asset_id: string,
  counter_user_id: string,
  trace_id: string,
  setLoading: any,
  $t: any,
  success: any = undefined,
) => {
  const res = await ApiCheckIsPaid({
    amount,
    asset_id,
    counter_user_id,
    trace_id,
  })

  if (res.status === "paid") {
    if (typeof success === "function") success()
    else ToastSuccess($t("manager.asset.depositSuccess"))
    setLoading(false)
  } else {
    await delay()
    checkPaid(
      amount,
      asset_id,
      counter_user_id,
      trace_id,
      setLoading,
      $t,
      success,
    )
  }
}
