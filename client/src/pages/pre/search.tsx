import React, { useEffect, useState } from "react"
import styles from "./search.less"
import { BackHeader } from "@/components/BackHeader"
import { ApiGetGroupList, IGroupItem } from "@/apis/group"
import { history, useIntl } from "umi"
import { $get, $set } from "@/stores/localStorage"
import { base64Encode, getMixinCtx } from "@/assets/ts/tools"
import { get$t } from "@/locales/tools"

export default () => {
  const $t = get$t(useIntl())
  const [groupList, setGroupList] = useState<Array<IGroupItem>>(
    [] as IGroupItem[],
  )
  const [search, setSearch] = useState("")
  const mixinCtx = getMixinCtx()
  useEffect(() => {
    const groupList = $get("groupList")
    if (groupList) setGroupList(groupList)
  }, [])

  useEffect(() => {
    if (search)
      ApiGetGroupList().then((list) =>
        setGroupList(
          list.filter((item) =>
            item.name.toLowerCase().includes(search.toLowerCase()),
          ),
        ),
      )
    else
      ApiGetGroupList().then((groupList) => {
        $set("groupList", groupList)
        setGroupList(groupList)
      })
  }, [search])

  const handleClickShared = () => {
    let schema = `mixin://send?category=app_card&data=`
    const join_url = `${location.origin}/explore`
    schema += base64Encode({
      app_id: process.env.CLIENT_ID,
      icon_url: `https://mixin-images.zeromesh.net/AuKlAvXRYK1XHfvCIDeq07ThLYfgzmYUYIHw8UO-na-BGv-prgczmqptvHVvufwJA2WUnQ1cSXgNF5A_NS6F-hzZn0BJxfLAJASf=s256`,
      title: $t("pre.explore.button"),
      description: $t("pre.explore.title"),
      action: join_url,
    })
    window.location.href = schema
  }

  return (
    <div className={styles.container}>
      <BackHeader
        name={$t("join.title")}
        action={
          mixinCtx && (
            <i
              onClick={handleClickShared}
              className={`iconfont iconic_share ${styles.shareIcon}`}
            />
          )
        }
      />
      <div className={styles.search}>
        <i className="iconfont iconsearch" />
        <input
          type="text"
          placeholder={$t("join.search.name")}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>
      <ul>{groupList.map((item, idx) => GroupItem(item, idx, $t))}</ul>
    </div>
  )
}

const GroupItem = (group: IGroupItem, idx: number, $t: any) => {
  let content = $t("join.search.holder")
  group.check?.forEach(({ symbol, amount }, idx) => {
    if (idx !== 0) content += ` ${$t("join.search.or")}`
    content += ` ${amount} ${symbol}`
  })
  return (
    <li
      className={styles.group_item}
      key={idx}
      onClick={() =>
        history.push({
          pathname: `/join/${group.group_number}`,
          query: { from: "search" },
        })
      }
    >
      <img src={group.icon_url} alt="" />
      <p>{group.name}</p>
      <span className={styles.group_item_c}>
        {content.length === 4 ? "可预约空投" : content}
      </span>
      {content.length !== 4 && (
        <span className={styles.group_item_p}>{`${group.people} ${$t(
          "join.search.people",
        )}`}</span>
      )}
    </li>
  )
}
