import React, { useEffect, useState } from "react"
import styles from "./tradeList.less"
import { BackHeader } from "@/components/BackHeader"
import { ApiCheckGroup } from "@/apis/conversation"
import { getConversationId } from "@/assets/ts/tools"
import { ApiGetGroupList, IGroupItem } from "@/apis/group"
import { ApiGetSwapList, ISwapItem, ISwapResp } from "@/apis/transfer"
import { get$t } from "@/locales/tools"
import { history, useIntl } from "umi"

const getGroupInfo = async (setSharesList: any) => {
  let groupList: IGroupItem[] = await ApiGetGroupList()
  updateSharesList(groupList, setSharesList)
}

const updateSharesList = async (
  groupList: IGroupItem[],
  setSharesList: any,
) => {
  const groupInfo = await ApiCheckGroup(getConversationId()!)
  if (groupList && Array.isArray(groupList)) {
    const idx = groupList.findIndex(
      (item) => item.group_id === groupInfo.group.group_id,
    )
    if (idx !== -1 && groupList[idx].check.length > 0) {
      let swapList = await Promise.all(
        groupList[idx].check.map((item) => ApiGetSwapList(item.asset_id!)),
      )
      swapList = swapList.filter((item) => item.list)
      setSharesList(swapList)
    }
  }
}

const swapList = ["ExinSwap", "4Swap", "ExinOne", "ExinLocal"]
export default () => {
  const $t = get$t(useIntl())

  const [shareList, setShareList] = useState([] as ISwapResp[])

  useEffect(() => {
    getGroupInfo(setShareList)
  }, [])

  const getMethods = (swapItem: ISwapItem[] | undefined): string => {
    const t: any = {}
    swapItem?.forEach((item) => {
      t[item.type] = true
    })
    let str = ""
    for (const key in t) {
      str += swapList[Number(key)] + "、"
    }

    str = str.slice(0, -1)
    return str
  }

  return (
    <div className={styles.container}>
      <BackHeader name="持仓币种交易" />

      <ul>
        {shareList.map(({ list, asset }) => (
          <li
            key={asset.asset_id}
            onClick={() => history.push(`/trade/${asset.asset_id}`)}
          >
            <img src={asset.icon_url} alt="" className={styles.icon} />
            <p className={styles.title}>
              {$t("home.trade")} {asset.symbol}
            </p>
            <span className={styles.price}>${fixRate(asset.price_usd!)}</span>
            <span className={styles.exchange}>{getMethods(list)}</span>
            <span
              className={`${styles.rate} ${
                Number(asset.change_usd) > 0 ? "green" : "red"
              }`}
            >
              {fixRate(asset.change_usd!)}%
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}

function fixRate(num: string): string {
  return (Number(num) * 100).toFixed(2)
}
