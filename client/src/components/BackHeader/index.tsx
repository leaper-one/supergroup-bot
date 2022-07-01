import React from 'react';
import styles from './index.less';
import { history } from 'umi';
import { Icon } from '../Icon';

interface Props {
  name: string;
  noBack?: Boolean;
  action?: JSX.Element | undefined;
  onClick?: () => void | undefined;
  isWhite?: boolean;
  backHome?: boolean;
  className?: string;
}

export const BackHeader = (props: Props) => {
  let name = props.name;
  if (name.includes('<br/>')) name = name.replace('<br/>', '');
  return (
    <div className={`${styles.header} ${props.className} ${props.isWhite && styles.white}`}>
      {!props.noBack && <Icon i="ic_return" className={styles.back} onClick={() => (props.backHome ? history.push('/') : history.go(-1))} />}
      <span onClick={props.onClick}>{name}</span>
      {props.action && <div className={styles.action}>{props.action}</div>}
    </div>
  );
};
