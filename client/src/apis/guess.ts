import {
  GuessResponse,
  GuessPageInitData,
  GuessType,
  GuessRecord,
} from "@/types"
import { apis } from "./http"

export const ApiGetGuessPageData = (id: string): Promise<GuessPageInitData> =>
  apis.get(`/guess/${id}`).then((x: GuessPageInitData<GuessResponse>) => ({
    ...x,
    rules: x.rules.split(";"),
    explain: x.explain.split(";"),
  }))

export interface CreateGuessForm {
  guess_id: string
  guess_type: GuessType
}

export const ApiCreateGuess = (data: CreateGuessForm) =>
  apis.post("/guess", data)

export const ApiGetGuessRecord = (id: string): Promise<GuessRecord[]> =>
  apis.get("/guess/record", { id })
