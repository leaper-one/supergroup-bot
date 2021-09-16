import { BackHeader } from "@/components/BackHeader"
import { Radio } from "@/components/Radio"
import { get$t } from "@/locales/tools"
import React, { FC } from "react"
import { useIntl } from "react-intl"
import styles from "./records.less"

interface GuessItemProps {
  label: string
  name: string
  onChange?: React.ChangeEventHandler
}

const GuessItem: FC<GuessItemProps> = ({ label, name, onChange }) => {
  return (
    <div>
      <div />
      <span>{label}</span>
      <Radio name={name} onChange={onChange} />
    </div>
  )
}

export default function GuessRecordsPage() {
  const t = get$t(useIntl())

  return (
    <div>
      <BackHeader name="猜价格赢 TRX" />
      <h1 className={styles.header}>TRX 今日价格竞猜</h1>
      <p className={styles.description}>
        今日 8:00 UTC+8 波场 TRX 价格为 $0.0150 (价格来自Coingecko.com)
        请预测明日 8:00 UTC+8 波场价格与今日比是：
      </p>
      <div className={styles.content}>
        <GuessItem label={t("guess.up")} name="up" />
        <GuessItem label={t("guess.down")} name="down" />
        <GuessItem label={t("guess.flat")} name="flat" />
      </div>
    </div>
  )
}
