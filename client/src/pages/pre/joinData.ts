import { getAuthUrl, getCodeUrl } from "@/apis/http"
import { ICheck } from "@/apis/conversation"
import { IGroup } from "@/apis/group"
import { history } from "umi"
import { ToastFailed } from "@/components/Sub"

export const modalNoAuthData = ($t: any) => ({
  title: $t("join.modal.auth"),
  desc: $t("join.modal.authDesc"),
  icon: "shouquanjiance",
  button: $t("join.modal.authBtn"),
  descStyle: "red",
  buttonAction: () => (window.location.href = getAuthUrl()),
})

export const modalForbidData = (status: string, setShowModal: any) => ({
  title: "禁止加入群组",
  icon: "jinzhijiaru",
  desc:
    status === "8"
      ? "24 小时内不能加入群组，请联系管理员或等 24 小时再进入群。"
      : "你被禁止入群，想要加入群组请联系管理员。",
  button: "知道了",
  descStyle: "red",
  buttonAction: () => setShowModal(false),
})

let activeAsset: any
export const modalNoShares = (
  check_list: Array<ICheck>,
  handleClickBtn: any,
  $t: any,
) => ({
  icon: "chicangjiance1",
  title: $t("join.modal.shares"),
  desc: getCheckMsg(check_list, $t),
  button: $t("join.modal.sharesBtn"),
  descStyle: "red",
  buttonAction: async () => {
    let t = await handleClickBtn()
    t && ToastFailed($t("join.modal.sharesFail"), 2)
  },
  tips: $t("join.modal.sharesTips"),
  tipsStyle: "blank",
  tipsAction: () => history.push(`/transfer/${activeAsset}`),
})

export const modalAppointSuccess = (
  icon_url: string,
  setShowModal: any,
  $t: any,
) => ({
  icon_url,
  title: $t("join.modal.appoint"),
  desc: $t("join.modal.appointDesc"),
  descStyle: "blank",
  button: $t("join.modal.appointBtn"),
  buttonStyle: "submit",
  isAirdrop: true,
  buttonAction: () => {
    setShowModal(false)
    window.location.href = `mixin://users/${process.env.CLIENT_ID}`
  },
  tips: $t("join.modal.appointTips"),
})

export const modalReceive = (
  groupInfo: IGroup,
  clickModalReceived: any,
  $t: any,
) => ({
  icon_url: groupInfo.icon_url,
  title: $t("join.modal.receive"),
  desc: $t("join.modal.receiveDesc"),
  descStyle: "blank",
  button: $t("join.modal.receiveBtn"),
  buttonStyle: "submit",
  isAirdrop: true,
  buttonAction: () => clickModalReceived(groupInfo),
})

export const modalReceived = (groupInfo: IGroup, comment: string, $t: any) => ({
  icon_url: groupInfo.icon_url,
  title: $t("join.modal.receive"),
  desc: $t("join.modal.receivedDesc", { comment }),
  descStyle: "blank",
  button: $t("join.modal.receivedBtn"),
  disabled: true,
  isAirdrop: true,
  buttonAction: () => {},
})

export const mainJoin = (groupInfo: IGroup, handleClickBtn: any, $t: any) => ({
  groupInfo,
  button: $t("join.main.join"),
  buttonAction: handleClickBtn,
  tips: $t("join.main.joinTips"),
})

export const mainAirdropAppoint = (
  groupInfo: IGroup,
  clickAppointment: any,
  $t: any,
) => ({
  groupInfo,
  button: $t("join.main.appointBtn"),
  buttonAction: () => clickAppointment(groupInfo),
})

export const mainAirdropAppointed = (groupInfo: IGroup, $t: any) => ({
  groupInfo,
  buttonAction: () => {},
  button: $t("join.main.appointedBtn"),
  tips: $t("join.main.appointedTips"),
  disabled: true,
  tipsStyle: "blank",
  tipsAction: () =>
    (window.location.href = `mixin://users/${process.env.CLIENT_ID}`),
})

export const mainAirdropReceive = (
  groupInfo: IGroup,
  clickReceived: any,
  $t: any,
) => ({
  groupInfo,
  button: $t("join.main.receiveBtn"),
  buttonAction: () => clickReceived(groupInfo),
})

export const mainAirdropReceived = (
  groupInfo: IGroup,
  code_id: string,
  $t: any,
) => ({
  groupInfo,
  button: $t("join.main.receivedBtn"),
  disabled: true,
  buttonAction: () => {},
  tips: $t("join.main.receivedTips"),
  tipsStyle: "blank",
  tipsAction: () => (window.location.href = getCodeUrl(code_id)),
})

export const mainAirdropNoAccess = (groupInfo: IGroup, $t: any) => ({
  groupInfo,
  button: $t("join.main.noAccess"),
  disabled: true,
  buttonAction: () => {},
  tips: $t("join.title"),
  tipsStyle: "blank",
  tipsAction: () => history.push("/explore"),
})

export const mainAirdropOver = (groupInfo: IGroup, $t: any) => ({
  groupInfo,
  button: $t("join.main.appointOver"),
  disabled: true,
  buttonAction: () => {},
  tips: $t("join.title"),
  tipsStyle: "blank",
  tipsAction: () => history.push("/explore"),
})

function getCheckMsg(checkList: Array<ICheck>, $t: any): string {
  let str = $t("join.modal.sharesCheck")
  let { length } = checkList
  if (length > 1) {
    const usd: ICheck | undefined = checkList.find(
      (item) => item.asset_id === "4d8c508b-91c5-375b-92b0-ee702ed2dac5",
    )

    const pusd: ICheck | undefined = checkList.find(
      (item) => item.asset_id === "31d2ea9c-95eb-3355-b65b-ba096853bc18",
    )
    const others: ICheck[] | undefined = checkList.filter(
      (item) => item.symbol !== "USDT",
    )
    if (pusd && usd && others) {
      checkList = [pusd, usd, ...others]
      length = checkList.length
    }
  }
  activeAsset = checkList[0].asset_id
  checkList.forEach((item, idx) => {
    const { name, amount, symbol } = item
    str += `${name} ${$t("join.modal.sharesCheck1")} ${amount} ${symbol}`
    str += idx < length - 1 ? ` ${$t("join.modal.sharesCheck2")} ` : " 。"
  })
  return str
}
