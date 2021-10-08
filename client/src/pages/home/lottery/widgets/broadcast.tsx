import React, { FC } from "react"
import { Carousel } from "antd-mobile"
import styles from "./broadcast.less"
import { Lucker } from "@/types"
import { useIntl } from "react-intl"
import { get$t } from "@/locales/tools"

export interface BroadcastBoxProps {
  data: Lucker[]
}

export const BroadcastBox: FC<BroadcastBoxProps> = ({ data }) => {
  const t = get$t(useIntl())

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
            {data.map((item, idx) => (
              <div key={item.trace_id || idx} className={styles.item}>
                {item.full_name}&nbsp;
                {t("claim.drew")}
                &nbsp;{item.amount}&nbsp;
                {item.symbol}
                {Number(item.price_usd) > 0 &&
                  t("claim.worth", { value: item.price_usd, prefix: ", " })}
              </div>
            ))}
          </Carousel>
        )}
      </div>
    </div>
  )
}
