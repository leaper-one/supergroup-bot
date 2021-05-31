import React from "react"
import styles from "./pre.less"
import { BackHeader } from "@/components/BackHeader"
import { history } from "umi"

export default () => {
  return (
    <>
      <BackHeader name="群发红包" />
      <div className={styles.container}>
        <div
          className={styles.item}
          onClick={() => history.push("/red?type=instant")}
        >
          <i className="iconfont iconjishihongbao" />
          <span>即时红包</span>
          <i className={`iconfont iconic_arrow ${styles.right}`} />
        </div>
        <p>立刻将红包发到多个群</p>
        {/*<div*/}
        {/*  className={styles.item}*/}
        {/*  onClick={() => history.push("/red?type=timing")}*/}
        {/*>*/}
        {/*  <i className="iconfont icondingshihongbao" />*/}
        {/*  <span>定时红包</span>*/}
        {/*  <i className={`iconfont iconic_arrow ${styles.right}`} />*/}
        {/*</div>*/}
        {/*<p>定时定期发红包，可设置连续多日多次发红包</p>*/}
      </div>
    </>
  )
}
