// @ts-ignore
import qs from 'qs';
import React from 'react';
import { history } from 'umi';
import { getAuthUrl } from '@/apis/http';
import { FullLoading } from '@/components/Loading';
import { ApiAuth } from '@/apis/user';
import { $set } from '@/stores/localStorage';
import { ToastFailed } from '@/components/Sub';
import { ApiGetGroup } from '@/apis/group';
import { GlobalData } from '@/stores/store';

export default () => {
  const query: any = history.location.query;
  const { code, return_to, state } = query;
  if (code) {
    auth(code, decodeURIComponent(return_to), state);
  } else {
    ApiGetGroup().then((group) => {
      $set('group', group);
      window.location.href = getAuthUrl({ state });
    });
  }
  return <FullLoading />;
};

async function auth(code: string, return_to: string, state = '') {
  const { authentication_token, is_new, ...user } = await ApiAuth(code, state);
  if (!authentication_token) {
    ToastFailed('认证失败...');
    history.replace(`/auth`);
    return;
  }
  $set('token', authentication_token);
  $set('user', user);
  if (is_new) GlobalData.isNewUser = true;
  let url = return_to ? decodeURIComponent(return_to) : '/';
  let pathname = url;
  let query: any = {};
  if (url.includes('?')) {
    const [_pathname, _query] = url.split('?');
    query = qs.parse(_query);
    pathname = _pathname;
  }
  if (state) query.state = state;
  history.push({ pathname, query });
}
