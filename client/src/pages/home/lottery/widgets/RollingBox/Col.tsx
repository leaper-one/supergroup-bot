import React, { FC } from "react"
import styles from "./Col.less"

export const Col: FC = ({ children }) => {
  return <div className={styles.container}>{children}</div>
}
