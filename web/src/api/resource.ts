import request from '@/utils/request'
import type { ApiResponse } from '@/model/response'

export interface UploadFileResponse {
  message: string
  url: string
}

/**
 * 上传文件
 */
export const uploadFile = (data: FormData): Promise<ApiResponse<UploadFileResponse>> => {
  return request({
    url: '/action/resource',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}