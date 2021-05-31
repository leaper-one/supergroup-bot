import React, { useState } from "react"
import { BackHeader } from "@/components/BackHeader"
import styles from "./index.less"
import Statics from "./tab/statics"
import Setting from "@/pages/setting/index"
import { $get } from "@/stores/localStorage"
import { history, useIntl } from "umi"
import { staticUrl } from "@/apis/http"
import { GlobalData } from "@/stores/store"
import { Asset } from "./asset"
import { get$t } from "@/locales/tools"

const TabList = ["home", "statistics", "asset", "setting"]
const IconList = [
  "iconic_unselected_1",
  "iconic_unselected_4",
  "iconweixuanzhong",
  "iconic_unselected_5",
]
const IconSelectList = [
  "iconic_select_1",
  "iconic_select_4",
  "iconxuanzhong",
  "iconic_select_5",
]

export default () => {
  const [currentPage, setCurrentPage] = useState(GlobalData.managerCurrentTab)

  const $t = get$t(useIntl())

  return (
    <div className={`${styles.mainBox}`}>
      <BackHeader
        name="社群助手"
        noBack
        onClick={() => history.push(`/home`)}
        action={
          currentPage === "asset" ? (
            <i
              className={styles.addIcon + " iconfont iconic_add"}
              onClick={() => history.push(`/asset/deposit`)}
            />
          ) : undefined
        }
      />
      <div className={styles.container}>
        {currentPage === "home" && <Home $t={$t} />}
        {currentPage === "statistics" && <Statics />}
        {currentPage === "asset" && <Asset />}
        {currentPage === "setting" && <Setting />}
      </div>
      {
        <div className={styles.footer}>
          {TabList.map((item, idx) => (
            <i
              key={idx}
              className={`iconfont ${
                currentPage === item ? IconSelectList[idx] : IconList[idx]
              }`}
              onClick={() => {
                GlobalData.managerCurrentTab = item
                setCurrentPage(item)
              }}
            />
          ))}
        </div>
      }
    </div>
  )
}

interface IHomeProps {
  $t: any
}

const Home = (props: IHomeProps) => {
  const group = $get("group")
  if (!group) {
    history.replace("/")
    return <></>
  }
  const asset_id = group.asset_id

  return (
    <div className={styles.index}>
      {asset_id && (
        <li onClick={() => history.push("/transfer/" + asset_id)}>
          <img src={staticUrl + "home_0.png"} alt="" />
          <p>{props.$t("home.trade")}</p>
        </li>
      )}
      <li onClick={() => history.push("/invite")}>
        <img src={staticUrl + "home_1.png"} alt="" />
        <p>邀请入群</p>
      </li>
      <li onClick={() => history.push("/red/pre")}>
        <img src={staticUrl + "home_2.png"} alt="" />
        <p>群发红包</p>
      </li>
      <li onClick={() => history.push("/broadcast")}>
        <img src={staticUrl + "home_3.png"} alt="" />
        <p>公告</p>
      </li>
      {/*<li onClick={() => history.push("/explore")}>*/}
      {/*  <img src={staticUrl + "home_3.png"} alt="" />*/}
      {/*  <p>群发公告</p>*/}
      {/*</li>*/}
      {/*<li onClick={() => history.push("/findBot")}>*/}
      {/*  <img src={staticUrl + "home_5.png"} alt=""/>*/}
      {/*  <p>发现机器人</p>*/}
      {/*</li>*/}
      {/*<li onClick={() => history.push("/more")}>*/}
      {/*  <img src={staticUrl + "home_6.png"} alt=""/>*/}
      {/*  <p>更多活动</p>*/}
      {/*</li>*/}
      <li>
        <img src={staticUrl + "home_4.png"} alt="" />
        <p>敬请期待</p>
      </li>
    </div>
  )
}
