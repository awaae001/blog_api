import { api } from '@/utils/request'
import type { PaginatedResponse } from './response'

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

export const getImages = (params: ImageListParams) => api.get<PaginatedImages>('/action/image', params)
export const createImage = (data: CreateImagePayload) => api.post('/action/image', data)
export const updateImage = (id: number, data: UpdateImagePayload) => api.put(`/action/image/${id}`, data)
export const deleteImage = (id: number) => api.del(`/action/image/${id}`)
