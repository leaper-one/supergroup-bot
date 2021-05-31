import { apis } from "@/apis/http"
import { IGenerateTransfer } from "@/apis/conversation"
import { getTimeZone } from "@/locales/tools"
import { $get } from "@/stores/localStorage"

export interface IRedPacket {
  type: string
  packet_id?: string
  mode?: string
  asset_id?: string
  amount?: string
  people?: number
  memo?: string
  group_id?: string
  zone?: string

  rate?: string
  date_cycle?: string
  time_cycle?: string
  send_times?: number
}

export const ApiPostRedGenerate = (
  packet: IRedPacket,
): Promise<IGenerateTransfer> => {
  packet.zone = getTimeZone()
  packet.group_id = $get("group").group_id

  return apis.post(`/red/generate`, packet)
}

export interface IRedCheckPaid {
  payed: boolean
}

export const ApiGetRedCheckPaid = (trace_id: string): Promise<IRedCheckPaid> =>
  apis.get(`/red/checkPaid/${trace_id}`)
