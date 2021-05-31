import React, { useEffect, useState } from "react"
import styles from "./index.less"
import managerStyles from "@/pages/manager/index.less"
import { BackHeader } from "@/components/BackHeader"
import { getAuthUrl, staticUrl } from "@/apis/http"
import { history, useIntl } from "umi"
import VConsole from "vconsole"
import { getConversationId, getMixinCtx, setHeaderTitle } from "@/assets/ts/tools"
import { ApiCheckGroup } from "@/apis/conversation"
import { get$t } from "@/locales/tools"
import { ApiGetGroup } from "@/apis/group"
import { Confirm } from "@/components/Sub"
import { $get, $set } from "@/stores/localStorage";

export default () => {
  let t = 0
  const { asset_id, information_url, name } = $get("group") || {}
  const { avatar_url } = $get("user") || {}
  const [isImmersive, setImmersive] = useState(true)
  const $t = get$t(useIntl())
  useEffect(() => {
    ApiGetGroup().then(group => {
      $set('group', group)
    })
    const { immersive } = getMixinCtx() || {}
    if (immersive === false) setImmersive(false)
    setTimeout(() => {
      setHeaderTitle(name)
    })
  }, [])

  return (
    <div className={styles.mainBox}>
      {isImmersive === true && <BackHeader
        name={name}
        onClick={() => {
          if (t === 20) {
            new VConsole()
          } else if (t === 45) {
            ApiCheckGroup(getConversationId()!).then(console.log)
            history.push("/manager")
          }
          t++
        }}
        noBack
        action={
          <>
            {avatar_url ? (
              <img
                onClick={async () => {
                  const isConfirm = await Confirm("确认重新授权？")
                  isConfirm && (window.location.href = getAuthUrl())
                }}
                className={styles.avatar}
                src={avatar_url}
                alt=""
              />
            ) : (
              <i
                onClick={() => (window.location.href = getAuthUrl())}
                className={`iconfont iconshouquandenglu ${styles.avatar}`}
              />
            )}
          </>
        }
      />}
      <ul className={`${styles.container} ${managerStyles.index}`}>

        {information_url && <li onClick={() => location.href = (information_url)}>
          <img src={staticUrl + "home_7.png"} alt=""/>
          <p>{$t("home.article")}</p>
        </li>}
        {asset_id && (
          <li onClick={() => history.push("/transfer/" + asset_id)}>
            <img src={staticUrl + "home_0.png"} alt=""/>
            <p>{$t("home.trade")}</p>
          </li>
        )}
        <li onClick={() => history.push("/invite")}>
          <img src={staticUrl + "home_1.png"} alt=""/>
          <p>{$t("home.invite")}</p>
        </li>
        {/*<li onClick={() => history.push("/explore")}>*/}
        {/*  <img src={staticUrl + "home_3.png"} alt=""/>*/}
        {/*  <p>{$t("home.findGroup")}</p>*/}
        {/*</li>*/}
        <li onClick={() => history.push("/findBot")}>
          <img src={staticUrl + "home_5.png"} alt=""/>
          <p>{$t("home.findBot")}</p>
        </li>

        {/*{!isEnglish && (*/}
        {/*  <li onClick={() => history.push("/more")}>*/}
        {/*    <img src={staticUrl + "home_6.png"} alt=""/>*/}
        {/*    <p>{$t("home.more")}</p>*/}
        {/*  </li>*/}
        {/*)}*/}
      </ul>
    </div>
  )
}

// export const getCurrentGroup = async (): Promise<IGroupItem | undefined> => {
// const data = await ApiGetGroupList()
// const cur = $get("group").group_id
// return data.find((item) => item.group_id === cur)
// }
