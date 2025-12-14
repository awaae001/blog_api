import type { ApiResponse } from '@/model/response'

// Corresponds to model.StatusData in the backend
export interface StatusData {
  friend_link_count: number
  rss_count: number
  rss_post_count: number
}

// Corresponds to model.SystemStatus in the backend
export interface SystemStatus {
  uptime: string
  status_data: StatusData
  time: string
}

// The actual data structure within the main ApiResponse
export type SystemStatusResponse = ApiResponse<SystemStatus>