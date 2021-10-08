import React, {
  FC,
  ReactNode,
  useRef,
  TouchEvent,
  useState,
  useEffect,
  useLayoutEffect,
} from "react"
import styles from "./TouchBar.less"

const useIsomorphicEffect =
  typeof window === "undefined" ? useEffect : useLayoutEffect
const toPercent = (num: number) => Number(num.toFixed(4)) * 100
const noop = () => {}

export interface TouchBarProps {
  label?: ReactNode
  okLabel?: ReactNode
  onOk?(): void
  freeze?: boolean
  freezeOnOk?: boolean
}

export const TouchBar: FC<TouchBarProps> = ({
  label,
  okLabel,
  freeze,
  freezeOnOk = true,
  onOk = noop,
}) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const anchorRef = useRef<HTMLDivElement>(null)
  const [progress, setProgress] = useState(0)
  const [safeProgress, setSafeProgress] = useState(0) // UI状态, 因为使用的是transform 所以100% 按钮会被刚好移出一格 后续优化dom结构 可能就不需要了
  const [isActive, setIsActive] = useState(false) // UI 状态 控制按钮的边缘颜色(anchorRef的背景色)

  // TODO: 优化dom结构 这里大概就不需要了
  useIsomorphicEffect(() => {
    if (!containerRef.current || !anchorRef.current) return
    const { x, width } = containerRef.current.getBoundingClientRect()
    const anchorRect = anchorRef.current.getBoundingClientRect()
    const maxPercent = toPercent(anchorRect.width / width)
    setSafeProgress(maxPercent)
  }, [])

  const getPercent = (clientX: number) => {
    if (!containerRef.current || !anchorRef.current) return 0
    const { x, width } = containerRef.current.getBoundingClientRect()
    const anchorRect = anchorRef.current.getBoundingClientRect()

    const maxSpace = width - x - anchorRect.width // 起点拉到屏幕最左边
    const offset = clientX - x - anchorRect.width // 起点拉到屏幕最左边

    // Math.min(offset, maxSpace) maxSpace 这里是最大安全边界 不会超过这个数
    const cursor = Math.min(Math.max(offset, 0), maxSpace)

    // Math.max(progress, 0) 0 这里是最小安全边界, 负数将不会移动
    const result = Math.max(toPercent(cursor / maxSpace), 0)

    if (result > 80) return 100 // 某些情况下 toFixed 也无法拯救js的浮点数
    return result
  }

  const isFreeze = freeze || (freezeOnOk && progress === 100)
  const handleTouchStart = (e: TouchEvent<HTMLDivElement>) => {
    if (isFreeze) return
    setIsActive(true)
    setProgress(getPercent(e.touches[0].clientX))
  }

  const handleTouchMove = (e: TouchEvent<HTMLDivElement>) => {
    if (isFreeze) return
    setProgress(getPercent(e.touches[0].clientX))
  }

  const handleTouchEnd = () => {
    if (progress === 100) {
      return onOk()
    }

    setIsActive(false)
    setProgress(0)
  }

  return (
    <div
      ref={containerRef}
      className={styles.container}
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
    >
      <div
        className={styles.slider}
        style={
          safeProgress
            ? {
                transform: `translateX(${
                  progress === 100 ? progress - safeProgress : progress
                }%)`,
              }
            : undefined
        }
      >
        <div
          ref={anchorRef}
          className={`${styles.anchor} ${isActive ? styles.active : ""} ${
            progress === 0 && isActive ? styles.start : ""
          }`}
        >
          <button className={styles.button}>
            <div className={styles.line} />
            <div className={styles.line} />
            <div className={styles.line} />
          </button>
        </div>
        <div className={styles.mask} />
      </div>
      <span className={styles.label}>
        {progress === 100 ? okLabel || label : label}
      </span>
    </div>
  )
}
