import request from '@/utils/request'
import type { ApiResponse } from '@/model/response'
import type { ResourceEntry } from '@/model/resource'

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
const encodePath = (filePath: string) =>
  encodeURI(filePath).replace(/[?#]/g, (match) => encodeURIComponent(match))

export const deleteFile = (filePath: string, target: 'local' | 'oss'): Promise<ApiResponse> => {
  const encodedPath = encodePath(filePath)
  return request({
    url: `/action/resource/${target}/${encodedPath}`,
    method: 'delete'
  })
}

/**
 * 获取本地资源目录列表
 */
export const listResources = (path = ''): Promise<ApiResponse<ResourceEntry[]>> => {
  const normalizedPath = path.replace(/^\/+/, '')
  const encodedPath = encodePath(normalizedPath)
  const suffix = normalizedPath ? `/${encodedPath}` : '/'
  return request({
    url: `/action/resource${suffix}`,
    method: 'get'
  })
}
