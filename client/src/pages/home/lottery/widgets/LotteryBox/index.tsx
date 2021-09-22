import { ApiPostLottery, ApiGetLotteryReward } from "@/apis/claim"
import { get$t } from "@/locales/tools"
import React, { useEffect, useState, FC, useRef } from "react"
import { useIntl } from "umi"
import { Lottery } from "../../types"
import styles from "./lotteryBox.less"

interface LotteryBoxProps {
  data?: Lottery[]
  ticketCount?: number
  value?: string
  onEnd(): void
  onStart(): void
}

export const LotteryBox: FC<LotteryBoxProps> = ({
  data = [],
  ticketCount = 0,
  onStart,
  onEnd,
}) => {
  const t = get$t(useIntl())
  const [activeReward, setActiveReward] = useState("")
  const startRef = useRef<any>()
  const [disabled, setDisabled] = useState(false)

  useEffect(() => {
    if (data && data.length) {
      startRef.current = createLucyLottery(data)
    }
  }, [data.join()])

  const handleStartClick = () => {
    if (ticketCount <= 0) return
    if (disabled) return
    setDisabled(true)
    onStart()

    ApiPostLottery().then((x) => {
      startRef.current(
        x.lottery_id,
        (params: any) => setActiveReward(params.lottery_id),
        (params: any) => {
          setActiveReward(params.lottery_id)
          onEnd()
          setDisabled(false)
        },
      )
    })
  }

  return (
    <div className={styles.container}>
      <div className={styles.wrapper}>
        <div className={styles.lottery}>
          {data &&
            data.map((reward) => (
              <div
                key={reward.lottery_id}
                className={`${styles.reward} ${
                  activeReward === reward.lottery_id ? styles.active : ""
                }`}
              >
                <div className={styles.prize}>
                  <img src={reward.icon_url} alt="" />
                  <p>{reward.amount}</p>
                </div>
              </div>
            ))}
          <div className={styles.content}>
            <div className={styles.startWrapper}>
              <div className={styles.start}>
                <button
                  onClick={handleStartClick}
                  disabled={disabled}
                  className={
                    !disabled && ticketCount > 0
                      ? styles.active
                      : styles.default
                  }
                >
                  <div>{t("claim.now")}</div>
                  <div>{t("claim.title")}</div>
                </button>
                <span className={styles.tip}>
                  {t("claim.you")}&nbsp;
                  <span className={styles.count}>{ticketCount}</span>
                  &nbsp;{t("claim.ticketCount")}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
const nextIndex = [0, 1, 2, 3, 4, 6, 8, 10, 15, 14, 13, 12, 11, 9, 7, 5]
const nextMap: Record<number, number> = {
  0: 1,
  1: 2,
  2: 3,
  3: 4,
  4: 6,
  6: 8,
  8: 10,
  10: 15,
  15: 14,
  14: 13,
  13: 12,
  12: 11,
  11: 9,
  9: 7,
  7: 5,
  5: 0,
}
const createLucyLottery = (list: any) => {
  const cycleNumber = 5, //圈数
    defaultSpeed = 10,
    maxSpeed = 4
  let next: number = 0,
    myReq: any

  list = JSON.parse(JSON.stringify(list))
  for (let i = 0; i < 16; i++) {
    list[next].next = list[nextMap[next]]
    next = nextMap[next]
  }

  return (id: string, running: any, runend: any) => {
    let counter = 0 // 计数器
    let current = 0 // 当前数字值
    let currentObj = list[0]
    let endObj = list.findIndex((item: any) => item.lottery_id === id)
    let addCount = nextIndex.findIndex((item: any) => item === endObj)
    let allCount = cycleNumber * list.length + addCount
    running(currentObj)
    const step = () => {
      // 减速环节
      if (Math.sqrt(current) <= defaultSpeed - (allCount - counter)) {
        current = current + 2
      } else {
        current = 0
        // 往前移动一个；
        counter++
        currentObj = currentObj.next
        running(currentObj)
      }
      // 停止
      if (counter >= allCount) {
        runend(currentObj)
        cancelAnimationFrame(myReq!)
        myReq = undefined
        return
      }
      myReq = requestAnimationFrame(step)
    }
    myReq = requestAnimationFrame(step)
  }
}
