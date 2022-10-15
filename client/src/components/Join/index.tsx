import React, { useEffect } from 'react';
import styles from './index.less';
import { CodeURLIcon } from '@/components/CodeURL/icon';
import { IGroup } from '@/apis/group';
import { Button } from '@/components/Sub';

export interface IJoin {
  groupInfo: IGroup | undefined;
  button: string;
  buttonAction: () => void;
  tips?: string;
  tipsStyle?: string;
  tipsAction?: () => void;
  disabled?: boolean;
  loading?: boolean;
}

interface Props {
  props: IJoin | undefined;
}

let onceSubmit = false;
export const Join = (props: Props) => {
  if (!props.props || !props.props.groupInfo) {
    return <></>;
  }
  const { groupInfo, button, buttonAction, tips, disabled, tipsAction, tipsStyle, loading } = props.props;
  useEffect(() => {
    return () => {
      onceSubmit = false;
    };
  }, []);
  return (
    <div className={`${styles.container} safe-view`}>
      <header className={styles.header}>
        <img src={require('@/assets/img/join.png')} alt="" />
        <CodeURLIcon icon_url={groupInfo?.icon_url} />
      </header>
      <div className={styles.content}>
        <h3>{groupInfo?.name}</h3>
        <p className={styles.member}>{groupInfo?.total_people}</p>
        <p className={styles.desc}>{groupInfo?.description}</p>
        <Button
          loading={loading}
          disabled={disabled}
          className={styles.button}
          onClick={async () => {
            if (onceSubmit) return;
            onceSubmit = true;
            await buttonAction();
            onceSubmit = false;
          }}
        >
          {button}
        </Button>
        {tips && (
          <p className={styles.tips + ' ' + (tipsStyle ? tipsStyle : '')} onClick={tipsAction}>
            {tips}
          </p>
        )}
      </div>
    </div>
  );
};
