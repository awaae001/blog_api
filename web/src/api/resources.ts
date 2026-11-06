import request, { api } from '@/utils/request'

export interface ResourceEntry {
  name: string
  path: string
  is_dir: boolean
  size: number
  mod_time: string
}

export interface UploadFileResponse {
  message: string
  url: string
  local_path?: string
  objectKey?: string
}

const encodePath = (filePath: string) =>
  encodeURI(filePath).replace(/[?#]/g, (match) => encodeURIComponent(match))
const resourcePath = (path: string) => encodePath(path.replace(/^\/+/, ''))

export const uploadFile = (data: FormData, target: 'local' | 'oss' = 'local') =>
  api.post<UploadFileResponse>(`/action/resource/${target}`, data)
export const deleteFile = (filePath: string, target: 'local' | 'oss') =>
  api.del(`/action/resource/${target}/${encodePath(filePath)}`)
export const listResources = (path = '') => api.get<ResourceEntry[]>(`/action/resource/${resourcePath(path)}`)

/** 文件内容直接返回，不包 ApiResponse。 */
export const getResourceFileBlob = (path: string) =>
  request.get<never, Blob>(`/action/resource/${resourcePath(path)}`, { responseType: 'blob' })
