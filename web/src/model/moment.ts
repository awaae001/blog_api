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
