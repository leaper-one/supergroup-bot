import React from "react"
import styles from "@/pages/red/red.less"

interface IPeopleProps {
  people: string
  setPeople: (v: string) => void
  $t: (v: string) => string
  maxPeople: number
}

export const InstantPeopleItem = (props: IPeopleProps) => (
  <li key="iPeople" className={styles.persons}>
    <input
      value={props.people}
      onChange={(e) => props.setPeople(e.target.value)}
      type="number"
      placeholder={`${Math.ceil(props.maxPeople / 200)} ~ ${props.maxPeople}`}
    />
    <span>{props.$t("red.people")}</span>
  </li>
)

interface IAmountProps {
  amount: string
  setAmount: (v: string) => void
  $t: (v: string) => string
  noRight?: boolean
  placeholder?: string
  onBlur?: (v: string) => void
}

export const InstantAmountItem = (props: IAmountProps) => (
  <li key="iAmount" className={styles.amount}>
    <input
      value={props.amount}
      onChange={(e) => props.setAmount(e.target.value)}
      onBlur={(e) => props.onBlur && props.onBlur(e.target.value)}
      type="number"
      placeholder={props.placeholder || "最小 0.0001"}
    />
    {!props.noRight && <span>{props.$t("red.amount")}</span>}
  </li>
)
