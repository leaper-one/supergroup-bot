import { BackHeader } from "@/components/BackHeader"
import { Radio } from "@/components/Radio"
import { get$t } from "@/locales/tools"
import React, { FC, useCallback, useEffect, useState } from "react"
import { useIntl } from "react-intl"
import { useParams } from "umi"
import { ApiGetGuessRecord, ApiGetGuessPageData } from "@/apis/guess"
import { GuessRecord, GuessResult } from "@/types"

import styles from "./records.less"

interface GuessG {
  date: string
  result?: GuessResult
}

export default function GuessRecordsPage() {
  const t = get$t(useIntl())
  const { id } = useParams<{ id: string }>()
  const [records, setRecords] = useState<GuessG[]>()
  const [startAt, setStartAt] = useState<string>()
  const [endAt, setEndAt] = useState<string>()
  const [coin, setCoin] = useState<string>()

  const fetchPageDate = useCallback(() => {
    Promise.all([ApiGetGuessPageData(id), ApiGetGuessRecord(id)]).then(
      ([g, r]) => {
        setCoin(g.symbol)

        const start = new Date(g.start_at).getTime()
        const end = new Date(g.end_at).getTime()
        const days = (end - start) / (1000 * 3600 * 24)

        const tempRecords = Array.from(Array(days)).map((x, idx) => {
          let d = new Date(g.start_at)
          d.setDate(d.getDate() + (idx + 1))
          const date = d.toISOString().slice(0, 10)
          const record = r.find((x) => x.date === date)

          return {
            date,
            result: record?.result,
          }
        })

        setRecords(tempRecords)
      },
    )
  }, [id])

  useEffect(() => {
    fetchPageDate()
  }, [])

  const guessResult = (result?: GuessResult) => {
    if (result === undefined) {
      return t("guess.records.notplay")
    }
    if (result === GuessResult.Pending) {
      return t("guess.records.pending")
    }

    return t(`guess.records.${result === GuessResult.Win ? "win" : "lose"}`)
  }

  return (
    <div className={styles.container}>
      <BackHeader isWhite name={t("guess.records.name")} />
      <h1 className={styles.header}>{t("guess.records.history", { coin })}</h1>
      <div className={styles.content}>
        <p className={styles.tip}>
          {t("guess.records.play")}
          <span className={styles.day}>
            &nbsp;{records?.filter((x) => x.result !== undefined).length || 0}
            &nbsp;
          </span>
          {t("guess.records.playinfo")}
        </p>
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
    </div>
  )
}
