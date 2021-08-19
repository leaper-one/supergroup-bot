import React, { useEffect, useRef, useState } from "react"
import styles from "./index.less"
// @ts-ignore
import Qrcode from "qrious"
import { CodeURLIcon } from "@/components/CodeURL/icon"
import { IGroup } from "@/apis/group"
import { get$t } from "@/locales/tools"
import { useIntl } from "umi"
import { Button } from "../Sub"

interface Props {
  groupInfo: IGroup | undefined
  action: string
}

export const CodeURL = (props: Props) => {
  const $t = get$t(useIntl())
  const { groupInfo } = props
  if (!groupInfo) return <></>
  const [lang, setLang] = useState("en")
  const canvas: any = useRef()
  useEffect(() => {
    if (navigator.language.includes("zh") && process.env.LANG === "zh") setLang("zh")
    new Qrcode({
      element: canvas.current,
      value: window.location.href,
      level: "H",
      padding: 0,
      size: 300,
    })
  }, [])

  return (
    <>
      <div className={styles.container}>
        <CodeURLIcon icon_url={groupInfo?.icon_url} />
        <div className={styles.title}>{groupInfo.name}</div>
        <p>{groupInfo?.description}</p>

        <canvas className={styles.code} ref={canvas} />

        <span>
          {$t("join.code.invite")}
        </span>

        {<Button className={styles.openBtn} onClick={() => location.href = `mixin://apps/${groupInfo.client_id}?action=open`}>{$t('join.open')}</Button>}

        <a href={lang === 'en' ? "https://mixin-www.zeromesh.net/messenger" : "https://mixindl.com/#/"}>
          {$t(`join.code.${lang === 'en' ? "download" : "downloadXinsheng"}`)}
        </a>
      </div>
    </>
  )
}
