import { BackHeader } from "@/components/BackHeader"
import React from "react"
import styles from "./redRecord.less"

const redRecord = [
  {
    time: "03/28/2020",
    lists: [
      {
        avatar: "../../assets/img/avatar.png",
        type: "手气红包",
        memo: "3 个群 510 人抢到了红包",
        amount: "-0.00038",
        symbol: "BTC",
      },
      {
        avatar: "",
        type: "手气红包",
        memo: "3个切片",
        amount: "-0.00078",
        symbol: "BTC",
      },
    ],
  },
  {
    time: "12/12/2020",
    lists: [
      {
        avatar: "../../assets/img/avatar.png",
        type: "手气红包",
        memo: "3 个群 510 ",
        amount: "+0.00067",
        symbol: "EOS",
      },
      {
        avatar: "",
        type: "手气红包",
        memo: "3个哈哈哈哈",
        amount: "+0.00011",
        symbol: "SIT",
      },
    ],
  },
]

export default () => {
  return (
    <>
      <BackHeader name="红包记录" />
      <div className={styles.container}>
        <div>
          {redRecord.map((item, index) => (
            <>
              <div className={styles.time}>{item.time}</div>
              {item.lists.map((list, idx) => (
                <div key={idx} className={styles.records}>
                  <img src={require("../../assets/img/avatar.png")} alt="" />
                  <span className={styles.type}>{list.type}</span>
                  <div className={styles.desc}>
                    <span
                      className={`${
                        Number(list.amount) > 0 ? "green" : "red"
                      } ${styles.amount}`}
                    >
                      {list.amount}
                    </span>
                    <i>{list.symbol}</i>
                  </div>
                  <span className={styles.memo}>{list.memo}</span>
                </div>
              ))}
            </>
          ))}
        </div>
      </div>
    </>
  )
}
