import color from "./colors"
import { history } from 'umi'
import crypto from 'crypto'

export const getURLParams = (): any => history.location.query || {}
export const getTheme = (): string | undefined => {
  let metas = document.getElementsByTagName("meta")
  for (let i = 0; i < metas.length; i++) {
    if (metas[i].name === "theme-color") return metas[i].content
  }
}
export const getUUID = (): string => {
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === "x" ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}
export const getMixinCtx = () => {
  let ctx
  switch (environment()) {
    case "iOS":
      ctx = prompt("MixinContext.getContext()") as string
      return JSON.parse(ctx)
    case "Android":
      const w: any = window
      ctx = w.MixinContext.getContext()
      return JSON.parse(ctx)
    default:
      return undefined
  }
}


export const getConversationIdByUserIDs = (r1: string, r2: string) => {
  let [minId, maxId] = [r1, r2]
  if (minId > maxId) {
    [minId, maxId] = [r2, r1]
  }

  const hash = crypto.createHash('md5')
  hash.update(minId)
  hash.update(maxId)
  const bytes = hash.digest()

  bytes[6] = (bytes[6] & 0x0f) | 0x30
  bytes[8] = (bytes[8] & 0x3f) | 0x80

  // eslint-disable-next-line unicorn/prefer-spread
  const digest = Array.from(bytes, byte => `0${(byte & 0xff).toString(16)}`.slice(-2)).join('')
  const uuid = `${digest.slice(0, 8)}-${digest.slice(8, 12)}-${digest.slice(12, 16)}-${digest.slice(
    16,
    20
  )}-${digest.slice(20, 32)}`
  return uuid
}
export const getConversationId = (): string | undefined => {
  const ctx = getMixinCtx()
  return ctx?.conversation_id
}

export const getMixinVersion = () => {
  const ctx = getMixinCtx()
  return ctx?.app_version
}

export const changeTheme = (color: string) => {
  let head = document.getElementsByTagName("head")[0]
  let metas = document.getElementsByTagName("meta")
  let body = document.getElementsByTagName("body")[0]
  body.style.backgroundColor = color
  for (let i = 0; i < metas.length; i++) {
    if (metas[i].name === "theme-color") {
      head.removeChild(metas[i])
    }
  }
  let meta = document.createElement("meta")
  meta.name = "theme-color"
  meta.content = color
  head.appendChild(meta)
  const w: any = window
  switch (environment()) {
    case "iOS":
      return (
        w.webkit.messageHandlers.reloadTheme &&
        w.webkit.messageHandlers.reloadTheme.postMessage("")
      )
    case "Android":
      let { reloadTheme } = w.MixinContext
      if (typeof reloadTheme === "function") return w.MixinContext.reloadTheme()
  }
}
export const getAvatarColor = (id: string): string => color(id)

export const environment = (): string | undefined => {
  const w: any = window
  if (
    w.webkit &&
    w.webkit.messageHandlers &&
    w.webkit.messageHandlers.MixinContext
  ) {
    return "iOS"
  }
  if (w.MixinContext && w.MixinContext.getContext) {
    return "Android"
  }
  return (
    (process.env.NODE_ENV === "development" &&
      "Chrome") ||
    undefined
  )
}
export const playlist = (audios: string[]) => {
  const w: any = window
  switch (environment()) {
    case 'iOS':
      w.webkit && w.webkit.messageHandlers && w.webkit.messageHandlers.playlist && w.webkit.messageHandlers.playlist.postMessage(audios)
      return
    case 'Android':
    case 'Desktop':
      w.MixinContext && (typeof w.MixinContext.playlist === 'function') && w.MixinContext.playlist(audios)
      return
  }
}
export const delay = (number = 1500): Promise<undefined> => {
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve(undefined)
    }, number)
  })
}

export const pageToTop = () => window.scrollTo(0, 0)

export const copy = (message: string) => {
  const input = document.createElement("input")
  input.value = message
  document.body.appendChild(input)
  input.select()
  input.setSelectionRange(0, input.value.length), document.execCommand("Copy")
  document.body.removeChild(input)
}

export const base64Encode = (msg: Object | string): string => {
  if (typeof msg === "object") {
    msg = JSON.stringify(msg)
  }
  return encodeURIComponent(Buffer.from(msg as string).toString("base64"))
}

export const setHeaderTitle = (title: string) => {
  document.title = title
}