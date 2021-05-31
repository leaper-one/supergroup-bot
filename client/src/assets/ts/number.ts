import BigNumber from "bignumber.js"

type likeNumber = string | number | BigNumber

export const Times = (...r: likeNumber[]): BigNumber =>
  r.reduce((pre: BigNumber, cur: likeNumber) => {
    return pre.times(cur)
  }, new BigNumber(1))

export const Div = (s1: likeNumber, s2: likeNumber, dot = 8): string => {
  return new BigNumber(s1).div(s2).toFixed(dot)
}

export const Add = (...r: likeNumber[]): BigNumber =>
  r.reduce((pre: BigNumber, cur: likeNumber) => {
    return pre.plus(cur)
  }, new BigNumber(0))

export const getUsd = (n: likeNumber, isUsd = true, dot = 2): string => {
  let str = isUsd ? "$" : ""
  let num = new BigNumber(n)
  if (dot === 2 && num.lt(0.1)) {
    str += num.toFormat(4)
  } else {
    str += new BigNumber(n).toFormat(dot)
  }
  return str
}
