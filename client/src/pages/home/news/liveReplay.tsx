import React, { useEffect, useState } from 'react';
import styles from './liveReplay.less';
import { BackHeader } from '@/components/BackHeader';
import { get$t } from '@/locales/tools';
import { history, useIntl } from 'umi';
import { ApiGetLiveInfo, ApiGetLiveReplayList, ILive, IReplay } from '@/apis/live';
import { handleBroadcast } from '@/pages/home/news/index';
import { liveReplayPrefixURL } from '@/apis/http';
import { base64Encode, playlist } from '@/assets/ts/tools';
import { ApiGetGroup } from '@/apis/group';
import { $get, $set } from '@/stores/localStorage';
import { FullLoading } from '@/components/Loading';

let isStart = false;
export default function Page(props: any) {
  const $t = get$t(useIntl());
  const [live, setLive] = useState<ILive>();
  const [list, setList] = useState<IReplay[]>();
  const [showList, setShowList] = useState<IReplay[]>([]);
  const [loading, setLoading] = useState(false);
  const [loaded, setLoaded] = useState(false);
  const user = $get('_user');

  useEffect(() => {
    const id = props?.match?.params?.id;
    if (!id) return history.push('/');
    if (!$get('group')) {
      ApiGetGroup().then((group) => {
        $set('group', group);
        initPage(id);
      });
    } else {
      initPage(id);
    }
  }, []);

  const initPage = async (id: string) => {
    setLoading(true);
    const [list, live] = await Promise.all([ApiGetLiveReplayList(id), ApiGetLiveInfo(id)]);
    setLive(live);
    setList(list);
    setShowList(list.slice(0, 20));
    setLoaded(true);
    setLoading(false);
  };
  const handleClickShared = () => {
    let schema = `mixin://send?category=app_card&data=`;
    const group = $get('group');
    schema += base64Encode({
      app_id: group.client_id,
      icon_url: group.icon_url,
      title: live?.title,
      description: live?.description,
      action: location.href,
    });
    window.location.href = schema;
  };

  const audioList = list && list.filter((item) => item.category === 'PLAIN_AUDIO');

  const action = {
    onAudioPlay: (idx: number) => {
      for (let i = 0; i < list!.length; i++) {
        if (i !== idx && list![i].category === 'PLAIN_AUDIO') {
          const audioDOM = document.getElementById(`replay-${i}`) as HTMLAudioElement;
          if (!audioDOM) continue;
          if (!audioDOM.paused) audioDOM.pause();
        }
      }
      const audioDOM = document.getElementById(`replay-${idx}`) as HTMLAudioElement;
      audioDOM.play();
    },
    onAudioEnded: (idx: number) => {
      let nextIdx = -1;
      for (let i = idx + 1; i < list!.length; i++) {
        if (list![i].category === 'PLAIN_AUDIO') {
          nextIdx = i;
          break;
        }
      }
      if (nextIdx !== -1) {
        const audioDom = document.getElementById(`replay-${nextIdx}`) as HTMLAudioElement;
        audioDom.play();
      }
    },
  };

  return (
    <>
      <div className={styles.container}>
        <BackHeader
          name={$t('news.liveReplay.title')}
          action={
            <>
              {user && user.status === 9 && <i className={`iconfont iconbar-chart-2 ${styles.stat}`} onClick={() => history.push(`/news/liveStat`)} />}
              <i className={`iconfont iconic_share ${styles.share}`} onClick={() => handleClickShared()} />
            </>
          }
        />
        <div
          className={styles.content}
          onScroll={(e: any) => {
            const { offsetHeight, scrollHeight, scrollTop } = e.target;
            if (offsetHeight + scrollTop + 100 >= scrollHeight && showList.length < list!.length) {
              setShowList(list!.slice(0, showList.length + 20));
            }
          }}
        >
          <img src={live?.img_url} alt="" className={styles.main_image} />
          {showList && showList.map((item, idx) => msgItem(item, idx, action))}
        </div>
        {audioList && audioList.length >= 2 && loaded && (
          <img onClick={() => playlist(audioList.map((item) => liveReplayPrefixURL + item.data))} className={styles.backPlay} src={require('@/assets/img/back_play.png')} alt="" />
        )}
      </div>
      {loading && <FullLoading mask opacity />}
    </>
  );
}

function msgItem(msg: IReplay, idx: number, action: any) {
  switch (msg.category) {
    case 'PLAIN_TEXT':
      return <div key={idx} className={styles.text} dangerouslySetInnerHTML={{ __html: handleBroadcast(msg.data) }} />;
    case 'PLAIN_AUDIO':
      return (
        <div key={idx} className={styles.audio}>
          <audio id={'replay-' + idx} src={liveReplayPrefixURL + msg.data} controls onPlay={() => action.onAudioPlay(idx)} onEnded={() => action.onAudioEnded(idx)} />
        </div>
      );
    case 'PLAIN_IMAGE':
      return (
        <div key={idx} className={styles.image}>
          <img src={liveReplayPrefixURL + msg.data} alt="" />
        </div>
      );
    case 'PLAIN_VIDEO':
      return (
        <div key={idx} className={styles.video}>
          <video src={liveReplayPrefixURL + msg.data} controls />
        </div>
      );
  }
}
