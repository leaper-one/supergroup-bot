import React, { useEffect, useState } from "react"
import styles from "./index.less"
import { BackHeader } from "@/components/BackHeader"
import { CoinModal } from "@/components/PopupModal/coinSelect"
import { pageToTop } from "@/assets/ts/tools"
import { IAsset } from "@/apis/asset"
import { Modal, Toast } from "antd-mobile"
import { $get, $set } from "@/stores/localStorage"
import { history } from "umi"
import { Button } from "@/components/Sub"

type TLang = "zh" | "en"
const Lang: TLang[] = ["zh", "en"]
const langList = { zh: "中文", en: "English" }
const langDesc = { zh: "中文", en: "英文" }
const langTips = {
  zh: "名称后缀不要加“群”字，自动建群时会自动在后面加上。",
  en:
    "Please do not add “Group” by the end of the group name, because it will be added automatically once the group is created.",
}

export default () => {
  pageToTop()
  const [coinModal, setCoinModal] = useState(false)
  const [langModal, setLangModal] = useState(false)
  const [activeAsset, setActiveAsset] = useState<IAsset>()
  const [activeLang, setActiveLang] = useState<TLang>("zh")
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")

  useEffect(() => {
    const create = $get("create")
    if (create) {
      setActiveAsset({ icon_url: create.icon_url, asset_id: create.asset_id })
      setName(create.name)
      setDescription(create.description)
    }
  }, [])

  const clickConfirm = () => {
    if (!activeAsset) {
      return Toast.info("请先选择币种")
    }
    if (!name) {
      return Toast.info("请先填写社群名")
    }
    const create = $get("create")
    if (create) {
      $set("create", {
        ...create,
        name,
        description,
        icon_url: activeAsset.icon_url,
        asset_id: activeAsset.asset_id,
        lang: activeLang,
      })
    } else {
      $set("create", {
        name,
        description,
        icon_url: activeAsset.icon_url,
        asset_id: activeAsset.asset_id,
        lang: activeLang,
      })
    }
    history.push("/create/coin")
  }

  const select = (asset: IAsset | undefined) => {
    setActiveAsset(asset)
    setCoinModal(false)
  }

  return (
    <>
      <BackHeader name="创建社群" noBack />
      <div className={styles.container}>
        {activeAsset ? (
          <img
            key={1}
            onClick={() => setCoinModal(true)}
            src={activeAsset.icon_url}
            className={styles.asset_icon}
          />
        ) : (
          <img
            key={2}
            onClick={() => setCoinModal(true)}
            src={require("@/assets/img/svg/add.svg")}
          />
        )}
        <span className={styles.iconDesc}>群徽章</span>
        <input
          type="text"
          placeholder="群名称"
          onChange={(e) => setName(e.target.value)}
          value={name}
        />
        <p>{langTips[activeLang]}</p>
        <div className={styles.lang} onClick={() => setLangModal(true)}>
          <span>{langDesc[activeLang]}</span>
          <i className="iconfont iconic_down" />
        </div>
        <textarea
          placeholder="群简介"
          onChange={(e) => setDescription(e.target.value)}
          value={description}
        />
        <Button className="btn" onClick={clickConfirm}>
          创建
        </Button>
      </div>

      <Modal
        popup
        visible={coinModal}
        onClose={() => setCoinModal(false)}
        animationType="slide-up"
      >
        <CoinModal select={select} active={activeAsset} />
      </Modal>
      <Modal
        popup
        visible={langModal}
        onClose={() => setLangModal(false)}
        animationType="slide-up"
      >
        <div className={styles.langModal}>
          <h4>语言</h4>
          <i
            onClick={() => setLangModal(false)}
            className={`iconfont iconguanbi ${styles.close}`}
          />

          <ul className={styles.langList}>
            {Lang.map((item) => (
              <li
                key={item}
                className={styles.langItem}
                onClick={() => {
                  setLangModal(false)
                  setActiveLang(item)
                }}
              >
                <div className={styles.name}>{langList[item]}</div>
                <div className={styles.desc}>{langDesc[item]}</div>
                {activeLang === item && (
                  <i className={`iconfont iconcheck ${styles.select}`} />
                )}
              </li>
            ))}
          </ul>
        </div>
      </Modal>
    </>
  )
}
