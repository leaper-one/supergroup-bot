import React, { useEffect, useState } from 'react';
import { BackHeader } from "@/components/BackHeader";
import { get$t } from "@/locales/tools";
import { history, useIntl } from "umi";
import styles from './liveStat.less'
import { ApiGetLiveStat } from "@/apis/live";
import moment from "moment";
import { BigNumber } from "bignumber.js";
import { GlobalData } from "@/stores/store";

export default function Page() {
  const $t = get$t(useIntl())
  const [stat, setStat] = useState<any>()

  useEffect(() => {
    const { live_id } = GlobalData.live || {}
    if (!live_id) return history.replace("/news")
    ApiGetLiveStat(live_id).then(stat => {
      stat.duration = moment(stat.end_at).endOf(stat.start_at).minutes()
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
