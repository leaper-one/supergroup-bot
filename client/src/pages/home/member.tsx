import React from 'react'
import { useState } from 'react'
import styles from './member.less'
import { Modal } from 'antd-mobile'
import { BackHeader } from '@/components/BackHeader'
import { JoinModal } from '@/components/PopupModal/join'
import { Button } from '@/components/Sub'

const data = [
  { func: '接受全部聊天记录', has: true },
  { func: '参与抢红包', has: true },
  { func: '给管理员留言', has: true },
  { func: '发消息参与聊天', has: false },
]
export default function Page() {
  const [show, setShow] = useState(false)
  const [active, setActive] = useState(0)
  const [showNext, setShowNext] = useState(false)
  const [showGiveUp, setShowGiveUp] = useState(false)
  return (
    <div className={styles.container}>
      <BackHeader name="会员中心" isWhite />
      <div className={styles.content}>
        <MemberCard hasDot setShowGiveUp={setShowGiveUp} />
      </div>
      <div className={styles.foot}>
        {/* <div className={styles.btn} onClick={() => setShow(true)}>升级会员</div> */}
        <Button onClick={() => setShow(true)}>升级会员</Button>
        <div className={styles.time}>会员有效期截止到 2022-2-22，请到期后再续费。</div>
      </div>
      <Modal
        visible={show}
        popup
        onClose={() => {
          setShow(false)
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
              <MemberCard />
              <div className={styles.tip}>
                通过授权免费开通会定期访问并检查您的资产是否满足持仓要求，给更多说明请参见文
                <br />
                档：https://w3c.group/c/1589349785347025
              </div>
              <div className={styles.foot}>
                <Button onClick={() => setShowNext(true)}>授权免费获得</Button>
                <div className={styles.pay}>支付0.1MOB获得</div>
              </div>
            </>
            :
            <>
              <div className={`${styles.desc} ${active === 0 && styles.active}`} onClick={() => setActive(0)}>
                <div className={styles.title}>初级会员</div>
                <div className={styles.intro}>可发文字等 3 种类型消息，每分钟可发 5～10 条消息。</div>
                <div className={styles.price}>付费价值 1 美金的 MOB 获得 1 年会员或钱包余额始终大于或等于 10 MOB 授权免费领取永久有效。</div>
              </div>
              <div className={`${styles.desc} ${active === 1 && styles.active}`} onClick={() => setActive(1)}>
                <div className={styles.title}>资深会员</div>
                <div className={styles.intro}>可发文字等9种类型消息，每分钟可发10~20条消息。</div>
                <div className={styles.price}>付费价值5美金的MOB获得1年会员或钱包余额始终大于或等于200MOB授权免费领取永久有效。</div>
              </div>
              <div className={styles.foot}>
                <Button onClick={() => setShowNext(true)}>继续</Button>
              </div>
            </>
          }
        </div>
      </Modal>
      <Modal
        visible={showGiveUp}
        popup
        onClose={() => {
          setShowGiveUp(false)
        }}
        animationType="slide-up">
        <JoinModal modalProp={{
          title: "放弃会员资格",
          desc: "点下方放弃会员资格权按钮重新授权后你将失去会员资格，同时社群机器人将无法读取你的资产信息，你可以随时再次授权获得会员资格。",
          descStyle: styles.red,
          icon: "shenqingxuzhi",
          button: "放弃会员资格",
          tips: "取消"
        }} />
      </Modal>
    </div>
  )
}


interface IMemberPros {
  hasDot?: boolean
  setShowGiveUp?: (show: boolean) => void
}

const MemberCard = (props: IMemberPros) => <div className={styles.memberCard}>
  <div className={styles.card_head}>
    <div>未开通会员</div>
    {props.hasDot && <div className={styles.dots} onClick={() => props.setShowGiveUp!(true)}>...</div>}
  </div>
  {data.map((item, index) => (
    <div key={index} className={styles.func}>
      <i className={`iconfont iconcheck ${styles.icon} ${styles.iconHas}`} />
      <div>{item.func}</div>
    </div>
  ))}
  <img src={require('@/assets/img/member.png')} className={styles.img} />
</div>