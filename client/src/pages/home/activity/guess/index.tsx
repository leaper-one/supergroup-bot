import { BackHeader } from "@/components/BackHeader"
import { Radio } from "@/components/Radio"
import { get$t } from "@/locales/tools"
import { GuessType, GuessTypeKeys } from "@/types"
import React, { FC, useEffect, useState, memo, useMemo, useRef } from "react"
import { ApiGetGuessPageData, ApiCreateGuess } from "@/apis/guess"
import { useIntl } from "react-intl"
import { useParams, history } from "umi"
import { Button } from "./widgets/Button"
import { Modal } from "antd-mobile"
import { calcUtcHHMM, getUtcHHMM } from "@/utils/time"
import flagSrc from "@/assets/img/guess_flag.png"

import styles from "./guess.less"
import { Icon } from "@/components/Icon"
import { FullLoading } from "@/components/Loading"
import { changeTheme } from "@/assets/ts/tools"

interface NewLineProps {
  txt: string
  className?: string
  newLineClassName?: string
}

const NewLine: FC<NewLineProps> = ({ txt, className, newLineClassName }) => {
  let data = txt.split("\\n")
  if (data.length)
    return (
      <>
        {data.map((x) => (
          <p key={x} className={newLineClassName}>
            {x}
          </p>
        ))}
      </>
    )

  return <span className={className}>{txt}</span>
}

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
            <img className={styles.flag} src={flagSrc} />
            <div className={styles.tip}>
              <NewLine txt={x} newLineClassName={styles.p} />
            </div>
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
  disabled?: boolean
  onChange?: React.ChangeEventHandler<HTMLInputElement>
}

const GuessOption: FC<GuessOptionProps> = React.memo(
  ({ label, name, checked, disabled, onChange }) => {
    return (
      <div className={styles.option}>
        <div className={styles[GuessType[name]]} />
        <span className={styles.label}>{label}</span>
        <Radio
          name={GuessType[name]}
          onChange={onChange}
          checked={checked}
          disabled={disabled}
        />
      </div>
    )
  },
)

type GuessPageParams = {
  id: string
}

type ModalType =
  | "choose"
  | "confirm"
  | "success"
  | "missing"
  | "end"
  | "notstart"

export default function GuessPage() {
  const t = get$t(useIntl())
  const [choose, setChoose] = useState<GuessTypeKeys>()
  const { id } = useParams<GuessPageParams>()
  const [startAt, setStartAt] = useState<string>()
  const [endAt, setEndAt] = useState<string>()
  const [startTime, setStartTime] = useState<string>()
  const [endTime, setEndTime] = useState<string>()
  const [rules, setRules] = useState<string[]>()
  const [explains, setExplains] = useState<string[]>()
  const [disabled, setDisabled] = useState(false)
  const [modalType, setModalType] = useState<ModalType>()
  const prevModalTypeRef = useRef<ModalType>()
  const [isLoaded, setIsLoaded] = useState(false)

  const [usd, setUsd] = useState<string>()
  const [coin, setCoin] = useState<string>()

  useEffect(() => {
    changeTheme("#da1f27")
    return () => {
      changeTheme("#fff")
    }
  }, [])

  const fetchPageData = (cb?: () => void) => {
    ApiGetGuessPageData(id).then((x) => {
      setRules(x.rules)
      setExplains(x.explain)
      setCoin(x.symbol)
      setStartAt(x.start_at)
      setEndAt(x.end_at)
      setStartTime(calcUtcHHMM(x.start_time, 8))
      setEndTime(calcUtcHHMM(x.end_time, 8))

      if (x.today_guess) {
        setDisabled(true)
        setChoose(GuessType[x.today_guess!] as GuessTypeKeys)
      }

      setUsd(x.price_usd)
      setIsLoaded(true)
      if (cb) cb()
    })
  }

  useEffect(() => {
    fetchPageData()
  }, [fetchPageData])

  useEffect(() => {
    if (modalType && prevModalTypeRef.current !== modalType) {
      prevModalTypeRef.current = modalType
    }
  }, [modalType])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    // setChoose(e.target.name as GuessTypeKeys)
    setChoose(e.currentTarget.name as GuessTypeKeys)
  }

  const navToRecords = () => {
    history.push(`/activity/${id}/records`)
  }

  const handleSubmitValidate = () => {
    if (!startTime || !endTime) return
    if (!choose) {
      return setModalType("choose")
    }

    const nowTime = calcUtcHHMM(getUtcHHMM(), 8)
    const [nh, nm] = nowTime.split(":").map(Number)
    const [sh, sm] = startTime.split(":").map(Number)
    const [eh, em] = endTime.split(":").map(Number)

    const isDateNotStart = startAt && Date.parse(startAt) > Date.now()
    const isDateEnd = endAt && Date.parse(endAt) < Date.now()
    const isHHmmEnd = nh > eh || (nh >= eh && nm >= em)
    const isHHmmNotStart = nh < sh || (nh < sh && nm < sm)

    if (isDateNotStart || isHHmmNotStart) {
      return setModalType("notstart")
    }

    if (isDateEnd || (endAt && Date.parse(endAt) === Date.now() && isHHmmEnd)) {
      return setModalType("end")
    }

    if (isHHmmEnd) {
      return setModalType("missing")
    }

    setModalType("confirm")
  }

  const handleSubmit = () => {
    if (!choose) return
    ApiCreateGuess({ guess_id: id, guess_type: GuessType[choose] }).then(() => {
      setDisabled(true)
      setModalType("success")
    })
  }

  const handleModalClose = () => {
    setModalType(undefined)
  }

  const modalBtn = useMemo(() => {
    if (
      modalType === "confirm" ||
      (!modalType && prevModalTypeRef.current === "confirm")
    ) {
      return (
        <div className={styles.btnGroup}>
          <Button className={styles.btn} onClick={handleSubmit}>
            {t("guess.sure")}
          </Button>
          <Button
            className={styles.btn}
            kind="warning"
            onClick={handleModalClose}
          >
            {t("guess.notsure")}
          </Button>
        </div>
      )
    }

    return (
      <Button className={styles.btn} onClick={handleModalClose}>
        {t(modalType === "choose" ? "guess.goChoose" : "guess.okay")}
      </Button>
    )
  }, [modalType])

  return (
    <div className={styles.container}>
      <BackHeader
        name={"猜价格赢 " + coin}
        isWhite
        action={
          <Icon
            i="ic_file_text"
            className={styles.record}
            onClick={navToRecords}
          />
        }
      />
      <div className={styles.header}>
        <h1 className={styles.title}>{t("guess.todayGuess", { coin })} </h1>
        <p className={styles.description}>
          {t("guess.todyDesc", { coin, usd, time: startTime })}
        </p>
      </div>
      <div className={styles.content}>
        {/* onChange={handleChange} */}
        <div className={styles.card}>
          <div className={styles.guess}>
            <GuessOption
              label={t("guess.up")}
              name={GuessType.Up}
              checked={choose === "Up"}
              disabled={disabled}
              onChange={handleChange}
            />
            <GuessOption
              label={t("guess.down")}
              name={GuessType.Down}
              checked={choose === "Down"}
              disabled={disabled}
              onChange={handleChange}
            />
            <GuessOption
              label={t("guess.flat")}
              name={GuessType.Flat}
              checked={choose === "Flat"}
              disabled={disabled}
              onChange={handleChange}
            />
          </div>
          <Button
            className={styles.confirm}
            disabled={disabled}
            onClick={handleSubmitValidate}
          >
            {t("guess.sure")}
          </Button>
        </div>
        <TipList data={rules} label="活动规则" />
        <TipList data={explains} label="活动说明" />
      </div>
      <Modal visible={!!modalType} transparent onClose={handleModalClose}>
        {(modalType || prevModalTypeRef.current) && (
          <div className={styles.modal}>
            <div
              className={`${styles.emoji} ${
                styles[(modalType || prevModalTypeRef.current) as string]
              }`}
            />
            <p className={styles.tip}>
              {t(`guess.${modalType || prevModalTypeRef.current}.tip`)}
            </p>
            <p className={styles.info}>
              {t(`guess.${modalType || prevModalTypeRef.current}.info`, {
                coin,
                start: startTime,
                end: endTime,
              })}
            </p>

            {modalBtn}
          </div>
        )}
      </Modal>
      {!isLoaded && <FullLoading mask />}
    </div>
  )
}
