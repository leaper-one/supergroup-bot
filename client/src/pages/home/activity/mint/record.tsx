import { ApiGetMintRecord, ApiPostMintByID, IMintRecord } from '@/apis/mint'
import { changeTheme, getURLParams } from '@/assets/ts/tools'
import { BackHeader } from "@/components/BackHeader"
import { FullLoading } from '@/components/Loading'
import { ToastFailed } from '@/components/Sub'
import { get$t } from '@/locales/tools'
import React, { useEffect, useState } from "react"
import { useIntl } from 'react-intl'
import styles from "./record.less"



export default function () {
  let { id } = getURLParams()
  const [recordList, setRecordList] = useState<IMintRecord[]>([])
  const [allList, setAllList] = useState<IMintRecord[]>([])
  const [currentStatus, setCurrentStatus] = useState(0)
  const [showMask, setShowMask] = useState(false)
  const [isLoaded, setLoaded] = useState(false)
  const $t = get$t(useIntl())
  useEffect(() => {
    initPage()
    return () => {
      changeTheme('#fff')
    }
  }, [])
  useEffect(() => {
    if (currentStatus === 0) setRecordList(formatRecord(allList))
    else if (currentStatus === 1)
      setRecordList(formatRecord(allList.filter(item => item.status === 1)))
    else if (currentStatus === 2)
      setRecordList(formatRecord(allList.filter(item => item.status === 2)))
  }, [currentStatus])

  const initPage = async () => {
    const list = await ApiGetMintRecord(id)
    setAllList(list)
    setRecordList(formatRecord(list))
    changeTheme('#230d78')
    setLoaded(true)
  }

  return <div>
    <div className={styles.container}>
      <BackHeader name={$t('mint.record.title')} />
      <ul className={styles.tab}>
        {[0, 2, 1].map(status => <li
          key={status}
          className={`${styles.tabItem} ${currentStatus === status && styles.active}`}
          onClick={() => setCurrentStatus(status)}
        > {$t('mint.record.' + status)} </li>)}
      </ul>
      {recordList.map((record, idx) => <div className={styles.card} key={idx}>
        <div className={styles.cardTitle}>
          <span className={styles.pair}>{$t('mint.record.pair')}</span>
          <span className={styles.lp}>{$t('mint.record.lp')}</span>
          <span className={styles.per}>{$t('mint.record.per')}</span>
        </div>
        {record.items!.map(item =>
          <div className={styles.cardItem}>
            <span className={styles.pair}>{record.status === 3 ? '—' : item.symbol}</span>
            <span className={styles.lp}>{record.status === 3 ? '—' : formatNumber(item.amount!)}</span>
            <span className={styles.per}>{record.status === 3 ? '—' : formatPer(item.profit!)}</span>
          </div>
        )}
        <div className={styles.cardTime}>
          <div className={styles.time}>{record.date}</div>
          <div
            className={styles[`status${record.status}`]}
            onClick={async () => {
              if (record.status !== 1) return
              try {
                const res = await ApiPostMintByID(record.record_id!)
                if (res === 'success') {
                  setShowMask(true)
                  recordList[idx].status = 2
                  setRecordList([...recordList])
                } else {
                  ToastFailed($t('mint.record.wait'))
                  initPage()
                }
              } catch (e) {
                ToastFailed($t('mint.record.wait'))
                initPage()
              }
            }}>{$t('mint.record.' + record.status)}</div>
        </div>
      </div>)}
    </div>
    {showMask && <div className={styles.mask} onClick={() => setShowMask(false)}>
      <div className={styles.mask_content} onClick={(e) => e.stopPropagation()}>
        <div className={styles.mask_main}>{$t('mint.record.tips')}</div>
        <div className={styles.mask_btn}>
          <div className={styles.btn_item_fx_1} onClick={() => setShowMask(false)}>知道了</div>
        </div>
      </div>
    </div>}
    {!isLoaded && <FullLoading mask />}
  </div>
}

function formatNumber(n: number | string): string {
  n = Number(n)
  if (n < 0.01) return "<0.01"
  return String(Number(n.toFixed(2)))
}

function formatPer(n: number | string): string {
  n = Number(Number(n).toFixed(2)) * 100
  return n + '%'
}

function formatRecord(record: IMintRecord[]): IMintRecord[] {
  const res: IMintRecord[] = []
  let resMap: { [key: string]: IMintRecord[] } = {}
  record.forEach(({ date, record_id, symbol, amount, profit, status }) => {
    if (!resMap[record_id!]) resMap[record_id!] = []
    resMap[record_id!].push({ date, record_id, symbol, amount, profit, status })
  })
  Object.keys(resMap).map(_record_id => {
    const { date, record_id, status } = resMap[_record_id][0]
    let dailyRecord: IMintRecord = { date, record_id, status, items: [] }
    resMap[_record_id].forEach(({ symbol, amount, profit }) => {
      dailyRecord.items!.push({ symbol, amount, profit })
    })
    res.push(dailyRecord)
  })
  return res
}
