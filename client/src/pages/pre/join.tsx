import React, { useEffect, useState } from 'react';
import { ApiGetGroup } from '@/apis/group';
import { history, useIntl } from 'umi';
import { IJoin, Join } from '@/components/Join';
import { CodeURL } from '@/components/CodeURL';
import { environment, setHeaderTitle } from '@/assets/ts/tools';
import { get$t } from '@/locales/tools';
import { $get } from '@/stores/localStorage';
import BigNumber from 'bignumber.js';

export default () => {
  const $t = get$t(useIntl());
  const [joinProps, setJoinProps] = useState<IJoin>();
  const mixinCtx = environment();
  const { from, c } = history.location.query || {};
  const handleClickBtn = () => history.replace(`/auth?state=` + (c ? c : ''));
  const initPage = async () => {
    const groupInfo = await ApiGetGroup();
    setTimeout(() => setHeaderTitle(groupInfo.name));
    if (from === 'auth') handleClickBtn();
    groupInfo.total_people = `${new BigNumber(groupInfo.total_people).toFormat()} ${$t('join.main.member')}`;
    setJoinProps({
      groupInfo,
      button: $t('join.main.join'),
      buttonAction: handleClickBtn,
      tips: $t('join.main.joinTips'),
    });
  };

  useEffect(() => {
    if ($get('token')) return history.replace(`/`);
    initPage();
  }, []);

  return <>{mixinCtx ? <Join props={{ ...joinProps, loading: false } as IJoin} /> : joinProps?.groupInfo ? <CodeURL groupInfo={joinProps?.groupInfo} action="join" /> : null}</>;
};
