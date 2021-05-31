import React, { useEffect, useState } from "react"
import styles from "./coinSelect.less"
import {
  ApiGetAssetBySymbol,
  ApiGetMyAssets,
  ApiGetTop100,
  IAsset,
} from "@/apis/asset"

type CoinProps = IAsset | undefined

interface Props {
  select: (asset: CoinProps) => void
  closeWithAnimation?: number
  active?: CoinProps
  myAsset?: boolean
  avoid?: string[]
  assetList?: IAsset[]
}

export const CoinModal = (props: Props) => {
  const { select, active, avoid, myAsset } = props
  const [assetList, setAssetList] = useState<IAsset[]>([])
  const [search, setSearch] = useState("")

  useEffect(() => {
    if (myAsset) {
      if (!props.assetList)
        ApiGetMyAssets().then((list) =>
          setAssetList(
            list.filter(
              (item) =>
                item.name?.toLowerCase().includes(search.toLowerCase()) ||
                item.symbol?.toLowerCase().includes(search.toLowerCase()),
            ),
          ),
        )
      return
    }
    if (search)
      if (!props.assetList)
        ApiGetAssetBySymbol(search).then(
          (list) => Array.isArray(list) && setAssetList(list),
        )
      else setAssetInit()
  }, [search])

  const setAssetInit = () => {
    if (myAsset) {
      ApiGetMyAssets().then((item) => {
        setAssetList(item)
      })
    } else if (!props.assetList) {
      ApiGetTop100().then((list) => {
        if (avoid)
          list = list.filter((item: IAsset) => !avoid.includes(item.asset_id!))
        setAssetList(list.slice(0, 20))
      })
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.search + " " + "flex"}>
        <img src={require("@/assets/img/svg/search.svg")} alt="" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Name, Symbol"
          type="text"
        />
        <span onClick={() => select(active)}>取消</span>
      </div>
      <ul className={styles.list}>
        {(props.assetList || assetList).map((asset, idx) => (
          <li key={asset.asset_id} onClick={() => select(asset)}>
            <img src={asset.icon_url} alt="" />
            <p>{asset.name}</p>
            <i>{myAsset ? `${asset.balance} ${asset.symbol}` : asset.symbol}</i>
            <img
              className={
                (active &&
                  active.asset_id === asset.asset_id &&
                  styles.selected) +
                " " +
                styles.select
              }
              src={require("@/assets/img/svg/select.svg")}
              alt=""
            />
          </li>
        ))}
      </ul>
    </div>
  )
}
