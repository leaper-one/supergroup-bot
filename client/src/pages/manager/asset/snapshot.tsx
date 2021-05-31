import React, { useEffect, useState } from "react"
import styles from "./snapshot.less"
import { ApiGetAssetByID, IAsset } from "@/apis/asset"
import { history, useIntl, useParams } from "umi"
import { BackHeader } from "@/components/BackHeader"
import headStyles from "@/pages/home/trade.less"
import style from "@/pages/manager/asset/index.less"
import { get$t } from "@/locales/tools"
import { ApiGetGroupSnapshots, ISnapshotItem } from "@/apis/group"
import { staticUrl } from "@/apis/http"
import moment from "moment"
import { FullLoading } from "@/components/Loading"

export default (props: any) => {
  const [activeAsset, setActiveAsset] = useState<IAsset>()
  const { id = "" } = useParams<{ id: string }>() || {}
  const [snapshots, setSnapshots] = useState<ISnapshotItem[]>([])
  const [loading, setLoading] = useState(false)
  const $t = get$t(useIntl())

  useEffect(() => {
    setLoading(true)
    Promise.all([ApiGetAssetByID(id), ApiGetGroupSnapshots(id)]).then(
      ([asset, snapshots]) => {
        setActiveAsset(asset)
        setSnapshots(snapshots)
        setLoading(false)
      },
    )
  }, [])

  const totalAmount = snapshots.reduce(
    (pre, cur) => pre + Number(cur.amount),
    0,
  )

  return (
    <div className={styles.container}>
      <BackHeader name={(activeAsset?.symbol || "") + " Token"} />

      <header className={`${headStyles.price} ${style.head} ${styles.head}`}>
        <img className={styles.icon} src={activeAsset?.icon_url} alt="" />
        <div className={styles.amount}>
          <span className={styles.num}>{Number(totalAmount.toFixed(8))}</span>
          <span className={style.btc}>{activeAsset?.symbol}</span>
          <div className={styles.usd}>
            {" "}
            ≈{" "}
            {Number(Number(activeAsset?.price_usd || 0) * totalAmount).toFixed(
              2,
            )}
          </div>
        </div>
        <div className={styles.action}>
          <span onClick={() => history.push(`/asset/deposit`)}>
            {$t("manager.asset.deposit")}
          </span>
          <span onClick={() => history.push(`/asset/withdrawal`)}>
            {$t("manager.asset.withdrawal")}
          </span>
        </div>
      </header>

      <ul className={styles.list}>
        {snapshots.map((s, i) => (
          <li key={i} className={styles.item}>
            <img
              className={styles.icon}
              src={`${staticUrl}snapshot/${s.origin}.png`}
              alt=""
            />
            <div className={styles.info}>
              <p className={styles.title}>{$t(`manager.asset.${s.origin}`)}</p>
              <span className={styles.desc}>
                {moment(s.created_at).format("YYYY-MM-DD HH:mm")}
              </span>
            </div>
            <div className={styles.amountDesc}>
              {Number(s.amount) > 0 && (
                <span className={`green ${styles.add}`}>+</span>
              )}
              <span
                className={`${styles.amount} ${
                  Number(s.amount) > 0 ? "green" : "red"
                }`}
              >
                {s.amount}
              </span>
              <span className={styles.symbol}>{activeAsset?.symbol}</span>
            </div>
          </li>
        ))}
      </ul>

      {loading && <FullLoading mask />}
    </div>
  )
}
