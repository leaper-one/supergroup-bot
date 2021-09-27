import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { BackHeader } from "@/components/BackHeader"
import { getAddUserURL, getAuthUrl, staticUrl } from "@/apis/http"
import { history, useIntl } from "umi"
import {
  environment,
  getConversationId,
  getMixinCtx,
  setHeaderTitle,
} from "@/assets/ts/tools"
import { ApiCheckGroup } from "@/apis/conversation"
import { get$t } from "@/locales/tools"
import { ApiGetGroup, IGroupInfo1 } from "@/apis/group"
import { $get, $set } from "@/stores/localStorage"
import BigNumber from "bignumber.js"
import { ApiGetMe } from "@/apis/user"
import { GlobalData } from "@/stores/store"
import { JoinModal } from "@/components/PopupModal/join"
import { Modal } from "antd-mobile"

export default () => {
  let t = 0
  const userCache = $get("_user") || {}
  const [isImmersive, setImmersive] = useState(true)
  const [group, setGroup] = useState<IGroupInfo1>($get("group"))
  const [modal, setModal] = useState(false)
  const [hasAsset, setHasAsset] = useState($get("hasAsset") || false)
  const [avatarUrl, setAvatarUrl] = useState(() => userCache.avatar_url)
  const [isClaim, setIsClaim] = useState(() => userCache.is_claim)
  const [isBlock, setIsBlock] = useState(() => userCache.is_block)

  const $t = get$t(useIntl())

  useEffect(() => {
    if (!environment() || !$get("token")) {
      history.push(`/join`)
      return
    }
    ApiGetGroup().then((group) => {
      $set("group", group)
      setHasAsset(!!group.asset_id)
      $set("hasAsset", !!group.asset_id)
      if (!group.asset_id) {
        group.asset_id =
          group.client_id === "47cdbc9e-e2b9-4d1f-b13e-42fec1d8853d"
            ? "c94ac88f-4671-3976-b60a-09064f1811e8"
            : "c6d0c728-2624-429b-8e0d-d9d19b6592fa"
      }

      setGroup(group)
      ApiGetMe().then((user) => {
        $set("_user", user)
        setAvatarUrl(user.avatar_url)
        setIsBlock(user.is_block)
        setIsClaim(user.is_claim)
      })
      if (GlobalData.isNewUser)
        setTimeout(() => {
          setModal(true)
          GlobalData.isNewUser = false
        })
    })
    const { immersive } = getMixinCtx() || {}
    if (immersive === false) setImmersive(false)
    setTimeout(() => {
      setHeaderTitle(group?.name || "")
    })
  }, [])

  let price = 0
  if (group) {
    const usd = Number(group.price_usd)
    if (usd < 10) {
      price = Number(usd.toFixed(4))
    } else {
      price = Number(usd.toFixed(2))
    }
  }

  return (
    <div className={styles.mainBox}>
      {isImmersive && (
        <BackHeader
          name={group?.name}
          onClick={() => {
            if (t === 20) {
            } else if (t === 45) {
              ApiCheckGroup(getConversationId()!).then(console.log)
              history.push("/manager")
            }
            t++
          }}
          noBack
          action={
            <>
              {avatarUrl ? (
                <i
                  onClick={() => {
                    const user = $get("_user")
                    let route =
                      user && user.status === 9
                        ? "/manager/setting"
                        : "/setting"
                    history.push(route)
                  }}
                  className={`iconfont iconic_unselected_5 ${styles.avatar}`}
                />
              ) : (
                <i
                  onClick={() => (window.location.href = getAuthUrl())}
                  className={`iconfont iconshouquandenglu ${styles.avatar}`}
                />
              )}
            </>
          }
        />
      )}
      <div className={styles.statistic}>
        <img
          className={styles.bg}
          src={require("@/assets/img/asset_bg.png")}
          alt=""
        />
        <div className={styles.content}>
          <div className={styles.content_item}>
            <span className={styles.title}>
              {group?.symbol} {$t("transfer.price")}
            </span>
            <span
              className={`${styles.price} ${price === 0 && styles.priceZero}`}
            >
              {price === 0 ? $t("transfer.noPrice") : `$ ${price}`}
            </span>
            <span
              className={`${styles.rate} ${
                Number(group?.change_usd) > 0 ? styles.green : styles.red
              }`}
            >
              {Number((Number(group?.change_usd) * 100).toFixed(2))}%
            </span>
          </div>
          <div className={`${styles.content_item} ${styles.right}`}>
            <span className={styles.title}>{$t("home.people_count")}</span>
            <span className={styles.people}>
              {new BigNumber(group?.total_people).toFormat()}
            </span>
            <span className={styles.info}>
              {$t("home.week")} +{group?.week_people}
            </span>
          </div>
        </div>
      </div>
      <div className={styles.navList}>
        <div className={styles.navItem} onClick={() => history.push(`/invite`)}>
          <div className={styles.navItemInner}>
            <img src={require("@/assets/img/invite.png")} alt="" />
          </div>
          <p>{$t("home.invite")}</p>
        </div>
        <div
          className={styles.navItem}
          onClick={() =>
            window.open(
              `mixin://apps/${
                process.env.RED_PACKET_ID
              }?action=open&conversation=${getConversationId()}`,
            )
          }
        >
          <div className={styles.navItemInner}>
            <img src={require("@/assets/img/red-packet.png")} alt="" />
          </div>
          <p>{$t("home.redPacket")}</p>
        </div>
        {group?.has_reward && (
          <div
            className={styles.navItem}
            onClick={() => history.push(`/reward`)}
          >
            <div className={styles.navItemInner}>
              <img src={require("@/assets/img/reward.png")} alt="" />
            </div>
            <p>{$t("home.reward")}</p>
          </div>
        )}
        {!isBlock && (
          <div
            className={styles.navItem}
            onClick={() => history.push(`/lottery`)}
          >
            <div
              className={`${styles.navItemInner} ${
                isClaim === false && styles.lottery
              }`}
            >
              <img src={require("@/assets/img/reward.png")} alt="" />
            </div>
            <p>{$t("claim.lottery")}</p>
          </div>
        )}
        <div
          className={styles.navItem}
          onClick={() => history.push("/activity")}
        >
          <div className={styles.navItemInner}>
            <img src={require("@/assets/img/active.png")} alt="" />
          </div>
          <p>{$t("home.activity")}</p>
        </div>
        {group?.speak_status === 1 && (
          <div
            className={styles.navItem}
            onClick={() => history.push(`/member`)}
          >
            <div className={styles.navItemInner}>
              <img src={require("@/assets/img/member-icon.png")} alt="" />
            </div>
            <p>{$t("member.center")}</p>
          </div>
        )}
        <div
          className={styles.navItem}
          onClick={() => (location.href = getAddUserURL(group?.client_id))}
        >
          <div className={styles.navItemInner}>
            <img src={require("@/assets/img/open-chat.png")} alt="" />
          </div>
          <p>{$t("home.open")}</p>
        </div>
      </div>
      <ul className={`${styles.container} ${styles.index}`}>
        <li onClick={() => history.push(`/news`)}>
          <img src={staticUrl + "home_7.png"} alt="" />
          <p>{$t("home.article")}</p>
        </li>
        {group && group.asset_id && (
          <li onClick={() => history.push("/transfer/" + group.asset_id)}>
            <img src={staticUrl + "home_0.png"} alt="" />
            <p>{$t("home.trade")}</p>
          </li>
        )}
        <li onClick={() => history.push("/explore")}>
          <img src={staticUrl + "home_3.png"} alt="" />
          <p>{$t("home.findGroup")}</p>
        </li>
        <li onClick={() => history.push("/findBot")}>
          <img src={staticUrl + "home_5.png"} alt="" />
          <p>{$t("home.findBot")}</p>
        </li>
      </ul>
      <Modal
        visible={modal}
        animationType="slide-up"
        popup
        onClose={() => setModal(false)}
      >
        <JoinModal
          modalProp={{
            title: `${group?.name} ${$t("home.joinSuccess")}`,
            desc: group?.description,
            icon_url: group?.icon_url,
            button: $t("home.enterChat"),
            buttonAction: () => {
              location.href = getAddUserURL(group?.client_id)
            },
            tips: $t("home.enterHome"),
            tipsAction: () => setModal(false),
            tipsStyle: "blank-btn",
            isAirdrop: true,
          }}
        />
      </Modal>
    </div>
  )
}
