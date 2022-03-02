import vConsole from "vconsole"
import { requestConfig } from "./apis/http"
import { $get, $set } from "@/stores/localStorage"

export const request = requestConfig
if (
  process.env.NODE_ENV === "development" &&
  navigator.userAgent.includes("Mixin")
)
  new vConsole()

$set("umi_locale", process.env.LANG)
export function modifyClientRenderOpts(memo: any) {
  return {
    ...memo,
    rootElement: memo.rootElement,
  }
}
