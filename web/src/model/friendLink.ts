/**
 * 定义友链的数据结构，与后端模型保持一致
 */
export interface FriendLink {
  id: number
  name: string
  link: string
  avatar: string
  description: string
  status: 'survival' | 'timeout' | 'error' | 'pending'
  enable_rss: boolean
  skip_health_check: boolean
  is_died?: boolean
  email?: string
  times?: number
  updated_at: number
}

/**
 * 分页查询参数
 */
export interface FriendLinkListParams {
  page?: number
  page_size?: number
  is_died?: boolean
  status?: string
  search?: string
}

/**
 * 分页响应数据结构
 */
export interface PaginatedFriendLinks {
  items: FriendLink[]
  total: number
  page: number
  page_size: number
}

/**
 * 创建友链的请求体
 */
export interface CreateFriendLinkPayload {
  name: string
  link: string
  avatar?: string
  description?: string
  email?: string
  enable_rss?: boolean
}

/**
 * 更新友链的请求体
 */
export type UpdateFriendLinkData = Partial<Omit<FriendLink, 'id' | 'updated_at'>> & {
  website_name?: string
  website_url?: string
  website_icon_url?: string
}

export interface UpdateFriendLinkPayload {
  data: UpdateFriendLinkData
  opt?: {
    overwrite_if_blank?: boolean
  }
}

/**
 * 一次巡查任务的进度（后端内存态，进程重启后重置）
 */
export interface FriendLinkInspectionProgress {
  running: boolean
  label?: string
  started_at?: number
  finished_at?: number
  total: number
  processed: number
  survival: number
  failed: number
  database_failures: number
  rss_discovered: number
  error?: string
}

/**
 * 单条友链在当前/最近一轮巡查中的位置
 */
export interface FriendLinkRecheckState {
  id: number
  in_run: boolean
  done: boolean
  run_status?: string
  run: FriendLinkInspectionProgress
  link: FriendLink
}
