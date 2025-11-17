import request from '@/utils/request'

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  code: number
  message: string
  data: {
    token: string
    expires_at: string
  }
}

export const authApi = {
  login(data: LoginRequest) {
    return request.post<any, LoginResponse>('/verify', data)
  }
}
