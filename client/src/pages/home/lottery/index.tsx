import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import React, { useEffect, useState, useCallback, useMemo, useRef } from "react"
import { useIntl } from "react-intl"
import {
  ApiGetClaimPageData,
  ApiPostClain,
  ApiPostLotteryExchange,
  ApiGetLotteryReward,
} from "@/apis/claim"
import { Modal } from "antd-mobile"
import { BroadcastBox } from "./widgets/broadcast"
import { Energy } from "./widgets/energy"

import styles from "./lottery.less"
import { LotteryBox } from "./widgets/LotteryBox"
import { Prize } from "./types"
import { ToastSuccess } from "@/components/Sub"
import { history } from "umi"
import { Lucker } from "@/types"

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
  const [reward, setReward] = useState<Prize>()
  const [luckers, setLuckers] = useState<Lucker[]>([])
  const [isReceiving, setIsReceiving] = useState(false)
  const [isPlayMusic, setIsPlayMusic] = useState(false)
  const [isCliamed, setIsClaimed] = useState(false)
  const [music, setMusic] = useState<string>()

  const audioRef = useRef<HTMLAudioElement>()

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
        setReward(x.receiving)
      }
    })

  useEffect(() => {
    fetchPageData()
    audioRef.current = new Audio()
    audioRef.current.src = BG.idle
    audioRef.current.muted = true
    audioRef.current.autoplay = true
  }, [])

  const handleExchangeClick = useCallback(() => {
    ApiPostLotteryExchange().then(() => {
      fetchPageData()
    })
  }, [])

  const handleCheckinClick = useCallback(() => {
    ApiPostClain().then(() => {
      fetchPageData()
    })
  }, [])

  const handleLotteryEnd = useCallback(() => {
    fetchPageData(() => setMusic(BG.success))
  }, [])

  const handleLotteryStart = useCallback(() => {
    setMusic(BG.runing)
  }, [])

  const handleRewardClick = () => {
    if (!reward?.trace_id) return
    setIsReceiving(true)
    ApiGetLotteryReward(reward.trace_id)
      .then(() => {
        ToastSuccess(t("receiveSuccess"))
        setReward(undefined)
      })
      .finally(() => {
        setIsReceiving(false)
        setMusic(BG.idle)
      })
  }

  const handleMusicToggle = () => {
    setIsPlayMusic((prev) => !prev)
  }

  return (
    <div className={styles.container}>
      <BackHeader
        name={t("claim.title")}
        isWhite
        action={
          <>
            <button className={styles.action_music} onClick={handleMusicToggle}>
              <i className={`iconfont iconic_music_open ${styles.icon}`} />
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
        visible={!!reward && !reward.is_received}
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
              <i className={`iconfont iconloding ${styles.loading}`} />
            ) : (
              <span>{t("claim.receive")}</span>
            )}
          </button>
        </div>
      </Modal>
      <audio autoPlay muted={!isPlayMusic} src={music} />
    </div>
  )
}
