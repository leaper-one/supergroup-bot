import React from 'react';
import styles from './exit.less';
import { get$t } from "@/locales/tools";
import { useIntl } from "@@/plugin-locale/localeExports";

export default function Page() {
  const $t = get$t(useIntl())
  return (
    <div>
      <h3 className={styles.title}>{$t('setting.exited')}</h3>
      <p className={styles.desc}>{$t('setting.exitedDesc')}</p>
    </div>
  );
}
