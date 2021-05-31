import { BackHeader } from "@/components/BackHeader"
import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { Modal } from "antd-mobile"
import { $get } from "@/stores/localStorage"
import { base64Encode, copy } from "@/assets/ts/tools"
import { IGroup } from "@/apis/group"
import { useIntl } from "umi"
import { get$t } from "@/locales/tools"
import { ApiGetInviteCount } from "@/apis/invite"
import { FullLoading } from "@/components/Loading"

export default () => {
  const $t = get$t(useIntl())
  const group: IGroup = $get("group")
  const join_url = `${location.origin}/join`
  const { invite_status } = $get("setting") || {}
  const open = invite_status === "1"

  const [count, setCount] = useState(0)
  const [loading, setLoading] = useState(true)
  const copyGroupURL = () => {
    copy(join_url)
    Modal.alert($t("action.tips"), $t("success.copy"))
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
    ApiGetInviteCount(group.group_id!).then((c) => {
      setCount(c.people)
      setLoading(false)
    })
  }, [])

  return (
    <>
      <BackHeader name={$t("invite.title")}/>
      <div className={styles.container}>
        <ul>
          <li onClick={sendInviteCard}>
            <i className="iconfont iconic_share"/>
            <span>{$t("invite.card")}</span>
          </li>
          <li onClick={copyGroupURL}>
            <i className="iconfont iconfuzhiruqunlianjie"/>
            <span>{$t("invite.link")}</span>
          </li>
        </ul>
        <p>{$t("invite.tip1")}</p>

        {/*<ul>*/}
        {/*  <li className={styles.my} onClick={() => history.push("/invite/my")}>*/}
        {/*    <i className="iconfont iconwodeyaoqing" />*/}
        {/*    <span className={styles.name}>{$t("invite.my.title")}</span>*/}
        {/*    <div>*/}
        {/*      <span>{count}</span>*/}
        {/*      <i className="iconfont iconic_arrow" />*/}
        {/*    </div>*/}
        {/*  </li>*/}
        {/*</ul>*/}
        {/*<p*/}
        {/*  className={open ? styles.red : ""}*/}
        {/*  dangerouslySetInnerHTML={{*/}
        {/*    __html: $t(`invite.tip${open ? "" : "Not"}Open`),*/}
        {/*  }}*/}
        {/*/>*/}
      </div>
      {loading && <FullLoading mask/>}
    </>
  )
}
