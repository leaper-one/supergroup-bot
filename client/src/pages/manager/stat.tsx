import React, { useEffect, useState } from "react"
import styles from "./stat.less"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { ApiGetGroupStat } from "@/apis/group"
import { getOptions, getStatisticsDate } from "./statData"
import * as echarts from "echarts"
import { BigNumber } from "bignumber.js"
import { Icon } from "@/components/Icon"

interface IStat {
  totalUser: number
  highUser: number
  weekUser: number
  weekActiveUser: number
  totalMessage: number
  weekMessage: number
}

const statList = [
  "totalUser",
  "highUser",
  "weekUser",
  "weekActiveUser",
  "totalMessage",
  "weekMessage",
]
const mode = ["all", "month", "week"]

// let userCharts: any
// let messageCharts: any
let charts: any = {}
let data: any
export default function Page() {
  const $t = get$t(useIntl())
  const [stat, setStat] = useState<any>()
  const [showUserModal, setShowUserModal] = useState(false)
  const [userMode, setUserMode] = useState("all")
  const [messageMode, setMessageMode] = useState("all")
  const [showMessageModal, setShowMessageModal] = useState(false)
  useEffect(() => {
    ApiGetGroupStat().then((res) => {
      data = getStatisticsDate(res, $t)
      setStat(data[2])
      initCharts(userMode, "user")
      initCharts(userMode, "message")
    })
    return () => {
      charts = {}
    }
  }, [])
  const initCharts = (time: string, chart: string) => {
    if (!charts[chart])
      charts[chart] = echarts.init(document.getElementById(chart)!)
    let _data: any = chart === "user" ? data[0] : data[1]
    let duration: number = 0
    if (time === "week") {
      duration = 7 * 24 * 60 * 60 * 1000
    } else if (time === "month") {
      duration = 30 * 24 * 60 * 60 * 1000
    }
    if (duration) {
      _data = JSON.parse(JSON.stringify(_data))
      for (let i = 0; i < _data.data.length; i++) {
        const { list } = _data.data[i]
        _data.data[i].list = list.filter(
          (item: any) => Date.now() - Number(new Date(item[0])) < duration,
        )
      }
    }
    charts[chart].setOption(getOptions(_data))
  }

  return (
    <div className={styles.container}>
      <BackHeader name={$t("stat.title")} />
      <div className={styles.list}>
        {statList.map((item) => (
          <div key={item} className={styles.item}>
            <p>{$t(`stat.${item}`)}</p>
            <h3>{stat && new BigNumber(stat[item]).toFormat()}</h3>
          </div>
        ))}
      </div>
      <div className={styles.charts}>
        <div id="user" style={{ width: "100%", height: "100%" }} />
        <Select
          showModal={showUserModal}
          setShowModal={setShowUserModal}
          mode={userMode}
          setMode={setUserMode}
          $t={$t}
          initCharts={initCharts}
          chart="user"
        />
      </div>
      <div className={styles.charts}>
        <div id="message" style={{ width: "100%", height: "100%" }} />
        <Select
          showModal={showMessageModal}
          setShowModal={setShowMessageModal}
          mode={messageMode}
          setMode={setMessageMode}
          $t={$t}
          initCharts={initCharts}
          chart="message"
        />
      </div>
    </div>
  )
}

interface ISelectProps {
  showModal: boolean
  setShowModal: (v: boolean) => void
  mode: string
  setMode: (v: string) => void
  $t: any
  initCharts: (m: string, c: string) => void
  chart: string
}

const Select = (props: ISelectProps) => (
  <div className={styles.select}>
    <div onClick={() => props.setShowModal(!props.showModal)}>
      <span>{props.$t("stat." + props.mode)}</span>
      <Icon i="ic_down" />
    </div>
    {props.showModal && (
      <div className={styles.selectModal}>
        {mode.map((item) => (
          <span
            key={item}
            className={styles.selectModalItem}
            onClick={() => {
              props.initCharts(item, props.chart)
              props.setMode(item)
              props.setShowModal(false)
            }}
          >
            {props.$t("stat." + item)}
          </span>
        ))}
      </div>
    )}
  </div>
)
