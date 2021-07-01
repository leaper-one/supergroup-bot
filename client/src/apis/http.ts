import { history, request, RequestConfig } from "umi"
import { $get } from "@/stores/localStorage";

let baseUrl = process.env.SERVER_URL

export const mixinBaseURL = "https://mixin-api.zeromesh.net"
export const liveReplayPrefixURL = "https://super-group-cdn.mixinbots.com/live-replay/"

export const getAuthUrl = () => {
  let { pathname, search, query } = history.location
  if (search && !search.startsWith("?")) search = "?" + search
  let returnTo: string =
    pathname === "/auth"
      ? (query?.return_to as string) || "/"
      : pathname + search
  if (!returnTo.startsWith("/join")) {
    returnTo = "/"
  }


  return `https://mixin-www.zeromesh.net/oauth/authorize?client_id=${getClientURL()}&scope=PROFILE:READ+ASSETS:READ+MESSAGES:REPRESENT&response_type=code&return_to=${returnTo}`
}

function getClientURL() {
  let clientID = process.env.CLIENT_ID
  if (!clientID) {
    clientID = $get('group').client_id
  }
  return clientID
}

export const staticUrl = `https://taskwall.zeromesh.net/group-manager/`

export const getCodeUrl = (code_id: string) =>
  `https://mixin.one/codes/${code_id}`

export const payUrl = ({
                         trace = "",
                         recipient = getClientURL(),
                         asset = process.env.ASSET_ID,
                         amount = process.env.AMOUNT,
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
      if (!baseUrl) {
        const [t1] = location.host.split('.')
        baseUrl = `https://${t1}-api.mixinbots.com`
      }
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
          return (window.location.href = getAuthUrl())
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
