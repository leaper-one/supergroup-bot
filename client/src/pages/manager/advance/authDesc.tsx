import React, { useEffect, useState } from 'react'
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { Switch } from 'antd-mobile'
import styles from '@/pages/home/setting.less'
import { Button, ToastSuccess } from '@/components/Sub'
import { getURLParams } from '@/assets/ts/tools'
import { $get } from '@/stores/localStorage'
import { ApiPutGroupMemberAuth } from '@/apis/user'

const authList = [
  "plain_text",
  "plain_sticker",
  "lucky_coin",
  "plain_image",
  "plain_video",
  "plain_post",
  "plain_data",
  "plain_live",
  "plain_contact",
  "plain_transcript",
  "url"
]

export default function Page() {
  const $t = get$t(useIntl())
  const [status] = useState(getURLParams()?.s || "1")
  const [auth, setAuth] = useState($get("auth")[status])
  const memberStatus = $t('advance.member.' + status)
  return (
    <div>
      <BackHeader name={memberStatus} />
      <ul className={styles.list}>
        {authList.map(item => <li key={item} className={styles.formItem}>
          <div className={styles.formItemLeft}>
            <p>{$t('advance.' + item)}</p>
          </div>
          <Switch
            color="black"
            checked={auth[item]}
            onChange={() => setAuth({ ...auth, [item]: !auth[item] })}
          />
        </li>)}
      </ul>
      <Button onClick={async () => {
        const res = await ApiPutGroupMemberAuth(auth)
        if (res === 'success') {
          ToastSuccess($t('success.save'))
          setTimeout(() => {
            history.go(-1)
          }, 500)
        }
      }}>{$t('action.save')}</Button>
      <p className={styles.tips}>{$t('advance.member.tips', {
        status: memberStatus,
        count: auth?.limit
      })}</p>
    </div>
  )
}
