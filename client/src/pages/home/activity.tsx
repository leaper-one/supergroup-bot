import React from 'react'
import { BackHeader } from "@/components/BackHeader"
import { get$t } from "@/locales/tools"
import { useIntl } from "@@/plugin-locale/localeExports"
import { $get } from "@/stores/localStorage"
import { IActivity } from "@/apis/group"
import styles from './activity.less'
import { ApiAirdropReceived, ApiGetAirdrop } from '@/apis/airdrop'
import { useState } from 'react'
import { useEffect } from 'react'
import { ToastFailed, ToastSuccess } from '@/components/Sub'
import { FullLoading } from '@/components/Loading'

export default function Page() {
  const $t = get$t(useIntl())
  const [activity, setActivity] = useState<IActivity[]>([])
  const [loaded, setLoaded] = useState(false)
  useEffect(() => {
    initPage()
  }, [])

  const initPage = async () => {
    const now = new Date()
    let a: IActivity[] = $get("group")?.activity || []
    const airdropIdx = a.findIndex(item => item.action.startsWith("airdrop"))
    a = a.map(item => ({
      ...item,
      isExpire: now > new Date(item.expire_at)
    }))
    if (airdropIdx === -1) setActivity(a)
    else await checkAirdrop(a, airdropIdx, setActivity)
    setLoaded(true)
  }

  return (
    <div className={`${styles.container}`}>
      <BackHeader name={$t('home.activity')} />

      {loaded ? <div className={styles.content}>
        {
          activity.length > 0 ?
            activity.map(item =>
              <img key={item.activity_index}
                src={(item.isExpire ? item.expire_img_url : item.img_url) + '?t=1'}
                className={styles.card}
                onClick={() => {
                  if (item.isExpire) return
                  if (item.action.startsWith('http')) return location.href = item.action
                  if (item.action.startsWith('airdrop')) return handleAirdrop(item.action, $t, initPage)
                }}
                alt=""
              />) :
            <div className={styles.noData}>
              <img src={require('@/assets/img/no-events.png')} alt="" />
              <p>{$t('home.noActive')}</p>
            </div>
        }
      </div>
        : <FullLoading />}
    </div>
  )
}

const checkAirdrop = async (activities: IActivity[], idx: number, setActivity: any) => {
  const [_, airdropID] = activities[idx].action.split(':')
  if (!airdropID || airdropID.length !== 36) return setActivity(activities)
  const airdrop = await ApiGetAirdrop(airdropID)
  if (airdrop.status >= 2) activities[idx].isExpire = true
  setActivity(activities)
}


const handleAirdrop = async (action: string, $t: any, reloadList: any) => {
  const [_, airdropID] = action.split(':')
  if (!airdropID || airdropID.length !== 36) return
  const airdrop = await ApiAirdropReceived(airdropID)
  if (airdrop === 2) {
    reloadList()
    return ToastSuccess($t('airdrop.success'))
  } else return ToastFailed($t('airdrop.failed'))
}