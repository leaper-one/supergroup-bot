export type RollingBoxState = {
  ticket: number
  status: "running" | "idle"
}

export type TicketActionTypes = "increment_ticket" | "decrement_ticket"

export type StatusActionTypes = "running" | "idle"

export type RollingBoxAction = {
  type: TicketActionTypes | StatusActionTypes
}

export const rollingBoxReducer = (
  state: RollingBoxState,
  action: RollingBoxAction,
): RollingBoxState => {
  switch (action.type) {
    case "increment_ticket":
      return {
        ...state,
        ticket: state.ticket + 1,
      }

    case "decrement_ticket":
      if (state.ticket <= 0) return state

      return {
        ...state,
        ticket: state.ticket - 1,
      }

    case "idle":
    case "running":
      return { ...state, status: action.type }

    default:
      return state
  }
}
