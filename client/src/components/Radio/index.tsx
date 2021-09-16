import React, { ReactNode, FC } from "react"
import styles from "./radio.less"

interface RadioProps {
  label?: ReactNode
  name?: string
  checked?: boolean
  onChange?: React.ChangeEventHandler
}

export const Radio: FC<RadioProps> = ({ label, name, checked, onChange }) => {
  return (
    <label className={styles.label}>
      <span>
        <input
          type="radio"
          name={name}
          checked={checked}
          onChange={onChange}
          className={styles.radio}
        />
        {label}
      </span>
    </label>
  )
}
