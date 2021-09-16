import React, { FC } from "react"

export type AppIcons = ""

interface IconProps extends React.HTMLProps<HTMLElement> {
  i: AppIcons
}
export const Icon: FC<IconProps> = ({ className, i, ...rest }) => (
  <i className={`iconfont icon${i} ${className}`} {...rest} />
)
