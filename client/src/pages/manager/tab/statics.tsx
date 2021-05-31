import React, { useEffect, useState } from "react"
import styles from "./statics.less"
import { ApiGetGroupStat } from "@/apis/group"
import { useIntl } from "umi"
import { get$t } from "@/locales/tools"

const staticKey = ["members", "broadcasts", "conversations"]

export default () => {
  const [stat, setStat] = useState<any>({})

  const $t = get$t(useIntl())

  useEffect(() => {
    initPage().then()
  }, [])
  const initPage = async () => {
    const stat = await ApiGetGroupStat()
    setStat(stat)
  }

  return (
    <div className={styles.container}>
      <ul className={styles.data}>
        {staticKey.map((item, index) => (
          <li key={index}>
            <span className={styles.title}>{$t("manager." + item)}</span>
            <span className={styles.amount}>{stat[item] || 0}</span>
          </li>
        ))}
      </ul>
      <div className={styles.chart}>
        <h3>{$t("manager.list")}</h3>
        <span>{(stat.list && stat.list[0]?.count) || 0}</span>
      </div>
    </div>
  )
}
