import React, { FC } from "react"
import { RollingBoxProvider } from "./context"

interface RollingBoxOptions {}

export const RollingBox: FC = ({ children }) => {
  return <RollingBoxProvider>{children}</RollingBoxProvider>
}
