import React, { useEffect, useState } from 'react';
import { environment, getConversationId } from '@/assets/ts/tools';
import { history, useIntl } from 'umi';
import { FullLoading, Loading } from '@/components/Loading';
import { get$t } from '@/locales/tools';
import { $get } from '@/stores/localStorage';

export default () => {
  const [content, setContent] = useState('');
  const $t = get$t(useIntl());
  useEffect(() => {
    if (environment()) {
      checkGroup();
    } else {
      setContent($t('error.mixin'));
    }
  }, []);

  return <>{content ? <Loading content={content} noCancel /> : <FullLoading />}</>;
};

async function checkGroup() {
  const nextPage = $get('token') ? '/home' : '/join';
  history.push(nextPage);
}
