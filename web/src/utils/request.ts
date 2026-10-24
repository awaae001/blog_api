import axios, { type AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error: AxiosError<{ message?: string }>) => {
    if (axios.isCancel(error)) return Promise.reject(error)

    const response = error.response
    let message = response?.data?.message || (response ? '请求失败' : '网络错误，请检查连接')
    if (response?.status === 401) {
      if (response.config.url === '/verify/passwd') {
        message = response.data?.message || '用户名或密码错误'
      } else {
        localStorage.removeItem('token')
        router.push('/login')
        message = '登录已过期，请重新登录'
      }
    }
    ElMessage.error(message)
    return Promise.reject(error)
  }
)

export default request
