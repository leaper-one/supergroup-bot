import { history, request, RequestConfig } from "umi"
import { $get } from "@/stores/localStorage";

let baseUrl = process.env.SERVER_URL

export const mixinBaseURL = "https://mixin-api.zeromesh.net"

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


const hostClientMap = {
  "cnb": "f6deb534-13bd-45f0-9b34-0d618827f500",
  "mobilecoin": "47b0b809-2bb5-4c94-becd-35fb93f5c6fe",
  "bitcoin": "d0828d93-c4e2-4b5f-a801-d2aabbd80424",
  "eos": "c7d5a9a8-916f-4583-86e6-56014b1ab673",
  "ethereum": "2bc4c914-bab8-4e77-9d3c-50635827deec",
  "horizen": "ae90deb6-2737-4e09-b82a-ff41d565dc35"
}

function getClientURL() {
  let clientID = process.env.CLIENT_ID
  if (!clientID) {
    const [t1] = location.host.split('.')
    if (Object.keys(hostClientMap).includes(t1))
      clientID = hostClientMap[t1]
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
  post: (url: string, data: any = {}): Promise<any> =>
    request(url, { method: "POST", data }),
  put: (url: string, data: any = {}): Promise<any> =>
    request(url, { method: "PUT", data }),
  delete: (url: string, params: any = {}, options: any = {}): Promise<any> =>
    request(url, { method: "DELETE", params, ...options }),
}
