import { ApiGetLotteryRecord, ApiGetClaimRecord } from "@/apis/claim"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { RecordByDate } from "@/types"
import React, { FC, useEffect, useState } from "react"
import { useIntl } from "react-intl"
import styles from "./records.less"

export default function Histories() {
  const t = get$t(useIntl())
  const [isClaimTab, setIsClaimTab] = useState(false)
  const [lotteryRecords, setLotteryRecords] = useState<RecordByDate[]>([])
  const [claimRecords, setClaimRecords] = useState<RecordByDate[]>([])

  const fetchList = (isLottery?: boolean) => {
    if (isLottery) {
      return ApiGetLotteryRecord().then(setLotteryRecords)
    }

    ApiGetClaimRecord().then(setClaimRecords)
  }

  useEffect(() => {
    fetchList(!isClaimTab)
  }, [isClaimTab])

  const renderList = (data: RecordByDate[]) => (
    <ul className={styles.records}>
      {data.map(([date, list]) => (
        <>
          <li key={date} className={styles.date}>
            {date}
          </li>
          {list.map((r, idx) => (
            <li key={r.lottery_id || idx} className={styles.record}>
              <div className={styles.recordLeft}>
                <div className={styles.logo}>
                  {r.icon_url ? (
                    <img src={r.icon_url} />
                  ) : (
                    <i
                      className={`iconfont ${
                        r.power_type === "cliam"
                          ? "iconic_qiandao"
                          : "iconic_yaoqing"
                      }`}
                    />
                  )}
                </div>
                <span className={styles.name}>
                  {r.symbol
                    ? "抽奖"
                    : r.power_type == "claim"
                    ? "签到"
                    : "能量兑换"}
                </span>
              </div>
              <div className={styles.recordRight}>
                <span
                  className={
                    Number(r.amount) < 0 ? styles.negative : styles.plus
                  }
                >
                  {Number(r.amount) < 0 ? "" : "+"}
                  {r.amount}
                </span>
                <span className={styles.desc}>
                  {r.symbol ? r.symbol : "能量"}
                </span>
              </div>
            </li>
          ))}
        </>
      ))}
    </ul>
  )

  const lotteryList = renderList(lotteryRecords)

  const energyList = renderList(claimRecords)

  return (
    <>
      <BackHeader name={t("claim.records.title")} />
      <div className={styles.page}>
        <TabSwitchBar activeRight={isClaimTab} onSwitch={setIsClaimTab} />
        {/* {renderList(isClaimTab ? claimRecords : lotteryRecords)} */}
        <div
          className={`${styles.tabPanel} ${isClaimTab ? styles.switch : ""}`}
        >
          <div className={styles.content}>
            <div className={styles.left}>{lotteryList}</div>
            <div className={styles.right}>{energyList}</div>
          </div>
        </div>
      </div>
    </>
  )
}

interface TabSwitchBarProps {
  activeRight: boolean
  onSwitch(param: boolean): void
}

const TabSwitchBar: FC<TabSwitchBarProps> = ({ activeRight, onSwitch }) => {
  return (
    <div className={styles.tabbar}>
      <div
        className={`${styles.switch} ${activeRight ? styles.active : ""}`}
        onClick={() => {
          onSwitch(!activeRight)
        }}
      >
        <div className={styles.item}>中奖记录</div>
        <div className={styles.item}>能量记录</div>
      </div>
    </div>
  )
}
