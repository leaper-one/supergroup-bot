import React, { useEffect, useState } from "react"
import tradeStyles from "../trade.less"
import { BackHeader } from "@/components/BackHeader"
import styles from "./my.less"
import { useIntl } from "umi"
import { get$t } from "@/locales/tools"
import { ApiGetInviteList, IInviteItem } from "@/apis/invite"
import { Flex } from "antd-mobile"
import { FullLoading } from "@/components/Loading"
import { $get } from '@/stores/localStorage'

export default () => {
  const $t = get$t(useIntl())
  const [total] = useState($get("invitation"))
  const [list, setList] = useState<IInviteItem[]>([] as IInviteItem[])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    ApiGetInviteList().then((list) => {
      setList(list)
      setLoading(false)
    })
  }, [])

  return (
    <>
      <div className={`${tradeStyles.container} ${styles.container} safe-view`}>
        <BackHeader
          name={$t("invite.my.title")}
          action={<i
            className={`iconfont iconbangzhu ${styles.helpIcon}`}
            onClick={() => (window.location.href = `https://w3c.group/c/1611914754694662`)}
          />}
        />
        <section className={tradeStyles.price}>
          <div className={tradeStyles.title}>
            <span>{$t("invite.my.reward")}</span>
            <span>{$t("invite.my.people")}</span>
          </div>
          <div className={tradeStyles.amount}>
            <span>
              {total.power}<i className={styles.symbol}>{$t('claim.energy.title')}</i>
            </span>
            <span className={styles.green}>{total.count}</span>
          </div>
        </section>
        {list.length > 0 ? (
          <div className={styles.list}>
            {list.map((item, idx) => (
              <div key={idx} className={styles.item}>
                <img src={item.avatar_url} alt="" />
                <span>{item.full_name}</span>
                <span>
                  {/* {$t("invite.my." + item.status)} */}
                  {item.amount + ' ' + $t('claim.energy.title')}
                </span>
                <span>{item.identity_number}</span>
                <span>{item.created_at}</span>
              </div>
            ))}
          </div>
        ) : (
          <Flex
            className={styles.noInvited}
            direction="column"
            justify="center"
          >
            <img
              src="https://taskwall.zeromesh.net/group-manager/no_invited.png"
              alt=""
            />
            <span>
              {$t("invite.my.noInvited")}，
              <a href="https://w3c.group/c/1611914754694662">
                {$t("invite.my.rule")}
              </a>
            </span>
          </Flex>
        )}
      </div>
      {loading && <FullLoading mask />}
    </>
  )
}

