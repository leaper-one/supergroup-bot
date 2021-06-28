import React from 'react';
import { BackHeader } from "@/components/BackHeader";
import { get$t } from "@/locales/tools";
import { useIntl } from "@@/plugin-locale/localeExports";
import { $get } from "@/stores/localStorage";
import { IActivity } from "@/apis/group";
import styles from './activity.less'

export default function Page() {
  const $t = get$t(useIntl())
  let activity: IActivity[] = $get("group").activity
  const now = new Date()
  activity = activity.map(item => ({
    ...item,
    isExpire: now > new Date(item.expire_at)
  }))


  return (
    <div className={styles.container}>
      <BackHeader name={$t('home.activity')} isWhite/>

      <div className={styles.content}>
        {activity.map(item =>
          <img key={item.activity_index}
               src={(item.isExpire ? item.expire_img_url : item.img_url) + '?t=1'}
               onClick={() => {
                 if (item.isExpire) return
                 location.href = item.action
               }}
               alt=""
          />)}
      </div>
    </div>
  );
}
