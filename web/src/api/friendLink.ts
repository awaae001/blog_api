import request from '@/utils/request'
import type { ApiResponse } from '@/model/response'
import type {
  FriendLinkListParams,
  PaginatedFriendLinks,
  CreateFriendLinkPayload,
  UpdateFriendLinkPayload
} from '@/model/friendLink'

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
  return request({
    url: `/action/friend/${id}`,
    method: 'put',
    data: payload
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