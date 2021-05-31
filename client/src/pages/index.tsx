import React, { useEffect, useState } from "react"
import { environment, getConversationId } from "@/assets/ts/tools"
import { history, useIntl } from "umi"
import { FullLoading, Loading } from "@/components/Loading"
import { get$t } from "@/locales/tools"
import { $get } from "@/stores/localStorage";

export default () => {
  const conversation_id = getConversationId()
  const [content, setContent] = useState("")
  const $t = get$t(useIntl())
  useEffect(() => {
    if (environment()) {
      checkGroup(conversation_id).then()
    } else {
      setContent($t("error.mixin"))
    }
  }, [])

  return (
    <>{content ? <Loading content={content} noCancel/> : <FullLoading/>}</>
  )
}

async function checkGroup(id: string | undefined) {
  // let nextPage = "/home"
  // if (id) nextPage = await getNextPage(id)
  const nextPage = $get('token') ? '/home' : '/join'
  history.push(nextPage)
}

async function getNextPage(id: string): Promise<string> {
  // const { is_owner, is_manager, group, setting } = await ApiCheckGroup(id)
  // let nextPage = ""
  // if (is_manager) {
  //   nextPage = "/manager"
  // } else {
  //   nextPage = is_owner ? "/home" : "/pre"
  // }
  // $set("is_manager", is_manager)
  // group && $set("group", group)
  // setting && $set("setting", setting)
  // return nextPage
  return ""
}
