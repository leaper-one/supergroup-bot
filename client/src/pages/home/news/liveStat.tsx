import React, { useEffect, useState } from 'react';
import { BackHeader } from "@/components/BackHeader";
import { get$t } from "@/locales/tools";
import { history, useIntl } from "umi";
import styles from './liveStat.less'
import { ApiGetLiveStat } from "@/apis/live";
import { BigNumber } from "bignumber.js";
import { $get } from "@/stores/localStorage";

export default function Page() {
  const $t = get$t(useIntl())
  const [stat, setStat] = useState<any>()

  useEffect(() => {
    const { live_id } = $get("active_live")
    if (!live_id) return history.replace("/news")
    ApiGetLiveStat(live_id).then(stat => {
      stat.duration = (Number(new Date(stat.end_at)) - Number(new Date(stat.start_at))) / 1000 / 60 | 0
      setStat(stat)
    })
  }, [])

  return <div>
    <BackHeader name={$t("news.stat.title")}/>
    <ul className={styles.list}>
      {["read_count", "deliver_count", "duration", "user_count", "msg_count"].map(item =>
        <li
          key={item}
          className={styles.item}>
          <p>{$t(`news.stat.${item}`)}</p>
          <h4>{stat && new BigNumber(stat[item]).toFormat(0)}</h4>
        </li>)}
    </ul>
  </div>
}
