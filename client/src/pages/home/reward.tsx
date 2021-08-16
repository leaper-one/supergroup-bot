import { useIntl } from '@/.umi/plugin-locale/localeExports'
import { ApiCheckIsPaid, ApiGetAssetByID, ApiGetTop100, IAsset } from '@/apis/asset'
import { ApiGetGroupStatus } from '@/apis/group'
import { payUrl } from '@/apis/http'
import { ApiGetAdminAndGuest, IUser } from '@/apis/user'
import { delay, getURLParams, getUUID } from '@/assets/ts/tools'
import { BackHeader } from '@/components/BackHeader'
import { FullLoading } from '@/components/Loading'
import { PopAdminAndGuestModal, PopCoinModal } from '@/components/PopupModal/coinSelect'
import { Button, Confirm, ToastFailed, ToastSuccess } from '@/components/Sub'
import { get$t } from '@/locales/tools'
import { $get } from '@/stores/localStorage'
import { GlobalData } from '@/stores/store'
import React, { useState, useEffect } from 'react'
import styles from './reward.less'


export default function Page() {
  const $t = get$t(useIntl())
  const [isLoading, setLoading] = useState(false)
  const [coinModal, setCoinModal] = useState(false)
  const [userModal, setUserModal] = useState(false)
  const [activeCoin, setActiveCoin] = useState<IAsset>()
  const [activeUser, setActiveUser] = useState<IUser>()
  const [amount, setAmount] = useState("")
  const group = $get('group') || {}

  useEffect(() => {
    initPage()
  }, [])

  const initPage = async () => {
    setLoading(true)
    const [assetList, rewardList] = await Promise.all([ApiGetTop100(), ApiGetAdminAndGuest()])
    if (!group.asset_id) {
      group.asset_id = group.client_id === '47cdbc9e-e2b9-4d1f-b13e-42fec1d8853d' ?
        'c94ac88f-4671-3976-b60a-09064f1811e8' : 'c6d0c728-2624-429b-8e0d-d9d19b6592fa'
    }
    let asset = assetList.find(item => item.asset_id === group.asset_id)
    if (!asset) {
      asset = await ApiGetAssetByID(group.asset_id)
      asset.balance = "0"
    }
    setActiveCoin(asset)
    const { uid } = getURLParams() || {}
    if (uid) {
      let activeUser = rewardList.find(u => u.identity_number === uid)
      if (activeUser) setActiveUser(activeUser)
    }
    setLoading(false)
    // const valuesList = assetList.filter(item => Number(item.balance) * Number(item.price_usd) > 1)
    // if (valuesList.length > 0) setHasCoinList(true)
  }

  const clickReward = async () => {
    if (isLoading) return
    if (!activeUser) return ToastFailed($t('reward.who'))
    if (!activeCoin) return
    if (Number(activeCoin.price_usd) * Number(amount) < 1) return ToastFailed($t('reward.less'))
    setLoading(true)
    const status = await ApiGetGroupStatus()
    if (status === "2") {
      Confirm($t('action.tips'), $t('reward.isLiving'))
      return setLoading(false)
    }

    const trace = getUUID()
    location.href = payUrl({
      trace,
      asset: activeCoin.asset_id,
      recipient: group.client_id,
      amount,
      memo: encodeURIComponent(JSON.stringify({ type: "reward", reward: activeUser!.user_id }))
    })
    const res = await checkPaid(amount, activeCoin!.asset_id!, activeUser.user_id!, trace, $t)
    if (res === 'paid') {
      await delay(2000)
      ToastSuccess($t('reward.success'))
      GlobalData.MyAssetList = undefined
      setAmount("")
      initPage()
    }
  }

  return (
    <>
      <div className={styles.container}>
        <BackHeader name={$t('reward.title')} />
        {activeCoin && <div className={`${styles.coin} ${styles.item}`} onClick={() => setCoinModal(true)}>
          <img src={activeCoin.icon_url} alt="" />
          <div>
            <p>{activeCoin.name}</p>
            {/* <span>{activeCoin.balance} {activeCoin.symbol}</span> */}
          </div>
          <i className={`iconfont iconic_down ${styles.icon}`} />
        </div>}

        <div className={`${styles.user} ${styles.item}`} onClick={() => setUserModal(true)}>
          <p className={!activeUser ? styles.noUser : ""}>{activeUser ? `${activeUser.full_name} (${activeUser.identity_number})` : $t('reward.who')}</p>
          <i className={`iconfont iconic_down ${styles.icon}`} />
        </div>

        <div className={`${styles.amount} ${styles.item}`}>
          <input type="number" placeholder={$t('reward.amount')} value={amount} onChange={e => setAmount(e.target.value)} />
          <p>{(Number(activeCoin?.price_usd) * Number(amount)).toFixed(2)} USD</p>
        </div>

        <Button className={styles.button} onClick={() => clickReward()}>
          {$t('reward.title')}
        </Button>

        <PopCoinModal
          coinModal={coinModal}
          setCoinModal={setCoinModal}
          activeCoin={activeCoin}
          setActiveCoin={setActiveCoin}
        />

        <PopAdminAndGuestModal
          activeUser={activeUser}
          setActiveUser={setActiveUser}
          userModal={userModal}
          setUserModal={setUserModal}
          $t={$t}
        />
      </div>
      {isLoading && <FullLoading mask opacity />}
    </>
  )
}

export const checkPaid = async (amount: string, asset_id: string, counter_user_id: string, trace_id: string, $t: any): Promise<string> => new Promise(async resolve => {
  const check = async () => {
    const res = await ApiCheckIsPaid({
      amount,
      asset_id,
      counter_user_id,
      trace_id,
    })
    if (res.status === "paid") {
      resolve('paid')
    } else {
      await delay()
      check()
    }
  }
  check()
})
