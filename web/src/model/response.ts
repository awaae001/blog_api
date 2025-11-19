/**
 * Unified API response structure, corresponding to the backend's model.ApiResponse
 */
export interface ApiResponse<T = any> {
  code: number    // HTTP status code
  message: string // Response message
  data: T         // Response data
}