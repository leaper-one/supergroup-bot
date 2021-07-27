import { BackHeader } from "@/components/BackHeader"
import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { Modal } from "antd-mobile"
import { $get } from "@/stores/localStorage"
import { base64Encode, copy } from "@/assets/ts/tools"
import { useIntl } from "umi"
import { get$t } from "@/locales/tools"
import { FullLoading } from "@/components/Loading"

export default () => {
  const $t = get$t(useIntl())
  const join_url = `${location.origin}/join`
  const [loading, setLoading] = useState(false)
  const copyGroupURL = () => {
    copy(join_url)
    Modal.alert($t("action.tips"), $t("success.copy"), [{ text: $t("action.confirm") }])
  }
  const sendInviteCard = async () => {
    let schema = `mixin://send?category=app_card&data=`
    const { client_id: app_id, name: title, icon_url } = $get('group')
    schema += base64Encode({
      app_id,
      icon_url,
      title,
      description: $t("invite.desc"),
      action: join_url,
    })
    window.location.href = schema
  }

  useEffect(() => {
  }, [])

  return (
    <>
      <BackHeader name={$t("invite.title")} />
      <div className={styles.container}>
        <ul>
          <li onClick={sendInviteCard}>
            <i className="iconfont iconic_share" />
            <span>{$t("invite.card")}</span>
          </li>
          <li onClick={copyGroupURL}>
            <i className="iconfont iconfuzhiruqunlianjie" />
            <span>{$t("invite.link")}</span>
          </li>
        </ul>
        <p>{$t("invite.tip1")}</p>
      </div>
      {loading && <FullLoading mask />}
    </>
  )
}
