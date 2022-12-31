import React, { useEffect, useState } from 'react';
import { history } from 'umi';
import styles from './index.less';
import { get$t } from '@/locales/tools';
import { useIntl } from '@@/plugin-locale/localeExports';
import { BackHeader } from '@/components/BackHeader';
import { ApiGetLiquidityByID, ApiPostLiquidityJoin, LiquidityResp } from '@/apis/mint';
import { getAuthUrl } from '@/apis/http';
import { formatTime } from '@/utils/time';
import FullLoading from '@/components/Loading';
import mintStyles from '../mint/index.less';
import { changeTheme, getURLParams } from '@/assets/ts/tools';

const msgMap = {
  success: '您已成功参与，参与其间请不要取消授权，否则将影响奖励瓜分。',
  limit: '您的 LP token 不足，请在今日 24:00（UTC+0）前准备足够的 LP token，否则将无法参与下期奖励瓜分。',
  auth: '请授权读取资产和交易记录，否则无法参与奖励瓜分。',
  miss: '您已错过本期奖励瓜分报名，请获取足够的 LP token 后于每月 1 日 24:00（UTC+0）前点击报名参与活动。',
};

export default function Page() {
  const $t = get$t(useIntl());
  const [resp, setResp] = useState<LiquidityResp>();
  const [showKnowModal, setShowKnowModal] = useState(false);
  const [msg, setMsg] = useState('');
  const [showContinueModal, setContinueModal] = useState(false);
  const { id } = getURLParams();

  useEffect(() => {
    changeTheme('#a650de');
    ApiGetLiquidityByID(id).then((res) => setResp(res));
    return () => {
      changeTheme('#fffff');
    };
  }, []);

  const clickJoin = () => {
    if (resp!.is_join) {
      setMsg(msgMap.success);
      setShowKnowModal(true);
      return;
    }
    const date = new Date().getUTCDate();
    if (date !== 1) {
      setMsg(msgMap.miss);
      setShowKnowModal(true);
      return;
    }
    ApiPostLiquidityJoin(id).then((res) => {
      if (['success', 'limit', 'miss'].includes(res)) {
        let t = msgMap[res as keyof typeof msgMap];
        setMsg(t);
      } else if (res.code === 403) {
        setMsg(msgMap.auth);
      }
      setShowKnowModal(true);
    });
  };

  return (
    <div className={styles.bg}>
      <BackHeader
        name=""
        isWhite
        className={styles.headerBg}
        action={<i className={`iconfont iconbangzhu ${styles.iconHelp}`} onClick={() => (location.href = `https://quill.im/articles/856e17e5-4600-43cd-98be-771185ad94a9`)} />}
      />
      {resp ? (
        <>
          <div className={styles.info}>
            <h3>{resp.info.title}</h3>

            <div className={styles.infoDesc}>
              <div className={styles.time}>
                {formatTime(resp.info.start_at)} - {formatTime(resp.info.end_at)}
              </div>
              <div className={styles.desc}>{resp.info.description}</div>
              <img className={styles.icon} src={require('@/assets/img/active/liquidity/img_lock.png')} alt="" />
              <div className={styles.infoDescBottom}>
                <h4 className={styles.amount}>{resp.yesterday_amount}</h4>
                <div className={styles.btn}>领取奖励</div>
                <p>累计（lptoken）</p>
                <p onClick={() => history.push(`/a/liquidity/records?id=${id}`)}>查看奖励记录 &gt;</p>
              </div>
            </div>

            <div className={styles.action}>
              <button onClick={() => setContinueModal(true)}>获得 LP Token</button>
              <button onClick={() => clickJoin()}>报名参加</button>
            </div>
          </div>

          <div className={styles.detail} style={{ marginTop: '-180px' }}>
            <h4>活动详情</h4>
            {resp.list.map((item, idx) => (
              <div className={styles.item} key={idx}>
                <div className={styles.itemTitle}>第{item.idx}期</div>
                <div className={styles.itemAmount}>
                  瓜分 <b>{item.amount}</b> {item.symbol}
                </div>
                <div className={styles.itemTime}>
                  {formatTime(item.start_at)} - {formatTime(item.end_at)}
                </div>
              </div>
            ))}
          </div>

          <div className={`${styles.detail}`}>
            <h4>参与方式</h4>
            <div className={styles.intro}>
              <p className={styles.introItem}>
                每月 <b>1 日 24:00（UTC+0）</b> 点前注入足够的 LP token 并在锁仓奖励页面点 <b>“报名参与”</b>授权活动页面读取 LP token 数量；超过时间注入将视为放弃本月参与机会。
              </p>
              <p className={styles.introItem}>
                每日机器人都会不定时监测 LP token 余额，若参与期间撤回流动性、或减少 LP token 持仓，其数量低于 <b>4360</b>，将无法参与本月奖励瓜分。
              </p>
              <p className={styles.introItem}>
                成功在 4swap 注入流动性后，请务必返回本页面，点击<b>“报名参与”</b>授权资产读取，否则无法记录您的 LP token 数量。
              </p>
              <p className={styles.introItem}>将根据参与活动的用户每日提供的 LP token 份额进行统计，每月瓜分一次奖励，即提供的流动性越多，瓜分奖励统计的份额越多。</p>
            </div>
          </div>

          <div className={styles.detail}>
            <h4>领取奖励</h4>

            <div className={styles.intro}>
              <p className={styles.introItem}>
                奖励将在<b>次月 1 日 02:00（UTC+0）</b> 点开放领取，需用户在活动页面手动领取。
              </p>
              <p className={styles.introItem}>
                奖励领取时间为<b>次月 1日至 10 日（UTC+0）</b>，错过领取时间将无法获得奖励。
              </p>
            </div>
          </div>
        </>
      ) : (
        <FullLoading />
      )}
      {showContinueModal && (
        <div className={mintStyles.mask} onClick={() => setContinueModal(false)}>
          <div className={mintStyles.mask_content} onClick={(e) => e.stopPropagation()}>
            <div className={mintStyles.mask_main} dangerouslySetInnerHTML={{ __html: resp!.info.lp_desc }} />
            <div className={mintStyles.mask_btn}>
              <div
                className={mintStyles.btn_item}
                onClick={() => {
                  location.href = resp!.info.lp_url;
                  setContinueModal(false);
                }}
              >
                {$t('mint.continue')}
              </div>
              <div className={mintStyles.btn_item} onClick={() => setContinueModal(false)}>
                {$t('mint.close')}
              </div>
            </div>
          </div>
        </div>
      )}
      {showKnowModal && (
        <div className={mintStyles.mask} onClick={() => setShowKnowModal(false)}>
          <div className={mintStyles.mask_content} onClick={(e) => e.stopPropagation()}>
            <div className={mintStyles.mask_main}>{msg}</div>
            <div className={mintStyles.mask_btn}>
              <div
                className={`${mintStyles.btn_item_fx_1} ${mintStyles.btn_item}`}
                onClick={() => {
                  setShowKnowModal(false);
                  if (msg === msgMap.auth) {
                    window.location.href = getAuthUrl({ hasAssets: true, hasSnapshots: true, returnTo: location.pathname });
                  }
                }}
              >
                {$t('action.know')}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
