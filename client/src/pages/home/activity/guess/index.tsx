import { BackHeader } from "@/components/BackHeader"
import { Radio } from "@/components/Radio"
import { get$t } from "@/locales/tools"
import { GuessType, GuessTypeKeys } from "@/types"
import React, { FC, useCallback, useEffect, useState, memo } from "react"
import { ApiGetGuessPageData } from "@/apis/guess"
import { useIntl } from "react-intl"
import { useParams } from "umi"
import { Button } from "./widgets/Button"

import styles from "./guess.less"

interface TipListProps {
  data?: string[]
  label: string
}
const TipList: FC<TipListProps> = memo(({ data, label }) => (
  <div className={`${styles.card} ${styles.tipList}`}>
    <h3 className={styles.label}>{label}</h3>
    <ul className={styles.list}>
      {data &&
        data.map((x) => (
          <li key={x} className={styles.item}>
            <div className={styles.flag} />
            <span>{x}</span>
          </li>
        ))}
    </ul>
  </div>
))

interface GuessOptionProps {
  label: string
  // logo: string
  name: GuessType
  checked?: boolean
  onChange?: React.ChangeEventHandler
}

const GuessOption: FC<GuessOptionProps> = ({
  label,
  name,
  checked,
  onChange,
}) => {
  return (
    <div className={styles.option}>
      <div className={styles[GuessType[name]]} />
      <span className={styles.label}>{label}</span>
      <Radio name={GuessType[name]} onChange={onChange} checked={checked} />
    </div>
  )
}

type GuessPageParams = {
  id: string
}

export default function GuessPage() {
  const t = get$t(useIntl())
  const [choose, setChoose] = useState<GuessTypeKeys>()
  const { id } = useParams<GuessPageParams>()
  const [startAt, setStartAt] = useState<string>()
  const [endAt, setEndAt] = useState<string>()
  const [rules, setRules] = useState<string[]>()
  const [explains, setExplains] = useState<string[]>()

  // console.log("111", id)

  const fetchPageData = useCallback(() => {
    // ApiGetGuessPageData(id).then(() => {
    // })
    setExplains([
      "活动规则活动规则活动规则活动规则活动规则活动规则活动规则活动规则活动规则",
      "活动规则活动规则活动规则活动规则活动规则活动规则活动规则活动规则活动规则",
    ])
    setRules([
      "活动规则活动规则活动规则活动规则活动规则活动规则活动规则活动规则活动规则",
      "活动规则活动规则活动规则活动规则活动规则活动规则活动规则活动规则活动规则",
    ])
  }, [id])

  useEffect(() => {
    fetchPageData()
  }, [fetchPageData])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    // setChoose(e.target.name as GuessTypeKeys)
    setChoose(e.currentTarget.name as GuessTypeKeys)
  }

  return (
    <div className={styles.container}>
      <BackHeader name="猜价格赢 TRX" />
      <h1 className={styles.header}>TRX 今日价格竞猜</h1>
      <p className={styles.description}>
        今日 8:00 UTC+8 波场 TRX 价格为 $0.0150 (价格来自Coingecko.com)
        请预测明日 8:00 UTC+8 波场价格与今日比是：
      </p>
      {/* onChange={handleChange} */}
      <div className={styles.card}>
        <div className={styles.guess}>
          <GuessOption
            label={t("guess.up")}
            name={GuessType.Up}
            checked={choose === "Up"}
            onChange={handleChange}
          />
          <GuessOption
            label={t("guess.down")}
            name={GuessType.Down}
            checked={choose === "Down"}
            onChange={handleChange}
          />
          <GuessOption
            label={t("guess.flat")}
            name={GuessType.Flat}
            checked={choose === "Flat"}
            onChange={handleChange}
          />
        </div>
        <Button className={styles.confirm}>{t("guess.confirm")}</Button>
      </div>
      <TipList data={rules} label="活动规则" />
      <TipList data={rules} label="活动说明" />
    </div>
  )
}
