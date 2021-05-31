import React, { useEffect, useState } from "react"
import { BackHeader } from "@/components/BackHeader"
import styles from "./index.less"
import { get$t } from "@/locales/tools"
import { useIntl } from "umi"
import { ApiGetPrsdigg, IPrsdiggArticle } from "@/apis/article"

import { Flex, PullToRefresh } from "antd-mobile"
import ReactDOM from "react-dom"
import moment from "moment"
import { FullLoading } from "@/components/Loading"
import myStyle from "@/pages/home/invite/my.less"
import { ApiGetAssetBySymbol } from "@/apis/asset";
import { $get } from "@/stores/localStorage";

interface IArticleProps {
  onClick: () => void
  avatar_url: string
  full_name: string
  icon_url: string
  amount: string
  symbol: string
  amount_usd: string
  created_at: string
  title: string
  content: string

  upvotes_count?: number
  comments_count?: number

  key_word?: string
  platform?: string
  rank?: number | string
}

export const Article = (props: IArticleProps) => {
  let icon = props.platform === "0" ? "iconbaidu1" : "icongoogle"

  return (
    <li className={styles.item} onClick={() => props.onClick()}>
      <header className={styles.head}>
        <img className={styles.avatar} src={props.avatar_url} alt=""/>
        <p className={styles.authorName}>{props.full_name}</p>
        <div className={styles.asset}>
          <img className={styles.coinAvatar} src={props.icon_url} alt=""/>
          <p>
            {props.amount} {props.symbol}
          </p>
        </div>
        <span>{moment(props.created_at).format("YYYY-MM-DD")}</span>
        <span className={styles.usd}>≈ $ {props.amount_usd}</span>
      </header>
      <h4 className={styles.title}>{props.title}</h4>
      <p className={styles.info}>{props.content}</p>
      <footer className={styles.foot}>
        {props.key_word ? (
          <>
            <div>
              <span>#{props.key_word}#</span>
            </div>
            <div>
              <i className={`iconfont ${icon} ${styles.icon}`}/>
              <span>{props.rank}</span>
            </div>
          </>
        ) : (
          <>
            <div>
              <i className={`iconfont iconzan3 ${styles.icon}`}/>
              <span>{props.upvotes_count}</span>
            </div>
            <div>
              <i className={`iconfont iconliuyan32 ${styles.icon}`}/>
              <span>{props.comments_count}</span>
            </div>
          </>
        )}
      </footer>
    </li>
  )
}

let ptr: any
let oneSubmit = false
let tmpLight = false
export default () => {
  const $t = get$t(useIntl())
  const [articleList, setArticleList] = useState<IPrsdiggArticle[]>([])
  const [isLoading, setLoading] = useState(false)
  const [height, setHeight] = useState(document.documentElement.clientHeight)
  const [isLoaded, setLoaded] = useState(false)

  useEffect(() => {
    let timmer: any = setInterval(() => {
      tmpLight = !tmpLight
    }, 300)
    return () => {
      clearInterval(timmer)
      timmer = null
    }
  }, [])

  useEffect(() => {
    initPage()
    if (articleList.length > 0) {
      const node: any = ReactDOM.findDOMNode(ptr)
      const hei = height - node.offsetTop
      setTimeout(() => setHeight(hei), 0)
    }
    return () => {
      oneSubmit = false
    }
  }, [])

  const initPage = async () => {
    if (oneSubmit) return
    oneSubmit = true
    let articles = await ApiGetPrsdigg(
      $get('group').symbol,
      articleList.length,
    )
    const obj: any = {}
    articles.forEach((item) => (obj[item.currency] = true))
    const assetInfo = await Promise.all(
      Object.keys(obj).map((item) => ApiGetAssetBySymbol(item)),
    )
    assetInfo.forEach(
      (assetList) => (obj[assetList[0].symbol!] = assetList[0].icon_url),
    )
    const newArticleList = articles.map((item) => ({
      ...item,
      asset_url: obj[item.currency],
    }))
    setArticleList([...articleList, ...newArticleList])
    oneSubmit = false
    setLoaded(true)
  }

  return (
    <div className={styles.container}>
      <BackHeader name={$t("home.article")}/>
      {!isLoaded ? (
        <FullLoading/>
      ) : articleList.length > 0 ? (
        <PullToRefresh
          className={styles.list}
          damping={60}
          ref={(el) => (ptr = el)}
          style={{
            height: height,
            overflow: "auto",
          }}
          indicator={{ deactivate: "下拉完成刷新" }}
          direction={"up"}
          refreshing={isLoading}
          onRefresh={async () => {
            setLoading(true)
            await initPage()
            setLoading(false)
          }}
        >
          {articleList.map((article, idx) => (
            <Article
              key={idx}
              onClick={() => (window.location.href = article.original_url)}
              avatar_url={article.author.avatar}
              full_name={article.author.name}
              icon_url={article.asset_url!}
              amount={String(article.price!)}
              symbol={article.currency}
              amount_usd={String(article.price_usd!)}
              created_at={article.created_at}
              title={article.title}
              content={article.intro}
              upvotes_count={article.upvotes_count}
              comments_count={article.comments_count}
            />
          ))}
        </PullToRefresh>
      ) : (
        <Flex className={myStyle.noInvited} direction="column" justify="center">
          <img
            src="https://taskwall.zeromesh.net/group-manager/no_invited.png"
            alt=""
          />
          <span>暂无文章...</span>
        </Flex>
      )}

      {/*{$get("setting").article_status === "1" && (*/}
      {/*  <img*/}
      {/*    className={styles.earn}*/}
      {/*    src={*/}
      {/*      light*/}
      {/*        ? require("@/assets/img/article_write1.png")*/}
      {/*        : require("@/assets/img/article_write2.png")*/}
      {/*    }*/}
      {/*    onClick={() => history.push(`/article/earn`)}*/}
      {/*  />*/}
      {/*)}*/}
    </div>
  )
}
