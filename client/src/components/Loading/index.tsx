import React from 'react';
import styles from './index.less';
import { get$t } from '@/locales/tools';
import { useIntl } from 'umi';

interface IProps {
  content?: string;
  noCancel?: boolean;
  cancel?: () => void;
}

export const Loading = (prop: IProps = {}) => {
  let { content, noCancel, cancel } = prop;
  const $t = get$t(useIntl());
  content = content || $t('modal.loading');

  return (
    <div className={styles.modal}>
      <div className={styles.mask} />
      <div className={styles.content}>
        <img src={require('@/assets/img/modalLoading.png')} alt="" />
        <p>{content}</p>
        {!noCancel && <button onClick={cancel}>{$t('action.cancel')}</button>}
      </div>
    </div>
  );
};

interface IFullLoadingProps {
  mask?: boolean;
  opacity?: boolean;
}

export const FullLoading = (props: IFullLoadingProps) => (
  <div className={styles.full_modal}>
    {props.mask && <div style={{ backgroundColor: props.opacity ? 'rgba(0,0,0,0.3)' : '#fff' }} className={styles.full_modal_mask} />}
    <div className={styles.spinner}>
      <div className={styles.rect1} />
      <div className={styles.rect2} />
      <div className={styles.rect3} />
      <div className={styles.rect4} />
      <div className={styles.rect5} />
    </div>
  </div>
);
