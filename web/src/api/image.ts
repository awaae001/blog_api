import request from '@/utils/request'
import type { ApiResponse } from '@/model/response'
import type {
  ImageListParams,
  PaginatedImages,
  CreateImagePayload,
  UpdateImagePayload
} from '@/model/image'

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