import React from 'react';
import { useState, useEffect } from 'react';
import styles from './member.less';
import { Modal } from 'antd-mobile';
import { BackHeader } from '@/components/BackHeader';
import { JoinModal } from '@/components/PopupModal/join';
import { Button, ToastSuccess } from '@/components/Sub';
import { get$t } from '@/locales/tools';
import { useIntl } from 'umi';
import { $get } from '@/stores/localStorage';
import { ApiGetMe, IUser } from '@/apis/user';
import moment from 'moment';
import { getAuthUrl, payUrl } from '@/apis/http';
import { ApiGetGroupVipAmount, IGroupInfo, IVipAmount } from '@/apis/group';
import { changeTheme, delay, getURLParams, getUUID } from '@/assets/ts/tools';
import { FullLoading, Loading } from '@/components/Loading';
import { checkPaid } from './reward';
import BigNumber from 'bignumber.js';
import { Icon } from '@/components/Icon';

const level: any = {
  2: { category: 3, min: 5, max: 10 },
  5: { category: 9, min: 10, max: 20 },
};

export default function Page() {
  const $t = get$t(useIntl());
  const [show, setShow] = useState(false);
  const [showNext, setShowNext] = useState(false);
  const [showGiveUp, setShowGiveUp] = useState(false);
  const [showFailed, setShowFailed] = useState(false);
  const [u, setUser] = useState<IUser>();
  const [selectList, setSelectList] = useState([2, 5]);
  const [selectStatus, setSelectStatus] = useState(0);
  const [vipAmount, setVipAmount] = useState<IVipAmount>();
  const [payLoading, setPayLoading] = useState(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const group: IGroupInfo = $get('group');
  if (!group.asset_id) {
    group.asset_id = '4d8c508b-91c5-375b-92b0-ee702ed2dac5';
    group.symbol = 'USDT';
  }

  useEffect(() => {
    changeTheme('#4A4A4D');
    ApiGetMe().then((u) => {
      setUser(u);
      if (u.status === 2) {
        setSelectList([5]);
        setSelectStatus(5);
      }
      const { state } = getURLParams();
      if (state && u.status! < Number(state)) {
        setShowFailed(true);
      }
      setIsLoaded(true);
    });
    ApiGetGroupVipAmount().then(setVipAmount);
    return () => {
      changeTheme('#fff');
    };
  }, []);

  const clickPay = async () => {
    const trace = getUUID();
    const amount = getPayAmount(selectStatus, vipAmount);
    const asset = group.asset_id;
    const recipient = group.client_id;
    location.href = payUrl({
      trace,
      amount,
      asset,
      recipient,
      memo: JSON.stringify({ type: 'vip' }),
    });
    setShow(false);
    setShowNext(false);
    setPayLoading(true);
    const t = await checkPaid(amount!, asset, recipient, trace, $t);
    if (t === 'paid') {
      while (true) {
        const u = await ApiGetMe();
        if (u.pay_status === selectStatus) {
          setPayLoading(false);
          setUser(u);
          if (u.status === 2) {
            setSelectList([5]);
            setSelectStatus(5);
          }
          ToastSuccess($t('success.operator'));
          return;
        }
        await delay(200);
      }
    }
  };
  return (
    <>
      <div className={styles.container}>
        <BackHeader name={$t('member.center')} isWhite backHome />
        <div className={styles.content}>{vipAmount && <MemberCard user={u} $t={$t} vipAmount={vipAmount} />}</div>
        {u && !isPay(u) && <div className={`${styles.tip1} ${styles.tip}`} dangerouslySetInnerHTML={{ __html: $t('member.authTips') }} />}
        <div className={styles.foot}>
          {(!u?.status || u?.status <= 2) && (
            <Button
              onClick={() => {
                if (isPay(u!)) {
                  setShowNext(true);
                  setSelectStatus(5);
                }
                setShow(true);
              }}
            >
              {$t('member.upgrade')}
            </Button>
          )}
          {u && u.pay_expired_at && isPay(u) && (
            <div className={styles.time}>
              {$t('member.expire', {
                date: moment(u.pay_expired_at).format('YYYY-MM-DD'),
              })}
            </div>
          )}
        </div>
        <Modal
          visible={show}
          animationType="slide-up"
          popup
          onClose={() => {
            setShow(false);
            setTimeout(() => setShowNext(false));
          }}
        >
          <div className={styles.modal_content}>
            <div className={styles.modal_close}>
              <img
                src={require('@/assets/img/svg/modalClose.svg')}
                onClick={() => {
                  setShow(false);
                  setTimeout(() => setShowNext(false));
                }}
              />
            </div>
            {showNext ? (
              <>
                {/* 确认验证模式 */}
                {vipAmount && <MemberCard showMode={selectStatus} $t={$t} vipAmount={vipAmount} />}
                {selectStatus === 0 && <div className={styles.tip} dangerouslySetInnerHTML={{ __html: $t('member.authTips') }} />}
                <div className={styles.foot}>
                  {selectStatus === 0 ? (
                    <Button onClick={() => (location.href = getAuthUrl({ returnTo: `/member`, hasAssets: true, state: '2' }))}>{$t('member.forFree')}</Button>
                  ) : (
                    <Button className={styles.pay} onClick={() => clickPay()}>
                      {$t('member.forPay', {
                        amount: getPayAmount(selectStatus, vipAmount),
                        symbol: group.symbol,
                      })}
                    </Button>
                  )}
                </div>
              </>
            ) : (
              <>
                {/* 选择验证模式... */}
                <div className={`${styles.desc} ${styles.desc0} ${selectStatus === 0 && styles.active}`} onClick={() => setSelectStatus(0)}>
                  <div className={styles.title}>{$t(`member.level${0}`)}</div>
                  <div className={styles.intro}>
                    {$t(`member.level${0}Sub`, {
                      lamount: formatNumber(group?.amount),
                      hamount: formatNumber(group?.large_amount),
                      symbol: group?.symbol,
                    })}
                  </div>
                </div>
                {group.asset_id &&
                  selectList.map((item) => (
                    <div key={item} className={`${styles.desc} ${styles[`desc${item}`]} ${selectStatus === item && styles.active}`} onClick={() => setSelectStatus(item)}>
                      <div className={styles.title}>{$t(`member.level${item}Pay`)}</div>
                      {/* <div className={styles.intro}>{$t(`member.level${item}Sub`)}</div> */}
                      <div className={styles.price}>
                        {$t(`member.levelPay`, {
                          amount: formatNumber(getPayAmount(item, vipAmount)),
                          symbol: group?.symbol,
                          category: level[item].category,
                          min: formatNumber(level[item].min),
                          max: formatNumber(level[item].max),
                          level: $t(`member.level${item}Pay`),
                        })}
                      </div>
                    </div>
                  ))}
                <div className={styles.foot}>
                  <Button onClick={() => setShowNext(true)}>{$t('action.continue')}</Button>
                </div>
              </>
            )}
          </div>
        </Modal>
        <Modal visible={showGiveUp} popup onClose={() => setShowGiveUp(false)} animationType="slide-up">
          <JoinModal
            modalProp={{
              title: $t('member.cancel'),
              desc: $t('member.cancelDesc'),
              descStyle: styles.red,
              icon: 'shenqingxuzhi',
              button: $t('member.cancel'),
              buttonAction: () => (location.href = getAuthUrl({ returnTo: '/member' })),
              tips: $t('action.cancel'),
              tipsAction: () => setShowGiveUp(false),
            }}
          />
        </Modal>
        <Modal visible={showFailed} popup onClose={() => setShowFailed(false)} animationType="slide-up">
          <JoinModal
            modalProp={{
              title: $t('member.failed'),
              desc: $t('member.failedDesc'),
              descStyle: styles.red,
              icon: 'a-huiyuankaitongshibai1',
              button: $t('action.know'),
              buttonAction: () => setShowFailed(false),
            }}
          />
        </Modal>
      </div>
      {payLoading && <Loading content={$t('member.checkPaid')} cancel={() => setPayLoading(false)} />}
      {!isLoaded && <FullLoading mask />}
    </>
  );
}

const getPayAmount = (selectStatus = 2, vipAmount: IVipAmount | undefined) => (selectStatus === 2 ? vipAmount?.level.fresh_amount : vipAmount?.level.large_amount);

const formatNumber = (num?: number | string) => {
  if (!num) return 0;
  return new BigNumber(num).toFormat();
};

const isPay = (u: IUser) => u.pay_expired_at && new Date(u.pay_expired_at) > new Date();

interface IMemberPros {
  user?: IUser;
  showMode?: number;
  $t?: any;
  vipAmount: IVipAmount;
}

const MemberCard = (props: IMemberPros) => {
  const { user, $t, showMode, vipAmount } = props;
  let sub = '',
    _status = 2;
  if (user && !user.status) {
    _status = 1;
  } else if (user) {
    let { pay_expired_at, status } = user;
    if ([3, 8, 9].includes(status!)) status = 5;
    if (new Date(pay_expired_at!) > new Date()) sub = 'Pay';
    else if (status !== 1) sub = 'Auth';
    _status = status!;
  } else if (typeof showMode === 'number') {
    _status = showMode;
    if (_status !== 0) sub = 'Pay';
  }
  let data: { label: string; allow: boolean }[] = [];
  if (_status === 0) {
    data = $t(`member.level${_status}Desc`)
      .split(',')
      .map((item: string) => {
        const [allow, label] = item.split('-');
        return {
          label,
          allow: allow === '1',
        };
      });
  } else if (_status === 1 || _status === 2 || _status === 5) {
    const { auth } = vipAmount;
    const { limit: max, plain_text } = auth[_status];
    const min = Math.floor(max / 2);
    const count = Object.values(auth[_status]).filter((v) => v === true).length;
    const desc: string = $t(`member.authDesc`, { min, max, count });
    data = desc.split('\n').map((v) => ({ label: v, allow: true }));
    if (!plain_text) {
      data.pop();
      data.push({ allow: false, label: $t('member.authReject') });
    }
  }
  return (
    <div className={`${styles.memberCard} ${showMode && styles.memberCardShort}`}>
      <div className={styles.cardHead}>
        <div>{$t(`member.level${_status}${sub}`)}</div>
        {_status !== 1 && <img className={styles.cardHeadIcon} src={require(`@/assets/img/member-vip-${_status}.png`)} />}
        {/* {sub === 'Auth' && <div className={styles.dots} onClick={() => setShowGiveUp!(true)}>...</div>} */}
      </div>
      {data.map((item, index) => (
        <div key={index} className={styles.func}>
          <Icon i={item.allow ? 'check' : 'guanbi2'} className={`${item.allow ? styles.iconHas : styles.iconNotHas} ${styles.icon}`} />
          <div>{item.label}</div>
        </div>
      ))}
    </div>
  );
};
