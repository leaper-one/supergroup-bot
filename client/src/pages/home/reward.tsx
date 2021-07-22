import { useIntl } from '@/.umi/plugin-locale/localeExports';
import { BackHeader } from '@/components/BackHeader';
import { CoinSelect } from '@/components/CoinSelect/tt';
import { get$t } from '@/locales/tools';
import React from 'react';
import styles from './reward.less';

export default function Page() {
  const $t = get$t(useIntl())
  return (
    <div className={styles.container}>
      <BackHeader name={$t('reward.title')} />
    </div>
  );
}
