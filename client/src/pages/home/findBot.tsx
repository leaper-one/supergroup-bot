import React from "react"
import { BackHeader } from "@/components/BackHeader"
import styles from "./findBot.less"
import { get$t } from "@/locales/tools"
import { useIntl } from "umi"

export default () => {
  const $t = get$t(useIntl())
  const url = process.env.LANG === 'zh' ? 'https://hot-bots.mixinbots.com' : 'https://bots.mixin.zone'

  return (
    <div className={styles.container}>
      <BackHeader name={$t("home.findBot")} />
      <iframe
        id="iframe"
        className={styles.iframe}
        src={url}
      />
    </div>
  )
}
