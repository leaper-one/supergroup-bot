import React, { useState } from 'react';
import styles from './addLive.less';
import { BackHeader } from '@/components/BackHeader';
import { get$t } from '@/locales/tools';
import { useIntl } from '@@/plugin-locale/localeExports';
import { Button, ToastSuccess } from '@/components/Sub';
import { Modal } from 'antd-mobile';
import { ApiUploadFile } from '@/apis/common';
import { ApiPostLive, ILive } from '@/apis/live';
import { $get } from '@/stores/localStorage';
import { Icon } from '@/components/Icon';

let onceClick = false;
export default function Page() {
  const $t = get$t(useIntl());
  const [form, setForm] = useState<ILive>($get('active_live'));
  const [show, setShow] = useState(false);

  return (
    <div>
      <BackHeader name={$t('news.sendLive')} />
      <div className={styles.content}>
        <div className={styles.img}>
          {form.img_url ? (
            <img className={styles.imgView} src={form.img_url} alt="" />
          ) : (
            <>
              <Icon i="tianjia" className={styles.imgIcon} />
              <span className={styles.imgDesc}>{$t('news.form.img')}</span>
            </>
          )}
          <input
            type="file"
            className={styles.imgFile}
            accept="image/*"
            onChange={async (e) => {
              const file = e.target.files![0];
              const { view_url } = (await ApiUploadFile(file)) || {};
              if (!view_url) return;
              setForm({ ...form, img_url: view_url });
            }}
          />
        </div>

        <div className={`${styles.item} ${styles.select}`} onClick={() => setShow(true)}>
          <span>{form.category ? $t('news.form.' + form.category) : $t('news.form.category')}</span>
          <Icon i="ic_down" />
        </div>

        <div className={`${styles.item} ${styles.input}`}>
          <input type="text" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} placeholder={$t('news.form.title')} />
        </div>

        <div className={`${styles.item} ${styles.textarea}`}>
          <textarea value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder={$t('news.form.desc')} />
        </div>

        <Button
          className={styles.btn}
          onClick={async () => {
            if (onceClick) return;
            onceClick = true;
            const res = await ApiPostLive(form);
            if (res === 'success') ToastSuccess($t('success.save'));
            setTimeout(() => {
              history.go(-1);
            }, 500);
            onceClick = false;
          }}
        >
          {$t('action.save')}
        </Button>
      </div>

      <NewsCategorySelectModal show={show} setShow={setShow} $t={$t} category={form.category} setCategory={(v) => setForm({ ...form, category: v })} />
    </div>
  );
}

interface INewsFormCategory {
  show: boolean;
  setShow: (v: boolean) => void;
  $t: any;
  category: number;
  setCategory: (v: number) => void;
}

const NewsCategorySelectModal = (props: INewsFormCategory) => (
  <Modal popup animationType="slide-up" visible={props.show} onClose={() => props.setShow(false)}>
    <div className={styles.categorySelect}>
      <Icon i="guanbi" className={styles.close} onClick={() => props.setShow(false)} />
      <h3>{props.$t('news.form.category')}</h3>
      <ul>
        {[1, 2].map((item) => (
          <li
            key={item}
            onClick={() => {
              props.setCategory(item);
              props.setShow(false);
            }}
          >
            <p>{props.$t('news.form.' + item)}</p>
            {item === props.category && <Icon i="check" />}
          </li>
        ))}
      </ul>
    </div>
  </Modal>
);
