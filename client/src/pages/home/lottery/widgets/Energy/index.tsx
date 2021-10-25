import React, { FC } from "react"
import styles from "./energy.less"
import { Progress } from "antd"
import { get$t } from "@/locales/tools"
import { useIntl } from "@/.umi/plugin-locale/localeExports"

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
  const $t = get$t(useIntl())

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
            percent={value}
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
          disabled={value < 100}
          onClick={onExchangeClick}
          className={`${styles.exchange} ${value >= 100 ? styles.active : styles.default}`}
        >
          {$t("claim.energy.exchange")}
        </button>
        <ul className={styles.job_list}>
          <li className={styles.job}>
            <div className={styles.icon}>
              <img src={require("@/assets/img/svg/ic_qiandao.svg")} alt="" />
            </div>
            <p className={styles.info}>{$t("claim.energy.checkin.describe")}</p>
            <div className={styles.jobBtn}>
              <button className={styles.btn} onClick={onCheckinClick} disabled={isCheckedIn}>
                {$t(`claim.energy.checkin.${isCheckedIn ? 'checked' : 'label'}`)}
              </button>
              <span className={styles.caption}>{$t("claim.energy.checkin.count", { count: checkinCount, })}</span>
            </div>
          </li>
        </ul>
      </div>
    </div>
  )
}
