import request from '@/utils/request'
import type { ApiResponse, PaginatedResponse } from './response'

export interface Image {
  id: number
  name: string
  url: string
  local_path: string
  is_local: number
  is_oss: number
  status: string
}

export interface ImageListParams {
  page?: number
  page_size?: number
  status?: string
  search?: string
}

export type PaginatedImages = PaginatedResponse<Image>

export interface CreateImagePayload {
  name: string
  url: string
  local_path?: string
  is_local?: number
  is_oss?: number
}

export interface UpdateImagePayload {
  name?: string
  url?: string
  status?: string
}

/**
 * 获取图片列表
 */
export const getImages = (params: ImageListParams): Promise<ApiResponse<PaginatedImages>> => {
  return request({
    url: '/action/image',
    method: 'get',
    params
  })
}

/**
 * 创建图片
 */
export const createImage = (data: CreateImagePayload): Promise<ApiResponse> => {
  return request({
    url: '/action/image',
    method: 'post',
    data
  })
}

/**
 * 更新图片
 */
export const updateImage = (id: number, payload: UpdateImagePayload): Promise<ApiResponse> => {
  return request({
    url: `/action/image/${id}`,
    method: 'put',
    data: payload
  })
}

/**
 * 删除图片
 */
export const deleteImage = (id: number): Promise<ApiResponse> => {
  return request({
    url: `/action/image/${id}`,
    method: 'delete'
  })
}