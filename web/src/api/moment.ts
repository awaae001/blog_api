import request from '@/utils/request'
import type { ApiResponse } from '@/model/response'
import type {
  MomentListParams,
  QueryMomentsResponse,
  CreateMomentPayload,
  UpdateMomentPayload,
  CreateMediaPayload
} from '@/model/moment'

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
