import React, { useEffect, useState } from 'react';
import styles from './member.less';
import { BackHeader } from "@/components/BackHeader";
import { get$t } from "@/locales/tools";
import { useIntl } from "@@/plugin-locale/localeExports";
import { SwipeAction } from 'antd-mobile';
import { ApiGetUserList, IUser } from "@/apis/user";
import moment from 'moment'
import { ToastWarning } from "@/components/Sub";

let page = 1
let loading = false
export default function Page() {
  const $t = get$t(useIntl())
  const [userList, setUserList] = useState<IUser[]>()
  useEffect(() => {
    loadList()
    return () => {
      page = 1
      loading = false
    }
  }, [])

  const clickSetGuestOrManager = (user_id: string) => {
  }

  const clickMute = (user_id: string) => {
  }

  const clickBlock = (user_id: string) => {
  }

  const loadList = () => {
    if (loading) return
    loading = true
    ApiGetUserList(page).then(users => {
      if (users.length === 0) return ToastWarning($t("member.done"))
      if (page > 1) setUserList([...userList!, ...users])
      else setUserList(users)
      page++
      loading = false
    })
  }
  return (
    <div className={`${styles.container} safe-view`}>
      <BackHeader name={$t('member.title')}/>
      <div className={styles.search}>
        <i className="iconfont iconsearch"/>
        <input type="text" placeholder="Mixin ID, Name"/>
      </div>
      <div className={styles.list} onScroll={event => {
        if (loading) return
        const { scrollTop, scrollHeight, clientHeight } = event.target as any
        if (scrollTop + clientHeight + 200 > scrollHeight) loadList()
      }}>
        {userList?.map((item, idx) =>
          <SwipeAction
            key={idx}
            right={[{
              text: "设为嘉宾",
              className: styles.action,
              onPress: () => clickSetGuestOrManager(item.user_id!)
            }, {
              text: "禁言",
              className: styles.action,
              onPress: () => clickMute(item.user_id!)
            }, {
              text: "拉黑",
              className: styles.action,
              onPress: () => clickBlock(item.user_id!)
            }]}
          >
            <div className={styles.item}>
              <img src={item.avatar_url} alt=""/>
              <div className={styles.itemName}>
                <h5>{item.full_name}</h5>
                {[8, 9].includes(item.status!) && <i>{$t(`member.status${item.status}`)}</i>}
              </div>
              <p>{getActiveTime($t, item.active_at!)}</p>
              <span>{item.identity_number}</span>
              <span>{moment(item.created_at).format("YYYY-MM-DD")}</span>
            </div>
          </SwipeAction>)
        }
      </div>
    </div>
  );
}

function getActiveTime($t: any, time: string): string {
  const hourDuration = Math.ceil((Date.now() - Number(new Date(time))) / 1000 / 3600)
  if (hourDuration < 24) return $t('member.hour', { n: hourDuration })
  const dayDuration = Math.ceil(hourDuration / 24)
  if (dayDuration < 30) return $t('member.day', { n: dayDuration })
  const monthDuration = Math.ceil(dayDuration / 30)
  if (monthDuration < 12) return $t('member.month', { n: monthDuration })
  const yearDuration = Math.ceil(monthDuration / 12)
  return $t('member.year', { n: yearDuration })
}
