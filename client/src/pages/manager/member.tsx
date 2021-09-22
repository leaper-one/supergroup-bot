import React, { useEffect, useState } from "react"
import styles from "./member.less"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { SwipeAction, Modal } from "antd-mobile"
import {
  ApiPostSearchUserList,
  ApiGetUserList,
  ApiPutUserStatus,
  ApiPutUserMute,
  ApiPutUserBlock,
  ApiGetUserStat,
  IUser,
  IClientUserStat,
} from "@/apis/user"
import moment from "moment"
import { Button, Confirm, ToastSuccess, ToastWarning } from "@/components/Sub"
import { $get } from "@/stores/localStorage"
import { Icon } from "@/components/Icon"

let page = 1
let loading = false
let timer: any = null
let tmpUser: any

export default function Page() {
  const $t = get$t(useIntl())
  const [userList, setUserList] = useState<IUser[]>()
  const [key, setKey] = useState<string>("")
  const [searchList, setSearchList] = useState<IUser[]>()
  const [muteTime, setMuteTime] = useState<string>("12")
  const [muteModal, setMuteModal] = useState<boolean>(false)
  const [statusModal, setStatusModal] = useState<boolean>(false)
  const [status, setStatus] = useState<string>("all")
  const [stat, setStat] = useState<IClientUserStat>()
  useEffect(() => {
    ApiGetUserStat().then(setStat)
    loadList(true)
    return () => {
      page = 1
      loading = false
      timer = null
      tmpUser = undefined
    }
  }, [status])

  const isOwner = $get("group").owner_id === $get("user").user_id
  const clickSetGuestOrManager = async (user: IUser) => {
    const status = isOwner ? 9 : 8
    const { full_name, identity_number, user_id } = user
    const isConfirm = await Confirm(
      $t("action.tips"),
      $t(
        `member.action.${
          user.status === status ? "confirmCancel" : "confirmSet"
        }`,
        {
          full_name,
          identity_number,
          c: $t(`member.action.${isOwner ? "admin" : "guest"}`),
        },
      ),
    )
    if (!isConfirm) return
    let is_cancel = false
    if (user.status === status) is_cancel = true
    const res = await ApiPutUserStatus(user_id!, status, is_cancel)
    if (res === "success") {
      ToastSuccess($t("success.operator"))
      loadList(true)
    }
  }

  const clickMute = async () => {
    setMuteModal(false)
    const { full_name, identity_number, user_id } = tmpUser
    const isConfirm = await Confirm(
      $t("action.tips"),
      $t("member.action.confirmMute", {
        full_name,
        identity_number,
        mute_time: muteTime,
      }),
    )
    if (!isConfirm) return
    const res = await ApiPutUserMute(user_id, muteTime)
    if (res === "success") {
      ToastSuccess($t("success.operator"))
      loadList(true)
    }
  }

  const clickBlock = async (user: IUser) => {
    const { full_name, identity_number, user_id } = user
    const isConfirm = await Confirm(
      $t("action.tips"),
      $t("member.action.confirmBlock", { full_name, identity_number }),
    )
    if (!isConfirm) return
    const res = await ApiPutUserBlock(user_id!, false)
    if (res === "success") {
      ToastSuccess($t("success.operator"))
      loadList(true)
    }
  }

  const loadList = async (init = false) => {
    if (loading) return
    loading = true
    if (init) {
      page = 1
    }
    const [users] = await Promise.all([ApiGetUserList(page, status)])
    if (!init && users.length === 0 && userList && userList.length > 0)
      return ToastWarning($t("member.done"))
    if (page > 1) setUserList([...userList!, ...users])
    else setUserList(users)
    page++
    loading = false
  }

  useEffect(() => {
    if (key.length === 0) return setSearchList(undefined)
    if (timer) clearTimeout(timer)
    timer = setTimeout(
      () => ApiPostSearchUserList(key).then(setSearchList),
      500,
    )
  }, [key])

  return (
    <div className={`${styles.container} safe-view`}>
      <BackHeader
        name={$t("member.title")}
        action={
          <i
            className={`iconfont iconshaixuan ${styles.filter}`}
            onClick={() => setStatusModal(true)}
          />
        }
      />
      <div className={styles.search}>
        <Icon i="search" />
        <input
          type="text"
          placeholder="Mixin ID, Name"
          value={key}
          onChange={(e) => setKey(e.target.value)}
        />
      </div>
      <div
        className={styles.list}
        onScroll={(event) => {
          if (loading || status != "all") return
          const { scrollTop, scrollHeight, clientHeight } = event.target as any
          if (scrollTop + clientHeight + 500 > scrollHeight) loadList()
        }}
      >
        {(searchList || userList)?.map((item, idx) => (
          <SwipeAction
            className={styles.swiper}
            key={idx}
            right={getActionList(
              $t,
              item,
              isOwner,
              clickSetGuestOrManager,
              setMuteModal,
              clickBlock,
            )}
          >
            <div className={styles.item}>
              <img src={item.avatar_url} alt="" />
              <div className={styles.itemName}>
                <h5>{item.full_name}</h5>
                {[8, 9].includes(item.status!) && (
                  <i>{$t(`member.status${item.status}`)}</i>
                )}
              </div>
              <p>{getActiveTime($t, item.active_at!)}</p>
              <span>{item.identity_number}</span>
              <span>{moment(item.created_at).format("YYYY-MM-DD")}</span>
            </div>
          </SwipeAction>
        ))}
      </div>
      <MuteModal
        muteModal={muteModal}
        setMuteModal={setMuteModal}
        $t={$t}
        muteTime={muteTime}
        setMuteTime={setMuteTime}
        clickMute={clickMute}
      />
      <StatusModal
        stat={stat}
        statusModal={statusModal}
        setStatusModal={setStatusModal}
        $t={$t}
        status={status}
        setStatus={setStatus}
      />
    </div>
  )
}

interface IMuteModalProps {
  muteModal: boolean
  setMuteModal: (muteModal: boolean) => void
  $t: Function
  muteTime: string
  setMuteTime: (muteTime: string) => void
  clickMute: () => void
}
const MuteModal = (props: IMuteModalProps) => (
  <Modal
    popup
    visible={props.muteModal}
    onClose={() => props.setMuteModal(false)}
    animationType="slide-up"
  >
    <div className={styles.modal}>
      <h4>{props.$t("member.action.mute")}</h4>
      <img
        className={styles.modalClose}
        src={require("@/assets/img/svg/closeBtn.svg")}
        alt=""
        onClick={() => props.setMuteModal(false)}
      />
      <div className={styles.modalInput}>
        <input
          type="text"
          value={props.muteTime}
          onChange={(e) => props.setMuteTime(e.target.value)}
        />
        <span className={styles.modalUint}>
          {props.$t("member.modal.unit")}
        </span>
      </div>
      <p className={styles.modalDesc}>{props.$t("member.modal.desc")}</p>
      <Button onClick={() => props.clickMute()}>
        {props.$t("action.continue")}
      </Button>
    </div>
  </Modal>
)

interface ITypeModalProps {
  statusModal: boolean
  setStatusModal: (statusModal: boolean) => void
  $t: Function
  status: string
  setStatus: (type: string) => void
  stat?: IClientUserStat
}
const statusList = ["all", "admin", "guest", "mute", "block"]
type TStatus = "all" | "admin" | "guest" | "mute" | "block"

const StatusModal = (props: ITypeModalProps) => (
  <Modal
    popup
    visible={props.statusModal}
    onClose={() => props.setStatusModal(false)}
    animationType="slide-up"
  >
    <div className={styles.modal}>
      <h4>{props.$t("member.status.title")}</h4>
      <img
        className={styles.modalClose}
        src={require("@/assets/img/svg/closeBtn.svg")}
        alt=""
        onClick={() => props.setStatusModal(false)}
      />
      <ul className={styles.modalStatusList}>
        {statusList.map((type) => (
          <li
            key={type}
            className={styles.modalStatusItem}
            onClick={() => {
              props.setStatus(type)
              props.setStatusModal(false)
            }}
          >
            <span>{props.$t(`member.status.${type}`)}</span>
            <p>
              {props.stat && props.stat[type as TStatus]}{" "}
              {props.$t(`member.status.people`)}
            </p>
            {props.status === type && (
              <i
                className={`iconfont iconcheck ${styles.modalStatusActive}`}
              ></i>
            )}
          </li>
        ))}
      </ul>
    </div>
  </Modal>
)

function getActiveTime($t: any, time: string): string {
  const hourDuration = Math.ceil(
    (Date.now() - Number(new Date(time))) / 1000 / 3600,
  )
  if (hourDuration < 24) return $t("member.hour", { n: hourDuration })
  const dayDuration = Math.ceil(hourDuration / 24)
  if (dayDuration < 30) return $t("member.day", { n: dayDuration })
  const monthDuration = Math.ceil(dayDuration / 30)
  if (monthDuration < 12) return $t("member.month", { n: monthDuration })
  const yearDuration = Math.ceil(monthDuration / 12)
  return $t("member.year", { n: yearDuration })
}

function getActionList(
  $t: any,
  item: IUser,
  isOwner: boolean,
  clickSetGuestOrManager: any,
  setMuteModal: any,
  clickBlock: any,
): any {
  let actionList = [
    {
      text: $t(`member.action.${item.status! > 5 ? "cancel" : "set"}`, {
        c: $t(`member.action.${isOwner ? "admin" : "guest"}`),
      }),
      className: styles.action,
      onPress: () => clickSetGuestOrManager(item),
    },
    {
      text: $t("member.action.mute"),
      className: styles.action,
      onPress: () => {
        tmpUser = item
        setMuteModal(true)
      },
    },
    {
      text: $t("member.action.block"),
      className: styles.action,
      onPress: () => clickBlock(item),
    },
  ]
  if (item.status === 9 && !isOwner) return []
  return actionList
}
