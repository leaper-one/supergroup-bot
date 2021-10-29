import { ApiGetLotteryRecord, ApiGetClaimRecord, RecordByDate } from "@/apis/claim"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import React, {
  FC,
  LegacyRef,
  ReactNode,
  useEffect,
  useRef,
  useState,
} from "react"
import useInfiniteScroll from "react-infinite-scroll-hook"
import { useIntl } from "react-intl"
import styles from "./records.less"
import { Icon } from "@/components/Icon"

const powerTypeMap: any = {
  "claim": "ic_qiandao",
  "claim_extra": "ic_qiandao",
  "lottery": "ic_choujiang",
  "invitation": "ic_yaoqing",
}

export default function RecordListPage() {
  const $t = get$t(useIntl())
  const [isClaimTab, setIsClaimTab] = useState(false)
  const [lotteryRecords, setLotteryRecords] = useState<RecordByDate[]>([])
  const [claimRecords, setClaimRecords] = useState<RecordByDate[]>([])
  const [lotteryLoading, setLotteryLoading] = useState(true)
  const [claimLoading, setClaimLoading] = useState(true)
  const [hasMoreLottery, setHasMoreLottery] = useState(true)
  const [hasMoreClaim, setHasMoreClaim] = useState(true)
  const [lotteryPage, setLotteryPage] = useState(1)
  const [claimPage, setClaimPage] = useState(1)

  const prevLotteryPageRef = useRef<number>()
  const prevClaimPageRef = useRef<number>()

  const fetchList = (isLottery?: boolean, p: number = 1) => {
    if (isLottery) {
      setLotteryLoading(true)
      return ApiGetLotteryRecord(p, lotteryRecords).then(
        ({ hasMore, list }) => {
          setLotteryLoading(false)
          setHasMoreLottery(hasMore)
          setLotteryRecords((prev) => (list.length ? list : prev))
        },
      )
    }

    setClaimLoading(true)
    ApiGetClaimRecord(p, claimRecords).then(({ hasMore, list }) => {
      setHasMoreClaim(hasMore)
      setClaimRecords((prev) => (list.length ? list : prev))
      setClaimLoading(false)
    })
  }

  const handlePageChange = () => {
    if (isClaimTab) {
      setClaimPage((prev) => prev + 1)
      return
    }

    setLotteryPage((prev) => prev + 1)
  }

  const [lotterySentryRef, lotteryScrollCtx] = useInfiniteScroll({
    loading: lotteryLoading,
    hasNextPage: hasMoreLottery,
    onLoadMore: handlePageChange,
    rootMargin: "0px 0px 200px 0px",
  })

  const [claimSentryRef, claimScrollCtx] = useInfiniteScroll({
    loading: claimLoading,
    hasNextPage: hasMoreClaim,
    onLoadMore: handlePageChange,
    rootMargin: "0px 0px 200px 0px",
  })

  useEffect(() => {
    if (isClaimTab && claimPage && claimPage != prevClaimPageRef.current) {
      prevClaimPageRef.current = claimPage
      fetchList(false, claimPage)
      return
    }

    if (!isClaimTab && lotteryPage !== prevLotteryPageRef.current) {
      prevLotteryPageRef.current = lotteryPage
      fetchList(true, lotteryPage)
    }
  }, [isClaimTab, lotteryPage, claimPage])

  useEffect(() => {
    document.body.classList.add(styles.disablescroll)
    return () => {
      document.body.classList.remove(styles.disablescroll)
    }
  }, [])

  const renderList = <T extends HTMLDivElement>(
    data: RecordByDate[],
    hasMore: boolean = true,
    ref: LegacyRef<T>,
    sentryRef: LegacyRef<T>,
  ) => (
    <div ref={ref} className={styles.scrollableContainer}>
      {data.map(([date, list]) => (
        <ul key={date} className={styles.records}>
          <li key={date} className={styles.date}>
            {date}
          </li>
          {list.map((r, idx) => (
            <li key={idx} className={styles.record}>
              <div className={styles.recordLeft}>
                <div className={styles.logo}>
                  {r.icon_url ? <img src={r.icon_url} />
                    : <Icon i={powerTypeMap[r.power_type!]} />}
                </div>
                <span className={styles.name}>{$t("claim.records." + (r.power_type ? "power_" + r.power_type : "lottery"))}</span>
              </div>
              <div className={styles.recordRight}>
                <span className={Number(r.amount) < 0 ? styles.negative : styles.plus}>
                  {Number(r.amount) < 0 ? "" : "+"}
                  {r.amount}
                </span>
                <span className={styles.desc}>
                  {r.symbol ? r.symbol : $t("claim.energy.title")}
                </span>
              </div>
            </li>
          ))}
        </ul>
      ))}
      {hasMore && (
        <div ref={sentryRef}>
          <Icon i="ic_load" className={styles.loading} />
        </div>
      )}
    </div>
  )

  const lotteryList = renderList(
    lotteryRecords,
    hasMoreLottery,
    lotteryScrollCtx.rootRef,
    lotterySentryRef,
  )

  const energyList = renderList(
    claimRecords,
    hasMoreClaim,
    claimScrollCtx.rootRef,
    claimSentryRef,
  )

  return (
    <>
      <BackHeader name={$t("claim.records.title")} />
      <div className={styles.page}>
        <TabSwitchBar
          activeRight={isClaimTab}
          onSwitch={setIsClaimTab}
          leftLabel={$t("claim.records.winning")}
          rightLabel={$t("claim.records.energy")}
        />
        <div
          className={`${styles.tabPanel} ${isClaimTab ? styles.switch : ""}`}
        >
          <div className={styles.content}>
            <div className={styles.left} id="scrollableLeft">
              {lotteryList}
            </div>
            <div className={styles.right} id="scrollableRight">
              {energyList}
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

interface TabSwitchBarProps {
  activeRight: boolean
  onSwitch(param: boolean): void
  leftLabel: ReactNode
  rightLabel: ReactNode
}

const TabSwitchBar: FC<TabSwitchBarProps> = ({
  activeRight,
  onSwitch,
  leftLabel,
  rightLabel,
}) => {
  return (
    <div className={styles.tabbar}>
      <div
        className={`${styles.switch} ${activeRight ? styles.active : ""}`}
        onClick={() => {
          onSwitch(!activeRight)
        }}
      >
        <div className={styles.item}>{leftLabel}</div>
        <div className={styles.item}>{rightLabel}</div>
      </div>
    </div>
  )
}
