import React from "react"
import { BackHeader } from "@/components/BackHeader"
import styles from "./findBot.less"
import { get$t } from "@/locales/tools"
import { useIntl } from "umi"

export default () => {
  const $t = get$t(useIntl())

  return (
    <div className={styles.container}>
      <BackHeader name={$t("home.findBot")} />
      <iframe
        id="iframe"
        className={styles.iframe}
        src={$t('home.findBotURL')}
      />
    </div>
  )
}
