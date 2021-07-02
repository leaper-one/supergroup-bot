import React from "react"
import styles from "./index.less"
import { Modal, Toast } from "antd-mobile"

interface Props {
  children?: any
  loading?: boolean
  disabled?: boolean
  onClick?: () => void
  className?: string
  type?: string
}

export const Button = (props: Props) => {
  const { children, loading, disabled, onClick, className, type } = props
  return (
    <button
      className={`${styles.button} ${
        disabled && styles.disabled
      } ${className} ${type && styles[type]}`}
      onClick={() => {
        if (loading) return
        if (typeof onClick === "function") onClick()
      }}
    >
      {loading ? (
        <img
          className={styles.loading}
          src={require("@/assets/img/btnLoading.png")}
          alt=""
        />
      ) : (
        children
      )}
    </button>
  )
}

export const ToastSuccess = (content = "Successful", duration = 2) => {
  Toast.info(toast("successful", content), duration)
}

export const ToastFailed = (content = "Failed", duration = 2) => {
  Toast.info(toast("failed", content), duration)
}

export const ToastWarning = (content = "Warning", duration = 2) => {
  Toast.info(toast("warning", content), duration)
  return false
}

const toast = (icon: string, content: string) => (
  <div className={styles.toast}>
    <img
      className={styles.icon}
      src={require(`@/assets/img/${icon}.png`)}
      alt=""
    />
    <span>{content}</span>
  </div>
)

export const Confirm = (title: string, content = "") =>
  new Promise((resolve) => {
    Modal.alert(title, content, [
      {
        text: "取消",
        onPress: () => {
          resolve(false)
        },
      },
      {
        text: "确认",
        onPress: () => {
          resolve(true)
        },
      },
    ])
  })

export const Prompt = (title: string, content: ""): Promise<string> =>
  new Promise(resolve => {
    Modal.prompt(title, content, [
      { text: "取消" },
      { text: "提交", onPress: resolve }
    ])
  })
