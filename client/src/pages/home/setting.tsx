import React, { useEffect, useState } from 'react';
import { BackHeader } from "@/components/BackHeader";
import { history } from 'umi'
import { get$t } from "@/locales/tools";
import { useIntl } from "@@/plugin-locale/localeExports";
import redStyle from "@/pages/red/red.less";
import styles from "@/pages/setting/invite.less";
import { Switch } from 'antd-mobile';
import { NumberConfirm } from "@/components/BottomkModal/number";
import { Confirm, ToastFailed, ToastSuccess } from "@/components/Sub";
import { ApiGetMe, ApiPostChatStatus, IUser } from '@/apis/user';
import settingStyle from './setting.less'
import { ApiDeleteGroup } from "@/apis/group";
import { $get, $set } from "@/stores/localStorage";

export default function Page() {
  const $t = get$t(useIntl())
  const [show, setShow] = useState(false)
  const [user, setUser] = useState<IUser>($get('_user'))

  const toggleReceive = async () => {
    const res = await ApiPostChatStatus(!user?.is_received, user!.is_notice_join)
    if (res.user_id) {
      ToastSuccess($t("success.operator"))
      setUser(res)
      $set("_user", res)
      setShow(false)
    }
  }
  const toggleNoticeJoin = async () => {
    if (!user?.is_received) {
      ToastFailed($t("setting.receivedFirst"))
      return
    }
    const res = await ApiPostChatStatus(user!.is_received, !user!.is_notice_join)
    if (res.user_id) {
      ToastSuccess($t("success.operator"))
      $set("_user", res)
      setUser(res)
    }
  }


  useEffect(() => {
    ApiGetMe().then(user => {
      setUser(user)
      $set("_user", user)
    })
  }, [])


  return (
    <div>
      <BackHeader name={$t('setting.title')}/>
      <ul className={redStyle.list}>
        <li className={styles.formItem}>
          <p>{$t('setting.accept')}</p>
          <Switch
            color="black"
            checked={user ? user.is_received : true}
            onChange={() => {
              if (user?.is_received) return setShow(true)
              toggleReceive()
            }}
          />
        </li>
        <p className={styles.desc}>{$t('setting.acceptTips')}</p>
        <li className={styles.formItem}>
          <p>{$t('setting.newNotice')}</p>
          <Switch
            color="black"
            checked={user ? user.is_notice_join : true}
            onChange={toggleNoticeJoin}
          />
        </li>
        <li
          className={`${styles.formItem} ${settingStyle.blue}`}
          onClick={async () => {
            const isConfirm = await Confirm($t('action.tips'), $t('setting.authConfirm'))
            if (isConfirm) {
              localStorage.clear()
              history.push(`/auth`)
            }
          }}
        >{$t('setting.auth')}</li>
        <li
          className={`${styles.formItem} ${settingStyle.red}`}
          onClick={async () => {
            const isConfirm = await Confirm($t('action.tips'), $t('setting.exitConfirm'))
            if (isConfirm) {
              const res = await ApiDeleteGroup()
              if (res === "success") {
                ToastSuccess($t('success.operator'))
                localStorage.clear()
                history.push(`/exit`)
              }
            }
          }}
        >{$t('setting.exit')}</li>
      </ul>
      <NumberConfirm
        show={show}
        setShow={setShow}
        title={$t("setting.cancel.title")}
        content={$t('setting.cancel.content')}
        confirm={() => toggleReceive()}
      />

    </div>
  );
}
