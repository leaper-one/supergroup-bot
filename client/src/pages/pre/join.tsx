import React, { useEffect, useState } from "react"
import { ApiGetGroup } from "@/apis/group"
import { history, useIntl } from "umi"
import { IJoin, Join } from "@/components/Join"
import { CodeURL } from "@/components/CodeURL"
import { environment, setHeaderTitle } from "@/assets/ts/tools"
import { mainJoin, } from "@/pages/pre/joinData"
import { get$t } from "@/locales/tools"
import { $get } from "@/stores/localStorage"
import BigNumber from 'bignumber.js'

export default () => {
  const $t = get$t(useIntl())
  const [joinProps, setJoinProps] = useState<IJoin>()
  const mixinCtx = environment()
  const handleClickBtn = () => history.push(`/auth`)

  const initPage = async () => {
    const group = await ApiGetGroup()
    setTimeout(() => {
      setHeaderTitle(group.name)
    })
    if (from === "auth") handleClickBtn()
    group.total_people = `${new BigNumber(group.total_people).toFormat()} ${$t('join.main.member')}`
    return setJoinProps(mainJoin(group, handleClickBtn, $t))
  }

  const from = history.location.query?.from
  useEffect(() => {
    if ($get('token')) return history.push(`/`)
    initPage()
  }, [])

  return <>
    {mixinCtx ? (
      <Join props={{ ...joinProps, loading: false } as IJoin} />
    ) : (
      <CodeURL groupInfo={joinProps?.groupInfo} action="join" />
    )}
  </>

}
