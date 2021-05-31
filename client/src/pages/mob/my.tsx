import { BackHeader } from "@/components/BackHeader"
import React from "react"
import styles from "./my.less"

const list = [
  {
    icon_url:
      "https://ss0.bdstatic.com/70cFuHSh_Q1YnxGkpoWK1HF6hhy/it/u=2919410099,2874855165&fm=26&gp=0.jpg",
    title: "空投奖励",
    symbol: "MOB",
    amount: "+0.000038",
  },
]

export default () => {
  return (
    <div className={styles.container}>
      <BackHeader name="我的空投" />

      {list.length === 0 ? (
        <div>
          <img
            src="https://taskwall.zeromesh.net/group-manager/no-kong.png"
            alt=""
          />
          <p>没有空投</p>
        </div>
      ) : (
        <ul>
          {list.map((item, idx) => (
            <li key={idx} className="flex">
              <div className="flex">
                <img src={item.icon_url} alt="" />
                <span>{item.title}</span>
              </div>
              <div className="flex">
                <span className="green">{item.amount}</span>
                <i>{item.symbol}</i>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
