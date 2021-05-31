import React, { useEffect, useState } from "react"
import { IAsset } from "@/apis/asset"
import headStyles from "@/pages/home/trade.less"
import style from "./index.less"
import { get$t } from "@/locales/tools"
import { history, useIntl } from "umi"
import { ApiGetBtcPrice, ApiGetGroupAssets } from "@/apis/group"

export const Asset = () => {
  const [assetList, setAssetList] = useState<IAsset[]>([])
  const [totalBtc, setTotalBtc] = useState(0)
  const [totalUsd, setTotalUsd] = useState(0)
  const $t = get$t(useIntl())

  useEffect(() => {
    ApiGetBtcPrice().then()
    Promise.all([ApiGetGroupAssets(), ApiGetBtcPrice()]).then(
      ([assets, btcPrice]) => {
        setAssetList(assets)
        const usd = assets.reduce((pre, cur) => {
          return pre + Number(cur.price_usd) * Number(cur.balance)
        }, 0)
        const btc = usd / Number(btcPrice)
        setTotalBtc(btc)
        setTotalUsd(usd)
      },
    )
  }, [])

  return (
    <div>
      <header className={`${headStyles.price} ${style.head}`}>
        <div className={style.title}>{$t("manager.asset.total")}</div>

        <div className={style.amount}>
          <span>{totalBtc.toFixed(8)}</span>
          <span className={style.btc}>BTC</span>
        </div>

        <div className={style.usd}>≈ ${totalUsd.toFixed(2)}</div>
      </header>

      <ul className={style.assets}>
        {assetList.map((asset) => (
          <li
            key={asset.asset_id}
            className={style.asset}
            onClick={() => history.push(`/snapshots/${asset.asset_id}`)}
          >
            <img className={style.icon} src={asset.icon_url} alt="" />
            <h4>
              {asset.balance} {asset.symbol}
            </h4>
            <p className={style.green}>
              {(Number(asset.change_usd) * 100).toFixed(2)}%
            </p>
            <span>
              ≈ ${(Number(asset.price_usd) * Number(asset.balance)).toFixed(2)}
            </span>
            <span className={style.price}>${asset.price_usd}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

interface IModalProps {
  showSelectCoin: boolean
  setSelectCoin: (v: boolean) => void
  select: (v: IAsset | undefined) => void
  avoidCoin: string[]
}

// export const AssetModal = (props: IModalProps) => {
//   const {showSelectCoin, select, setSelectCoin, avoidCoin} = props
//   return <Modal
//     popup
//     visible={showSelectCoin}
//     onClose={() => setSelectCoin(false)}
//     animationType="slide-up"
//   >
//     <CoinModal select={select} avoid={avoidCoin}/>
//   </Modal>
// }
