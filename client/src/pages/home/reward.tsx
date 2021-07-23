import { useIntl } from '@/.umi/plugin-locale/localeExports'
import { ApiCheckIsPaid, ApiGetMyAssets, IAsset } from '@/apis/asset'
import { payUrl } from '@/apis/http'
import { ApiGetAdminAndGuest, IUser } from '@/apis/user'
import { delay, getURLParams, getUUID } from '@/assets/ts/tools'
import { BackHeader } from '@/components/BackHeader'
import { FullLoading } from '@/components/Loading'
import { PopAdminAndGuestModal, PopCoinModal } from '@/components/PopupModal/coinSelect'
import { Button, ToastFailed, ToastSuccess } from '@/components/Sub'
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
  const groupClientID = $get('group').client_id

  useEffect(() => {
    initPage()
  }, [])

  const initPage = async () => {
    setLoading(true)
    const [assetList, rewardList] = await Promise.all([ApiGetMyAssets(), ApiGetAdminAndGuest()])
    setActiveCoin(assetList[0])
    const { uid } = getURLParams() || {}
    if (uid) {
      let activeUser = rewardList.find(u => u.identity_number === uid)
      if (activeUser) setActiveUser(activeUser)
    }
    setLoading(false)
  }

  return (
    <div className={styles.container}>
      <BackHeader name={$t('reward.title')} />
      {activeCoin && <div className={`${styles.coin} ${styles.item}`} onClick={() => setCoinModal(true)}>
        <img src={activeCoin.icon_url} alt="" />
        <p>{activeCoin.name}</p>
        <span>{activeCoin.balance} {activeCoin.symbol}</span>
      </div>}

      <div className={`${styles.user} ${styles.item}`} onClick={() => setUserModal(true)}>
        <p className={!activeUser ? styles.noUser : ""}>{activeUser ? `${activeUser.full_name} (${activeUser.identity_number})` : $t('reward.who')}</p>
        <i className={`iconfont iconic_down ${styles.icon}`} />
      </div>

      <div className={`${styles.amount} ${styles.item}`}>
        <input type="number" placeholder={$t('reward.amount')} value={amount} onChange={e => setAmount(e.target.value)} />
        <p>{(Number(activeCoin?.price_usd) * Number(amount)).toFixed(2)} USD</p>
      </div>

      <Button className={styles.button} onClick={async () => {
        if (isLoading) return
        if (!activeUser) return ToastFailed($t('reward.who'))
        const trace = getUUID()
        location.href = payUrl({
          trace,
          asset: activeCoin!.asset_id,
          recipient: groupClientID,
          amount,
          memo: encodeURIComponent(JSON.stringify({ reward: activeUser!.user_id }))
        })
        setLoading(true)
        const res = await checkPaid(amount, activeCoin!.asset_id!, activeUser.user_id!, trace, $t)
        if (res === 'paid') {
          await delay(2000)
          ToastSuccess($t('reward.success'))
          GlobalData.MyAssetList = undefined
          setAmount("")
          initPage()
        }
      }}>
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
      {isLoading && <FullLoading mask opacity />}
    </div>
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
