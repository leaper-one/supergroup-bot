import { history, request, RequestConfig } from "umi"
import { $get } from "@/stores/localStorage"
import auth from '@/pages/auth'
import { getConversationIdByUserIDs } from '@/assets/ts/tools'

export const mixinBaseURL = process.env.MIXIN_BASE_URL
export const liveReplayPrefixURL = process.env.LIVE_REPLAY_URL
export const serverURL = process.env.SERVER_URL

export const getAuthUrl = (returnTo = '', hasAuth = false, state = "") => {
  let { pathname, search, query } = history.location
  if (search && !search.startsWith("?")) search = "?" + search
  if (!returnTo) {
    returnTo =
      pathname === "/auth"
        ? (query?.return_to as string) || "/"
        : pathname + search
  }
  return `https://mixin-www.zeromesh.net/oauth/authorize?client_id=${getClientID()}&scope=PROFILE:READ+MESSAGES:REPRESENT${hasAuth ? '+ASSETS:READ' : ''}&response_type=code&return_to=${returnTo}&state=${state}`
}

function getClientID() {
  if ($get('group')) return $get('group').client_id
  else auth()
}

const checkVersion = (): boolean => {
  let reg = /Mixin\/([0-9]+)\.([0-9]+)\.([0-9]+)/
  const [_, a, b, c] = navigator.userAgent.match(reg) || []
  if (Number(a) > 0 || Number(b) > 31) return true
  if (Number(b) == 31 && Number(c) >= 1) return true
  return false
}

export const getAddUserURL = (userID: string) => {
  if (navigator.userAgent.includes("Mixin") && checkVersion()) {
    const userID = $get('user').user_id
    const clientID = getClientID()
    return `mixin://conversations/${getConversationIdByUserIDs(userID, clientID)}?user=${clientID}`
  }
  return `mixin://users/${userID}`
}



export const staticUrl = `https://taskwall.zeromesh.net/group-manager/`

export const getCodeUrl = (code_id: string) =>
  `https://mixin.one/codes/${code_id}`

export const payUrl = ({
  trace = "",
  recipient = getClientID(),
  asset = "",
  amount = "",
  memo = "",
} = {}) =>
  `https://mixin.one/pay?recipient=${recipient}&asset=${asset}&amount=${amount}&memo=${encodeURIComponent(
    memo,
  )}&trace=${trace}`

// export const getExinOtcUrl = (id: string) =>
//   `https://exinone.com/#/exchange/otc/otcDetail?id=${id}`

export const getExinOtcUrl = (id: string) =>
  `https://eiduwejdk.com/?from=${getClientID()}&type=bot#/exchange/otc/otcDetail?id=${id}`
export const getExinLocalUrl = (id: string) =>
  `https://hk.exinlocal.com/#/exchange?side=sell&&uuid=${id}`

export const getExinSwapUrl = (base: string, quote: string) =>
  `https://mixswap.exchange/#/swap?pay=${base}&receive=${quote}`
export const get4SwapUrl = (base: string, quote: string) =>
  `https://mixswap.exchange/#/swap?pay=${base}&receive=${quote}`
export const get4SwapNormalUrl = (base: string, quote: string) =>
  `https://mixswap.exchange/#/swap?pay=${base}&receive=${quote}`

//
// export const getExinSwapUrl = (id: string) =>
//   `https://app.exinswap.com/#/pairs/${id}`
// export const get4SwapUrl = (base: string, quote: string) =>
//   `https://f1-4swap-mtg.firesbox.com/#/pair-info?base=${base}&quote=${quote}`
// export const get4SwapNormalUrl = (base: string, quote: string) =>
//   `https://f1-uniswap.firesbox.com/#/pair-info?base=${base}&quote=${quote}`

export const requestConfig: RequestConfig = {
  timeout: 15 * 1000,
  requestInterceptors: [
    function (url, options) {
      const baseUrl = process.env.NODE_ENV === 'development' ?
        serverURL :
        `https://${location.host.split('.')[0]}-api.${serverURL}`
      url.startsWith("/") && (url = baseUrl + url)
      const token = $get('token')
      if (token) options.headers = {
        Authorization: "Bearer " + token,
        ...options.headers,
      }

      options.params = {
        ...options.params,
        t: Date.now(),
      }
      return { ...options, url }
    },
  ],
  responseInterceptors: [
    async (res) => {
      const data = await res.clone().json()
      if (data.error) {
        if ([401].includes(data.error.code)) {
          window.localStorage.clear()
          auth()
          return
        }
      }
      return data.data || data.error || data
    },
  ],
}

export const apis = {
  get: (url: string, params: any = {}, options: any = {}): Promise<any> =>
    request(url, { params, ...options }),
  post: (url: string, data: any = {}, options: any = {}): Promise<any> =>
    request(url, { method: "POST", data, ...options }),
  put: (url: string, data: any = {}): Promise<any> =>
    request(url, { method: "PUT", data }),
  delete: (url: string, params: any = {}, options: any = {}): Promise<any> =>
    request(url, { method: "DELETE", params, ...options }),
}
