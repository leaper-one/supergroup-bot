import React, { FC } from "react"
import styles from "./Row.less"

export const Row: FC = ({ children }) => (
  <div className={styles.container}>{children}</div>
)
