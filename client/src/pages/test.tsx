import React, { useEffect, useState } from 'react'
import styles from './test.less'

export default function MePage() {
  const [rewards, setRewards] = useState([
    { lottery_id: "SAT50", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 50, icon_url: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", client_id: "" },
    { lottery_id: "SAT100", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 100, icon_url: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", client_id: "" },
    { lottery_id: "SAT200", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 200, icon_url: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", client_id: "" },
    { lottery_id: "SAT500", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 500, icon_url: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", client_id: "" },
    { lottery_id: "SAT99999", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 99999, icon_url: "https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128", client_id: "" },
    { lottery_id: "AKITA5000", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 5000, icon_url: "https://mixin-images.zeromesh.net/JSxN4FxhH3LNDowo22bEV3fGMdrGmKrYzGyNqGbYe72GFEitLVFfwmxrjEE8ZDzqAc14LWUcuHtHiO8l7ODyExmnLwM3aPdx8D0Z=s128", client_id: "" },
    { lottery_id: "AKITA10000", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 10000, icon_url: "https://mixin-images.zeromesh.net/JSxN4FxhH3LNDowo22bEV3fGMdrGmKrYzGyNqGbYe72GFEitLVFfwmxrjEE8ZDzqAc14LWUcuHtHiO8l7ODyExmnLwM3aPdx8D0Z=s128", client_id: "" },
    { lottery_id: "AKITA50000", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 50000, icon_url: "https://mixin-images.zeromesh.net/JSxN4FxhH3LNDowo22bEV3fGMdrGmKrYzGyNqGbYe72GFEitLVFfwmxrjEE8ZDzqAc14LWUcuHtHiO8l7ODyExmnLwM3aPdx8D0Z=s128", client_id: "" },
    { lottery_id: "EPC100", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 100, icon_url: "https://mixin-images.zeromesh.net/HMXlpSt6KF9i-jp_ZQix9wFcMD27DrYox5kDrju6KkjvlQjQPZ2zimKKFYBJwecRTw5YAaMt4fpHXd1W0mwIxQ=s128", client_id: "" },
    { lottery_id: "EPC50", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 50, icon_url: "https://mixin-images.zeromesh.net/HMXlpSt6KF9i-jp_ZQix9wFcMD27DrYox5kDrju6KkjvlQjQPZ2zimKKFYBJwecRTw5YAaMt4fpHXd1W0mwIxQ=s128", client_id: "" },
    { lottery_id: "SHIB1000", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 1000, icon_url: "https://mixin-images.zeromesh.net/fgSEd6CY07BiZP76--7JA9P-rKIWRoXD8Eis8RUL6mP85_QPsbMoyJtWJ6MjE9jWFEjabNF0AKb8i2QOfdbCS6BJMntySps-8GfvJQ=s128", client_id: "" },
    { lottery_id: "SHIB5000", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 5000, icon_url: "https://mixin-images.zeromesh.net/fgSEd6CY07BiZP76--7JA9P-rKIWRoXD8Eis8RUL6mP85_QPsbMoyJtWJ6MjE9jWFEjabNF0AKb8i2QOfdbCS6BJMntySps-8GfvJQ=s128", client_id: "" },
    { lottery_id: "SHIB10000", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 10000, icon_url: "https://mixin-images.zeromesh.net/fgSEd6CY07BiZP76--7JA9P-rKIWRoXD8Eis8RUL6mP85_QPsbMoyJtWJ6MjE9jWFEjabNF0AKb8i2QOfdbCS6BJMntySps-8GfvJQ=s128", client_id: "" },
    { lottery_id: "DOGE1", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 1, icon_url: "https://mixin-images.zeromesh.net/gtz8ocdxuC4N2rgEDKGc4Q6sZzWWCIGDWYBT6mHmtRubLqpE-xafvlABX6cvZ74VXL4HjyIocnX-H_Vxrz3En9tMcIKED0c-2MhH=s128", client_id: "" },
    { lottery_id: "MOB0.1", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 0.1, icon_url: "https://mixin-images.zeromesh.net/eckqDQi50ZUCoye5mR7y6BvlbXX6CBzkP89BfGNNH6TMNuyXYcCUd7knuIDpV_0W7nT1q3Oo9ooVnMDGjl8-oiENuA5UVREheUu2=s128", client_id: "" },
    { lottery_id: "XIN0.1", asset_id: "965e5c6e-434c-3fa9-b780-c50f43cd955c", amount: 0.1, icon_url: "https://mixin-images.zeromesh.net/UasWtBZO0TZyLTLCFQjvE_UYekjC7eHCuT_9_52ZpzmCC-X-NPioVegng7Hfx0XmIUavZgz5UL-HIgPCBECc-Ws=s128", client_id: "" }
  ])
  const [activeReward, setActiveReward] = useState("")
  useEffect(() => {
    const run = createLucyLottery(rewards)
    run("SHIB10000",
      (params: any) => setActiveReward(params.lottery_id),
      (params: any) => {
        console.log(params)
        setActiveReward(params.lottery_id)
      })
  }, [])

  return <div>
    <div className={styles.lottery}>
      {rewards.map((reward) => <div
        key={reward.lottery_id}
        className={`${styles.reward} ${activeReward === reward.lottery_id ? styles.active : ''}`}
      >
        <img src={reward.icon_url} alt="" />
        <p>{reward.amount}</p>
      </div>)}
      <div className={styles.content}>
        test
      </div>
    </div>
  </div>
}
const nextIndex = [0, 1, 2, 3, 4, 6, 8, 10, 15, 14, 13, 12, 11, 9, 7, 5]
const nextMap = {
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
      if (Math.sqrt(current) <= (defaultSpeed - (allCount - counter))) {
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

// class LuckDraw {
//   DataArr: any
//   maxSpeed: number
//   myReq?: number
//   running?: any
//   runend?: any
//   constructor(DataArr: any) {
//     list = JSON.parse(JSON.stringify(DataArr))
//     maxSpeed = 4
//   }

//   run(id: string, running: any, runend: any) {
//   }
// }
