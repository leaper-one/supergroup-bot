let prefix: string[] = []

export function getI18n(origin: any, target: any) {
  const { length } = Object.keys(origin)
  let i = 0

  for (let key in origin) {
    if (typeof origin[key] === "string") {
      target[[...prefix, key].join(".")] = origin[key]
    } else {
      prefix.push(key)
      getI18n(origin[key], target)
      prefix.pop()
    }
    i++
  }
}

export const get$t = (intl: any) => (id: string, params: object = {}) =>
  intl.formatMessage({ id }, params)

export const getTimeZone = () => (0 - new Date().getTimezoneOffset()) / 60 + "h"
