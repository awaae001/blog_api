import { api } from '@/utils/request'

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

export const getMoments = (params: MomentListParams) => api.get<QueryMomentsResponse>('/action/moments', params)
export const createMoment = (data: CreateMomentPayload) => api.post('/action/moments', data)
export const updateMoment = (id: number, data: UpdateMomentPayload) => api.put(`/action/moments/${id}`, data)
export const deleteMoment = (id: number) => api.del(`/action/moments/${id}`)
export const createMomentMedia = (data: CreateMediaPayload) => api.post<MomentMedia>('/action/moments/media', data)
export const deleteMomentMedia = (id: number, hard = false) =>
  api.del(`/action/moments/media/${id}`, hard ? { hard: 1 } : undefined)
