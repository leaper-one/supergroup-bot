import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { ApiGetAssetBySymbol, ApiGetTop100, IAsset } from "@/apis/asset"

type CoinProps = IAsset | undefined

interface Props {
  select: (asset: CoinProps) => void
  closeWithAnimation?: number
  active?: CoinProps
  avoid?: string[]
}

export const Coin = (props: Props) => {
  const { select, active, avoid } = props
  const [assetList, setAssetLIst] = useState<IAsset[]>([])
  const [search, setSearch] = useState("")

  useEffect(() => {
    setAssetInit()
  }, [avoid])

  useEffect(() => {
    if (search)
      ApiGetAssetBySymbol(search).then((list) => {
        if (Array.isArray(list)) {
          setAssetLIst(list)
        }
      })
    else {
      setAssetInit()
    }
  }, [search])

  const setAssetInit = () => {
    ApiGetTop100().then((list) => {
      if (avoid)
        list = list.filter((item: IAsset) => !avoid.includes(item.asset_id!))
      setAssetLIst(list.slice(0, 20))
    })
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
        {assetList.map((asset, idx) => (
          <li
            key={asset.asset_id}
            onClick={() => {
              select(asset)
            }}
          >
            <img src={asset.icon_url} alt="" />
            <p>{asset.name}</p>
            <i>{asset.symbol}</i>
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
