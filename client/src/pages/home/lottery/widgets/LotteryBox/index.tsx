import { ApiPostLottery, ApiGetLotteryReward } from "@/apis/claim"
import { Modal } from "antd-mobile"
import React, { useEffect, useState, FC, useRef } from "react"
import { Lottery } from "../../types"
import styles from "./lotteryBox.less"

interface LotteryBoxProps {
  data?: Lottery[]
  ticketCount?: number
  value?: string
  onEnd(): void
}

export const LotteryBox: FC<LotteryBoxProps> = ({
  data = [],
  ticketCount = 0,
  onEnd,
}) => {
  const [activeReward, setActiveReward] = useState("")
  const [prize, setPrize] = useState<Lottery>()
  const startRef = useRef<any>()

  useEffect(() => {
    if (data && data.length) {
      startRef.current = createLucyLottery(data)
    }
  }, [data.join()])

  const handleStartClick = () => {
    // ticketCount <= 0 &&
    // if (!startRef.current) return
    ApiPostLottery().then((x) => {
      startRef.current(
        x.lottery_id,
        (params: any) => setActiveReward(params.lottery_id),
        (params: any) => {
          setPrize(params)
          setActiveReward(params.lottery_id)
          onEnd()
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
                <button onClick={handleStartClick}>
                  <div>立刻</div>
                  <div>抽奖</div>
                </button>
                <span className={styles.tip}>
                  您有&nbsp;<span className={styles.count}>{ticketCount}</span>
                  &nbsp;次抽奖机会
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
    defaultSpeed = 15,
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
