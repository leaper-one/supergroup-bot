import { BackHeader } from "@/components/BackHeader"
import React, { useEffect, useState } from "react"
import styles from "./red.less"
import { Button, ToastFailed, ToastSuccess } from "@/components/Sub"
import { ApiGetMyAssets, IAsset } from "@/apis/asset"
import { FullLoading, Loading } from "@/components/Loading"
import { history, useIntl } from "umi"
import { get$t } from "@/locales/tools"
import { ApiGetRedCheckPaid, ApiPostRedGenerate } from "@/apis/packet"
import { $get, $set } from "@/stores/localStorage"
import { TimingAmount, TimingPacketRate } from "@/pages/red/timing"
import {
  PopCoinModal,
  PopRateModal,
  PopRedTypeModal,
  TRedType,
} from "@/pages/red/modal"
import { ApiGetGroupList, getGroupID } from "@/apis/group"
import { payUrl } from "@/apis/http"

interface IAssetSelectProps {
  setCoinModal: (v: boolean) => void
  activeCoin: IAsset | undefined
}

export const AssetSelect = (props: IAssetSelectProps) => (
  <li className={styles.currency} onClick={() => props.setCoinModal(true)}>
    <img className={styles.assetIcon} src={props.activeCoin?.icon_url} alt="" />
    <span className={styles.name}>{props.activeCoin?.name}</span>
    <span
      className={styles.desc}
    >{`${props.activeCoin?.balance} ${props.activeCoin?.symbol}`}</span>
    <i className={`iconfont iconic_down ${styles.myIcon}`} />
  </li>
)

interface IMemoEditProps {
  memo: string
  setMemo: (v: string) => void
  $t: any
  noRight?: boolean
  placeholder?: string
}

export const MemoEdit = (props: IMemoEditProps) => (
  <li className={styles.words}>
    <input
      value={props.memo}
      onChange={(e) => props.setMemo(e.target.value)}
      type="text"
      placeholder={props.placeholder || "Bitcoin goes to the moon!"}
    />
    {!props.noRight && <span>{props.$t("red.memo")}</span>}
  </li>
)

export default () => {
  const [pageLoaded, setPageLoaded] = useState(false)
  const [coinModal, setCoinModal] = useState(false)
  const [typeModal, setTypeModal] = useState(false)
  const [payModal, setPayModal] = useState(false)
  const [rateModal, setRateModal] = useState(false)
  const [packetRate, setPacketRate] = useState("25")
  const [activeCoin, setActiveCoin] = useState<IAsset>()
  const [redType, setRedType] = useState<TRedType>("1")
  const [amount, setAmount] = useState("")
  const [memo, setMemo] = useState("")

  const [maxPeople, setMaxPeople] = useState(0)

  const _type = history.location.query?.type || "instant"

  const $t = get$t(useIntl())

  useEffect(() => {
    const t = $get("timing_packet")
    if (t) {
      setActiveCoin(t.asset)
      setRedType(t.mode)
      setPacketRate(t.rate)
      setAmount(t.amount)
      setMemo(t.memo)
    }

    Promise.all([ApiGetMyAssets(), ApiGetGroupList()]).then(
      ([assetList, groupList]) => {
        if (!(_type === "timing" && t)) setActiveCoin(assetList[0])
        setPageLoaded(true)
        const group = groupList.find(
          (item) => item.group_id === $get("group").group_id,
        )
        if (group) {
          setMaxPeople(group.people)
          $set("max_people", group.people)
        }
      },
    )
  }, [])

  const handleClickSendPacket = async (type: string) => {
    if (!amount) return ToastFailed($t("error.amount"), 1)
    if (type === "timing") {
      $set("timing_packet", {
        asset: activeCoin,
        mode: redType,
        rate: packetRate,
        amount,
        memo,
      })
      return history.push("/red/timing")
    }
    setPayModal(true)
    const { trace } = await ApiPostRedGenerate({
      type: "0",
      amount,
      rate: packetRate,
      mode: redType,
      memo,
      asset_id: activeCoin!.asset_id,
    })
    const sendPeople = ((maxPeople * Number(packetRate)) / 100) | 0 || 1
    const payAmount = sendPeople * Number(amount)
    window.location.href = payUrl({
      trace,
      asset: activeCoin?.asset_id,
      amount: payAmount.toFixed(8),
      memo: JSON.stringify({
        G: getGroupID(),
        O: "packet_send",
      }),
    })
    checkPaid(trace)
  }

  const checkPaid = async (trace_id: string) => {
    const { payed } = await ApiGetRedCheckPaid(trace_id)
    if (payed) {
      setPayModal(false)
      ToastSuccess("发送成功", 2)
      history.goBack()
    } else {
      setTimeout(() => checkPaid(trace_id), 1000)
    }
  }

  return (
    <>
      <div className={styles.container}>
        <BackHeader name={$t("red.title")} />
        <ul className={styles.list}>
          <AssetSelect activeCoin={activeCoin} setCoinModal={setCoinModal} />
          <li className={styles.redType} onClick={() => setTypeModal(true)}>
            <span>{$t("red.type." + redType)}</span>
            <i className={`iconfont iconic_down ${styles.myIcon}`} />
          </li>
          {[
            TimingPacketRate({ packetRate, setRateModal, $t }),
            TimingAmount({
              amount,
              setAmount,
              $t,
              priceUsd: activeCoin?.price_usd,
            }),
          ]}
          <MemoEdit memo={memo} setMemo={setMemo} $t={$t} />
        </ul>
        <footer>
          <Button
            type="red"
            onClick={() => handleClickSendPacket(_type as string)}
          >
            {$t(`red.${_type === "instant" ? "send" : "next"}`)}
          </Button>
          <p>{$t("red.tips")}</p>
        </footer>
      </div>
      <PopCoinModal
        coinModal={coinModal}
        setCoinModal={setCoinModal}
        activeCoin={activeCoin}
        setActiveCoin={setActiveCoin}
      />
      <PopRedTypeModal
        typeModal={typeModal}
        setTypeModal={setTypeModal}
        redType={redType}
        setRedType={setRedType}
        $t={$t}
      />
      <PopRateModal
        rateModal={rateModal}
        setRateModal={setRateModal}
        packetRate={packetRate}
        totalPeople={maxPeople}
        $t={$t}
        onSave={(r) => setPacketRate(r)}
      />
      {payModal && (
        <Loading
          content={$t("modal.check")}
          cancel={() => setPayModal(false)}
        />
      )}
      {!pageLoaded && <FullLoading mask />}
    </>
  )
}
