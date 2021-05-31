import React, { useEffect, useRef } from "react"
import styles from "./index.less"
// @ts-ignore
import Qrcode from "qrious"
import { CodeURLIcon } from "@/components/CodeURL/icon"
import { IGroup } from "@/apis/group"
import { get$t } from "@/locales/tools"
import { useIntl } from "umi"

interface Props {
  groupInfo: IGroup | undefined
  action: string
}

export const CodeURL = (props: Props) => {
  const $t = get$t(useIntl())
  const { groupInfo } = props
  if (!groupInfo) return <></>
  const canvas: any = useRef()
  useEffect(() => {
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
          {$t("join.code.invite", {
            action: $t("join.code.action." + props.action),
          })}
        </span>
        <a href="https://mixin-www.zeromesh.net/messenger">
          {$t("join.code.download")}
        </a>
      </div>
    </>
  )
}
