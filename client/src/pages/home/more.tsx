import React from "react"
import styles from "./more.less"
import { staticUrl } from "@/apis/http"
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { history, useIntl } from "umi"
import { $get } from "@/stores/localStorage"

export default () => {
  const $t = get$t(useIntl())
  return (
    <>
      <BackHeader name={$t("home.more")} />
      <div className={styles.container}>
        {$get("setting").article_status === "1" && (
          <img
            onClick={() => history.push(`/article/earn`)}
            src={staticUrl + "more-3.png"}
            alt=""
          />
        )}
        <img
          onClick={() =>
            (window.location.href =
              "mixin://users/e08207df-55de-4ad9-8463-af692824f988")
          }
          src={staticUrl + "more-1.png"}
          alt=""
        />
        <img
          onClick={() =>
            (window.location.href =
              "mixin://users/1da1124a-9c97-4f2b-b332-f11f77c7604a")
          }
          src={staticUrl + "more-2.png"}
          alt=""
        />
      </div>
    </>
  )
}
