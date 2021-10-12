import React, { FC, TouchEvent, useState } from "react"
import styles from "./TouchBar.less"

const transition = 'all 0.1s'
export interface TouchBarProps {
  label: string
  onOk(): void
}

export const TouchBar: FC<TouchBarProps> = ({
  label,
  onOk
}) => {
  const [progress, setProgress] = useState(0)
  const [btn, setBtn] = useState("4px")
  const [isBack, setIsBack] = useState(false)

  const getPercent = (e: any) => {
    const currentWidth = e.touches[0].clientX
    const totalWidth = e.target.parentElement.clientWidth
    if (currentWidth <= 32) setBtn("4px")
    else if (currentWidth >= totalWidth - 32) setBtn("calc(100% - 58px)")
    else setBtn(`${currentWidth - 27}px`)
    return currentWidth / totalWidth * 100 | 0
  }
  const handleTouchStart = (e: TouchEvent<HTMLDivElement>) => {
    document.addEventListener("touchmove", handleTouchMove)
    document.addEventListener("touchend", handleTouchEnd)
  }

  const handleTouchMove = (e: any) => {
    setProgress(getPercent(e))
  }

  const handleTouchEnd = () => {
    setIsBack(true)
    setTimeout(() => {
      setProgress(0)
      setBtn("4px")
      document.removeEventListener("touchmove", handleTouchMove)
      document.removeEventListener("touchend", handleTouchEnd)
      if (progress >= 100) onOk()
      setTimeout(() => {
        setIsBack(false)
      }, 60)
    }, 100)
  }

  return (
    <div className={styles.container} >
      <div
        className={styles.bar}
        style={{
          width: `${progress}%`,
          transition: isBack ? transition : ''
        }}
      />
      <div
        className={`${styles.button}`}
        style={{
          left: btn,
          transition: isBack ? transition : ''
        }}
        onTouchStart={handleTouchStart}
        onTouchEnd={handleTouchEnd}
      >
        <div className={styles.line} />
        <div className={styles.line} />
        <div className={styles.line} />
      </div>
      <span className={styles.label}>
        {label}
      </span>
    </div>
  )
}
