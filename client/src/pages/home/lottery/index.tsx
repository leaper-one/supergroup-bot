import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import React, { useEffect, useState } from "react"
import { useIntl } from "react-intl"
import {
  ApiGetClaimPageData,
  ApiPostClain,
  ApiPostLotteryExchange,
  ApiGetLotteryReward,
} from "@/apis/claim"
import { Modal } from "antd-mobile"
import { BroadcastBox } from "./widgets/broadcast"
import { Energy } from "./widgets/Energy"

import styles from "./lottery.less"
import { LotteryBox } from "./widgets/LotteryBox"
import { Prize } from "./types"
import { ToastSuccess } from "@/components/Sub"
import { history } from "umi"
import { Lucker } from "@/types"
import { changeTheme } from "@/assets/ts/tools"
import { Icon } from "@/components/Icon"

const BG = {
  idle: "https://super-group-cdn.mixinbots.com/lottery/bg.mp3",
  runing: "https://super-group-cdn.mixinbots.com/lottery/ing.mp3",
  success: "https://super-group-cdn.mixinbots.com/lottery/success.mp3",
}

export default function LotteryPage() {
  const t = get$t(useIntl())
  const [checkinCount, setCheckinCount] = useState(0)
  const [energy, setEnergy] = useState(0)
  const [prizes, setPrizes] = useState()
  const [times, setTimes] = useState(0)
  const [showReward, setShowReward] = useState(false)
  const [reward, setReward] = useState<Prize>()
  const [luckers, setLuckers] = useState<Lucker[]>([])
  const [isReceiving, setIsReceiving] = useState(false)
  const [isCliamed, setIsClaimed] = useState(false)
  const [hasMusic, setHasMusic] = useState(false)
  const [hasRunMusic, setHasRunMusic] = useState(false)
  const [hasSuccessMusic, setHasSuccessMusic] = useState(false)

  const fetchPageData = (cb?: () => void) =>
    ApiGetClaimPageData().then((x) => {
      setPrizes(x.lottery_list || [])
      setCheckinCount(x.count)
      setEnergy(x.power.balance)
      setTimes(x.power.lottery_times)
      setLuckers(x.last_lottery)
      setIsClaimed(x.is_claim)

      if (cb) cb()
      if (x.receiving) {
        setShowReward(true)
        setReward(x.receiving)
      }
    })

  useEffect(() => {
    fetchPageData()
    changeTheme("#2b120b")
    return () => {
      changeTheme("#fff")
    }
  }, [])

  const handleExchangeClick = () => {
    ApiPostLotteryExchange().then(() => {
      ToastSuccess(t("claim.energy.success"))
      fetchPageData()
    })
  }

  const handleCheckinClick = () => {
    ApiPostClain().then(() => {
      fetchPageData()
    })
  }

  const handleLotteryEnd = () => {
    fetchPageData(() => {
      setHasRunMusic(false)
      setHasSuccessMusic(true)
      setTimeout(() => {
        setHasSuccessMusic(false)
      }, 2000)
    })
  }

  const handleLotteryStart = () => {
    setHasRunMusic(true)
  }

  const handleRewardClick = () => {
    if (!reward?.trace_id) return
    setIsReceiving(true)
    ApiGetLotteryReward(reward.trace_id)
      .then(() => {
        ToastSuccess(t("claim.receiveSuccess"))
        setShowReward(false)
        setTimeout(() => {
          setReward(undefined)
        }, 1000)
      })
      .finally(() => {
        setIsReceiving(false)
      })
  }

  const handleMusicToggle = () => {
    setHasMusic((prev) => !prev)
  }

  return (
    <div className={styles.container}>
      <BackHeader
        name={t("claim.title")}
        isWhite
        action={
          <>
            <button className={styles.action_music} onClick={handleMusicToggle}>
              <i
                className={`iconfont ${
                  hasMusic ? "iconic_music_open" : "iconic_music_close"
                } ${styles.icon}`}
              />
            </button>
            <i
              className={`iconfont iconic_file_text ${styles.action_records}`}
              onClick={() => history.push("/lottery/records")}
            />
          </>
        }
      />
      <BroadcastBox data={luckers} />
      {prizes && (
        <LotteryBox
          data={prizes}
          ticketCount={times}
          onEnd={handleLotteryEnd}
          onStart={handleLotteryStart}
        />
      )}
      <Energy
        checkinCount={checkinCount}
        value={energy}
        isCheckedIn={isCliamed}
        onCheckinClick={handleCheckinClick}
        onExchangeClick={handleExchangeClick}
      />

      <Modal
        popup
        visible={showReward && !!reward && !reward.is_received}
        transparent
        maskClosable={false}
        animationType="slide-up"
      >
        <div className={styles.modal}>
          <div className={styles.header}>
            <img src={reward?.icon_url} className={styles.icon} />
            <h3>
              {reward?.amount} {reward?.symbol}
            </h3>
          </div>
          {reward && Number(reward.price_usd) > 0 && (
            <p className={styles.value}>≈ ${reward?.price_usd}</p>
          )}
          <p className={styles.description}>{reward?.description}</p>
          <button
            disabled={isReceiving}
            className={styles.btn}
            onClick={handleRewardClick}
          >
            {isReceiving ? (
              <Icon i="loading" className={styles.loading} />
            ) : (
              <span>{t("claim.receive")}</span>
            )}
          </button>
        </div>
      </Modal>
      {hasMusic && <audio autoPlay src={BG.idle} loop />}
      {hasMusic && hasRunMusic && <audio autoPlay src={BG.runing} loop />}
      {hasMusic && hasSuccessMusic && <audio autoPlay src={BG.success} />}
    </div>
  )
}
