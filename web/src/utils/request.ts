import axios, { AxiosError } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error: AxiosError<any>) => {
    if (error.response) {
      const { status, data, config } = error.response

      if (status === 401) {
        if (config.url === '/verify') {
          ElMessage.error(data?.message || '用户名或密码错误')
        } else {
          localStorage.removeItem('token')
          router.push('/panel/login')
          ElMessage.error('登录已过期，请重新登录')
        }
      } else {
        ElMessage.error(data?.message || '请求失败')
      }
    } else {
      ElMessage.error('网络错误，请检查连接')
    }

    return Promise.reject(error)
  }
)

export default request
