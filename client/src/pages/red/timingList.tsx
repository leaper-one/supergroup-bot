import React from "react"
import styles from "./timingList.less"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { SwipeAction } from "antd-mobile"
import { Button, Confirm } from "@/components/Sub"

export default () => {
  const $t = get$t(useIntl())
  return (
    <div className={styles.container}>
      <BackHeader name={$t("red.timingTitle")} />

      <ul className={styles.list}>
        <SwipeAction
          autoClose
          right={[
            {
              text: "取消",
              style: {
                backgroundColor: "#FA596D",
                color: "white",
                width: "100px",
                height: "50px",
              },
              onPress: async () => {
                const isConfirm = await Confirm(
                  "提示",
                  "是否取消这个定时红包？",
                )
                console.log(isConfirm)
              },
            },
          ]}
        >
          <li className={styles.item}>
            <img
              src="https://mixin-images.zeromesh.net/0sQY63dDMkWTURkJVjowWY6Le4ICjAFuu3ANVyZA4uI3UdkbuOT5fjJUT82ArNYmZvVcxDXyNjxoOv0TAYbQTNKS=s128"
              alt=""
            />
            <p className={styles.name}>0.0001 BOX</p>
            <span className={styles.status}>进行中</span>
            <span className={styles.info}>普通红包，每周三上午2:00</span>
            <span className={styles.process}>0/4</span>
          </li>
        </SwipeAction>
      </ul>

      <footer className={styles.footer}>
        <Button>添加定时红包</Button>
        <p>最多可添加 10 个定时红包</p>
      </footer>
    </div>
  )
}
