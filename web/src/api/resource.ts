import request from '@/utils/request'
import type { ApiResponse } from '@/model/response'

export interface UploadFileResponse {
  message: string
  url: string
  local_path?: string
  objectKey?: string
}

/**
 * 上传文件
 */
export const uploadFile = (data: FormData, target: 'local' | 'oss' = 'local'): Promise<ApiResponse<UploadFileResponse>> => {
  return request({
    url: `/action/resource/${target}`,
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

/**
 * 删除文件
 */
export const deleteFile = (filePath: string, target: 'local' | 'oss'): Promise<ApiResponse> => {
  return request({
    url: `/action/resource/${target}/${filePath}`,
    method: 'delete'
  })
}