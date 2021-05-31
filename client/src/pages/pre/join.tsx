import React, { useEffect, useState } from "react"
import { ApiGetGroup } from "@/apis/group"
import { history, useIntl } from "umi"
import { IJoin, Join } from "@/components/Join"
import { CodeURL } from "@/components/CodeURL"
import { environment, setHeaderTitle } from "@/assets/ts/tools"
import { mainJoin, } from "@/pages/pre/joinData"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import styles from "./join.less"
import { $get } from "@/stores/localStorage";

export default () => {
  const $t = get$t(useIntl())
  const [joinProps, setJoinProps] = useState<IJoin>()
  const mixinCtx = environment()
  const handleClickBtn = () => history.push(`/auth`)

  const initPage = async () => {
    const { name, icon_url, description } = await ApiGetGroup()
    setTimeout(() => {
      setHeaderTitle(name)
    })
    if (from === "auth") handleClickBtn()
    return setJoinProps(mainJoin({ name, icon_url: icon_url as string, description }, handleClickBtn, $t))
  }

  const from = history.location.query?.from
  useEffect(() => {
    if ($get('token')) return history.push(`/`)
    initPage()
  }, [])

  return (
    <div
      className={`${styles.container} ${
        from === "search" ? styles.hasHeader : ""
      }`}
    >
      {from === "search" && (
        <BackHeader name={joinProps?.groupInfo?.name || $t("join.title")}/>
      )}
      {mixinCtx ? (
        <Join props={{ ...joinProps, loading: false } as IJoin}/>
      ) : (
        <CodeURL groupInfo={joinProps?.groupInfo} action="join"/>
      )}
    </div>
  )
}
