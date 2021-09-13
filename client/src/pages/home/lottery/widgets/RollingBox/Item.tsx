import React, { PropsWithChildren, MouseEvent, ReactNode } from "react"
import { useRollingBox } from "./context"
import styles from "./Item.less"

export interface ItemProps<T extends string | number> {
  isActive?: boolean
  value?: T
  label?: ReactNode
  icon?: string
  onClick?(v?: T): void
}

export const Item = <T extends string | number>({
  label,
  value,
  onClick,
}: PropsWithChildren<ItemProps<T>>) => {
  const ctx = useRollingBox()

  const handleClick = (e: MouseEvent<HTMLDivElement>) => {
    e.preventDefault()

    if (onClick) onClick(value)
  }

  return (
    <div
      className={`${styles.container} ${
        ctx.cursor === value || (ctx.status === "idle" && ctx.value === value)
          ? styles.isActive
          : ""
      }`}
      data-value={value}
      onClick={handleClick}
    >
      {label}
    </div>
  )
}
