import { apis } from "@/apis/http"

export interface IAppointResp {
  status: number
}

export const ApiGetAirdrop = (airdropID: string): Promise<IAppointResp> =>
  apis.get(`/airdrop/${airdropID}`)

export const ApiAirdropReceived = (airdropID: string): Promise<number> =>
  apis.post(`/airdrop/${airdropID}`)
