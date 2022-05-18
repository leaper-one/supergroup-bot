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

export default function Page(props: any) {
  const $t = get$t(useIntl());
  const [live, setLive] = useState<ILive>();
  const [list, setList] = useState<IReplay[]>();
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

  const initPage = (id: string) => {
    ApiGetLiveReplayList(id).then(setList);
    ApiGetLiveInfo(id).then(setLive);
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

  return (
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
      <div className={styles.content}>
        <img src={live?.img_url} alt="" className={styles.main_image} />
        {list && list.map((item) => msgItem(item))}
      </div>
      {audioList && audioList.length >= 2 && (
        <img onClick={() => playlist(audioList.map((item) => liveReplayPrefixURL + item.data))} className={styles.backPlay} src={require('@/assets/img/back_play.png')} alt="" />
      )}
    </div>
  );
}

function msgItem(msg: IReplay) {
  switch (msg.category) {
    case 'PLAIN_TEXT':
      return <div className={styles.text} dangerouslySetInnerHTML={{ __html: handleBroadcast(msg.data) }} />;
    case 'PLAIN_AUDIO':
      return (
        <div className={styles.audio}>
          <audio src={liveReplayPrefixURL + msg.data} controls />
        </div>
      );
    case 'PLAIN_IMAGE':
      return (
        <div className={styles.image}>
          <img src={liveReplayPrefixURL + msg.data} alt="" />
        </div>
      );
    case 'PLAIN_VIDEO':
      return (
        <div className={styles.video}>
          <video src={liveReplayPrefixURL + msg.data} controls />
        </div>
      );
  }
}
