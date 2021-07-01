import React, { useEffect, useState } from 'react';
import styles from './liveReplay.less';
import { BackHeader } from "@/components/BackHeader";
import { get$t } from "@/locales/tools";
import { history, useIntl } from "umi";
import { ApiGetLiveReplayList, ILive, IReplay } from "@/apis/live";
import { handleBroadcast } from "@/pages/home/news/index";
import { liveReplayPrefixURL } from "@/apis/http";
import { $get } from "@/stores/localStorage";

export default function Page() {
  const $t = get$t(useIntl())
  const [live, setLive] = useState<ILive>($get("active_live"))
  const [list, setList] = useState<IReplay[]>()

  useEffect(() => {
    ApiGetLiveReplayList(live.live_id!).then(setList)
  }, [])


  return (
    <div className={styles.container}>
      <BackHeader
        name={$t("news.liveReplay.title")}
        action={<i
          className={`iconfont iconbar-chart-2 ${styles.stat}`}
          onClick={() => history.push(`/news/liveStat`)}
        />}
      />
      <div className={styles.content}>
        <img src={live.img_url} alt="" className={styles.main_image}/>
        {list && list.map(item => msgItem(item))}
      </div>
    </div>
  );
}

function msgItem(msg: IReplay) {
  switch (msg.category) {
    case "PLAIN_TEXT":
      return <div className={styles.text} dangerouslySetInnerHTML={{ __html: handleBroadcast(msg.data) }}/>
    case "PLAIN_AUDIO":
      return <div className={styles.audio}>
        <audio src={liveReplayPrefixURL + msg.data} controls/>
      </div>
    case "PLAIN_IMAGE":
      return <div className={styles.image}>
        <img src={liveReplayPrefixURL + msg.data} alt=""/>
      </div>
    case "PLAIN_VIDEO":
      return <div className={styles.video}>
        <video src={liveReplayPrefixURL + msg.data} controls/>
      </div>
  }
}

