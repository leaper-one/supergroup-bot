import React, { useEffect, useState } from "react"
import tradeStyles from "../trade.less"
import { BackHeader } from "@/components/BackHeader"
import styles from "./my.less"
import { useIntl } from "umi"
import { get$t } from "@/locales/tools"
import { ApiGetInviteList, IInviteItem } from "@/apis/invite"
import { IGroup, IGroupSetting } from "@/apis/group"
import { $get } from "@/stores/localStorage"
import { Flex } from "antd-mobile"
import moment from "moment"
import { FullLoading } from "@/components/Loading"

export default () => {
  const group: IGroup = $get("group")
  const setting: IGroupSetting = $get("setting")
  const $t = get$t(useIntl())
  const [list, setList] = useState<IInviteItem[]>([] as IInviteItem[])

  const [loading, setLoading] = useState(true)

  const open = setting?.invite_status === "1"

  useEffect(() => {
    ApiGetInviteList(group.group_id!, 0).then((list) => {
      setList(list)
      setLoading(false)
    })
  }, [])

  return (
    <>
      <div className={`${tradeStyles.container} ${styles.container}`}>
        <BackHeader
          name={$t("invite.my.title")}
          action={
            <i
              className={`iconfont iconbangzhu ${styles.helpIcon}`}
              onClick={() =>
                (window.location.href = `https://w3c.group/c/1611914754694662`)
              }
            />
          }
        />
        <section className={tradeStyles.price}>
          <div className={tradeStyles.title}>
            <span>{$t("invite.my.reward")}</span>
            <span>{$t("invite.my.people")}</span>
          </div>
          <div className={`${tradeStyles.amount} ${styles.amount}`}>
            {open ? (
              <span>
                0.000152<i className={styles.symbol}>BTC</i>
              </span>
            ) : (
              <span>{$t("invite.my.noTitle")}</span>
            )}
            <span className={styles.green}>{list.length}</span>
          </div>
          <span className={styles.usd}>
            {open ? `≈ $101.11` : $t("invite.my.noTips")}
          </span>
        </section>
        {list.length > 0 ? (
          list.map((item, idx) => (
            <div key={idx} className={styles.list}>
              <div className={styles.item}>
                <img src={item.avatar_url} alt="" />
                <span>{item.full_name}</span>
                <span className={item.status === "0" ? styles.red : ""}>
                  {$t("invite.my." + item.status)}
                </span>
                <span>{item.identity_number}</span>
                <span>{formatDate(item.created_at)}</span>
              </div>
            </div>
          ))
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

function formatDate(data: string): string {
  return moment(data).format("MM/DD")
}
