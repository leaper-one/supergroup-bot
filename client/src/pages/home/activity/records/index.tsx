import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import React, { useCallback, useEffect, useState } from "react"
import { useIntl } from "react-intl"
import { useParams } from "umi"
import { ApiGetGuessRecord, ApiGetGuessPageData } from "@/apis/guess"
import { GuessResult } from "@/types"

import styles from "./records.less"
import { changeTheme } from "@/assets/ts/tools"
import { FullLoading } from "@/components/Loading"

const isValidResult = (r?: GuessResult) =>
  r != undefined && r >= GuessResult.Pending

interface PlayDays {
  inrow: number // 连续的参于次数
  total: number // 总共参与次数
}

const calcPlayedDays = (min: number, data: GuessRecord[] = []) =>
  data.reduce(
    (acc: PlayDays, cur, idx, arr) => {
      if (!isValidResult(cur.result))
        return {
          inrow: acc.inrow >= min ? acc.inrow : 0,
          total: acc.total,
        }

      return {
        inrow: isValidResult(arr[Math.max(idx - 1, 0)].result) // 前一天没参加
          ? acc.inrow + 1
          : acc.inrow,
        total: acc.total + 1,
      }
    },
    { inrow: 0, total: 0 },
  )

interface GuessRecord {
  date: string
  result?: GuessResult
}

export default function GuessRecordsPage() {
  const t = get$t(useIntl())
  const { id } = useParams<{ id: string }>()
  const [records, setRecords] = useState<GuessRecord[]>()
  const [coin, setCoin] = useState<string>()
  const [isLoaded, setIsLoaded] = useState(false)

  const fetchPageDate = useCallback(() => {
    Promise.all([ApiGetGuessPageData(id), ApiGetGuessRecord(id)]).then(
      ([g, r]) => {
        setCoin(g.symbol)

        const millisecondsPerDay = 24 * 60 * 60 * 1000
        const start = new Date(g.start_at).getTime()
        const end = new Date(g.end_at).getTime()
        const days = (end - start) / millisecondsPerDay + 1

        const now = new Date()

        const tempRecords = Array.from(Array(days)).map((x, idx) => {
          let d = new Date(g.start_at)
          d.setDate(d.getDate() + idx)

          const date = d.toISOString().slice(0, 10)
          const record = r.find((x) => x.date === date)

          let result = record?.result

          if (
            d.getFullYear() >= now.getFullYear() &&
            d.getMonth() >= now.getMonth() && // 值是从0开始的
            d.getDate() > now.getDate() &&
            !result
          ) {
            result = GuessResult.NotStart
          }

          return {
            date,
            result,
          }
        })
        setIsLoaded(true)
        setRecords(tempRecords)
      },
    )
  }, [id])

  useEffect(() => {
    changeTheme("#da1f27")
    fetchPageDate()
    return () => {
      changeTheme("#fff")
    }
  }, [])

  const guessResult = (result?: GuessResult) => {
    switch (result) {
      case undefined:
        return t("guess.records.notplay")
      case GuessResult.NotStart:
        return t("guess.records.notstart")
      case GuessResult.Pending:
        return t("guess.records.pending")
      default:
        return t(`guess.records.${result === GuessResult.Win ? "win" : "lose"}`)
    }
  }

  const { total, inrow } = calcPlayedDays(3, records)

  return (
    <div className={styles.container}>
      <BackHeader isWhite name={t("guess.records.name")} />
      <h1 className={styles.header}>{t("guess.records.history", { coin })}</h1>
      <div className={styles.content}>
        <div className={styles.tip}>
          <p className={styles.tipContent}>
            {t(
              inrow > 1
                ? "guess.records.consecutiveplay"
                : "guess.records.play",
            )}
            <span className={styles.day}>
              &nbsp;{inrow > 1 ? inrow : total}
              &nbsp;
            </span>
            {t("guess.records.day")}
            {inrow >= 3 && t("guess.records.vip")}
            {t("guess.records.result")}
          </p>
        </div>
        <div className={styles.vs_title}>
          <span>{t("guess.records.win")}</span>
          <span>{t("guess.records.lose")}</span>
        </div>
        <div className={styles.vs_detail}>
          <span className={styles.num}>
            {records?.filter((x) => x.result === GuessResult.Win).length || 0}
          </span>
          <span className={styles.flag}>:</span>
          <span className={styles.num}>
            {records?.filter((x) => x.result === GuessResult.lose).length || 0}
          </span>
        </div>
        <ul className={styles.list}>
          <li className={styles.title}>
            <span>{t("guess.records.date")}</span>
            <span>{t("guess.records.end")}</span>
          </li>
          {records &&
            records.map((x) => (
              <li className={styles.item} key={x.date}>
                <span>{x.date.replace(/-/g, "/")}</span>
                <span>{guessResult(x.result)}</span>
              </li>
            ))}
        </ul>
      </div>
      {!isLoaded && <FullLoading mask opacity />}
    </div>
  )
}
