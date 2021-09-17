import React, { ReactNode, FC, HTMLProps } from "react"

import styles from "./radio.less"

interface RadioProps extends Omit<HTMLProps<HTMLInputElement>, "label"> {
  label?: ReactNode
}

export const Radio: FC<RadioProps> = ({
  label,
  name,
  checked,
  onChange,
  disabled,
}) => {
  return (
    <label
      className={`${styles.label} ${checked ? styles.checked : ""} ${
        disabled ? styles.disabled : ""
      }`}
    >
      <input
        type="radio"
        name={name}
        checked={checked}
        disabled={disabled}
        onChange={onChange}
        className={styles.radio}
      />
      {label}
    </label>
  )
}
