export const $set = (key: string, value: any) => {
  if (typeof value === "object") {
    value = JSON.stringify(value)
  }
  localStorage.setItem(key, value)
}

export const $get = (key: string) => {
  let v = localStorage.getItem(key)
  if (!v) return v
  try {
    return JSON.parse(v)
  } catch (e) {
    return v
  }
}

export const $remove = (key: string) => {
  localStorage.removeItem(key)
}
