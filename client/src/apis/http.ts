import { history, request, RequestConfig } from "umi"
import { $get } from "@/stores/localStorage"

export const mixinBaseURL = process.env.MIXIN_BASE_URL
export const liveReplayPrefixURL = process.env.LIVE_REPLAY_URL
export const serverURL = process.env.SERVER_URL

export const getAuthUrl = (returnTo = '') => {
  let { pathname, search, query } = history.location
  if (search && !search.startsWith("?")) search = "?" + search
  if (!returnTo) {
    returnTo =
      pathname === "/auth"
        ? (query?.return_to as string) || "/"
        : pathname + search
  }
  return `https://mixin-www.zeromesh.net/oauth/authorize?client_id=${getClientURL()}&scope=PROFILE:READ+ASSETS:READ+MESSAGES:REPRESENT&response_type=code&return_to=${returnTo}`
}

function getClientURL() {
  return $get('group').client_id
}

export const getAddUserURL = (userID: string) => `mixin://users/${userID}`

export const staticUrl = `https://taskwall.zeromesh.net/group-manager/`

export const getCodeUrl = (code_id: string) =>
  `https://mixin.one/codes/${code_id}`

export const payUrl = ({
  trace = "",
  recipient = getClientURL(),
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
  `https://eiduwejdk.com/?from=${getClientURL()}&type=bot#/exchange/otc/otcDetail?id=${id}`
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
        `http://192.168.2.153:7001` :
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
          window.location.href = getAuthUrl()
          window.localStorage.clear()
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
