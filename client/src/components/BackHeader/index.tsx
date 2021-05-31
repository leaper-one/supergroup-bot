import React from "react"
import styles from "./index.less"
import { history } from "umi"

interface Props {
  name: string
  noBack?: Boolean
  action?: JSX.Element | undefined
  onClick?: () => void | undefined
}

export const BackHeader = (props: Props) => {
  return (
    <div className={styles.header}>
      {!props.noBack && (
        <img
          onClick={() => history.go(-1)}
          src={require("@/assets/img/svg/left.svg")}
          alt=""
        />
      )}
      <span onClick={props.onClick}>{props.name}</span>
      {props.action && <div className={styles.action}>{props.action}</div>}
    </div>
  )
}
