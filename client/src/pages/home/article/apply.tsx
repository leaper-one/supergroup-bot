import React, { useEffect, useState } from "react"
import styles from "./apply.less"
import { BackHeader } from "@/components/BackHeader"
import { history, useIntl } from "umi"
import { Button, ToastFailed } from "@/components/Sub"
import { Modal } from "antd-mobile"
import { JoinModal } from "@/components/PopupModal/join"
import { ApiPostArticle } from "@/apis/article"
import { getGroupID } from "@/apis/group"
import { payUrl } from "@/apis/http"
import { Loading } from "@/components/Loading"
import { checkPaid } from "@/pages/manager/asset/assetChange"
import { get$t } from "@/locales/tools"
import { ApiGetAssetByID } from "@/apis/asset"

export default () => {
  const [showLimit, setShowLimit] = useState(false)
  const [title, setTitle] = useState("")
  const [link, setLink] = useState("")
  const [keyWord, setKeyWord] = useState("")
  const [loading, setLoading] = useState(false)

  const [showTips, setShowTips] = useState(false)
  const [showPay, setShowPay] = useState(false)
  const [showSuccess, setShowSuccess] = useState(false)

  const [mobPrice, setMobPrice] = useState("35")

  const $t = get$t(useIntl())

  const clickApply = async () => {
    if (!title) return ToastFailed("标题不能为空")
    if (!link) return ToastFailed("链接不能为空")
    if (!keyWord) return ToastFailed("关键字不能为空")
    if (keyWord.length > 30) return ToastFailed("关键字不能超过30个字")
    const key = keyWord.toLowerCase()
    if (!key.includes("mob") && !key.includes("mobilecoin"))
      return ToastFailed("请检查关键字")
    setShowTips(true)
  }

  const clickPay = async () => {
    const { code, trace_id, asset_id } = await ApiPostArticle({
      group_id: getGroupID(),
      title,
      link,
      key_word: keyWord,
    })
    if (code === 403) {
      ToastFailed("请入群 7 天后参与...")
    } else if (code === 405) {
      ToastFailed("链接已被占用")
    } else if (code === 20117) {
      setShowLimit(true)
    } else if (trace_id) {
      location.href = payUrl({
        trace: trace_id,
        asset: asset_id,
        amount: "0.01",
        memo: JSON.stringify({
          M: "article deposit",
          G: getGroupID(),
          O: "article",
        }),
      })
      setLoading(true)
      checkPaid(
        "0.01",
        asset_id,
        process.env.CLIENT_ID!,
        trace_id,
        setLoading,
        $t,
        () => {
          setLoading(false)
          setShowSuccess(true)
        },
      )
    }
  }

  useEffect(() => {
    ApiGetAssetByID("eea900a8-b327-488c-8d8d-1428702fe240").then((res) => {
      if (res.price_usd) setMobPrice(res.price_usd)
    })
  }, [])

  return (
    <div>
      <BackHeader
        name="申请参与"
        action={
          <i
            className={`iconfont iconbangzhu ${styles.help}`}
            onClick={() =>
              (location.href = `https://w3c.group/c/1618883636346675`)
            }
          />
        }
      />

      <ul className={styles.form}>
        <li className={styles.formItem}>
          <input
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            type="text"
            placeholder={"例如：MobileCoin 购买指南"}
          />
          <span>标题</span>
        </li>
        <li className={styles.formItem}>
          <input
            value={keyWord}
            onChange={(e) => setKeyWord(e.target.value)}
            type="text"
            placeholder={"搜索引擎关键字"}
          />
          <span>关键字</span>
        </li>
        <li className={styles.formItem}>
          <input
            value={link}
            onChange={(e) => setLink(e.target.value)}
            type="text"
            placeholder={"文章链接"}
          />
          <span>链接</span>
        </li>
      </ul>

      <ul className={styles.tips}>
        <li>文章内容必须包含 MobileCoin 持仓群链接</li>
        <li>文章内容必须包含 4swap 机器人ID 7000103537</li>
        <li>文章内容必须包含 4swap 交易链接 https://x.firesbox.com/9L1I2</li>
        <li>关键字必须包含 MOB 或者 MobileCoin</li>
        <li>关键字必须小于 30 个字符</li>
        <li>通过关键字搜索文章必须在搜索结果位置 100 以内</li>
      </ul>
      <p className={styles.tips2}>
        推荐在知乎、CSDN、博客园等权重较高的平台发表文章。
      </p>

      <Button className={styles.btn} onClick={clickApply}>
        申请
      </Button>

      <p className={styles.tips3}>
        参与活动需要支付 0.01 MOB 保证金，更多细节 参见 <a href="">文档</a> 。
      </p>

      <Modal
        className={styles.payModal}
        visible={showPay}
        animationType="slide-up"
        popup
      >
        <div className={styles.assetIcon}>
          <img
            src="https://mixin-images.zeromesh.net/JtSsCbZUzBpdDEI6JbLZ1-ZdAWUakaLBSpAp25gu0uHKoHC3kAeDTXZhsgMEOk_i3nFSAKI4QqFTEEPqv31QFD-hDQHpGA2zoG_A=s128"
            alt=""
          />
          <img
            src="https://mixin-images.zeromesh.net/JtSsCbZUzBpdDEI6JbLZ1-ZdAWUakaLBSpAp25gu0uHKoHC3kAeDTXZhsgMEOk_i3nFSAKI4QqFTEEPqv31QFD-hDQHpGA2zoG_A=s128"
            alt=""
          />
        </div>
        <div className={styles.amount}>
          <h4>0.01 MOB</h4>
          <p>≈ {(Number(mobPrice) * 0.01).toFixed(2)} USD</p>
        </div>

        <p className={styles.content}>
          为了减少水文恶意申请，参与活动需要支付 0.01 MOB 保证金，自申请 24
          小时后一旦确认搜索排名立刻返还，更多细节参见{" "}
          <a href="https://w3c.group/c/1618883636346675">文档</a> 。
        </p>

        <Button className={styles.payBtn} onClick={() => clickPay()}>
          支付
        </Button>
        <p className={styles.payTips} onClick={() => setShowPay(false)}>
          稍后
        </p>
      </Modal>

      <Modal
        visible={showLimit}
        animationType="slide-up"
        popup
        onClose={() => setShowLimit(false)}
      >
        <JoinModal
          modalProp={{
            title: "投稿已达上限",
            desc: "投稿已达上限，感谢您的支持。",
            descStyle: "red",
            button: "好的",
            buttonAction: () => setShowLimit(false),
            icon: "tougaoyidashangxian",
          }}
        />
      </Modal>

      <Modal
        visible={showTips}
        animationType="slide-up"
        popup
        onClose={() => setShowTips(false)}
      >
        <JoinModal
          modalProp={{
            title: "申请须知",
            desc: "",
            content: (
              <div className={styles.tipsModal}>
                <p>请确保您的申请符合以下条件</p>
                <ul>
                  <li>文章内容必须包含 4swap 机器人ID</li>
                  <li>文章内容必须包含 4swap 交易链接</li>
                  <li>文章内容包含 MobileCoin 持仓群 入群链接</li>
                  <li>文章内容包含自己的 Mixin ID</li>
                  <li>
                    确保自己的文章已经被百度或 Google
                    收录，且可以通过关键字搜索到。
                  </li>
                </ul>
                <p>申请一旦提交无法修改也无法撤销。</p>
              </div>
            ),
            button: "继续",
            buttonAction: () => {
              setShowTips(false)
              setTimeout(() => setShowPay(true), 50)
            },
            tips: "稍后",
            tipsStyle: "blank",
            tipsAction: () => setShowTips(false),
            icon: "shenqingxuzhi",
          }}
        />
      </Modal>

      <Modal
        visible={showSuccess}
        animationType="slide-up"
        popup
        onClose={() => setShowSuccess(false)}
      >
        <JoinModal
          modalProp={{
            title: "申请提交成功",
            desc:
              "感谢参与本次 #写文赚币# 活动，你的文章申请已提交成功，请打开推送留意社群助手最新消息通知。",
            button: "好的",
            buttonAction: () => {
              setShowSuccess(false)
              history.push(`/article/my`)
            },
            isAirdrop: true,
            icon_url:
              "https://mixin-images.zeromesh.net/JtSsCbZUzBpdDEI6JbLZ1-ZdAWUakaLBSpAp25gu0uHKoHC3kAeDTXZhsgMEOk_i3nFSAKI4QqFTEEPqv31QFD-hDQHpGA2zoG_A=s128",
          }}
        />
      </Modal>

      {loading && <Loading cancel={() => setLoading(false)} />}
    </div>
  )
}
