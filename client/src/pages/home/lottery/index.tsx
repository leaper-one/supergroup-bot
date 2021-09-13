import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import React, { useEffect, useState, useCallback } from "react"
import { useIntl } from "react-intl"
import {
  ApiGetClaimPageData,
  ApiPostClain,
  ApiPostLotteryExchange,
  ApiPostLottery,
} from "@/apis/claim"
import { BroadcastBox } from "./widgets/broadcast"
import { Start } from "./widgets/RollingBox/Start"
import { RollingBox, Item, Col, Row } from "./widgets/RollingBox"
import { Energy } from "./widgets/energy"

import styles from "./lottery.less"
import { LotteryBox } from "./widgets/LotteryBox"

export default function Lottery() {
  const t = get$t(useIntl())
  const [checkinCount, setCheckinCount] = useState(0)
  const [energy, setEnergy] = useState(0)
  const [prizes, setPrizes] = useState()
  const [times, setTimes] = useState(0)
  const [value, setValue] = useState("")

  const fetchPageData = () => {
    ApiGetClaimPageData().then((x) => {
      console.log(x)
      setPrizes(x.lottery_list || [])
      setCheckinCount(x.count)
      setEnergy(x.power.balance)
      setTimes(x.power.lottery_times)
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

  const handleStart = () => {
    ApiPostLottery().then((x) => {
      setValue(x.lottery_id)
    })
  }

  return (
    <div className={styles.container}>
      <BackHeader
        name={t("claim.title")}
        action={<i className="iconfont iconic_music_open" />}
      />
      <BroadcastBox>Crossle 抽到了 0.0000001 BTC，价值 $ 1.23</BroadcastBox>
      {prizes && <LotteryBox data={prizes} />}
      <Energy
        checkinCount={checkinCount}
        value={energy}
        onCheckinClick={handleCheckinClick}
        onExchangeClick={handleExchangeClick}
      />
    </div>
  )
}
