import request from '@/utils/request'
import type { SystemStatusResponse } from '@/model/stats'

export const statsApi = {
  getSystemStatus() {
    return request.get<any, SystemStatusResponse>('/status')
  }
}