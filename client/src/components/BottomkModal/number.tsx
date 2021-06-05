import React, { MutableRefObject, useEffect, useRef, useState } from "react";
import { ToastWarning } from "@/components/Sub";
import styles from "@/pages/manager/sendBroadcast.less";
import { Modal } from 'antd-mobile'
import { useIntl } from "@@/plugin-locale/localeExports";
import { get$t } from "@/locales/tools";


interface INumberConfrimProps {
  show: boolean
  setShow: (v: boolean) => void
  title: string
  content: string
  confirm: () => void
}

export const NumberConfirm = (props: INumberConfrimProps) => {
  const [code, setCode] = useState<number[]>([])
  const [myCode, setMyCode] = useState<Array<string>>(["", "", "", ""])
  const $t = get$t(useIntl())
  const codeRefList = [
    useRef<HTMLInputElement>(),
    useRef<HTMLInputElement>(),
    useRef<HTMLInputElement>(),
    useRef<HTMLInputElement>(),
  ]
  useEffect(() => {
    setCode(getRandomCode())
    setMyCode(["", "", "", ""])
    setTimeout(() => {
      codeRefList[0].current?.focus()
    }, 100)
  }, [props.show])
  useEffect(() => {
    const finished = myCode.every((item) => item.length > 0)
    if (!finished) return
    const valid = myCode.every((item, idx) => item === String(code[idx]))
    if (valid) {
      props.confirm()
    } else ToastWarning($t("manager.broadcast.checkNumber"))
  }, [myCode])


  return <Modal
    visible={props.show}
    popup
    onClose={() => props.setShow(false)}
    animationType="slide-up"
  >
    <div className={styles.dialog}>
      <img
        className={styles.close}
        src={require("@/assets/img/svg/closeBtn.svg")}
        alt=""
        onClick={() => {
          props.setShow(false)
        }}
      />
      <div className={styles.title}>{props.title}</div>
      <p className={styles.text} dangerouslySetInnerHTML={{ __html: props.content }}/>
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
              if (!e.target.value) return
              myCode[idx] = e.target.value[e.target.value.length - 1]
              setMyCode([...myCode])
              if (idx < 3) codeRefList[idx + 1].current?.focus()
              else codeRefList[idx].current?.blur()
            }}
          />
        ))}
      </div>
      <p>{$t("manager.broadcast.input")}</p>
    </div>
  </Modal>
}


interface IConfirmContent {
  data: string
  close: () => void
  $t: any
  setLoading: (loading: boolean) => void
}

const getRandomCode = () =>
  Array(4)
    .fill(1)
    .map((_) => (Math.random() * 10) | 0)
