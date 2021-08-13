import React from "react"
import styles from "./index.less"
import { history } from 'umi'

interface Props {
  name: string
  noBack?: Boolean
  action?: JSX.Element | undefined
  onClick?: () => void | undefined
  isWhite?: boolean
  backHome?: boolean
}

export const BackHeader = (props: Props) => {
  return (
    <div className={`${styles.header} ${props.isWhite && styles.white}`}>
      {!props.noBack && <i className={`iconfont iconic_return ${styles.back}`} onClick={() => props.backHome ? history.push('/') : history.go(-1)} />}
      <span onClick={props.onClick}>{props.name}</span>
      {props.action && <div className={styles.action}>{props.action}</div>}
    </div>
  )
}
