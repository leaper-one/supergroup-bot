import React from 'react';
import styles from './index.less';
import { ToastSuccess } from '@/components/Sub';

interface Props {
  icon_url: string;
}

let initState = 0;
export const CodeURLIcon = (props: Props) => (
  <div className={styles.icon}>
    <img src="https://taskwall.zeromesh.net/group-manager/groupCircle.svg" alt="" />
    <img src={props.icon_url} alt="" />
    <img
      onClick={() => {
        initState++;
        if (initState >= 10) {
          window.localStorage.clear();
          ToastSuccess('请重新进入');
        }
      }}
      src={require('@/assets/img/waving_hand.png')}
      alt=""
    />
  </div>
);
