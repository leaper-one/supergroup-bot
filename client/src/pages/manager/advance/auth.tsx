import React, { useEffect, useState } from "react"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { ApiGetGroupMemberAuth } from '@/apis/user'
import { $get, $set } from '@/stores/localStorage'
import { Manager, List } from '..'
import { FullLoading } from '@/components/Loading'


function getManagerList($t: any): Manager[][] {
  const group = $get("group")
  const managerList: Manager[][] = [
    [
      {
        icon: "feihuiyuan",
        type: $t("advance.member.1"),
        route: "/manager/advance/authDesc?s=1",
      },
    ],
    [
      {
        icon: "chujihuiyuan",
        type: $t("advance.member.2"),
        route: "/manager/advance/authDesc?s=2",
      },
    ],
    [
      {
        icon: "gaojihuiyuan",
        type: $t("advance.member.5"),
        route: "/manager/advance/authDesc?s=5",
      },
    ],
  ]
  if (group.speak_status === 1) managerList.shift()
  return managerList
}

export default () => {
  const [managerList, setManagerList] = useState<any[]>([])
  const [isLoaded, setIsLoaded] = useState(false)
  const $t = get$t(useIntl())

  useEffect(() => {
    initPage()
  }, [])
  const initPage = async () => {
    setManagerList(getManagerList($t))
    const auth = await ApiGetGroupMemberAuth()
    setIsLoaded(true)
    $set("auth", auth)
  }

  return (
    <>
      <BackHeader name={$t("advance.msgAuth")} />
      <List lists={managerList} />
      {!isLoaded && <FullLoading opacity mask />}
    </>
  )
}
