import { ApiPostLottery, LotteryRecord } from "@/apis/claim"
import { get$t } from "@/locales/tools"
import React, { useEffect, useState, FC, useRef } from "react"
import { useIntl } from "umi"
import styles from "./LotteryBox.less"

interface LotteryBoxProps {
  data: LotteryRecord[]
  onPrizeClick(prize: LotteryRecord): void
  ticketCount?: number
  value?: string
  onEnd(): void
  onStart(): void
  onImgLoad(): void
}

const dotOutPosition = [[9, 15]]
new Array(15).fill(0).forEach((_, i) => dotOutPosition.push([25 + i * 21, 11]))
dotOutPosition.push([336, 15])
new Array(17).fill(0).forEach((_, i) => dotOutPosition.push([340, 34 + i * 21]))
dotOutPosition.push([336, 388])
new Array(15).fill(0).forEach((_, i) => dotOutPosition.push([319 - i * 21, 392]))
dotOutPosition.push([9, 388])
new Array(17).fill(0).forEach((_, i) => dotOutPosition.push([5, 370 - i * 21]))

const dotInnerPosition = [[7, 7]]
new Array(6).fill(0).forEach((_, i) => dotInnerPosition.push([29 + i * 23, 6]))
dotInnerPosition.push([166, 7])
new Array(8).fill(0).forEach((_, i) => dotInnerPosition.push([167, 28 + i * 20]))
dotInnerPosition.push([166, 188])
new Array(6).fill(0).forEach((_, i) => dotInnerPosition.push([144 - i * 23, 189]))
dotInnerPosition.push([7, 188])
new Array(8).fill(0).forEach((_, i) => dotInnerPosition.push([6, 168 - i * 20]))

export const LotteryBox: FC<LotteryBoxProps> = ({
  data,
  ticketCount = 0,
  onStart,
  onEnd,
  onImgLoad,
  onPrizeClick,
}) => {
  const $t = get$t(useIntl())
  const [activeReward, setActiveReward] = useState("")
  const startRef = useRef<any>()
  const [disabled, setDisabled] = useState(false)
  const [imgLoadCount, setImgLoadCount] = useState(0)

  useEffect(() => {
    startRef.current = createLucyLottery(
      data,
      (params: any) => setActiveReward(params.lottery_id)
    )
  }, [])

  useEffect(() => {
  }, [imgLoadCount])

  const handleStartClick = async () => {
    if (ticketCount <= 0) return
    if (disabled) return
    setDisabled(true)
    onStart()

    const { lottery_id } = await ApiPostLottery() || {}
    if (lottery_id) startRef.current(
      lottery_id,
      (params: any) => setActiveReward(params.lottery_id),
      (params: any) => {
        setActiveReward(params.lottery_id)
        onEnd()
        setDisabled(false)
      },
    )
  }

  const handlePrizeClick = (e: React.MouseEvent<HTMLDivElement>) => {
    const lotteryId = e.currentTarget.dataset.id
    if (!lotteryId || !data) return
    const target = data.find((x) => x.lottery_id === lotteryId)
    onPrizeClick(target!)
  }

  return (
    <div className={styles.container} style={{
      backgroundImage: `url(${require('@/assets/img/lottery-box-bg.png')})`,
    }}>
      {getCircleDot(dotOutPosition, styles.dotOut)}
      <div className={styles.lottery}>
        {data?.map((reward) => (
          <div
            key={reward.lottery_id}
            className={`${styles.reward} ${activeReward === reward.lottery_id ? styles.active : ""}`}
          >
            <div className={styles.prize} onClick={handlePrizeClick} data-id={reward.lottery_id}>
              <img src={reward.icon_url} onLoad={() => {
                setImgLoadCount(imgLoadCount + 1)
                if (imgLoadCount + 1 === data.length) onImgLoad()
              }} className={styles.icon} />
              <p className={styles.amount}>
                {reward.asset_id == "c6d0c728-2624-429b-8e0d-d9d19b6592fa" ?
                  (Number(reward.amount) * 1e8).toFixed() : reward.amount}
              </p>
            </div>
          </div>
        ))}
        <div className={styles.content}>
          {getCircleDot(dotInnerPosition, styles.dotInner)}
          <div className={styles.startWrapper}>
            <div className={styles.start}>
              <button
                disabled={disabled}
                onClick={handleStartClick}
                className={(!disabled && ticketCount > 0 && styles.active) || ""}
              >
                <div>{$t("claim.now")}</div>
                <div>{$t("claim.title")}</div>
              </button>
              <span className={styles.tip}>
                {$t("claim.you")}&nbsp;
                <span className={styles.count}>{ticketCount}</span>
                &nbsp;{$t("claim.ticketCount")}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function getCircleDot(dotOutPosition: number[][], name: string) {
  return dotOutPosition.map(([x, y], i) => <div className={name} key={i} style={{ left: `${x}px`, top: `${y}px`, }} />)
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
const createLucyLottery = (list: any, run: any) => {
  const cycleNumber = 4, //圈数
    defaultSpeed = 8,
    maxSpeed = 1
  let next: number = 0,
    myReq: any

  list = JSON.parse(JSON.stringify(list))
  for (let i = 0; i < 16; i++) {
    list[next].next = list[nextMap[next]]
    next = nextMap[next]
  }
  let currentObj = list[0]
  let timer = setInterval(() => {
    run(currentObj)
    currentObj = currentObj.next
  }, 1000)

  return (id: string, running: any, runend: any) => {
    clearInterval(timer)
    let counter = 0 // 计数器
    let current = 0
    let startIdx = list.findIndex(
      (item: any) => item.lottery_id === currentObj.lottery_id,
    )
    let endIdx = list.findIndex((item: any) => item.lottery_id === id)
    let startCount = nextIndex.findIndex((item: any) => item === startIdx) // 当前数字值
    let endCount = nextIndex.findIndex((item: any) => item === endIdx)
    let allCount = cycleNumber * list.length + endCount - startCount
    var addSpeed = defaultSpeed - maxSpeed
    var reduceSpeed = allCount - (defaultSpeed - maxSpeed)
    running(currentObj)
    const step = () => {
      if (counter < addSpeed) {
        if (current < Math.pow(defaultSpeed - counter, 2)) {
          current = current + defaultSpeed / 2
        } else {
          current = 0
          counter++
          currentObj = currentObj.next
          running(currentObj)
        }
      } else if (counter >= addSpeed && counter < reduceSpeed) {
        if (current < maxSpeed) {
          current++
        } else {
          current = 0
          counter++
          currentObj = currentObj.next
          running(currentObj)
        }
      } else if (Math.sqrt(current) <= defaultSpeed - (allCount - counter)) {
        current = current + 2
      } else {
        current = 0
        counter++
        currentObj = currentObj.next
        running(currentObj)
      }
      // 停止
      if (counter >= allCount) {
        runend(currentObj)
        cancelAnimationFrame(myReq!)
        myReq = undefined

        setTimeout(() => {
          timer = setInterval(() => {
            run(currentObj)
            currentObj = currentObj.next
          }, 1000)
        }, 3500)

        return
      }
      myReq = requestAnimationFrame(step)
    }
    myReq = requestAnimationFrame(step)
  }
}
