import request from '@/utils/request'
import type { ApiResponse } from '@/model/response'

/**
 * 定义友链的数据结构，与后端模型保持一致
 */
export interface FriendLink {
  id: number
  website_name: string
  website_url: string
  website_icon_url: string
  description: string
  status: 'survival' | 'timeout' | 'error' | 'died' | 'pending' | 'ignored'
  email?: string
  times?: number
  updated_at: string
}

/**
 * 分页查询参数
 */
export interface FriendLinkListParams {
  page?: number
  page_size?: number
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
 * 创建友链的请求体
 */
export interface CreateFriendLinkPayload {
  website_name: string
  website_url: string
  website_icon_url?: string
  description?: string
  email?: string
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
 * 更新友链的请求体
 */
export interface UpdateFriendLinkPayload {
  id: number
  data: Partial<Omit<FriendLink, 'id' | 'updated_at'>>
  opt?: {
    overwrite_if_blank?: boolean
  }
}

/**
 * 更新友链
 */
export const updateFriendLink = (payload: UpdateFriendLinkPayload): Promise<ApiResponse> => {
  return request({
    url: '/action/friend',
    method: 'put',
    data: payload
  })
}

/**
 * 删除友链的请求体
 */
export interface DeleteFriendLinkPayload {
  ids: number[]
}

/**
 * 删除友链
 */
export const deleteFriendLink = (data: DeleteFriendLinkPayload): Promise<ApiResponse> => {
  return request({
    url: '/action/friend',
    method: 'delete',
    data
  })
}