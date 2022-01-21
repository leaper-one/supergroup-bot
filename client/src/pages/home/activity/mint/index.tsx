import { BackHeader } from "@/components/BackHeader"
import { history, useIntl } from "umi"
import { get$t, getDurationDays, getTime } from "@/locales/tools"
import React, {
  useState,
  useEffect,
} from "react"
import styles from "./index.less"
import { changeTheme, getURLParams } from '@/assets/ts/tools'
import { ApiGetMintByID, IMint } from '@/apis/mint'
import { FullLoading } from '@/components/Loading'
import { getAuthUrl } from '@/apis/http'

export default function () {
  let { id, state } = getURLParams()
  const $t = get$t(useIntl())
  if (!id && state) id = state
  const [mintData, setMintData] = useState<IMint>()
  const [showContinueModal, setContinueModal] = useState(false)
  const [showKnowModal, setShowKnowModal] = useState(false)

  useEffect(() => {
    ApiGetMintByID(id).then(d => {
      setMintData(d)
      changeTheme('#32004a')
    })
    return () => {
      changeTheme('#fff')
    }
  }, [])


  return <div>
    {mintData ? <div className={styles.intro}>
      <div className={styles.bg} style={{ backgroundImage: `url(${mintData.bg})` }}></div>
      <BackHeader className={styles.top} name={mintData.title} />
      <header className={`${styles.header} ${styles.top}`}>
        <div className={styles.title}>{mintData.title}</div>
      </header>
      <div className={`${styles.desc} ${styles.top}`}>{mintData.description}</div>
      <div className={styles.btn}>
        <div className={styles.btn_item} onClick={() => setContinueModal(true)}>{$t('mint.join')}</div>
        <div className={styles.btn_item} onClick={() => {
          console.log(mintData.status)
          if (['auth', 'pending'].includes(mintData.status)) return setShowKnowModal(true)
          else history.push('/mint/record?id=' + id)
        }}>{$t('mint.receive')}</div>
      </div>
      <Card $t={$t} mintData={mintData} type="first" />
      <Card $t={$t} mintData={mintData} type="daily" />
      <Card $t={$t} mintData={mintData} />
    </div> : <FullLoading mask />}
    {showContinueModal && <div className={styles.mask} onClick={() => setContinueModal(false)}>
      <div className={styles.mask_content} onClick={(e) => e.stopPropagation()}>
        <div className={styles.mask_main} dangerouslySetInnerHTML={{ __html: mintData!.join_tips }} />
        <div className={styles.mask_btn}>
          <div className={styles.btn_item} onClick={() => {
            location.href = mintData!.join_url
            setContinueModal(false)
          }}>{$t('mint.continue')}</div>
          <div className={styles.btn_item} onClick={() => setContinueModal(false)}>{$t('mint.close')}</div>
        </div>
      </div>
    </div>}
    {showKnowModal && <>
      {<div className={styles.mask} onClick={() => setShowKnowModal(false)}>
        <div className={styles.mask_content} onClick={(e) => e.stopPropagation()}>
          <div className={styles.mask_main}>{$t('mint.' + mintData?.status)}</div>
          <div className={styles.mask_btn}>
            <div className={`${styles.btn_item_fx_1} ${styles.btn_item}`} onClick={() => {
              setShowKnowModal(false)
              if (mintData?.status === 'auth') {
                location.href = getAuthUrl({ hasAssets: true, state: mintData?.mining_id })
              }
            }}>{$t('action.know')}</div>
          </div>
        </div>
      </div>}
    </>}
  </div >
}
interface CardProps {
  mintData: IMint
  type?: "first" | "daily"
  $t: any
}
const Card = (props: CardProps) => {
  const { $t, mintData, type } = props
  return <section className={`${styles.card} ${!type && styles.faq}`}>
    {type ? <>
      <div className={styles.title}>{$t('mint.' + type)}</div>
      {type === "first" && <div className={styles.theme}  >{$t("mint.theme")}</div>}
      {(() => {
        const start = mintData[`${type}_time`]
        const end = mintData[`${type}_end`]
        const [aY, aM, aD] = getTime(start)
        const [bY, bM, bD] = getTime(end)
        const d = getDurationDays(start, end)
        const param = { aY, aM, aD, bY, bM, bD, d }
        return <p>{$t('mint.time')}： {$t('mint.duration', param)}</p>
      })()}
      <p>{$t('mint.reward')}：{mintData[`${type}_desc`]}</p>
      <p>{$t('mint.receiveTime')}：{$t('mint.receiveTimeTips')}</p>
    </> : <>
      <div className={styles.title}>{$t('mint.faq')}</div>
      <p className={styles.faq_desc} dangerouslySetInnerHTML={{ __html: mintData!.faq }}></p>
    </>}
  </section>
}