import request from '@/utils/request'
import type { ApiResponse, PaginatedResponse } from './response'

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

export interface FriendLinkListParams {
  page?: number
  page_size?: number
  is_died?: boolean
  status?: string
  search?: string
}

export type PaginatedFriendLinks = PaginatedResponse<FriendLink>

export interface CreateFriendLinkPayload {
  name: string
  link: string
  avatar?: string
  description?: string
  email?: string
  enable_rss?: boolean
}

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

/** 一次巡查任务的进度，进程重启后重置。 */
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
 * 获取友链列表
 */
export const getFriendLinks = (params: FriendLinkListParams): Promise<ApiResponse<PaginatedFriendLinks>> => {
  return request({
    url: '/action/friend',
    method: 'get',
    params
  })
}


/**
 * 创建友链
 */
export const createFriendLink = (data: CreateFriendLinkPayload): Promise<ApiResponse> => {
  return request({
    url: '/action/friend',
    method: 'post',
    data
  })
}


/**
 * 更新友链
 */
export const updateFriendLink = (id: number, payload: UpdateFriendLinkPayload): Promise<ApiResponse> => {
  const fieldMap: Record<string, string> = {
    name: 'website_name',
    link: 'website_url',
    avatar: 'website_icon_url'
  }
  const mappedData: Record<string, unknown> = {}
  Object.entries(payload.data).forEach(([key, value]) => {
    const targetKey = fieldMap[key] ?? key
    mappedData[targetKey] = value
  })
  return request({
    url: `/action/friend/${id}`,
    method: 'put',
    data: {
      ...payload,
      data: mappedData
    }
  })
}


/**
 * 删除友链
 */
export const deleteFriendLink = (id: number): Promise<ApiResponse> => {
  return request({
    url: `/action/friend/${id}`,
    method: 'delete'
  })
}


/**
 * 手动重新巡查单条友链
 */
export const recheckFriendLink = (id: number): Promise<ApiResponse> => {
  return request({
    url: `/action/friend/recheck/${id}`,
    method: 'post'
  })
}


/**
 * 触发全量巡查（后台异步执行）
 */
export const recheckAllFriendLinks = (): Promise<ApiResponse<FriendLinkInspectionProgress>> => {
  return request({
    url: '/action/friend/recheck',
    method: 'post'
  })
}


/**
 * 查询全量巡查进度
 */
export const getRecheckProgress = (): Promise<ApiResponse<FriendLinkInspectionProgress>> => {
  return request({
    url: '/action/friend/recheck',
    method: 'get'
  })
}
