import React, { FC, useState } from 'react';
import styles from './Energy.less';
import { Progress } from 'antd';
import { get$t } from '@/locales/tools';
import { useIntl, history } from 'umi';
import { ClaimData } from '@/apis/claim';
import { IGroup } from '@/apis/group';
import { $get } from '@/stores/localStorage';
import { Button } from '@/components/Sub';

interface EnergyProps {
  claim?: ClaimData;
  onExchangeClick(): void;
  onCheckInClick(): Promise<void>;
  onModalOpen(group: IGroup): void;
  onVoucherClick(): void;
}

export const Energy: FC<EnergyProps> = ({ claim, onExchangeClick, onCheckInClick, onModalOpen, onVoucherClick }) => {
  const $t = get$t(useIntl());
  console.log(claim)
  console.log(claim?.is_claim)
  const [loading, setLoading] = useState(false);
  const { power, count = 0, invite_count = 0, is_claim = false, double_claim_list = [] } = claim || {};
  const process = Number(power?.balance) || 0;
  const group = $get('group');
  const isDouble = double_claim_list.find((v) => v.client_id === group.client_id);

  return (
    <div className={styles.container}>
      <div className={styles.main}>
        <div className={styles.title}>
          <div className={styles.line_left} />
          <span className={styles.content}>{$t('claim.energy.title')}</span>
          <div className={styles.line_right} />
        </div>
        <div className={styles.tip}>{$t('claim.tag')}</div>
        <div className={styles.progress}>
          <Progress
            percent={process}
            showInfo={false}
            strokeWidth={14}
            strokeColor={{
              '0%': '#FC602E',
              '83.25%': '#F68934',
              '95.86%': '#FDDE8B',
            }}
          />
          <p className={styles.info}>{$t('claim.energy.describe')}</p>
        </div>
        <Button
          loading={loading}
          disabled={process < 100}
          onClick={async () => {
            setLoading(true);
            await onExchangeClick();
            setLoading(false);
          }}
          className={`${styles.exchange} ${process >= 100 ? styles.active : styles.default}`}
        >
          {$t('claim.energy.exchange')}
        </Button>
        <ul className={styles.job_list}>
          <TaskItem
            icon="iconic_qiandao"
            title={$t('claim.checkin.describe')}
            btn={$t(`claim.checkin.${is_claim ? 'checked' : 'label'}`)}
            disabled={is_claim}
            tips={$t('claim.checkin.count', { count })}
            action={onCheckInClick!}
            isDouble={!!isDouble}
            $t={$t}
          />
          {double_claim_list
            .filter((v) => v.client_id != group.client_id)
            .map((client) => (
              <TaskItem key={client.client_id} icon="iconxiaoxiquanxian" title={client.welcome!} btn={$t('action.open')} action={() => onModalOpen(client)} />
            ))}
          <TaskItem
            icon="iconic_yaoqing"
            title={$t('invite.claim.title')}
            btn={$t(`invite.claim.btn`)}
            tips={$t('invite.claim.count', { count: invite_count })}
            action={() => history.push('/invite')}
          />
          <TaskItem icon="iconic_duihuan" title={$t('claim.voucher.label')} btn={$t('claim.voucher.btn')} action={() => onVoucherClick()} />
        </ul>
      </div>
    </div>
  );
};

interface TaskItemProps {
  icon: string;
  title: string;
  btn: string;
  tips?: string;
  action: () => void;
  isDouble?: boolean;
  disabled?: boolean;
  $t?: (key: string) => string;
}
const TaskItem: FC<TaskItemProps> = ({ icon, title, btn, tips, action, isDouble, disabled, $t }) => {
  const [loading, setLoading] = useState(false);

  console.log('btn status:::', disabled);
  return (
    <li className={styles.job}>
      {isDouble && <div className={styles.double}>{$t!('claim.energy.title')} X2</div>}
      <div className={styles.icon}>
        <i className={`iconfont ${icon} ${styles.jobIcon}`} />
      </div>
      <p className={styles.info}>{title}</p>
      <div className={styles.jobBtn}>
        <Button
          className={styles.btn}
          loading={loading}
          onClick={async () => {
            setLoading(true);
            await action();
            setLoading(false);
          }}
          disabled={disabled}
        >
          {btn}
        </Button>
        {tips && <span className={styles.caption}>{tips}</span>}
      </div>
    </li>
  );
};
