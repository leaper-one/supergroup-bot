import React, { useState } from "react"
import style from "./index.less"
import { history, useIntl } from "umi"
import { ApiCheckTransfer, ApiGenerateTransfer } from "@/apis/conversation"
import { payUrl } from "@/apis/http"
import { delay } from "@/assets/ts/tools"
import { Button } from "@/components/Sub"
import { get$t } from "@/locales/tools"

const getDesc = (
  isPayPage: boolean,
  $t: any,
): [string, string, string, string] => {
  return isPayPage
    ? [
        $t("pre.create.title"),
        $t("pre.create.desc"),
        $t("pre.create.button"),
        "",
      ]
    : [
        $t("pre.explore.title"),
        $t("pre.explore.desc"),
        $t("pre.explore.button"),
        "info",
      ]
}

let checkCircle = true
export default () => {
  const [isPayPage, setPayPage] = useState(false)
  const $t = get$t(useIntl())
  const [title, content, button, status] = getDesc(isPayPage, $t)
  return (
    <div className={style.container}>
      <img
        className={style.mainImg}
        src="https://taskwall.zeromesh.net/group-manager/pre-home.png"
        alt=""
      />
      <div className={style.content}>
        <h3>{title}</h3>
        <p>{content}</p>
      </div>
      <footer>
        {!isPayPage ? (
          <>
            <Button type="info" onClick={() => history.push("/pre/search")}>
              {button}
            </Button>
            {/*<span*/}
            {/*  className={style.create}*/}
            {/*  onClick={async () => {*/}
            {/*    const can_create = await checkPay()*/}
            {/*    !can_create && setPayPage(true)*/}
            {/*  }}*/}
            {/*>{$t("pre.create.action")}</span>*/}
          </>
        ) : (
          <Button
            onClick={() => {
              checkCircle = true
              pay()
            }}
          >
            {button}
          </Button>
        )}
      </footer>
    </div>
  )
}

const checkPay = async () => {
  const { can_create } = await ApiCheckTransfer()
  can_create && history.push("/create")
  return can_create
}

const pay = async () => {
  const { trace } = await ApiGenerateTransfer()
  window.location.href = payUrl({ trace, memo: "buy group" })
  let can_create = false
  while (!can_create) {
    if (!checkCircle) return
    can_create = await checkPay()
    await delay()
  }
}

// 初始化的页面： 开通社群
// 用户点击：创建社群 -> 支付 ->
