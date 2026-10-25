import request from '@/utils/request'
import type { ApiResponse } from './response'

export interface Moment {
  id: number
  content: string
  status: 'visible' | 'hidden' | 'deleted'
  guild_id?: number
  channel_id?: number
  message_id?: number
  message_link?: string
  created_at: number
  updated_at: number
}

export interface MomentMedia {
  id: number
  moment_id: number
  name?: string
  media_url: string
  media_type: 'image' | 'video'
  is_local: number
  is_deleted: number
}

export interface MomentWithMedia extends Moment {
  media: MomentMedia[]
  reactions?: Record<string, number>
  selected_reaction?: string
}

export interface QueryMomentsResponse {
  moments: MomentWithMedia[]
  total: number
}

export interface MomentListParams {
  page: number
  page_size: number
  status?: string
}

export interface CreateMomentPayload {
  content: string
  media: Array<{
    media_url: string
    media_type: 'image' | 'video'
  }>
  guild_id?: number
  channel_id?: number
  message_id?: number
  message_link?: string
}

export interface UpdateMomentPayload {
  content?: string
  status?: 'visible' | 'hidden' | 'deleted'
  message_link?: string
}

export interface CreateMediaPayload {
  moment_id: number
  name?: string
  media_url: string
  media_type: 'image' | 'video'
  is_local: number
}

export const getMoments = (params: MomentListParams): Promise<ApiResponse<QueryMomentsResponse>> => {
  return request({
    url: '/action/moments',
    method: 'get',
    params
  })
}

export const createMoment = (payload: CreateMomentPayload): Promise<ApiResponse> => {
  return request({
    url: '/action/moments',
    method: 'post',
    data: payload
  })
}

export const updateMoment = (id: number, payload: UpdateMomentPayload): Promise<ApiResponse> => {
  return request({
    url: `/action/moments/${id}`,
    method: 'put',
    data: payload
  })
}

export const deleteMoment = (id: number): Promise<ApiResponse> => {
  return request({
    url: `/action/moments/${id}`,
    method: 'delete'
  })
}

export const createMomentMedia = (payload: CreateMediaPayload): Promise<ApiResponse> => {
  return request({
    url: '/action/moments/media',
    method: 'post',
    data: payload
  })
}

export const deleteMomentMedia = (id: number, hard = false): Promise<ApiResponse> => {
  return request({
    url: `/action/moments/media/${id}`,
    method: 'delete',
    params: hard ? { hard: 1 } : undefined
  })
}
