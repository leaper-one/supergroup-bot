import React from "react"
import { history } from "umi"
import { getAuthUrl } from "@/apis/http"
import { FullLoading } from "@/components/Loading"
import { ApiAuth } from "@/apis/user"
import { $set } from "@/stores/localStorage"
import qs from "qs"
import { ToastFailed } from "@/components/Sub";
import { ApiGetGroup } from "@/apis/group";

export default () => {
  const query: any = history.location.query
  const { code, return_to } = query
  if (code) {
    auth(code, return_to)
  } else {
    window.location.href = getAuthUrl()
  }
  return <FullLoading/>
}

async function auth(code: string, return_to: string) {
  const { authentication_token, ...user } = await ApiAuth(code)
  if (!authentication_token) {
    ToastFailed('认证失败...')
    history.push(`/auth`)
    return
  }
  $set("token", authentication_token)
  $set("user", user)
  const group = await ApiGetGroup()
  $set('group', group)
  let url = return_to ? decodeURIComponent(return_to) : "/"
  let pathname = url
  let query = {}
  if (url.includes("?")) {
    const [_pathname, _query] = url.split("?")
    query = qs.parse(_query)
    pathname = _pathname
  }
  history.push({ pathname, query: { ...query, from: "auth" } })
}
