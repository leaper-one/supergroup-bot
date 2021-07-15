import { apis } from "@/apis/http"
import { Div } from "@/assets/ts/number"
import { IAsset } from "@/apis/asset"
import { GlobalData } from "@/stores/store"

type TSwapType = "0" | "1" | "2" | "3" | "4"

export interface ISwapItem {
  lp_asset: string
  asset0: string
  asset0_price: string
  asset1: string
  asset1_price: string
  type: TSwapType
  pool: string
  earn: string
  amount: string
  icon_url: string
  asset0_symbol: string
  asset0_amount: string
  asset1_symbol: string
  asset1_amount: string

  price?: string
  rate?: string

  asset_id?: string
  buy_max?: string
  exchange?: string
  otc_id?: string
  price_usd?: string
}

export interface IExinAd {
  id?: number
  assetId?: string
  nickname?: string
  avatarUrl?: string
  isCertification?: boolean
  isLandun: boolean
  in5minRate: string
  maxPrice: string
  minPrice: string
  price: string
  orderSuccessRank: string
  multisigOrderCount: string
  payMethods: {
    id: number
    name: string
    symbol: "bank" | "alipay" | "wechatpay"
  }[]
}

export interface ISwapResp {
  list?: ISwapItem[]
  asset: IAsset
  ad: IExinAd[]
}

export const ApiGetSwapList = async (asset_id: string): Promise<ISwapResp> => {
  if (!GlobalData.swapList) GlobalData.swapList = {}
  if (GlobalData.swapList[asset_id]) return GlobalData.swapList[asset_id]
  const resp: ISwapResp = await apis.get(`/swapList/${asset_id}`)
  resp.list = resp.list?.map((item) => {
    let price = item.asset1 === asset_id ? item.asset1_price : item.asset0_price
    let rate
    switch (item.type) {
      case "1":
      case "4":
        rate = Div(item.asset0_amount, item.asset1_amount)
        break
      case "0":
        rate = Div(item.asset1_price, item.asset0_price)
    }
    return {
      rate,
      price,
      ...item,
    }
  })
  GlobalData.swapList[asset_id] = resp
  return GlobalData.swapList[asset_id]
}
