import React from 'react'
import { useState, useEffect } from 'react'
import styles from './member.less'
import { Modal } from 'antd-mobile'
import { BackHeader } from '@/components/BackHeader'
import { JoinModal } from '@/components/PopupModal/join'
import { Button, ToastSuccess } from '@/components/Sub'
import { get$t } from '@/locales/tools'
import { useIntl } from 'umi'
import { $get } from '@/stores/localStorage'
import { ApiGetMe, IUser } from '@/apis/user'
import moment from 'moment'
import { getAuthUrl, payUrl } from '@/apis/http'
import { ApiGetGroupVipAmount, IGroupInfo1, IVipAmount } from '@/apis/group'
import { changeTheme, delay, getURLParams, getUUID } from '@/assets/ts/tools'
import { Loading } from '@/components/Loading'
import { checkPaid } from './reward'

export default function Page() {
  const $t = get$t(useIntl())
  const [show, setShow] = useState(false)
  const [showNext, setShowNext] = useState(false)
  const [showGiveUp, setShowGiveUp] = useState(false)
  const [showFailed, setShowFailed] = useState(false)
  const [u, setUser] = useState<IUser>()
  const [selectList, setSelectList] = useState([2, 5])
  const [selectStatus, setSelectStatus] = useState(2)
  const [vipAmount, setVipAmount] = useState<IVipAmount>()
  const [payLoading, setPayLoading] = useState(false)
  const group: IGroupInfo1 = $get('group')

  useEffect(() => {
    changeTheme('#4A4A4D')
    ApiGetMe().then(u => {
      setUser(u)
      if (u.status === 2) {
        setSelectList([5])
        setSelectStatus(5)
      }
      const { state } = getURLParams()
      if (state && (u.status! < Number(state))) {
        setShowFailed(true)
      }


    })
    ApiGetGroupVipAmount().then(setVipAmount)
    return () => {
      changeTheme('#fff')
    }
  }, [])

  const clickPay = async () => {
    const trace = getUUID()
    const amount = getPayAmount(selectStatus, vipAmount)
    const asset = group.asset_id
    const recipient = group.client_id
    location.href = payUrl({
      trace, amount, asset,
      recipient,
      memo: JSON.stringify({ type: 'vip' })
    })
    setShow(false)
    setShowNext(false)
    setPayLoading(true)
    const t = await checkPaid(amount!, asset, recipient, trace, $t)
    if (t === 'paid') {
      while (true) {
        const u = await ApiGetMe()
        if (u.pay_status === selectStatus) {
          setPayLoading(false)
          setUser(u)
          if (u.status === 2) {
            setSelectList([5])
            setSelectStatus(5)
          }
          ToastSuccess($t('success.operator'))
          return
        }
        await delay(200)
      }
    }
  }

  return (
    <>
      <div className={styles.container}>
        <BackHeader name={$t('member.center')} isWhite />
        <div className={styles.content}>
          <MemberCard user={u} $t={$t} />
        </div>
        <div className={styles.foot}>
          {u && u.status && u.status < 5 && <Button onClick={() => {
            if (!group.asset_id) location.href = getAuthUrl('/member', true, "2")
            else setShow(true)
          }}>{$t('member.upgrade')}</Button>}
          {
            (u && u.pay_expired_at && new Date(u.pay_expired_at) > new Date()) &&
            <div className={styles.time}>{$t('member.expire', { date: moment(u.pay_expired_at).format('YYYY-MM-DD') })}</div>
          }
        </div>
        <Modal
          visible={show}
          popup
          onClose={() => {
            setShow(false)
            setTimeout(() => setShowNext(false))
          }}
          animationType="slide-up"
        >
          <div className={styles.modal_content}>
            <div className={styles.modal_close}>
              <img src={require('@/assets/img//svg/modalClose.svg')} onClick={() => {
                setShow(false)
                setTimeout(() => setShowNext(false))
              }} />
            </div>
            {showNext ?
              <>
                <MemberCard showMode={selectStatus} $t={$t} />
                <div className={styles.tip} dangerouslySetInnerHTML={{ __html: $t('member.authTips') }} />
                <div className={styles.foot}>
                  <Button onClick={() => {
                    location.href = getAuthUrl(`/member`, true, String(selectStatus))
                  }}>{$t('member.forFree')}</Button>
                  <div className={styles.pay} onClick={() => clickPay()}>{
                    $t('member.forPay', {
                      amount: getPayAmount(selectStatus, vipAmount),
                      symbol: group.symbol
                    })}</div>
                </div>
              </>
              :
              <>
                {selectList.map(item => (<div
                  key={item}
                  className={`${styles.desc} ${selectStatus === item && styles.active}`}
                  onClick={() => setSelectStatus(item)}
                >
                  <div className={styles.title}>{$t(`member.level${item}`)}</div>
                  <div className={styles.intro}>{$t(`member.level${item}Sub`)}</div>
                  <div className={styles.price}>{$t(`member.levelPay`, {
                    payAmount: getPayAmount(item, vipAmount),
                    amount: getAuthAmount(item, group),
                    symbol: group?.symbol
                  })}</div>
                </div>))}
                <div className={styles.foot}>
                  <Button onClick={() => setShowNext(true)}>{$t('action.continue')}</Button>
                </div>
              </>
            }
          </div>
        </Modal>
        <Modal
          visible={showGiveUp}
          popup
          onClose={() => setShowGiveUp(false)}
          animationType="slide-up">
          <JoinModal modalProp={{
            title: $t('member.cancel'),
            desc: $t('member.cancelDesc'),
            descStyle: styles.red,
            icon: "shenqingxuzhi",
            button: $t('member.cancel'),
            buttonAction: () => location.href = getAuthUrl('/member'),
            tips: $t('action.cancel'),
            tipsAction: () => setShowGiveUp(false),
          }} />
        </Modal>
        <Modal
          visible={showFailed}
          popup
          onClose={() => setShowFailed(false)}
          animationType="slide-up">
          <JoinModal modalProp={{
            title: $t('member.failed'),
            desc: $t('member.failedDesc'),
            descStyle: styles.red,
            icon: "a-huiyuankaitongshibai1",
            button: $t('action.know'),
            buttonAction: () => setShowFailed(false),
          }} />
        </Modal>
      </div>
      {payLoading && <Loading content={$t('member.checkPaid')} cancel={() => setPayLoading(false)} />}
    </>
  )
}

const getPayAmount = (selectStatus = 2, vipAmount: IVipAmount | undefined) =>
  selectStatus === 2 ? vipAmount?.fresh_amount : vipAmount?.large_amount

const getAuthAmount = (selectStatus = 2, group: IGroupInfo1) =>
  selectStatus === 2 ? group?.amount : group?.large_amount


interface IMemberPros {
  user?: IUser
  showMode?: number
  $t?: any
}

const MemberCard = (props: IMemberPros) => {
  const { user, $t, showMode } = props
  if (!user && !showMode) return <div />
  let sub = '', _status = 2
  if (user) {
    let { pay_expired_at, status } = user
    if ([3, 8, 9].includes(status!)) status = 5
    if (new Date(pay_expired_at!) > new Date()) sub = 'Pay'
    else if (status !== 1) sub = 'Auth'
    _status = status!
  } else if (showMode) {
    _status = showMode
  }
  const data: { label: string, isCheck: boolean }[] = $t(`member.level${_status}Desc`).split(',').map((item: string) => {
    const [isCheck, label] = item.split('-')
    return {
      label,
      isCheck: isCheck === '1'
    }
  })

  return <div className={`${styles.memberCard} ${showMode && styles.memberCardShort}`}>
    <div className={styles.cardHead}>
      <div>{$t(`member.level${_status}${sub}`)}</div>
      {_status !== 1 && <img className={styles.cardHeadIcon} src={require(`@/assets/img/member-vip-${_status}.png`)} />}
      {/* {sub === 'Auth' && <div className={styles.dots} onClick={() => setShowGiveUp!(true)}>...</div>} */}
    </div>
    {data.map((item, index) => (
      <div key={index} className={styles.func}>
        <i className={`iconfont ${item.isCheck ? 'iconcheck' : 'iconguanbi2'} ${styles.icon} ${item.isCheck ? styles.iconHas : styles.iconNotHas}`} />
        <div>{item.label}</div>
      </div>
    ))}
  </div>
}