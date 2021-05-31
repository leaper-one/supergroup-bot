import React, { useEffect, useState } from "react"
import earnStyles from "./earn.less"
import { BackHeader } from "@/components/BackHeader"
import { history } from "umi"
import { Article } from "@/pages/home/article/index"
import styles from "@/pages/home/article/index.less"
import { ApiGetArticle, IArticle } from "@/apis/article"

let tmpLight = false
export default () => {
  const [articleList, setArticleList] = useState([] as IArticle[])
  const [activeTab, setActiveTab] = useState(0)

  const [light, setLight] = useState(false)

  useEffect(() => {
    let timmer: any = setInterval(() => {
      tmpLight = !tmpLight
      setLight(tmpLight)
    }, 300)
    return () => {
      clearInterval(timmer)
      timmer = null
      tmpLight = false
    }
  }, [])

  useEffect(() => {
    getArticle("list")
  }, [])

  const getArticle = async (status: string) => {
    const articles = await ApiGetArticle(status)
    setArticleList(articles)
  }

  return (
    <div className={earnStyles.container}>
      <BackHeader
        name="#写文赚币#"
        action={
          <i
            className={`iconfont iconfile-text ${earnStyles.my}`}
            onClick={() => history.push(`/article/my`)}
          />
        }
      />
      <ul className={earnStyles.tab}>
        {["热门文章", "最新提交"].map((item, idx) => (
          <li
            key={item}
            className={`${earnStyles.tabItem} ${
              (activeTab === idx && earnStyles.active) || ""
            }`}
            onClick={() => {
              setActiveTab(idx)
              getArticle(idx === 0 ? "list" : "new")
            }}
          >
            {item}
          </li>
        ))}
      </ul>

      {articleList.map((article, idx) => (
        <Article
          key={idx}
          onClick={() => (window.location.href = article.link)}
          avatar_url={article.avatar_url}
          full_name={article.full_name}
          icon_url={article.icon_url}
          amount={String(article.amount)}
          symbol={article.symbol}
          amount_usd={String(article.amount_usd!)}
          created_at={article.created_at}
          title={article.title}
          content={article.link}
          platform={article.platform}
          key_word={article.key_word}
          rank={article.rank}
        />
      ))}

      <img
        className={styles.earn}
        src={
          tmpLight
            ? require("@/assets/img/article_apply1.png")
            : require("@/assets/img/article_apply2.png")
        }
        onClick={() => history.push(`/article/apply`)}
      />
    </div>
  )
}
