import React from 'react';
import styles from './join.less';
import { Button } from '@/components/Sub';
import { Icon, AppIcons } from '../Icon';

export interface JoinModalProps {
  title?: string;
  titleDesc?: string;
  desc?: string;
  button?: string;
  tips?: string;
  tipsStyle?: string;
  tipsAction?: () => void;
  buttonAction?: () => void;
  isAirdrop?: boolean;
  icon?: AppIcons;
  icon_url?: string;
  buttonStyle?: string;
  descStyle?: string;
  disabled?: boolean;
  loading?: boolean;
  content?: JSX.Element;
}

interface Props {
  modalProp: JoinModalProps | undefined;
}

export const JoinModal = (props: Props) => {
  if (!props.modalProp) return <></>;
  const { title, desc, buttonAction, button, tips, tipsStyle, tipsAction, isAirdrop, icon, icon_url, buttonStyle, descStyle, disabled, loading, content, titleDesc } = props.modalProp;

  return (
    <div className={`${styles.modal} ${isAirdrop ? styles.airdrop : ''}`}>
      {isAirdrop && <img className={styles.bg} src={require('@/assets/img/svg/modal_bg.svg')} alt="" />}
      {icon ? <Icon i={icon} className={styles.icon} /> : <img className={styles.icon} src={icon_url} alt="" />}
      <h4 className={styles.title}>{title}</h4>
      {titleDesc && <p className={styles.titleDesc}>{titleDesc}</p>}
      {content ? content : <p className={`${styles.desc} ${descStyle}`}>{desc}</p>}
      <Button className={styles.button} loading={loading} type={buttonStyle} onClick={buttonAction} disabled={disabled}>
        {' '}
        {button}
      </Button>
      {tips && (
        <span className={`${styles.tips} ${tipsStyle ? styles[tipsStyle] : ''}`} onClick={tipsAction}>
          {tips}
        </span>
      )}
    </div>
  );
};
