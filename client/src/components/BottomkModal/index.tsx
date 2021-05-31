import React, { useRef, useState } from "react"
import styles from "./index.less"

interface Props {
  content: JSX.Element
  close: () => void
  closeWithAnimation?: number
  top?: number
  height?: number
}

let tmpClose = 0

interface IStyle {
  top?: string
  height?: string
}

export const BottomModal = (props: Props) => {
  const { content, close, closeWithAnimation, top, height } = props
  const styleObj: IStyle = {}
  top && (styleObj.top = top + "px")
  height && (styleObj.height = height + "px")
  const [closeAnimate, setCloseAnimate] = useState(false)
  const closeModal = () => {
    setCloseAnimate(true)
    setTimeout(() => {
      document.body.style.overflow = ""
      close()
    }, 300)
  }
  if (closeWithAnimation !== undefined && closeWithAnimation != tmpClose) {
    tmpClose = closeWithAnimation
    closeModal()
  }
  document.body.style.overflow = "hidden"
  return (
    <div className={"modal"}>
      <div className="mask" onClick={closeModal} />
      <div
        className={`${styles.content} ${closeAnimate ? styles.close : ""} ${
          top ? styles.hasTop : ""
        }`}
        style={styleObj}
      >
        {content}
      </div>
    </div>
  )
}
