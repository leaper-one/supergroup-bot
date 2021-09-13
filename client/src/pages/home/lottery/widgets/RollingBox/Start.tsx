import React, { FC } from "react"
import styled from "styled-components"
import { useRollingBox } from "./context"

const Container = styled.div`
  display: flex;
  justify-content: center;
  align-items: center;
`

export interface StartProps {}
export const Start: FC<StartProps> = ({ children }) => {
  const { start, status } = useRollingBox()

  const handleClick = (e: React.MouseEvent) => {
    e.preventDefault()
    if (status === "idle") {
      start()
    }
  }

  return <Container onClick={handleClick}>{children}</Container>
}
