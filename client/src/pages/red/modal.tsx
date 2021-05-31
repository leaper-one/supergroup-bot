import React, { useState } from "react"
import { Modal, Slider } from "antd-mobile"
import { CoinModal } from "@/components/PopupModal/coinSelect"
import { IAsset } from "@/apis/asset"
import redStyle from "@/pages/red/red.less"
import { Button, ToastFailed } from "@/components/Sub"
import styles from "./modal.less"
import { JoinModal } from "@/components/PopupModal/join"

interface IPopCoinModalProps {
  coinModal: boolean
  setCoinModal: (v: boolean) => void
  activeCoin: IAsset | undefined
  setActiveCoin: (a: IAsset | undefined) => void
  assetList?: IAsset[]
}

export const PopCoinModal = (props: IPopCoinModalProps) => {
  return (
    <Modal
      popup
      animationType="slide-up"
      visible={props.coinModal}
      onClose={() => props.setCoinModal(false)}
    >
      <CoinModal
        active={props.activeCoin}
        select={(asset) => {
          props.setActiveCoin(asset)
          props.setCoinModal(false)
        }}
        myAsset
        assetList={props.assetList}
      />
    </Modal>
  )
}

interface IRedType {
  "0": string
  "1": string
}

export type TRedType = keyof IRedType

interface IPopRedTypeModalProps {
  typeModal: boolean
  setTypeModal: (v: boolean) => void
  redType: TRedType
  setRedType: (t: TRedType) => void
  $t: (v: string) => string
}

export const PopRedTypeModal = (props: IPopRedTypeModalProps) => (
  <Modal
    popup
    animationType="slide-up"
    visible={props.typeModal}
    onClose={() => props.setTypeModal(false)}
  >
    <div className={redStyle.redTypeModal}>
      <i
        className={`iconfont iconguanbi ${redStyle.close}`}
        onClick={() => props.setTypeModal(false)}
      />
      <h3>{props.$t("red.type.title")}</h3>
      <ul>
        {["0", "1"].map((item) => (
          <li
            key={item}
            onClick={() => {
              props.setRedType(item as TRedType)
              props.setTypeModal(false)
            }}
          >
            <p>{props.$t("red.type." + item)}</p>
            <span>{props.$t(`red.type.${item}Desc`)}</span>
            {item === props.redType && (
              <i className={`iconfont iconcheck ${redStyle.selected}`} />
            )}
          </li>
        ))}
      </ul>
    </div>
  </Modal>
)

interface IPopRateModalProps {
  rateModal: boolean
  setRateModal: (v: boolean) => void

  packetRate: string
  $t: (v: string, obj?: object) => string
  totalPeople: number
  onSave: (rate: string) => void
}

export const PopRateModal = (props: IPopRateModalProps) => {
  const [rate, setRate] = useState(Number(props.packetRate))
  return (
    <Modal
      className={styles.rateModal}
      popup
      animationType="slide-up"
      visible={props.rateModal}
      onClose={() => props.setRateModal(false)}
    >
      <h4 className={styles.title}>{props.$t("red.rate")}</h4>

      <div className={styles.slider}>
        <Slider
          min={1}
          max={100}
          step={1}
          value={rate}
          onChange={(e) => setRate(e || 1)}
          railStyle={{
            backgroundColor: "#E5E7EB",
            height: "8px",
            borderRadius: "4px",
          }}
          trackStyle={{
            backgroundColor: "#222",
            height: "8px",
            borderRadius: "4px",
          }}
          handleStyle={{
            borderColor: "#606067",
            height: "14px",
            width: "14px",
            marginTop: "-4px",
            marginLeft: "-5.5px",
            backgroundColor: "#606067",
          }}
        />

        <span style={{ left: `calc(${rate}%)` }} className={styles.sliderRate}>
          {rate}%
        </span>

        <span className={styles.sliderMin}>1%</span>
        <span className={styles.sliderMax}>100%</span>
      </div>

      <p className={styles.desc}>
        {props.$t("red.rateDesc", {
          people: props.totalPeople,
          rate: rate,
          receive: ((props.totalPeople * rate) / 100).toFixed(0),
        })}
      </p>

      <Button
        className={styles.btn}
        onClick={() => {
          props.onSave(String(rate))
          props.setRateModal(false)
        }}
      >
        保存
      </Button>
    </Modal>
  )
}

interface IPopTimingModalProps {
  timingModal: boolean
  setTimingModal: (v: boolean) => void

  date_cycle: string
  setDateCycle: (v: string) => void

  time_cycle: string
  setTimeCycle: (v: string) => void

  $t: (v: string, obj?: object) => string

  onSave: (date: string, time: string) => void
}

// 设置红包时间...
export const PopTimingModal = (props: IPopTimingModalProps) => {
  let [h, m] = props.time_cycle.split(":") || []
  let _mora = "morning"
  if (h && Number(h) > 12) {
    _mora = "afternoon"
    h = String(Number(h) - 12)
  }

  const [mora, setMora] = useState(_mora)
  const [hour, setHour] = useState(h)
  const [minute, setMinute] = useState(m)

  const date_cycle = props.date_cycle ? props.date_cycle.split(",") : []
  const [dateCycle, setDateCycle] = useState(date_cycle)

  const { $t } = props

  return (
    <Modal
      className={styles.timingModal}
      popup
      animationType="slide-up"
      visible={props.timingModal}
      onClose={() => props.setTimingModal(false)}
    >
      <h4>{$t("red.timing.title")}</h4>
      <p>{$t("red.timing.time")}</p>
      <div className={styles.time}>
        <div className={styles.timeMORA}>
          {["morning", "afternoon"].map((item) => (
            <span
              key={item}
              className={mora === item ? styles.active : ""}
              onClick={() => setMora(item)}
            >
              {$t("red.timing." + item)}
            </span>
          ))}
        </div>
        <div className={styles.timeTime}>
          <input
            type="text"
            onChange={(e) => setHour(e.target.value)}
            value={hour || ""}
          />
          <span>{$t("red.timing.hour")}</span>
          <input
            type="text"
            onChange={(e) => setMinute(e.target.value)}
            value={minute || ""}
          />
          <span>{$t("red.timing.minute")}</span>
        </div>
      </div>

      <p>{$t("red.timing.repeat")}</p>
      <ul className={styles.repeat}>
        {["everyday", "1", "2", "3", "4", "5", "6", "0"].map((item) => (
          <li
            key={item}
            className={
              dateCycle.some((date) => date === item) ? styles.active : ""
            }
            onClick={() => {
              if (item === "everyday")
                return setDateCycle(
                  dateCycle.length !== 7
                    ? ["1", "2", "3", "4", "5", "6", "0"]
                    : [],
                )
              let idx = dateCycle.findIndex((date) => item === date)
              if (idx === -1) {
                setDateCycle([...dateCycle, item])
              } else {
                dateCycle.splice(idx, 1)
                setDateCycle([...dateCycle])
              }
            }}
          >
            {$t(`red.timing.${item}`)}
          </li>
        ))}
      </ul>

      <Button
        className={styles.btn}
        onClick={() => {
          if (dateCycle.length === 0) return ToastFailed("重复错误")
          if (!hour || !minute) return ToastFailed("时间错误")
          let h = hour,
            m = minute
          if (mora === "afternoon") h = String(Number(h) + 12)

          props.onSave(dateCycle.join(","), [h, m].join(":"))
          props.setTimingModal(false)
        }}
      >
        保存
      </Button>
    </Modal>
  )
}

interface IPacketSuccessModalProps {
  successModal: boolean
  setSuccessModal: (v: boolean) => void

  time: string
  rate: string
  amount: string
  times: number
  asset: IAsset
  maxPeople: number
}

export const AddTimingPacketSuccessModal = (
  props: IPacketSuccessModalProps,
) => {
  const { asset, time, rate, amount, times, maxPeople } = props
  const total = Number(
    (
      Math.ceil((Number(maxPeople) * Number(rate)) / 100) *
      Number(amount) *
      times
    ).toFixed(8),
  )
  return (
    <Modal
      popup
      animationType="slide-up"
      visible={props.successModal}
      onClose={() => props.setSuccessModal(false)}
    >
      <JoinModal
        modalProp={{
          isAirdrop: true,
          title: "定时红包添加成功",
          icon_url: asset.icon_url,
          desc: `${time} 发红包，每次 ${rate}% 的人能抢到红包，人均 ${amount} ${asset.symbol}，总共发 ${times} 次，预计总花费 ${total} ${asset.symbol}`,
          button: "充值 " + asset.symbol,
          tips: "稍后充值",
          tipsStyle: "blank",
          tipsAction: () => props.setSuccessModal(false),
        }}
      />
    </Modal>
  )
}
