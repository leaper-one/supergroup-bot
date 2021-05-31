import vConsole from "vconsole"
import { requestConfig } from "./apis/http"
import { $get, $set } from "@/stores/localStorage"

export const request = requestConfig
if (
  process.env.NODE_ENV === "development" &&
  navigator.userAgent.includes("Mixin")
)
  new vConsole()

$set("umi_locale", process.env.LANG === "zh" ? "zh-CN" : "en-US")

const version = $get("version")
if (version !== "1.0.0" && process.env.LANG === "en") {
  localStorage.clear()
  $set("version", "1.0.0")
}

export function modifyClientRenderOpts(memo: any) {
  return {
    ...memo,
    rootElement: memo.rootElement,
  }
}
