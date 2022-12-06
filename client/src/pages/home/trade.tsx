import React, { useEffect, useState } from 'react';
import styles from './trade.less';
import { BackHeader } from '@/components/BackHeader';
import { get$t } from '@/locales/tools';
import { useIntl } from 'umi';
import { Flex } from 'antd-mobile';
import { ApiGetSwapList, IExinAd, ISwapItem } from '@/apis/transfer';
import { getUsd } from '@/assets/ts/number';
import { ApiGetAssetByID, IAsset } from '@/apis/asset';
import { getMixSwapUrl, getExinLocalUrl, getExinOtcUrl } from '@/apis/http';
import { FullLoading } from '@/components/Loading';
import { setHeaderTitle } from '@/assets/ts/tools';
import { Icon } from '@/components/Icon';

const assetIDSymbolMap: any = {};
export default (props: any) => {
  const $t = get$t(useIntl());
  const { id } = props.match.params;
  const [swapList, setSwapList] = useState<ISwapItem[]>([] as ISwapItem[]);
  const [asset, setAsset] = useState<IAsset>();
  const [loading, setLoading] = useState(false);
  const [current, setCurrent] = useState('coin');
  const [exinAd, setExinAd] = useState<IExinAd[]>();
  const [hasOTC] = useState<string>(process.env.HAS_OTC as string);
  let timer: any;

  useEffect(() => {
    setLoading(true);
    initPage().then(() => setLoading(false));
    return () => {
      clearTimeout(timer);
    };
  }, []);

  const initPage = async () => {
    let { list, asset, ad } = await ApiGetSwapList(id);
    if (id === '31d2ea9c-95eb-3355-b65b-ba096853bc18') {
      list = list?.filter((item) => item.asset1_symbol.includes('USD') || ['2', '3'].includes(item.type));
    }
    if (list) setSwapList(list);
    setHeaderTitle($t('transfer.title', { name: asset?.symbol }));
    setAsset(asset);
    timer = setTimeout(() => {
      initPage();
    }, 5 * 1000);

    if (ad && ad.length > 0) {
      const assets: string[] = [];
      for (let i = 0; i < ad.length; i++) {
        if (!assets.includes(ad[i].assetId!)) {
          assets.push(ad[i].assetId!);
        }
      }
      await Promise.all(assets.map((item) => ApiGetAssetByID(item))).then((res) => res.forEach((item) => (assetIDSymbolMap[item.asset_id!] = item.symbol)));
      setExinAd(ad);
    }
  };

  const change_usd = Number(asset?.change_usd);
  const color = change_usd > 0 ? 'green' : 'red';

  let price: any = '';
  if (asset) {
    const p = getUsd(asset.price_usd!, false);
    if (Number(p) === 0) price = $t('transfer.noPrice');
    else price = `$ ${p}`;
  }

  return (
    <>
      <div className={`${styles.container} ${hasOTC === '0' ? styles.noOtc : ''}`}>
        <BackHeader name={$t('transfer.title', { name: asset?.symbol })} />
        {asset && (
          <section className={styles.price}>
            <div className={styles.title}>
              <span>{$t('transfer.price')}</span>
              <span>24h</span>
            </div>
            <div className={styles.amount}>
              <span className={price.startsWith('$') ? '' : styles.noPrice}>{price}</span>
              <span className={styles[color]}>{(Number(change_usd) * 100).toFixed(2)}%</span>
            </div>
          </section>
        )}
        {exinAd && (
          <div className={styles.category}>
            {['coin', 'otc'].map((item) => (
              <span key={item} className={current === item ? styles.active : ''} onClick={() => setCurrent(item)}>
                {$t(`transfer.${item}`)}
              </span>
            ))}
          </div>
        )}
        {current === 'coin' ? (
          <ul className={styles.swapList}>{swapList.map((item, idx) => tradeCard(item, asset!, idx, $t))}</ul>
        ) : (
          <ul className={styles.swapList}>{exinAd?.map((item) => adCard(item, $t))}</ul>
        )}
      </div>
      {loading && <FullLoading mask />}
    </>
  );
};

const tradeCard = (item: ISwapItem, asset: IAsset, idx: number, $t: any) => {
  switch (item.type) {
    case '0':
    case '1':
    case '4':
      return swapCard(item, $t);
    case '2':
    case '3':
      return transferCard(item, asset, idx, $t);
  }
};

const payIconMap = {
  bank: 'iconyinhangka',
  alipay: 'iconzhifubao',
  wechatpay: 'iconweixinzhifu',
};

const adCard = (item: IExinAd, $t: any) => (
  <li className={styles.swapCard} key={item.id} onClick={() => (location.href = `https://www.tigaex.com/#/exchange/takeOrder?id=${item.id}`)}>
    <div className={styles.adCardHead}>
      <img src={item.avatarUrl} alt="" />
      <h3>{item.nickname}</h3>
      <span className={styles.adCardHeadMsg}>
        {
          <>
            {item.isLandun && (
              <>
                <Icon i="landun" />
                <span>{$t('transfer.auth')}</span>
              </>
            )}
            {item.isCertification && (
              <>
                <Icon i="yishiming" />
                <span>{$t('transfer.identity')}</span>
              </>
            )}
          </>
        }
      </span>
      <p className={styles.adCardHeadPrice}>{item.price} CNY</p>
    </div>
    <div className={styles.adCardContent}>
      <div className={styles.coinInfo}>
        <Flex className={styles.grep} justify="between">
          <span>{$t('transfer.payMethod')}</span>
          <span>{$t('transfer.category')}</span>
        </Flex>
        <Flex className={styles['m-t-4']} justify="between">
          <p className={styles.payMethods}>
            {item.payMethods?.map((item) => (
              <>
                <i className={`iconfont ${payIconMap[item.symbol]} ${styles[item.symbol]}`} />
                <span>{$t(`transfer.${item.symbol}`)}</span>
              </>
            ))}
          </p>
          <p>{assetIDSymbolMap[item.assetId!]}</p>
        </Flex>
        <Flex className={`${styles.grep} ${styles['m-t-10']}`} justify="between">
          <span>{$t('transfer.limit')}</span>
          <span>{$t('transfer.in5minRate')}</span>
        </Flex>
        <Flex className={styles['m-t-4']} justify="between">
          <p>{`${Number(item.minPrice)} ~ ${Number(item.maxPrice)} CNY`}</p>
          <p>{Number((Number(item.in5minRate) * 100).toFixed(2))}%</p>
        </Flex>
        <Flex className={`${styles.grep} ${styles['m-t-10']}`} justify="between">
          <span>{$t('transfer.multisigOrderCount')}</span>
          <span>{$t('transfer.orderSuccessRank')}</span>
        </Flex>
        <Flex className={styles['m-t-4']} justify="between">
          <p>{item.multisigOrderCount}</p>
          <p>{Number((Number(item.orderSuccessRank) * 100).toFixed(2))}%</p>
        </Flex>
      </div>
    </div>
  </li>
);

const swapCard = (item: ISwapItem, $t: any) => {
  return (
    <li className={styles.swapCard} key={item.lp_asset} onClick={() => (location.href = getMixSwapUrl(item.asset0, item.asset1))}>
      <div className={styles.coin}>
        <img className={styles.coinIcon} src={item.icon_url} alt="" />
        <h4 className={styles.coinTitle}>{[item.asset0_symbol, item.asset1_symbol].join('-')}</h4>
        <span className={styles.grep}>{`1 ${item.asset1_symbol} ≈ ${item.rate} ${item.asset0_symbol}`}</span>
        <span className={styles.coinPrice}>{getUsd(item.price!)}</span>
      </div>
      <div className={styles.coinInfo}>
        <Flex className={styles.grep} justify="between">
          <span>{$t('transfer.pool')}</span>
          <span>{$t('transfer.earn')}</span>
        </Flex>
        <Flex className={styles['m-t-4']} justify="between">
          <p>{getUsd(item.pool)}</p>
          <p>{item.earn}</p>
        </Flex>
        <Flex className={`${styles.grep} ${styles['m-t-10']}`} justify="between">
          <span>{$t('transfer.amount')}</span>
          <span>{$t('transfer.method')}</span>
        </Flex>
        <Flex className={styles['m-t-4']} justify="between">
          <p>{getUsd(item.amount)}</p>
          <p>{$t('transfer.maker')}</p>
        </Flex>
      </div>
    </li>
  );
};

const transferCard = (swap: ISwapItem, asset: IAsset, idx: number, $t: any) => {
  const url = swap.type === '2' ? getExinOtcUrl(swap.otc_id!) : getExinLocalUrl(swap.asset_id!);
  const symbol = swap.type === '2' ? 'CNY' : swap.asset1_symbol;
  const price = swap.type === '2' ? getUsd(swap.price_usd === '0' ? asset?.price_usd! : swap.price_usd!) : `${swap.price_usd} ${symbol}`;
  const exchange =
    swap.type === '2'
      ? $t('transfer.taker', {
          exchange: swap.exchange ? $t(`transfer.${swap.exchange}`) : $t('transfer.exchange'),
        })
      : $t('transfer.sign');
  return (
    <li key={idx} className={styles.transferCard} onClick={() => (window.location.href = url)}>
      <div className={styles.coin}>
        <img className={styles.coinIcon} src={swap.icon_url} alt="" />
        <h4 className={styles.coinTitle}>{swap.asset0_symbol}</h4>
        <span className={styles.grep}>{$t('transfer.order', { amount: swap.buy_max, symbol })}</span>
        <span className={styles.coinPrice}>{price}</span>
        <span className={`${styles.coinMethod} ${styles.grep}`}>{exchange}</span>
      </div>
    </li>
  );
};
