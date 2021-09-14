import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import React, { useEffect, useState, useCallback } from "react"
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

export default function LotteryPage() {
  const t = get$t(useIntl())
  const [checkinCount, setCheckinCount] = useState(0)
  const [energy, setEnergy] = useState(0)
  const [prizes, setPrizes] = useState()
  const [times, setTimes] = useState(0)
  const [reward, setReward] = useState<Prize>()
  const [isReceiving, setIsReceiving] = useState(false)

  const fetchPageData = () => {
    ApiGetClaimPageData().then((x) => {
      setPrizes(x.lottery_list || [])
      setCheckinCount(x.count)
      setEnergy(x.power.balance)
      setTimes(x.power.lottery_times)
      if (x.receiving) {
        setReward(x.receiving)
      }
    })
  }

  useEffect(() => {
    fetchPageData()
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
    fetchPageData()
  }, [])

  const handleRewardClick = () => {
    if (!reward?.trace_id) return
    setIsReceiving(true)
    ApiGetLotteryReward(reward.trace_id)
      .then(() => {
        ToastSuccess("领取成功，稍后给您转账")
        setReward(undefined)
      })
      .finally(() => {
        setIsReceiving(false)
      })
  }

  return (
    <div className={styles.container}>
      <BackHeader
        name={t("claim.title")}
        isWhite
        action={
          <>
            <i className={`iconfont iconic_music_open ${styles.music}`} />
            <i
              className="iconfont iconic_file_text"
              onClick={() => history.push("/lottery/records")}
            />
          </>
        }
      />
      <BroadcastBox>Crossle 抽到了 0.0000001 BTC，价值 $ 1.23</BroadcastBox>
      {prizes && (
        <LotteryBox
          data={prizes}
          ticketCount={times}
          onEnd={handleLotteryEnd}
        />
      )}
      <Energy
        checkinCount={checkinCount}
        value={energy}
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
          <p className={styles.value}>≈ ${reward?.price_usd}</p>
          <p className={styles.description}>
            MobileCoin（MOB）
            是一个隐私支付协议，专注为移动通讯应用程序提供安全、隐私、极简的加密货币服务，现已集成拥有
            4000 万月活用户的 Signal。
          </p>
          <button
            disabled={isReceiving}
            className={styles.btn}
            onClick={handleRewardClick}
          >
            {isReceiving ? (
              <i className={`iconfont iconloding ${styles.loading}`} />
            ) : (
              <span>领取奖品</span>
            )}
          </button>
        </div>
      </Modal>
    </div>
  )
}
