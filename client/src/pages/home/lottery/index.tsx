import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import React, { useEffect, useRef, useState } from "react"
import { useIntl } from "react-intl"
import {
  ApiGetClaimPageData,
  ApiPostClain,
  ApiPostLotteryExchange,
  ApiGetLotteryReward,
} from "@/apis/claim"
import { ApiGetGroupList, IGroupItem } from "@/apis/group"
import { Modal } from "antd-mobile"
import { BroadcastBox } from "./widgets/broadcast"
import { Energy } from "./widgets/Energy"
import { LotteryBox } from "./widgets/LotteryBox"
import { Prize } from "./types"
import { ToastSuccess } from "@/components/Sub"
import { history } from "umi"
import { Lucker } from "@/types"
import { changeTheme } from "@/assets/ts/tools"
import { Icon } from "@/components/Icon"
import { FullLoading } from "@/components/Loading"
import { ApiGetAssetByID, IAsset } from "@/apis/asset"

import styles from "./lottery.less"

const BG = {
  idle: "https://super-group-cdn.mixinbots.com/lottery/bg.mp3",
  runing: "https://super-group-cdn.mixinbots.com/lottery/ing.mp3",
  success: "https://super-group-cdn.mixinbots.com/lottery/success.mp3",
}

type ModalType = "preview" | "receive"
type Assets = Record<string, Required<IAsset>>

export default function LotteryPage() {
  const t = get$t(useIntl())
  const [checkinCount, setCheckinCount] = useState(0)
  const [energy, setEnergy] = useState(0)
  const [prizes, setPrizes] = useState<Prize[]>()
  const [assets, setAssets] = useState<Record<string, Required<IAsset>>>()
  const [times, setTimes] = useState(0)
  const [modalType, setModalType] = useState<ModalType>()
  const [groupList, setGroupList] = useState<IGroupItem[]>([])
  const [reward, setReward] = useState<Prize>()
  const [luckers, setLuckers] = useState<Lucker[]>([])
  const [isReceiving, setIsReceiving] = useState(false)
  const [isCliamed, setIsClaimed] = useState(false)
  const [hasMusic, setHasMusic] = useState(false)
  const [hasRunMusic, setHasRunMusic] = useState(false)
  const [hasSuccessMusic, setHasSuccessMusic] = useState(false)
  const [loading, setLoading] = useState(true)
  const prevModalTypeRef = useRef<ModalType>()

  const fetchAssets = (prizeList?: Prize[]) =>
    new Promise<Record<string, Required<IAsset>>>((resolve, reject) => {
      if (!prizeList) return resolve({})

      Promise.all(prizeList.map((x) => ApiGetAssetByID(x.asset_id)))
        .then((data: any) => {
          const result = data.reduce(
            (acc: Record<string, Required<IAsset>>, cur: Required<IAsset>) => {
              if (!acc[cur.asset_id]) {
                return Object.assign(acc, { [cur.asset_id]: cur })
              }
              return acc
            },
            {},
          )

          setAssets(result)
          resolve(result)
        })
        .catch(reject)
    })

  const transfromPrize = async (
    type: ModalType,
    prize: Prize,
    prizeList?: Prize[],
  ) => {
    let tempAssets = assets

    if (!assets) {
      tempAssets = await fetchAssets(prizeList)
    }

    const targetAsset = tempAssets![prize.asset_id]
    const targetGroup = groupList.find((x) => x.client_id === prize.client_id)

    let symbol = prize.symbol
    let price_usd = prize.price_usd

    if (targetAsset) {
      symbol = targetAsset.symbol === "BTC" ? "SAT" : targetAsset.symbol
      price_usd = Number(
        (Number(targetAsset.price_usd) * Number(prize.amount)).toFixed(8),
      ).toString()
    }

    setModalType(type)
    setReward(
      Object.assign({}, targetGroup, prize, targetAsset, {
        amount:
          prize.asset_id == "c6d0c728-2624-429b-8e0d-d9d19b6592fa"
            ? (Number(prize.amount) * 1e8).toFixed()
            : prize.amount,
        icon_url: prize.icon_url,
        symbol,
        price_usd,
      }),
    )
  }

  const fetchPageData = (cb?: () => void) =>
    Promise.all([ApiGetClaimPageData(), ApiGetGroupList()]).then(([x, y]) => {
      setGroupList(y)
      setPrizes(x.lottery_list || [])
      setCheckinCount(x.count)
      setEnergy(x.power.balance)
      setTimes(x.power.lottery_times)
      setLuckers(x.last_lottery)
      setIsClaimed(x.is_claim)

      if (cb) cb()
      if (x.receiving) {
        transfromPrize("receive", x.receiving, x.lottery_list)
      }
    })

  useEffect(() => {
    fetchPageData()
  }, [])

  useEffect(() => {
    if (prizes && prizes.length) {
      fetchAssets(prizes)
    }
  }, [prizes])

  useEffect(() => {
    if (!loading) {
      changeTheme("#2b120b")
      document.body.classList.add(styles.bg)
    }

    return () => {
      changeTheme("#fff")
      document.body.classList.remove(styles.bg)
    }
  }, [loading])

  useEffect(() => {
    if (modalType !== prevModalTypeRef.current) {
      prevModalTypeRef.current = modalType
    }
  }, [modalType])

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

  const handleReceiveClick = () => {
    if (modalType === "preview") {
      return setModalType(undefined)
    }

    if (!reward?.trace_id) return
    setIsReceiving(true)
    ApiGetLotteryReward(reward.trace_id).then((x) => {
      if (typeof x === "object") {
        setIsReceiving(false)
        setModalType("preview")

        return
      }

      if (x === "success") {
        // return setTimeout(() => {
        ToastSuccess(t("claim.receiveSuccess"))
        setModalType(undefined)
        setIsReceiving(false)
        // }, 300)
      }
    })
  }

  const handleMusicToggle = () => {
    setHasMusic((prev) => !prev)
  }

  const handleAllImgLoad = () => {
    setLoading(false)
  }

  const handlePrizeClick = (prize: Prize) => {
    transfromPrize("preview", prize)
  }

  const handleJoinClick = () => {
    setModalType(undefined)
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
      <LotteryBox
        data={prizes}
        ticketCount={times}
        onEnd={handleLotteryEnd}
        onStart={handleLotteryStart}
        onImgLoad={handleAllImgLoad}
        onPrizeClick={handlePrizeClick}
      />
      {!loading && (
        <Energy
          checkinCount={checkinCount}
          value={energy}
          isCheckedIn={isCliamed}
          onCheckinClick={handleCheckinClick}
          onExchangeClick={handleExchangeClick}
        />
      )}

      {!loading && (
        <Modal
          popup
          visible={
            !!reward &&
            (modalType === "preview" ||
              (modalType === "receive" && !reward.is_received))
          }
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
              disabled={modalType === "receive" && isReceiving}
              className={`${styles.btn} ${
                modalType || prevModalTypeRef.current
                  ? styles[modalType || prevModalTypeRef.current!]
                  : ""
              }`}
              onClick={handleReceiveClick}
            >
              {isReceiving ? (
                <Icon i="loding" className={styles.loading} />
              ) : (
                <span>
                  {t(`claim.${modalType === "receive" ? "receive" : "ok"}`)}
                </span>
              )}
            </button>
            {modalType === "preview" && reward && (
              <a
                className={styles.join}
                href={`mixin://apps/${reward.client_id}?action=open`}
                onClick={handleJoinClick}
              >
                {t("claim.join")}
              </a>
            )}
          </div>
        </Modal>
      )}

      {loading && <FullLoading mask />}
      {hasMusic && <audio autoPlay src={BG.idle} loop />}
      {hasMusic && hasRunMusic && <audio autoPlay src={BG.runing} loop />}
      {hasMusic && hasSuccessMusic && <audio autoPlay src={BG.success} />}
    </div>
  )
}
