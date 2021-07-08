import React, { useEffect } from 'react';
import styles from './stat.less';
import { BackHeader } from "@/components/BackHeader";
import { get$t } from "@/locales/tools";
import { useIntl } from "@@/plugin-locale/localeExports";
import { ApiGetGroupStat } from "@/apis/group";
import { getOptions, getStatisticsDate } from './statData'
import * as echarts from 'echarts'

interface IStat {
  totalUser: string
  highUser: string
  weekUser: string
  weekActiveUser: string
  totalMessage: string
  weekMessage: string
}

const statList = ["totalUser", "highUser", "weekUser", "weekActiveUser", "totalMessage", "weekMessage"]

export default function Page() {
  const $t = get$t(useIntl())
  useEffect(() => {
    ApiGetGroupStat().then(res => {
      const t = getStatisticsDate(res)
      echarts.init(document.getElementById("user")!).setOption(getOptions(t[0]))
      echarts.init(document.getElementById("message")!).setOption(getOptions(t[1]))
    })
  }, [])

  return (
    <div className={styles.container}>
      <BackHeader name={$t('stat.title')}/>
      <div className={styles.list}>
        {statList.map(item => (<div key={item} className={styles.item}>
          <p>{$t(`stat.${item}`)}</p>
          <h3>1,678</h3>
        </div>))}
      </div>
      <div id="user" className={styles.charts}/>
      <div id="message" className={styles.charts}/>
    </div>
  );
}
