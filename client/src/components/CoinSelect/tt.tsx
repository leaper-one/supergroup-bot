import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { BottomModal } from "@/components/BottomkModal"
import { ApiGetTop100, IAsset } from "@/apis/asset"
import { Modal } from "antd-mobile"

interface Props {
  select: (asset: IAsset | undefined) => void
  close: () => void
  closeWithAnimation?: number
  active: IAsset | undefined
}

export const CoinSelect = (props: Props) => {
  const [closeModal, setCloseModal] = useState(0)
  return (
    <BottomModal
      content={coin(props)}
      close={props.close}
      closeWithAnimation={closeModal || props.closeWithAnimation}
      top={48}
    />
  )
}

const coin = (props: Props) => {
  const { active, close } = props
  const [assetList, setAssetLIst] = useState<IAsset[]>([])

  useEffect(() => {
    ApiGetTop100().then(setAssetLIst)
  }, [])

  return (
    <div className={styles.container}>
      <div className={styles.search + " " + "flex"}>
        <img src={require("@/assets/img/svg/search.svg")} alt="" />
        <input placeholder="Name, Symbol" type="text" />
        <span
          onClick={() => {
            props.select(active)
          }}
        >
          取消
        </span>
      </div>
      <ul className={styles.list}>
        {assetList.map((asset, idx) => (
          <li
            key={idx}
            onClick={() => {
              props.select(asset)
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
