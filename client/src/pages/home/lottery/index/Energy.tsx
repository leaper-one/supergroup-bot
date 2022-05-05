import React, { FC } from "react"
import styles from "./Energy.less"
import { Progress } from "antd"
import { get$t } from "@/locales/tools"
import { useIntl, history } from "umi"
import { ClaimData } from '@/apis/claim'
import { IGroup } from '@/apis/group'
import { $get } from '@/stores/localStorage'

interface EnergyProps {
  claim?: ClaimData
  onExchangeClick?(): void
  onCheckinClick?(): void
  onModalOpen?(group: IGroup): void
}

export const Energy: FC<EnergyProps> = ({
  claim,
  onExchangeClick,
  onCheckinClick,
  onModalOpen,
}) => {
  const $t = get$t(useIntl())
  const { power, count = 0, invite_count = 0, is_claim = false, double_claim_list = [] } = claim || {}
  const process = Number(power?.balance) || 0
  const group = $get('group')
  const isDouble = double_claim_list.find(v => v.client_id === group.client_id)
  return (
    <div className={styles.container}>
      <div className={styles.main}>
        <div className={styles.title}>
          <div className={styles.line_left} />
          <span className={styles.content}>{$t("claim.energy.title")}</span>
          <div className={styles.line_right} />
        </div>
        <div className={styles.tip}>{$t("claim.tag")}</div>
        <div className={styles.progress}>
          <Progress
            percent={process}
            showInfo={false}
            strokeWidth={14}
            strokeColor={{
              "0%": "#FC602E",
              "83.25%": "#F68934",
              "95.86%": "#FDDE8B",
            }}
          />
          <p className={styles.info}>{$t("claim.energy.describe")}</p>
        </div>
        <button
          disabled={process < 100}
          onClick={onExchangeClick}
          className={`${styles.exchange} ${process >= 100 ? styles.active : styles.default}`}
        >
          {$t("claim.energy.exchange")}
        </button>
        <ul className={styles.job_list}>
          <TaskItem
            icon='iconic_qiandao'
            title={$t("claim.energy.checkin.describe")}
            btn={$t(`claim.energy.checkin.${is_claim ? 'checked' : 'label'}`)}
            disabled={is_claim}
            tips={$t("claim.energy.checkin.count", { count: count })}
            action={onCheckinClick!}
            isDouble={!!isDouble}
            $t={$t}
          />
          {double_claim_list
            .filter(v => v.client_id != group.client_id)
            .map(client => <TaskItem
              key={client.client_id}
              icon='iconxiaoxiquanxian'
              title={client.welcome!}
              btn={$t('action.open')}
              action={() => onModalOpen!(client)}
            />)}
          <TaskItem
            icon='iconic_yaoqing'
            title={$t("invite.claim.title")}
            btn={$t(`invite.claim.btn`)}
            tips={$t("invite.claim.count", { count: invite_count })}
            action={() => history.push('/invite')}
          />
        </ul>
      </div>
    </div>
  )
}

interface TaskItemProps {
  icon: string
  title: string
  btn: string
  tips?: string
  action: () => void
  isDouble?: boolean
  disabled?: boolean
  $t?: (key: string) => string
}
const TaskItem: FC<TaskItemProps> = ({
  icon, title, btn, tips, action, isDouble, disabled, $t
}) => {
  return <li className={styles.job}>
    {isDouble && <div className={styles.double}>{$t!('claim.energy.title')} X2</div>}
    <div className={styles.icon}>
      <i className={`iconfont ${icon} ${styles.jobIcon}`} />
    </div>
    <p className={styles.info}>{title}</p>
    <div className={styles.jobBtn}>
      <button
        className={styles.btn}
        onClick={action}
        disabled={disabled}
      >
        {btn}
      </button>
      {tips && <span className={styles.caption}>{tips}</span>}
    </div>
  </li>
}