import React, { useEffect, useState } from 'react'
import { BackHeader } from "@/components/BackHeader"
import { history } from 'umi'
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { Switch } from 'antd-mobile'
import { SliderConfirm } from "@/components/BottomkModal/number"
import { ToastSuccess } from "@/components/Sub"
import { ApiGetGroupAdvanceSetting, ApiPutGroupAdvanceSetting, IAdvanceSetting } from '@/apis/user'
import styles from '@/pages/home/setting.less'
import { $get, $set } from "@/stores/localStorage"
import { Icon } from '@/components/Icon'

export default function Page() {
  const $t = get$t(useIntl())
  const [showSlider, setShowSlider] = useState(false)
  const [setting, setSetting] = useState<IAdvanceSetting>($get('advance_setting') || {})
  const [active, setActive] = useState("")
  const [tips, setTips] = useState("")

  const confirmAction = async () => {
    console.log(active)
    const p: IAdvanceSetting = {
      new_member_notice: "",
      conversation_status: ""
    }
    if (active === "mute") {
      p.conversation_status = setting.conversation_status === "1" ? "0" : "1"
    } else if (active === "new") {
      p.new_member_notice = setting.new_member_notice === "1" ? "0" : "1"
    }
    const res = await ApiPutGroupAdvanceSetting(p)
    if (res === 'success') {
      ToastSuccess($t('success.operator'))
      await initPage()
      setShowSlider(false)
    }
  }

  const initPage = async () => {
    const setting = await ApiGetGroupAdvanceSetting()
    setSetting(setting)
    $set("advance_setting", setting)
  }

  useEffect(() => {
    initPage()
  }, [])


  return (
    <div>
      <BackHeader name={$t('advance.title')} />
      <ul className={styles.list}>
        <li className={styles.formItem}>
          <div className={styles.formItemLeft}>
            <Icon i="a-chicangbizhong2" />
            <p>{$t('advance.mute')}</p>
          </div>
          <Switch
            color="black"
            checked={setting ? setting.conversation_status === "1" : true}
            onChange={() => {
              setShowSlider(true)
              const operatorAction = setting.conversation_status === "1" ? "close" : "open"
              const action = $t("advance." + operatorAction)
              setTips($t("advance.muteConfirm", {
                action,
                tips: operatorAction === "open" ? $t("advance.muteTips") : ""
              }))
              setActive("mute")
            }}
          />
        </li>
        <li className={styles.formItem}>
          <div className={styles.formItemLeft}>
            <Icon i="a-chicangbizhong2" />
            <p>{$t('advance.newMember')}</p>
          </div>
          <Switch
            color="black"
            checked={setting ? setting.new_member_notice === "1" : true}
            onChange={() => {
              setShowSlider(true)
              const action = $t("advance." + (setting.new_member_notice === "1" ? "close" : "open"))
              setTips($t("advance.newMemberConfirm", { action, tips: "" }))
              setActive("new")
            }}
          />
        </li>
        <li
          className={styles.formItem}
          onClick={() => history.push('/manager/advance/auth')}
        >
          <div className={styles.formItemLeft}>
            <Icon i="a-ruqunhuanyingyu1" />
            <p>{$t('advance.msgAuth')}</p>
          </div>
          <Icon i="ic_arrow" />
        </li>
      </ul>
      <SliderConfirm
        show={showSlider}
        setShow={setShowSlider}
        title={tips}
        content={$t('advance.sliderConfirm')}
        confirm={confirmAction}
      />
    </div>
  )
}
