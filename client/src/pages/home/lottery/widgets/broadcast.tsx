import React, { FC } from "react"
import styles from "./broadcast.less"

export interface BroadcastBoxProps {
  // uname: string
  // content: string
}

export const BroadcastBox: FC<BroadcastBoxProps> = ({ children }) => {
  return (
    <div className={styles.container}>
      <div className={styles.content}>{children}</div>
    </div>
  )
}
