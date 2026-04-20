import axios, { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { message } from 'antd'
import { useAuthStore } from '../store/useAuthStore'

// API response type
interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// Create axios instance
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
request.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    if (data.code === 0) {
      return response
    }
    // Business error
    message.error(data.message || 'Request failed')
    return Promise.reject(new Error(data.message || 'Request failed'))
  },
  (error: AxiosError<ApiResponse>) => {
    if (error.response) {
      const { status, data } = error.response

      switch (status) {
        case 401:
          message.error('Login expired, please login again')
          useAuthStore.getState().logout()
          window.location.href = '/login'
          break
        case 403:
          message.error('No permission')
          break
        case 404:
          message.error('Resource not found')
          break
        case 500:
          message.error('Server error')
          break
        default:
          message.error(data?.message || 'Request failed')
      }
    } else if (error.request) {
      message.error('Network error, please check your connection')
    } else {
      message.error(error.message || 'Request failed')
    }
    return Promise.reject(error)
  }
)

// Generic request methods
export const get = <T = unknown>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
  return request.get(url, config)
}

export const post = <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
  return request.post(url, data, config)
}

export const put = <T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
  return request.put(url, data, config)
}

export const del = <T = unknown>(url: string, config?: AxiosRequestConfig): Promise<AxiosResponse<ApiResponse<T>>> => {
  return request.delete(url, config)
}

export default request
