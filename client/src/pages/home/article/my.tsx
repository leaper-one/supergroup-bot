import React, { useEffect, useState } from "react"
import { BackHeader } from "@/components/BackHeader"
import { Article } from "@/pages/home/article/index"
import { history } from "umi"
import { ApiGetArticle, IArticle } from "@/apis/article"

export default () => {
  const [articleList, setArticleList] = useState([] as IArticle[])

  useEffect(() => {
    ApiGetArticle("my").then(setArticleList)
  }, [])

  return (
    <div>
      <BackHeader
        name="我的文章"
        action={<i className={`iconfont iconic_add`} onClick={() => history.push(`/article/apply`)}/>}
      />

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
          rank={article.status === "1" ? "待审核" : article.rank}
        />
      ))}
    </div>
  )
}
