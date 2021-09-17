import React, { FC, useMemo, memo } from "react"
import { Carousel } from "antd-mobile"
import styles from "./broadcast.less"
import { Lucker } from "@/types"
import { useIntl } from "react-intl"
import { get$t } from "@/locales/tools"

export interface BroadcastBoxProps {
  // uname: string
  // content: string
  data: Lucker[]
}

export const BroadcastBox: FC<BroadcastBoxProps> = memo(({ data }) => {
  const t = get$t(useIntl())
  console.log(data)

  const widgets = useMemo(
    () =>
      data.map((item, idx) => (
        <div key={item.trace_id || idx} className={styles.item}>
          {item.full_name}&nbsp;
          {t("claim.drew")}
          &nbsp;{item.amount}&nbsp;
          {item.symbol}
          {Number(item.price_usd) > 0 &&
            t("claim.worth", { value: item.price_usd, prefix: ", " })}
        </div>
      )),
    [data],
  )

  return (
    <div className={styles.container}>
      <div className={styles.content}>
        {!!data.length && (
          <Carousel
            vertical
            autoplay
            dots={false}
            infinite
            autoplayInterval={3000}
            className={styles.carousel}
          >
            {widgets}
          </Carousel>
        )}
      </div>
    </div>
  )
})
