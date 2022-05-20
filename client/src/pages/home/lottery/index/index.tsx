import { BackHeader } from '@/components/BackHeader';
import { get$t } from '@/locales/tools';
import React, { useEffect, useRef, useState } from 'react';
import { useIntl } from 'react-intl';
import { ApiGetClaimPageData, ApiPostClaim, ApiPostLotteryExchange, ApiGetLotteryReward, ClaimData, LotteryRecord, ApiPostVoucher } from '@/apis/claim';
import { Modal, Carousel } from 'antd-mobile';
import { LotteryBox } from './LotteryBox';
import { Button, ToastFailed, ToastSuccess } from '@/components/Sub';
import { history } from 'umi';
import { changeTheme } from '@/assets/ts/tools';
import { Icon } from '@/components/Icon';
import { FullLoading } from '@/components/Loading';

import styles from './index.less';
import { JoinModal } from '@/components/PopupModal/join';
import { Energy } from './Energy';
import { IGroup } from '@/apis/group';

const BG = {
  idle: 'https://super-group-cdn.mixinbots.com/lottery/bg.mp3',
  runing: 'https://super-group-cdn.mixinbots.com/lottery/ing.mp3',
  success: 'https://super-group-cdn.mixinbots.com/lottery/success.mp3',
};

type ModalType = 'preview' | 'receive';

export default function LotteryPage() {
  const $t = get$t(useIntl());
  const [modalType, setModalType] = useState<ModalType>();
  const [reward, setReward] = useState<LotteryRecord>({} as LotteryRecord);
  const [isReceiving, setIsReceiving] = useState(false);
  const [hasMusic, setHasMusic] = useState(false);
  const [hasRunMusic, setHasRunMusic] = useState(false);
  const [hasSuccessMusic, setHasSuccessMusic] = useState(false);
  const [isLoaded, setLoaded] = useState(false);

  const [showDoubleModal, setShowDoubleModal] = useState(false);
  const [doubleGroup, setDoubleGroup] = useState<IGroup | null>(null);

  const [showVoucherModal, setShowVoucherModal] = useState(false);
  const [voucher, setVoucher] = useState('');

  const [btnLoading, setBtnLoading] = useState(false);
  const [claim, setClaim] = useState<ClaimData>({
    count: 0,
    invite_count: 0,
    is_claim: false,
    last_lottery: [],
    lottery_list: [],
    double_claim_list: [],
    power: {
      lottery_times: 0,
      balance: '0',
    },
  });
  const initPageData = async () => {
    const [claim] = await Promise.all([ApiGetClaimPageData()]);
    setClaim(claim);

    const lotteryList: { [asset_id: string]: boolean } = {};
    claim.lottery_list.forEach((lottery) => (lotteryList[lottery.asset_id] = true));

    if (claim.receiving) {
      setReward(claim.receiving);
      setShowDoubleModal(true);
      setModalType('receive');
    }
  };

  useEffect(() => {
    changeTheme('#2b120b');
    document.body.classList.add(styles.bg);
    initPageData();

    return () => {
      changeTheme('#fff');
      document.body.classList.remove(styles.bg);
    };
  }, []);

  const handleReceiveClick = async () => {
    if (modalType === 'preview') return setShowDoubleModal(false);
    if (!reward?.trace_id) return;
    setIsReceiving(true);
    const res = await ApiGetLotteryReward(reward.trace_id);
    if (res === 'success') {
      ToastSuccess($t('claim.receiveSuccess'));
      setShowDoubleModal(false);
      setIsReceiving(false);
    } else if (res && res.client_id) {
      setIsReceiving(false);
      setShowDoubleModal(false);
      setTimeout(() => {
        setShowDoubleModal(true);
        setModalType('preview');
        setReward({ ...reward, client_id: res.client_id });
      }, 200);
    }
  };

  const handleClickVoucher = async () => {
    setBtnLoading(true);
    const { status } = (await ApiPostVoucher(voucher)) || {};
    if (status === 3) {
      ToastSuccess($t('claim.voucher.status.' + status));
      setShowVoucherModal(false);
      initPageData();
      setVoucher('');
    } else {
      ToastFailed($t('claim.voucher.status.' + status));
    }
    setBtnLoading(false);
  };

  return (
    <div className={styles.container}>
      <BackHeader
        name={$t('claim.title')}
        isWhite
        action={
          <>
            <Icon i={hasMusic ? 'ic_music_open' : 'ic_music_close'} className={styles.headerIcon} onClick={() => setHasMusic(!hasMusic)} />
            <Icon className={styles.headerIcon} i="ic_file_text" onClick={() => history.push('/lottery/records')} />
          </>
        }
      />
      <div className={styles.broadcast}>
        <div className={styles.content}>
          {!!claim?.last_lottery.length && (
            <Carousel vertical autoplay dots={false} infinite className={styles.carousel}>
              {claim?.last_lottery.map((item, idx) => (
                <div key={item.trace_id || idx} className={styles.item}>
                  {item.full_name}&nbsp;
                  {$t('claim.drew')}
                  &nbsp;{item.amount}&nbsp;
                  {item.symbol}
                  {Number(item.price_usd) > 0 && $t('claim.worth', { value: item.price_usd, prefix: ', ' })}
                </div>
              ))}
            </Carousel>
          )}
        </div>
      </div>
      {claim?.lottery_list.length && (
        <LotteryBox
          data={claim?.lottery_list}
          ticketCount={claim?.power.lottery_times}
          onPrizeClick={(lottery: LotteryRecord) => {
            setShowDoubleModal(true);
            setModalType('preview');
            setReward(lottery);
          }}
          onImgLoad={() => setLoaded(true)}
          onStart={() => setHasRunMusic(true)}
          onEnd={async () => {
            await initPageData();
            setHasRunMusic(false);
            setHasSuccessMusic(true);
            setTimeout(() => {
              setHasSuccessMusic(false);
            }, 2000);
          }}
        />
      )}
      <Energy
        claim={claim}
        onCheckinClick={async () => {
          const res = await ApiPostClaim();
          res === 'success' && initPageData();
        }}
        onExchangeClick={async () => {
          const res = await ApiPostLotteryExchange();
          if (res === 'success') {
            ToastSuccess($t('claim.energy.success'));
            initPageData();
          }
        }}
        onModalOpen={(group: IGroup) => {
          setShowDoubleModal(true);
          setDoubleGroup(group);
        }}
        onVoucherClick={() => {
          setShowVoucherModal(true);
        }}
      />
      <Modal visible={showVoucherModal} animationType="slide-up" popup onClose={() => setShowVoucherModal(false)}>
        <div className={styles.voucher}>
          <div className={styles.voucherHeader}>
            <h3>{$t('claim.voucher.title')}</h3>
            <i onClick={() => setShowVoucherModal(false)} className="iconfont iconguanbi" />
          </div>
          <input type="text" placeholder={$t('claim.voucher.placeHolder')} value={voucher} onChange={(e) => setVoucher(e.target.value.toUpperCase())} />
          <Button onClick={handleClickVoucher} loading={btnLoading} type="submit" className={styles.btn} disabled={voucher.length !== 6}>
            {$t('claim.voucher.btn')}
          </Button>
        </div>
      </Modal>
      <Modal visible={showDoubleModal} animationType="slide-up" popup onClose={() => setShowDoubleModal(false)}>
        {doubleGroup ? (
          <JoinModal
            modalProp={{
              title: doubleGroup.name,
              titleDesc: 'Mixin ID: ' + doubleGroup.identity_number,
              desc: doubleGroup.description,
              icon_url: doubleGroup.icon_url,
              button: $t('claim.open'),
              buttonAction: () => (location.href = `mixin://apps/${doubleGroup.client_id}?action=open `),
              isAirdrop: true,
            }}
          />
        ) : (
          <JoinModal modalProp={getModalProps(reward, modalType, isReceiving, $t, setShowDoubleModal, handleReceiveClick)} />
        )}
      </Modal>
      {!isLoaded && <FullLoading mask />}
      {hasMusic && <audio autoPlay src={BG.idle} loop />}
      {hasMusic && hasRunMusic && <audio autoPlay src={BG.runing} loop />}
      {hasMusic && hasSuccessMusic && <audio autoPlay src={BG.success} />}
    </div>
  );
}

function getModalProps(reward: LotteryRecord, modalType: ModalType | undefined, isLoading: boolean, $t: any, setShowModal: any, receivedAction: any) {
  const title = getTitle(reward);
  const priceUsd = Number(reward.amount) * Number(reward.price_usd);
  let titleDesc = '';
  if (priceUsd >= 1e-8) titleDesc = '≈ $' + Number(priceUsd.toFixed(8));
  const isPreview = modalType === 'preview';
  return {
    title,
    titleDesc,
    desc: reward!.description,
    isAirdrop: true,
    icon_url: reward!.icon_url,
    button: $t(`claim.${isPreview ? 'ok' : 'receive'}`),
    buttonAction: () => (isPreview ? setShowModal(false) : receivedAction()),
    buttonStyle: isPreview ? '' : 'submit',
    tips: isPreview ? $t('claim.join') : '',
    tipsStyle: 'blank',
    loading: isLoading,
    tipsAction: () => (location.href = `mixin://apps/${reward!.client_id}?action=open`),
  };
}

function getTitle(reward: LotteryRecord) {
  if (reward.symbol === 'BTC') return (Number(reward.amount) * 1e8).toFixed() + ' SAT';
  return reward.amount + ' ' + reward.symbol;
}
