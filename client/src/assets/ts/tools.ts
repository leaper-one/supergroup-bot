import color from "./colors"

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
export const getConversationId = (): string | undefined => {
  const ctx = getMixinCtx()
  if (ctx) {
    return ctx.conversation_id
  } else if (process.env.CLIENT_ID === "11efbb75-e7fe-44d7-a14f-698535289310") {
    // return undefined
    return "309428e2-15a2-43dc-aa47-6739ff75b746"
  }
}

export const getMixinVersion = () => {
  const ctx = getMixinCtx()
  return (
    (ctx && ctx.app_version) ||
    (process.env.CLIENT_ID === "11efbb75-e7fe-44d7-a14f-698535289310" &&
      "27.0.0")
  )
}

export const changeTheme = (color: string) => {
  let head = document.getElementsByTagName("head")[0]
  let metas = document.getElementsByTagName("meta")
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
    (process.env.CLIENT_ID === "11efbb75-e7fe-44d7-a14f-698535289310" &&
      "Chrome") ||
    undefined
  )
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