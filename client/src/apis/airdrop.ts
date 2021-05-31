import { apis } from "@/apis/http"

export interface IAppointResp {
  status: string
  is_received?: boolean
  code_id?: string
  comment?: string
}

export const ApiAppointment = (groupNumber: string): Promise<IAppointResp> =>
  apis.get(`/airdrop/appointment/${groupNumber}`)

export const ApiAppointmentStatus = (
  groupNumber: string,
): Promise<IAppointResp> =>
  apis.get(`/airdrop/appointment/status/${groupNumber}`)

export const ApiAirdropReceived = (
  groupNumber: string,
): Promise<IAppointResp> => apis.get(`/airdrop/received/${groupNumber}`)
