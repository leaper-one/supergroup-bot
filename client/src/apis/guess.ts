import {
  RecordByDate,
  LotteryRecord,
  Record,
  GuessResponse,
  Guess,
  GuessPageInitData,
} from "@/types"
import { apis } from "./http"

export const ApiGetGuessPageData = (id: string): Promise<GuessPageInitData> =>
  apis.get(`/guess/${id}`).then((x: GuessPageInitData<GuessResponse>) => ({
    today_guess: x.today_guess,
    guess: {
      ...x.guess,
      rules: x.guess.rules.split(";"),
      explain: x.guess.explain.split(";"),
    },
  }))

export interface CreateGuessForm {
  trace_id: string
  guess_type: string
}

const ApiCreateGuess = (data: CreateGuessForm) => apis.post("/guess", data)
const ApiGetRecordsByGuessId = (id: string) => apis.get(`/guess/record/${id}`)
