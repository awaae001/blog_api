import request from '@/utils/request'
import type { ApiResponse } from './response'

export interface FriendLinkStatusCount {
  status: string
  count: number
}

export interface RssPostCountMonthly {
  month: string
  count: number
}

export interface StatusData {
  friend_link_count: number
  rss_count: number
  rss_post_count: number
  friend_link_status_pie: FriendLinkStatusCount[]
  rss_post_count_monthly: RssPostCountMonthly[]
}

export interface KVCacheStatus {
  entry_count: number
  payload_bytes: number
}

/** 对应后端 model.SystemStatus。 */
export interface SystemStatus {
  uptime: string
  status_data: StatusData
  database_size_bytes: number
  data_folder_size_bytes: number
  kv_cache: KVCacheStatus
  time: number
}

export const getSystemStatus = (): Promise<ApiResponse<SystemStatus>> => {
  return request.get('/status')
}
