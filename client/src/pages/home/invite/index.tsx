import { BackHeader } from "@/components/BackHeader"
import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { Modal } from "antd-mobile"
import { $get, $set } from "@/stores/localStorage"
import { base64Encode, copy } from "@/assets/ts/tools"
import { useIntl, history } from "umi"
import { get$t } from "@/locales/tools"
import { FullLoading } from "@/components/Loading"
import { Icon } from "@/components/Icon"
import { ApiGetInvitation, IInvitationResp } from '@/apis/invite'

export default () => {
  const $t = get$t(useIntl())
  const [loading, setLoading] = useState(true)
  const [invitation, setInvitation] = useState<IInvitationResp>()
  const copyGroupURL = () => {
    copy(getInviteUrl())
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
      action: getInviteUrl(),
    })
    window.location.href = schema
  }

  const getInviteUrl = () => `${location.origin}/join?c=${invitation?.code}`

  useEffect(() => {
    initPage()
  }, [])
  const initPage = async () => {
    const invitation = await ApiGetInvitation()
    setInvitation(invitation)
    $set('invitation', invitation)
    setLoading(false)
  }

  return (
    <>
      <BackHeader name={$t("invite.title")} />
      <div className={styles.container}>
        <ul>
          <li onClick={sendInviteCard}>
            <Icon i="ic_share" />
            <span>{$t("invite.card")}</span>
          </li>
          <li onClick={copyGroupURL}>
            <Icon i="fuzhiruqunlianjie" />
            <span>{$t("invite.link")}</span>
          </li>
        </ul>
        <p>{$t("invite.tip1")}</p>
        <ul>
          <li className={styles.my} onClick={() => history.push(`/invite/my`)}>
            <span>
              <Icon i="ic_yaoqing" />
              <span>{$t("invite.my.title")}</span>
            </span>
            <div className={styles.right}>
              <span>{invitation?.count || 0}</span>
              <Icon i="ic_arrow" />
            </div>
          </li>
        </ul>
        <p className={styles.red}>{$t("invite.tip2")}</p>
        <p className={styles.red}>{$t("invite.tip3")}</p>
      </div>
      {loading && <FullLoading mask />}
    </>
  )
}

