import React, {
  createContext,
  useContext,
  PropsWithChildren,
  useRef,
  useState,
} from "react"
import anime from "animejs"

export type Status = "running" | "idle"

type Start = () => void
type End<T> = () => T

interface RollingBoxStore<T = any> {
  cursor: number
  value: number

  items: number[]
  status: Status
  start: Start
}

const RollingBoxContext = createContext<RollingBoxStore | null>(null)

export interface RollingBoxProviderProps<T = any> {
  cursor?: number
  defaultCursor?: number
  value?: number
  speed?: number
  maxSpeed?: number
  minSpeed?: number
  duration?: number
  minTimes?: number
  onStart?: () => void | boolean

  isClockwise?: boolean
  options?: number[]
}

type CSYS = {
  x: number
  y: number
}

const bezier = (p: [CSYS, CSYS, CSYS], t: number) => ({
  x:
    Math.pow(1 - t, 2) * p[0].x +
    2 * t * (1 - t) * p[1].x +
    Math.pow(t, 2) * p[2].x,
  y:
    Math.pow(1 - t, 2 * p[0].y) +
    2 * t * (1 - t) * p[1].y +
    Math.pow(t, 2) * p[2].y,
})

function easeIn(t: number, b: number, c: number, d: number) {
  if (t >= d) t = d
  return c * (t /= d) * t + b
}

function easeOut(t: number, b: number, c: number, d: number) {
  if (t >= d) t = d
  return -c * (t /= d) * (t - 2) + b
}

export function RollingBoxProvider<T extends string | number>({
  value = 16,
  speed = 500,
  maxSpeed = 100,
  minSpeed = 200,
  minTimes = 2,
  duration = 4000,
  onStart,

  children,
  defaultCursor = 0,
  options = [15, 16, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14],
}: PropsWithChildren<RollingBoxProviderProps<T>>) {
  const [isStart, setIsStart] = useState(false)
  const [isError, setIsError] = useState(false)
  const [isSpeedup, setIsSpeedUp] = useState(false)
  const [cursor, setCursor] = useState(-1)

  const startTimeRef = useRef<number>(null)
  const loopTimerRef = useRef<NodeJS.Timer>()
  const cursorRef = useRef(-1)
  const speedRef = useRef<number>(500)
  const countRef = useRef(0)

  const maxIdx = options.length - 1
  const subscribers = []

  const loop = () => {
    loopTimerRef.current = setTimeout(() => {
      const overing = !!(!isStart && (value || value === 0))
      if (cursorRef.current === options.length - 1) {
        countRef.current += 1
      }

      // publish
      if (cursorRef.current <= maxIdx) {
        cursorRef.current += 1
      } else {
        cursorRef.current = 0
      }

      if (
        (!(overing && cursorRef.current === value) && !isError) ||
        countRef.current < minTimes
      ) {
        if (isSpeedup) {
          const point = bezier(
            [
              { x: 500, y: 500 },
              { x: 50, y: 50 },
              { x: 100, y: 100 },
            ],
            maxSpeed / speedRef.current,
          )
          speedRef.current = point.x
        } else {
          const point = bezier(
            [
              { x: 100, y: 100 },
              { x: 700, y: 700 },
              { x: 1000, y: 1000 },
            ],
            speedRef.current / minSpeed,
          )
          speedRef.current = Math.min(point.x, minSpeed)
        }
        setCursor(cursorRef.current)
        loop()
      } else {
        if (isError) return

        setCursor(-1)
        setIsStart(false)
        setIsSpeedUp(false)
        countRef.current = 0
        // publich
      }
    }, speedRef.current)
  }

  const stop = () => {}
  const run = () => {}
  const _slowDown = () => {}
  const _speedUp = () => {}
  const _uniform = () => {}

  const _start = () => {
    let a = { cursor: 0 }

    anime({
      targets: a,
      cursor: options.length - 1,
      loop: minTimes,
      duration: 3000,
      update(t: anime.AnimeInstance) {
        console.log(a)
        // t.duration = 50
      },
    })
  }

  const start = () => {
    setIsStart(true)
    setIsSpeedUp(true)
    loop()
  }

  const store: RollingBoxStore = {
    cursor,
    status: isStart ? "running" : "idle",
    value,
    start,
    items: options,
  }

  return (
    <RollingBoxContext.Provider value={store}>
      {children}
    </RollingBoxContext.Provider>
  )
}

export const useRollingBox = () => {
  const ctx = useContext(RollingBoxContext)
  if (!ctx) {
    throw Error("useRollingBox must be used within a RollingBoxProvider")
  }

  return ctx
}
