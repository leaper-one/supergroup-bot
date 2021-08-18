import React, { useEffect, useState } from 'react'
import styles from './index.less'
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { history, useIntl } from "umi"
import { ApiGetBroadcastList, IBroadcast } from "@/apis/broadcast"
import moment from "moment"
import { Modal, SwipeAction } from "antd-mobile"
import {
  ApiGetCancelTopNews,
  ApiGetLiveList,
  ApiGetStartLive,
  ApiGetStopLive,
  ApiGetTopNews,
  ILive
} from "@/apis/live"
import { $get, $set } from "@/stores/localStorage"
import { Confirm, Prompt, ToastSuccess } from "@/components/Sub"
import { GlobalData } from "@/stores/store"

export default function Page() {
  const $t = get$t(useIntl())
  const [show, setShow] = useState(false)
  const [activeTab, setActiveTab] = useState(GlobalData.activeLiveTabs || 0)
  const [broadcastList, setBroadcastList] = useState<IBroadcast[]>([])
  const [liveList, setLiveList] = useState<ILive[]>([])
  const [isManager, setIsManager] = useState(false)
  useEffect(() => {
    initList()
    const { status } = $get("_user") || {}
    if (status === 9) setIsManager(true)
  }, [])

  const initList = () => {
    setLiveList([])
    setBroadcastList([])
    ApiGetBroadcastList().then(setBroadcastList)
    ApiGetLiveList().then(setLiveList)
  }
  let newsList: any
  let liveTop: ILive[] = [],
    normal: (ILive | IBroadcast)[] = [],
    top: (ILive | IBroadcast)[] = []

  for (let i = 0; i < liveList.length; i++) {
    if (liveList[i].status !== 2) liveTop.push(liveList[i])
    else if (Number(new Date(liveList[i].top_at)) > 0) top.push({ ...liveList[i], is_top: true })
    else normal.push(liveList[i])
  }
  for (let i = 0; i < broadcastList.length; i++) {
    if (Number(new Date(broadcastList[i].top_at)) > 0) top.push({ ...broadcastList[i], is_top: true })
    else normal.push(broadcastList[i])
  }

  const topList = top.sort(
    (a, b) => Number(new Date(b.top_at)) - Number(new Date(a.top_at))
  )
  const lowList = normal.sort(
    (a, b) => Number(new Date(b.created_at!)) - Number(new Date(a.created_at!))
  )
  if (activeTab === 0) {
    newsList = [...liveTop, ...topList, ...lowList]
  } else if (activeTab === 1) newsList = liveList.filter(item => item.status === 2).sort(
    (a, b) => {
      const hasTop = Number(new Date(b.top_at!)) - Number(new Date(a.top_at))
      if (hasTop !== 0) return hasTop
      return Number(new Date(b.created_at!)) - Number(new Date(a.created_at!))
    }
  )
  else if (activeTab === 2) newsList = broadcastList.sort(
    (a, b) => {
      const hasTop = Number(new Date(b.top_at!)) - Number(new Date(a.top_at))
      if (hasTop !== 0) return hasTop
      return Number(new Date(b.created_at!)) - Number(new Date(a.created_at!))
    }
  )

  return <div className="safe-view">
    <BackHeader name={$t('home.article')} action={isManager ? <i
      className={`iconfont iconic_add`}
      onClick={() => setShow(true)}
    /> : <></>} />

    <div className={styles.tabs}>
      {["all", "replay", "broadcast"].map((v, i) => (
        <span
          key={i}
          className={activeTab === i ? styles.tabs_active : ''}
          onClick={() => {
            GlobalData.activeLiveTabs = i
            setActiveTab(i)
          }}
        >{$t(`news.${v}`)}</span>
      ))}
    </div>

    {newsList.length > 0 ? <div className={styles.list}>
      {newsList.map((item: any, idx: number) =>
        <div key={idx}>
          {item.img_url ?
            <SwipeAction
              disabled={!isManager}
              right={item.status === 0 ? [
                {
                  text: $t("news.action.edit"),
                  className: styles.actionEdit,
                  onPress() {
                    $set("active_live", item)
                    history.push(`/news/addLive`)
                  }
                },
                {
                  text: $t("news.action.start"),
                  className: styles.actionStart,
                  async onPress() {
                    let res = "", url = ""
                    if (item.category === 1) {
                      url = await Prompt($t("action.tips"), $t("news.prompt"))
                      if (!url) return
                    } else {
                      const isConfirm = await Confirm($t("action.tips"), $t("news.confirmStart"))
                      if (!isConfirm) return
                    }
                    res = await ApiGetStartLive(item.live_id, encodeURIComponent(url))
                    if (res === "success") {
                      ToastSuccess($t("success.operator"))
                      await initList()
                    }
                  }
                },
              ] : item.status === 1 ? [
                {
                  text: $t("news.action.stop"),
                  className: styles.actionNormal,
                  async onPress() {
                    const isConfirm = await Confirm($t("action.tips"), $t("news.confirmEnd"))
                    if (!isConfirm) return
                    const res = await ApiGetStopLive(item.live_id)
                    if (res === "success") {
                      ToastSuccess($t("success.operator"))
                      await initList()
                    }
                  }
                }
              ] : [
                {
                  text: item.is_top ? $t("news.action.cancelTop") : $t("news.action.top"),
                  className: styles.actionNormal,
                  async onPress() {
                    const isTop = item.is_top
                    const isConfirm = await Confirm($t("action.tips"), isTop ? $t("news.confirmTop") : $t("news.confirmCancelTop"))
                    if (!isConfirm) return
                    let res
                    if (isTop) {
                      res = await ApiGetCancelTopNews(item.live_id)
                    } else {
                      res = await ApiGetTopNews(item.live_id)
                    }

                    if (res === "success") {
                      ToastSuccess($t("success.operator"))
                      await initList()
                    }
                  }
                }
              ]}
              style={{ marginBottom: "10px" }}
            >
              <div className={`${styles.item} ${styles.live}`} onClick={() => {
                $set("active_live", item)
                if (item.status === 2) return history.push(`/news/liveReplay/${item.live_id}`)
                else return history.push(`/news/liveDesc`)
              }}>
                <img src={item.img_url} alt="" />
                <div className={styles.liveSticker}>
                  <img src={require(`@/assets/img/sticker-${item.status === 2 ? 'green' : 'red'}.png`)} alt="" />
                  <span>{item.status === 2 ? $t("news.replay") : $t("news.live")}</span>
                </div>
              </div>
            </SwipeAction>
            : <SwipeAction
              disabled={!isManager}
              style={{ marginBottom: "10px" }}
              right={[
                {
                  text: item.is_top ? $t("news.action.cancelTop") : $t("news.action.top"),
                  className: styles.actionNormal,
                  async onPress() {
                    const isTop = item.is_top
                    const isConfirm = await Confirm($t("action.tips"), isTop ? $t("news.confirmTop") : $t("news.confirmCancelTop"))
                    if (!isConfirm) return
                    let res
                    if (isTop) {
                      res = await ApiGetCancelTopNews(item.message_id)
                    } else {
                      res = await ApiGetTopNews(item.message_id)
                    }
                    if (res === "success") {
                      ToastSuccess($t("success.operator"))
                      await initList()
                    }
                  }
                }
              ]}
            >
              <div className={`${styles.broadcast} ${styles.item}`}>
                <p dangerouslySetInnerHTML={{ __html: handleBroadcast(item.data) }} />
                <div className={styles.footer}>
                  <div className={styles.footer_l}>
                    {Number(new Date(item.top_at)) > 0 && <i className="iconfont iconic_notice" />}
                    <span>{$t("news.broadcast")}</span>
                  </div>
                  <div className={styles.footer_r}>{moment(item.created_at).format("YYYY/MM/DD")}</div>
                </div>
              </div>
            </SwipeAction>}
        </div>
      )}
    </div> : <div className={styles.empty}>
      <img src={require('@/assets/img/no-news.png')} />
      <p>{$t('home.noNews')}</p>
    </div>}
    <NewsTypeActionModal show={show} setShow={setShow} $t={$t} />
  </div>
}

interface INewsActionProps {
  show: boolean
  setShow: (v: boolean) => void
  $t: any
}

const NewsTypeActionModal = (props: INewsActionProps) => (
  <Modal
    popup
    animationType="slide-up"
    visible={props.show}
    onClose={() => props.setShow(false)}
  >
    <div className={styles.addSelect}>
      <ul>
        {[props.$t("news.sendBroadcast"), props.$t("news.sendLive"), props.$t("action.cancel")].map((item, index) => (
          <li key={index} onClick={() => {
            if (index === 0) return history.push(`/broadcast/send`)
            if (index === 1) {
              $set("active_live", { img_url: "", category: 0, title: "", description: "" })
              return history.push(`/news/addLive`)
            }
            props.setShow(false)
          }}>
            <p>{item}</p>
          </li>
        ))}
      </ul>
    </div>
  </Modal>
)

export function handleBroadcast(s: string): string {
  const urls = httpString(s)
  for (let i = 0; i < urls.length; i++) {
    const url = urls[i]
    s = s.replace(url, `<a href="${url}">${url}</a>`)
  }
  return s
}

function httpString(s: string): string[] {
  const reg = /(https?|http|ftp|file):\/\/[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]/g
  const res = s.match(reg)
  return res && res || []
}