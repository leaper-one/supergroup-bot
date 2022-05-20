import vConsole from 'vconsole';
import { requestConfig } from './apis/http';
import { $set } from '@/stores/localStorage';

export const request = requestConfig;
if (process.env.NODE_ENV === 'development' && navigator.userAgent.includes('Mixin')) new vConsole();

let envLang = process.env.LANG;
let lang = navigator.language;

if (envLang === 'en') {
  if (lang.includes('zh')) {
    $set('umi_locale', 'zh');
  } else if (lang.includes('ja')) {
    $set('umi_locale', 'ja');
  } else {
    $set('umi_locale', 'en');
  }
} else {
  $set('umi_locale', envLang);
}

export function modifyClientRenderOpts(memo: any) {
  return {
    ...memo,
    rootElement: memo.rootElement,
  };
}
