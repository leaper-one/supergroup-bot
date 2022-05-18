import React from 'react';
import styles from './liveDesc.less';
import { GlobalData } from '@/stores/store';
import { ILive } from '@/apis/live';
import { get$t } from '@/locales/tools';
import { useIntl } from '@@/plugin-locale/localeExports';
import { BackHeader } from '@/components/BackHeader';
import { $get } from '@/stores/localStorage';

export default function Page() {
  const live: ILive = $get('active_live');
  const $t = get$t(useIntl());
  return (
    <div className={`safe-view ${styles.container}`}>
      <BackHeader name={$t('news.livePreview')} />
      <img className={styles.img} src={live.img_url} alt="" />
      <div className={styles.content}>
        <h4>{live.title}</h4>
        <p>{live.description}</p>
      </div>
    </div>
  );
}
