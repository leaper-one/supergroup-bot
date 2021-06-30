import React from 'react';
import styles from './liveReplay.less';
import { BackHeader } from "@/components/BackHeader";
import { get$t } from "@/locales/tools";
import { history, useIntl } from "umi";

export default function Page() {
  const $t = get$t(useIntl())


  return (
    <div className={styles.container}>
      <BackHeader
        name={$t("news.liveReplay.title")}
        action={<i
          className={`iconfont iconbar-chart-2 ${styles.stat}`}
          onClick={() => history.push(`/news/liveStat`)}
        />}
      />

    </div>
  );
}
