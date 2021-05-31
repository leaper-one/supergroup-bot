import React, { useEffect, useState } from "react"
import styles from "./article.less"
import { history } from "umi"
import { Modal } from "antd-mobile"
import { Confirm, ToastFailed, ToastSuccess } from "@/components/Sub"
import { $set } from "@/stores/localStorage"
import {
  ApiGetArticle,
  ApiGetArticleCrawlList,
  ApiGetArticleCrawlListByID,
  ApiPutArticle,
  ApiPutArticleCrawl,
  ApiPutArticlePass,
  IArticle,
} from "@/apis/article"
import moment from "moment"

let articleList: IArticle[] = []
const tabs = ["全部", "申请中", "未打分", "已打分"]
const mainTabs = ["文章管理", "抓取管理"]
let articleMap: any = {}
export default () => {
  const [activeList, setActiveList] = useState([] as IArticle[])
  const [activeArticle, setActiveArticle] = useState({} as IArticle)
  const [modal, setModal] = useState(false)
  const [score, setScore] = useState("")
  const [comment, setComment] = useState("")
  const [activeTab, setActiveTab] = useState(0)
  const [activeMainTab, setActiveMainTab] = useState(1)
  const [articleCrawlList, setArticleCrawlList] = useState([] as IArticle[])

  const [crawlModal, setCrawlModal] = useState(false)
  const [activeCrawl, setActiveCrawl] = useState({} as IArticle)
  const [link, setLink] = useState("")
  const [platform, setPlatform] = useState("0")
  const [rank, setRank] = useState(0)
  const reloadCrawlList = (id = "") =>
    id
      ? ApiGetArticleCrawlListByID(id).then(setArticleCrawlList)
      : ApiGetArticleCrawlList().then(setArticleCrawlList)

  useEffect(() => {
    const group_id = history.location.query?.id
    if (group_id) $set("group", { group_id })
    reloadList()
    reloadCrawlList()
  }, [])

  useEffect(() => {
    if (score === "") return
    const s = Number(score)
    if (s === 1) return setComment("")
    if (s === 0) {
      setComment(
        `你的文章《${activeArticle.title}》${activeArticle.link} 不符合活动内容要求，奖励已取消。`,
      )
    } else if (s < 1) {
      setComment(
        `你的文章《${activeArticle.title}》${activeArticle.link} 推荐系统从 1 降为 ${score}，文章质量有待提升，感谢参与。`,
      )
    } else if (s < 2) {
      setComment(
        `你的文章《${activeArticle.title}》${activeArticle.link} 推荐系统从 1 升为 ${score}，文章不错继续加油！`,
      )
    } else {
      setComment(
        `你的文章《${activeArticle.title}》${activeArticle.link} 被选为优秀文章，推荐系统从 1 升为 ${score}，同时立刻获得 1 篇投稿资格，你还可以再多投 1 篇文章。`,
      )
    }
  }, [score])

  useEffect(() => {
    switchTab()
  }, [activeTab])

  const reloadList = async (status = "all") => {
    const allList = await ApiGetArticle(status)
    articleList = allList
    for (const item of allList) {
      articleMap[item.article_id] = item
    }
    switchTab()
  }

  const switchTab = () => {
    switch (activeTab) {
      case 0:
        return setActiveList(articleList)
      case 1:
        return setActiveList(articleList.filter((item) => item.status === "1"))
      case 2:
        return setActiveList(
          articleList.filter(
            (item) => item.status === "2" && item.score === "1",
          ),
        )
      case 3:
        return setActiveList(
          articleList.filter(
            (item) => item.status === "2" && item.score !== "1",
          ),
        )
    }
  }

  return (
    <div>
      <ul className={styles.head}>
        {mainTabs.map((item, idx) => (
          <li
            key={idx}
            className={(activeMainTab === idx && styles.active) || ""}
            onClick={() => {
              setActiveMainTab(idx)
            }}
          >
            {item}
          </li>
        ))}
      </ul>
      {/* 文章管理 */}
      {activeMainTab === 0 && (
        <>
          <ul className={styles.head}>
            {tabs.map((item, idx) => (
              <li
                key={idx}
                className={(activeTab === idx && styles.active) || ""}
                onClick={() => {
                  setActiveTab(idx)
                }}
              >
                {item}
              </li>
            ))}
          </ul>
          <ul className={styles.list}>
            <li>
              <span>用户名</span>
              <span>ID</span>
              <span>关键字</span>
              <span>标题</span>
              <span>链接</span>
              <span>分数</span>
              <span>操作</span>
            </li>
            {activeList.map((item) => (
              <li key={item.article_id}>
                <div>
                  <img src={item.avatar_url} alt="" />
                  <span>{item.full_name}</span>
                </div>
                <span title={item.identity_number}>{item.identity_number}</span>
                <span title={item.key_word}>
                  {item.key_word}
                  <a
                    href={`https://www.google.com/search?q=${encodeURIComponent(
                      item.key_word,
                    )}`}
                    target="_blank"
                  >
                    <i className="iconfont icongoogle" />
                  </a>
                  <a
                    href={`https://www.baidu.com/s?wd=${encodeURIComponent(
                      item.key_word,
                    )}`}
                    target="_blank"
                  >
                    <i className="iconfont iconbaidu1" />
                  </a>
                </span>
                <span title={item.title}>{item.title}</span>
                <span title={item.link}>
                  <a href={item.link} target="_blank">
                    {item.link}
                  </a>
                </span>
                <span title={item.score}>{item.score}</span>
                <span>
                  {item.status === "2" ? (
                    <button
                      onClick={() => {
                        setActiveArticle(item)
                        setScore(item.score || "1")
                        if (item.comment) {
                          setComment(item.comment)
                        }
                        setModal(true)
                      }}
                    >
                      打分
                    </button>
                  ) : (
                    <button
                      onClick={async () => {
                        const isConfirm = await Confirm(
                          "提示",
                          "确认审核通过吗？",
                        )
                        if (isConfirm) {
                          const res = await ApiPutArticlePass(item.article_id)
                          if (res === true) {
                            ToastSuccess("审核成功")
                            await reloadList()
                          }
                        }
                      }}
                    >
                      审核
                    </button>
                  )}
                </span>
              </li>
            ))}
          </ul>
        </>
      )}
      {/* 抓取管理 */}
      {activeMainTab === 1 && (
        <>
          <ul className={`${styles.c_list} ${styles.list}`}>
            <li>
              <span>用户名</span>
              <span>关键字</span>
              <span>标题</span>
              <span>链接</span>
              <span>平台</span>
              <span>排名</span>
              <span>时间</span>
              <span>操作</span>
            </li>
            {articleCrawlList.map((item, idx) => (
              <li key={idx}>
                <div>
                  <img src={articleMap[item.article_id]?.avatar_url} alt="" />
                  <span onClick={() => item.rank && reloadCrawlList()}>
                    {articleMap[item.article_id]?.full_name}
                  </span>
                </div>
                <span title={articleMap[item.article_id]?.key_word}>
                  {articleMap[item.article_id]?.key_word}

                  <a
                    href={`https://www.google.com/search?q=${encodeURIComponent(
                      articleMap[item.article_id]?.key_word,
                    )}`}
                    target="_blank"
                  >
                    <i className="iconfont icongoogle" />
                  </a>
                  <a
                    href={`https://www.baidu.com/s?wd=${encodeURIComponent(
                      articleMap[item.article_id]?.key_word,
                    )}`}
                    target="_blank"
                  >
                    <i className="iconfont iconbaidu1" />
                  </a>
                </span>
                <span
                  title={articleMap[item.article_id]?.title}
                  onClick={() => !item.rank && reloadCrawlList(item.article_id)}
                >
                  {articleMap[item.article_id]?.title}
                </span>
                <span title={articleMap[item.article_id]?.link}>
                  <a href={articleMap[item.article_id]?.link} target="_blank">
                    {articleMap[item.article_id]?.link}
                  </a>
                </span>
                {item.rank && (
                  <>
                    <span>{item.platform === "0" ? "百度" : "谷歌"}</span>
                    <span>{item.rank}</span>
                    <span>{moment(item.created_at).format("MM-DD")}</span>
                    <span>
                      <button
                        onClick={() => {
                          setCrawlModal(true)
                          setActiveCrawl(item)
                          setRank(item.rank!)
                          setPlatform(item.platform!)
                          setLink(articleMap[item.article_id].link)
                        }}
                      >
                        修改
                      </button>
                    </span>
                  </>
                )}
              </li>
            ))}
          </ul>
        </>
      )}

      <Modal
        visible={crawlModal}
        transparent
        maskClosable={false}
        onClose={() => setCrawlModal(false)}
        title={articleMap[activeCrawl.article_id]?.title}
        footer={[
          {
            text: "取消",
            onPress: () => setCrawlModal(false),
          },
          {
            text: "确认",
            onPress: async () => {
              const { article_id, created_at } = activeCrawl
              const res = await ApiPutArticleCrawl({
                article_id,
                created_at,
                link,
                platform,
                rank,
              })
              if (res === true) {
                ToastSuccess("操作成功")
                setCrawlModal(false)
                reloadCrawlList()
              }
            },
          },
        ]}
      >
        <div className={styles.modal}>
          <label>
            链接：
            <input
              type="text"
              placeholder="1"
              value={link}
              onChange={(e) => setLink(e.target.value)}
            />
          </label>
          <p>0：百度 1：谷歌</p>
          <label>
            平台：
            <input
              type="text"
              placeholder="1"
              value={platform}
              onChange={(e) => setPlatform(e.target.value)}
            />
          </label>
          <label>
            排名：
            <input
              type="text"
              placeholder="1"
              value={rank}
              onChange={(e) => setRank(Number(e.target.value))}
            />
          </label>
          {/*<label>链接：<input type="text" placeholder="1" value={link} onChange={e => setLink(e.target.value)}/></label>*/}
        </div>
      </Modal>
      <Modal
        visible={modal}
        transparent
        maskClosable={false}
        onClose={() => setModal(false)}
        title="打分"
        footer={[
          {
            text: "取消",
            onPress: () => {
              setModal(false)
            },
          },
          {
            text: "确认",
            onPress: async () => {
              const data = await ApiPutArticle({
                ...activeArticle,
                score,
                comment,
              })
              if (data === true) {
                ToastSuccess("操作成功")
                setModal(false)
                reloadList()
              } else if (data.code === 405) {
                ToastFailed("请勿重复提交")
                setModal(false)
              }
            },
          },
        ]}
      >
        <div className={styles.modal}>
          <label>
            分数：
            <input
              type="number"
              placeholder="1"
              value={score}
              onChange={(e) => {
                if (Number(e.target.value) > 5) {
                  ToastFailed("最大为5倍")
                  setScore("5")
                } else {
                  setScore(e.target.value)
                }
              }}
            />
          </label>
          <label>
            备注：
            <textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              placeholder="请输入发送给用户的消息"
            />
          </label>
        </div>
      </Modal>
    </div>
  )
}
