import { BackHeader } from "@/components/BackHeader"
import { Radio } from "@/components/Radio"
import { get$t } from "@/locales/tools"
import React, { FC, useCallback, useEffect } from "react"
import { useIntl } from "react-intl"
import { useParams } from "umi"
import { ApiGetGuessRecord } from "@/apis/guess"
import styles from "./records.less"

export default function GuessRecordsPage() {
  const t = get$t(useIntl())
  const { id } = useParams<{ id: string }>()

  const fetchPageDate = useCallback(() => {
    ApiGetGuessRecord(id).then((x) => console.log(x))
  }, [id])

  useEffect(() => {
    fetchPageDate()
  }, [])

  return (
    <div className={styles.container}>
      <BackHeader isWhite name={t("guess.records.name")} />
      <h1 className={styles.header}>{t("guess.records.history")}</h1>
      <div className={styles.content}>
        <p>
          {t("guess.records.play")} <span className={styles.day}>3</span>
          {t("guess.records.playinfo")}
        </p>
        <div className={styles.vs_title}>
          <span>{t("guess.records.win")}</span>
          <span>{t("guess.records.lose")}</span>
        </div>
        <div className={styles.vs_detail}>
          <span className={styles.num}>2</span>
          <span className={styles.flag}>:</span>
          <span className={styles.num}>1</span>
        </div>
        <ul className={styles.list}>
          <li className={styles.title}>
            <span>{t("guess.records.date")}</span>
            <span>{t("guess.records.end")}</span>
          </li>
          <li className={styles.item}>
            <span>2021/09/10</span>
            <span>赢</span>
          </li>
          <li className={styles.item}>
            <span>2021/09/10</span>
            <span>赢</span>
          </li>
          <li className={styles.item}>
            <span>2021/09/10</span>
            <span>赢</span>
          </li>
          <li className={styles.item}>
            <span>2021/09/10</span>
            <span>赢</span>
          </li>
        </ul>
      </div>
    </div>
  )
}
