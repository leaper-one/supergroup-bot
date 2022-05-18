import { BackHeader } from '@/components/BackHeader';
import React, { MutableRefObject, useEffect, useRef, useState } from 'react';
import styles from './sendBroadcast.less';
import { Button, ToastSuccess, ToastWarning } from '@/components/Sub';
import { Modal } from 'antd-mobile';
import { ApiPostBroadcast } from '@/apis/broadcast';
import { history, useIntl } from 'umi';
import { get$t } from '@/locales/tools';
import { FullLoading } from '@/components/Loading';

interface IConfirmContent {
  data: string;
  close: () => void;
  $t: any;
  setLoading: (loading: boolean) => void;
}

const ConfirmContent = (props: IConfirmContent) => {
  const [code, setCode] = useState<number[]>([]);
  const [myCode, setMyCode] = useState<Array<string>>(['', '', '', '']);
  const $t = props.$t;
  const codeRefList = [useRef<HTMLInputElement>(), useRef<HTMLInputElement>(), useRef<HTMLInputElement>(), useRef<HTMLInputElement>()];
  useEffect(() => {
    setCode(getRandomCode());
    codeRefList[0].current?.focus();
  }, []);
  useEffect(() => {
    const finished = myCode.every((item) => item.length > 0);
    if (!finished) return;
    const valid = myCode.every((item, idx) => item === String(code[idx]));
    if (valid) {
      props.setLoading(true);
      ApiPostBroadcast(props.data).then((res) => {
        if (res) {
          ToastSuccess($t('success.send'));
          props.close();
          history.goBack();
          props.setLoading(false);
        }
      });
    } else ToastWarning($t('broadcast.checkNumber'));
  }, [myCode]);

  return (
    <div className={styles.dialog}>
      <img className={styles.close} src={require('@/assets/img/svg/closeBtn.svg')} alt="" onClick={props.close} />
      <div className={styles.title}>{$t('broadcast.sent')}</div>
      <p className={styles.text}>{props.data}</p>
      <div className={styles.validate}>
        {code.map((item, idx) => (
          <span key={idx}>{item}</span>
        ))}
      </div>
      <div className={styles.inputBox}>
        {myCode.map((codeItem, idx) => (
          <input
            ref={codeRefList[idx] as MutableRefObject<HTMLInputElement>}
            key={idx}
            type="number"
            value={myCode[idx]}
            onChange={(e) => {
              myCode[idx] = e.target.value;
              setMyCode([...myCode]);
              if (!e.target.value) return;
              if (idx < 3) codeRefList[idx + 1].current?.focus();
              else codeRefList[idx].current?.blur();
            }}
          />
        ))}
      </div>
      <p>{$t('broadcast.input')}</p>
    </div>
  );
};

export default () => {
  const [confirmModal, setConfirmModal] = useState(false);
  const [data, setData] = useState('');
  const [loading, setLoading] = useState(false);
  const $t = get$t(useIntl());
  return (
    <>
      {loading && <FullLoading opacity={true} mask />}
      <>
        <BackHeader name={$t('broadcast.sent')} />
        <div className={styles.container}>
          <textarea value={data} onChange={(e) => setData(e.target.value)} placeholder={$t('broadcast.holder')} />
          <Button
            className="btn"
            onClick={async () => {
              if (data.length === 0) return ToastWarning($t('broadcast.fill'), 2);
              setConfirmModal(true);
            }}
          >
            {$t('broadcast.send')}
          </Button>
        </div>
        <Modal visible={confirmModal} popup onClose={() => setConfirmModal(false)} animationType="slide-up">
          <ConfirmContent data={data} close={() => setConfirmModal(false)} $t={$t} setLoading={setLoading} />
        </Modal>
      </>
    </>
  );
};
const getRandomCode = () =>
  Array(4)
    .fill(1)
    .map((_) => (Math.random() * 10) | 0);
