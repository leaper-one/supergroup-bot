import React, { useEffect, useState } from 'react';
import { BackHeader } from '@/components/BackHeader';
import { List } from './';
import { get$t } from '@/locales/tools';
import { useIntl } from '@@/plugin-locale/localeExports';

function getManagerList($t: any) {
  return [
    [
      {
        icon: 'iconshequnxinxi',
        type: $t('manager.description'),
        mount: '',
        route: '/manager/hello?status=description',
        // route: "/setting/group",
      },
      {
        icon: 'iconruqunhuanyingyu',
        type: $t('manager.welcome'),
        route: '/manager/hello?status=welcome',
        // route: "/setting/hello",
      },
    ],
  ];
}

export default () => {
  const [managerList, setManagerList] = useState<any[]>([]);
  const $t = get$t(useIntl());

  useEffect(() => {
    initPage();
  }, []);
  const initPage = async () => {
    setManagerList(getManagerList($t));
  };

  return (
    <>
      <BackHeader name={$t('manager.base')} />
      <List lists={managerList} />
      {/*/!* : *!/*/}
      {/*{container(settingLists)}*/}
      {/*}*/}
    </>
  );
};
