import { api } from '@/utils/request'

export interface LoginRequest {
  username: string
  password: string
  turnstile_token?: string
}

export const authApi = {
  login: (data: LoginRequest) => api.post<{ token: string; expires_at: string }>('/verify/passwd', data)
}
