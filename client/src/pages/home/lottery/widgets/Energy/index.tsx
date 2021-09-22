import React, { FC, ReactNode } from "react"
import styles from "./energy.less"
import { Progress } from "antd"
import { ReactComponent as Qiandao } from "@/assets/img/svg/ic_qiandao.svg"
import { get$t } from "@/locales/tools"
import { useIntl } from "@/.umi/plugin-locale/localeExports"
// import { ReactComponent as Yaoqing } from "@/assets/img/svg/ic_yaoqing.svg"

interface JobBtnProps {
  label: ReactNode
  caption: ReactNode
  disabled?: boolean
  onClick?(): void
}

const JobBtn: FC<JobBtnProps> = ({ label, caption, disabled, onClick }) => (
  <div className={styles.jobBtn}>
    <button className={styles.btn} onClick={onClick} disabled={disabled}>
      {label}
    </button>
    <span className={styles.caption}>{caption}</span>
  </div>
)

interface JobProps {
  icon: ReactNode
  info: string
  action: ReactNode
}

const Job: FC<JobProps> = ({ action, icon, info }) => {
  return (
    <div className={styles.job}>
      <div className={styles.icon}>
        {icon}
        {/* <i className={`iconfont icon-${icon}`} /> */}
      </div>
      <p className={styles.info}>{info}</p>
      {action}
    </div>
  )
}

interface EnergyProps {
  value?: number
  checkinCount?: number
  isCheckedIn?: boolean
  onExchangeClick?(): void
  onCheckinClick?(): void
}

export const Energy: FC<EnergyProps> = ({
  checkinCount = 0,
  value = 0,
  isCheckedIn,
  onExchangeClick,
  onCheckinClick,
}) => {
  const t = get$t(useIntl())

  return (
    <div className={styles.container}>
      <div className={styles.main}>
        <div className={styles.title}>
          <div className={styles.line_left} />
          <span className={styles.content}>{t("claim.energy.title")}</span>
          <div className={styles.line_right} />
        </div>
        <div className={styles.tip}>{t("claim.tag")}</div>
        <div className={styles.progress}>
          <Progress
            percent={value}
            showInfo={false}
            strokeWidth={14}
            strokeColor={{
              "0%": "#FC602E",
              "83.25%": "#F68934",
              "95.86%": "#FDDE8B",
            }}
          />
          <p className={styles.info}>{t("claim.energy.describe")}</p>
        </div>
        <button
          disabled={value < 100}
          onClick={onExchangeClick}
          className={`${styles.exchange} ${
            value >= 100 ? styles.active : styles.default
          }`}
        >
          {t("claim.energy.exchange")}
        </button>
        <ul className={styles.job_list}>
          <Job
            icon={<Qiandao />}
            info={t("claim.energy.checkin.describe")}
            action={
              <JobBtn
                onClick={onCheckinClick}
                disabled={isCheckedIn}
                label={t(
                  isCheckedIn
                    ? "claim.energy.checkin.checked"
                    : "claim.energy.checkin.label",
                )}
                caption={t("claim.energy.checkin.count", {
                  count: checkinCount,
                })}
              />
            }
          />
          {/* <Job
            icon={<Yaoqing />}
            info="每邀请 1 人加入任意社群，获得 50 能量"
            action={<JobBtn label="邀请" caption="已邀请1" />}
          /> */}
        </ul>
      </div>
    </div>
  )
}
