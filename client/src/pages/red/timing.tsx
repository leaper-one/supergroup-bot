import React, { useState } from "react"

import redStyles from "./red.less"
import styles from "./timing.less"
import { BackHeader } from "@/components/BackHeader"
import { Button } from "@/components/Sub"
import { $get } from "@/stores/localStorage"
import { ApiPostRedGenerate } from "@/apis/packet"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { AddTimingPacketSuccessModal, PopTimingModal } from "@/pages/red/modal"

interface ITimingPacketRate {
  packetRate: string
  setRateModal: (b: boolean) => void
  $t: (v: string) => string
}

export const TimingPacketRate = (props: ITimingPacketRate) => (
  <li
    key="packet-rate"
    className={redStyles.redType}
    onClick={() => props.setRateModal(true)}
  >
    <span>{`${props.$t("red.rate")} ${props.packetRate}%`}</span>
    <i className={`iconfont iconic_down ${redStyles.myIcon}`} />
  </li>
)

interface ITimingAmountProps {
  amount: string
  setAmount: (v: string) => void
  $t: (v: string) => string
  priceUsd: string | undefined
}

export const TimingAmount = (props: ITimingAmountProps) => (
  <li key="amount" className={styles.amount}>
    <div className={styles.title}>
      <input
        value={props.amount}
        onChange={(e) => props.setAmount(e.target.value)}
        type="text"
        placeholder={props.$t("red.amount")}
      />
      <span>
        {(Number(props.amount) * Number(props.priceUsd))?.toFixed(2) || "0.00"}{" "}
        USD
      </span>
    </div>
    <p>{props.$t("red.amountDesc")}</p>
  </li>
)

export default function TimingPacketTime() {
  const [timingModal, setTimingModal] = useState(false)
  const [successModal, setSuccessModal] = useState(false)
  const [loading, setLoading] = useState(false)

  const [date_cycle, setDateCycle] = useState("")
  const [time_cycle, setTimeCycle] = useState("")
  const [send_times, setSendTimes] = useState(1)

  const $t = get$t(useIntl())
  const t = $get("timing_packet")

  const handelClickAddTimingPacket = async () => {
    if (loading) {
      return
    }

    setLoading(true)
    const { trace } = await ApiPostRedGenerate({
      type: "1",
      asset_id: t.asset.asset_id,
      amount: t.amount,
      memo: t.memo,
      mode: t.mode,
      rate: t.rate,
      date_cycle,
      time_cycle,
      send_times,
    })
    setLoading(false)

    if (trace) {
      setSuccessModal(true)
    }
  }

  let time = ""
  if (time_cycle) {
    if (date_cycle.split(",").length === 7) {
      time = "每天"
    } else {
      time = "每周"
      time += date_cycle
        .split(",")
        .map((item) => $t("red.week." + item))
        .join("、")
    }
    time += " "
    let [h, m] = time_cycle.split(":")
    let mora = "morning"
    if (Number(h) > 12) {
      mora = "afternoon"
      h = String(Number(h) - 12)
    }
    if (Number(h) < 10) h = "0" + h
    if (Number(m) < 10) m = "0" + m
    time += `${$t("red.timing." + mora)} ${h}:${m}`
  }

  return (
    <>
      <div className={`${styles.container} ${redStyles.container}`}>
        <BackHeader name={$t("red.timingTitle")} />
        <ul>
          <li
            className={redStyles.redType}
            onClick={() => setTimingModal(true)}
          >
            <span className={time_cycle ? "" : styles.grep}>
              {time_cycle ? `${time}` : $t("red.packetTime")}
            </span>
            <i className={`iconfont iconic_down ${redStyles.myIcon}`} />
          </li>
          <li className={redStyles.amount}>
            <input
              type="number"
              value={send_times}
              onChange={(e) => setSendTimes(Number(e.target.value))}
            />
            <span>{$t("red.times")}</span>
          </li>
        </ul>
        <footer>
          <Button
            onClick={() => handelClickAddTimingPacket()}
            disabled={!(time_cycle && send_times)}
            type="red"
            loading={loading}
          >
            {$t("red.addTiming")}
          </Button>
        </footer>
      </div>
      <PopTimingModal
        timingModal={timingModal}
        setTimingModal={setTimingModal}
        date_cycle={date_cycle}
        setDateCycle={setDateCycle}
        time_cycle={time_cycle}
        setTimeCycle={setTimeCycle}
        $t={$t}
        onSave={(date, time) => {
          setDateCycle(date)
          setTimeCycle(time)
        }}
      />
      <AddTimingPacketSuccessModal
        asset={t.asset}
        successModal={successModal}
        setSuccessModal={setSuccessModal}
        time={time}
        rate={t.rate}
        maxPeople={$get("max_people")}
        amount={t.amount}
        times={send_times}
      />
    </>
  )
}
