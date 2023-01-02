import React, { useEffect, useState } from 'react';
import idxStyles from './index.less';
import styles from './record.less';
import { get$t } from '@/locales/tools';
import { useIntl } from '@@/plugin-locale/localeExports';
import { BackHeader } from '@/components/BackHeader';
import { changeTheme, getURLParams } from '@/assets/ts/tools';
import { ApiGetLiquidityRecordByID, LiquidityRecordResp } from '@/apis/mint';

const statusMap = {
  S: '已领取',
  W: '未领取',
  F: '未达标',
};

export default function Page() {
  const $t = get$t(useIntl());
  const { id } = getURLParams();
  const [activeTab, setActiveTab] = useState(0);
  const [recordList, setRecordList] = useState<LiquidityRecordResp[]>([]);
  const [allList, setAllList] = useState<LiquidityRecordResp[]>([]);

  const preHandle = (list: LiquidityRecordResp[]) =>
    list.map((durationItem) => {
      return {
        duration: durationItem.duration,
        status: durationItem.status,
        list: durationItem.list.slice(0, 2),
        is_open: false,
        has_more: durationItem.list.length > 2,
      };
    });

  useEffect(() => {
    changeTheme('#a650de');
    ApiGetLiquidityRecordByID(id).then((records) => {
      setAllList(records);
      setRecordList(preHandle(records));
    });
    return () => {
      changeTheme('#fffff');
    };
  }, []);

  const handleTab = (index: number) => {
    if (activeTab === index) return;
    if (index === 0) {
      setRecordList(preHandle(allList));
    } else if (index === 1) {
      setRecordList(preHandle(allList.filter((item) => item.status === 'S')));
    } else if (index === 2) {
      setRecordList(preHandle(allList.filter((item) => item.status === 'W')));
    }
    setActiveTab(index);
  };

  return (
    <div className={idxStyles.bg}>
      <BackHeader name="奖励记录" isWhite className={idxStyles.headerBg} />
      <div className={styles.content}>
        <div className={styles.tabs}>
          {['所有', '已领取', '未领取'].map((item, idx) => (
            <div className={`${styles.tab} ${activeTab === idx && styles.active}`} onClick={() => handleTab(idx)}>
              {item}
            </div>
          ))}
        </div>

        {recordList.map((durationItem, idx) => (
          <div className={styles.list}>
            <div className={styles.header}>
              <div>{durationItem.duration}</div>
              <button>{statusMap[durationItem.status as keyof typeof statusMap]}</button>
            </div>
            {durationItem.list.map((item) => (
              <div className={styles.item}>
                <div className={styles.itemDate}> {item.date} </div>
                <div className={styles.itemTitle}>交易对</div>
                <div className={styles.itemTitle}>LP 数量</div>
                <div className={styles.itemTitle}>收益占比</div>
                <div className={styles.itemAmount}>{item.lp_symbol}</div>
                <div className={styles.itemAmount}>{item.lp_amount}</div>
                <div className={styles.itemAmount}>100%</div>
              </div>
            ))}
            {!durationItem.is_open && durationItem.has_more && (
              <div
                className={styles.more}
                onClick={() => {
                  const newList = [...recordList];
                  newList[idx].is_open = true;
                  newList[idx].list = allList[idx].list;
                  setRecordList(newList);
                }}
              >
                <img src={require(`@/assets/img/active/liquidity/ic_down.png`)} alt="" className={styles.moreIcon} />
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
