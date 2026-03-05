import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'

// API 响应结构
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 分页响应结构（与后端 PageResponse 一致）
export interface PageResponse<T> {
  code: number
  message: string
  data: T[]
  total: number
  page: number
  size: number
}

// 错误响应结构
export interface ApiError {
  code: number
  message: string
}

// 默认配置
const DEFAULT_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api'
const DEFAULT_TIMEOUT = 30000

// 创建 axios 实例
const createHttpClient = (baseURL: string = DEFAULT_BASE_URL): AxiosInstance => {
  const instance = axios.create({
    baseURL,
    timeout: DEFAULT_TIMEOUT,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // 请求拦截器
  instance.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
      // 可以在这里添加 token 等认证信息
      // const token = localStorage.getItem('token')
      // if (token) {
      //   config.headers.Authorization = `Bearer ${token}`
      // }
      return config
    },
    (error) => {
      return Promise.reject(error)
    }
  )

  // 响应拦截器
  instance.interceptors.response.use(
    (response: AxiosResponse<ApiResponse>) => {
      const { data } = response

      // 检查业务状态码
      if (data.code === 0 || data.code === 200) {
        return response
      }

      // 业务错误处理
      const error: ApiError = {
        code: data.code,
        message: data.message || '请求失败',
      }
      return Promise.reject(error)
    },
    (error) => {
      // HTTP 错误处理
      let message = '网络错误，请稍后重试'

      if (error.response) {
        const status = error.response.status
        switch (status) {
          case 400:
            message = '请求参数错误'
            break
          case 401:
            message = '未授权，请重新登录'
            // 可以在这里处理登录过期逻辑
            break
          case 403:
            message = '拒绝访问'
            break
          case 404:
            message = '请求的资源不存在'
            break
          case 500:
            message = '服务器内部错误'
            break
          case 502:
            message = '网关错误'
            break
          case 503:
            message = '服务不可用'
            break
          case 504:
            message = '网关超时'
            break
          default:
            message = `请求失败 (${status})`
        }
      } else if (error.code === 'ECONNABORTED') {
        message = '请求超时，请稍后重试'
      } else if (error.message === 'Network Error') {
        message = '网络连接失败，请检查网络'
      }

      const apiError: ApiError = {
        code: error.response?.status || -1,
        message,
      }
      return Promise.reject(apiError)
    }
  )

  return instance
}

// HTTP 客户端类
class HttpClient {
  private instance: AxiosInstance

  constructor(baseURL?: string) {
    this.instance = createHttpClient(baseURL)
  }

  // GET 请求
  get<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.instance.get(url, config).then((res) => res.data)
  }

  // POST 请求
  post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.instance.post(url, data, config).then((res) => res.data)
  }

  // PUT 请求
  put<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.instance.put(url, data, config).then((res) => res.data)
  }

  // PATCH 请求
  patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.instance.patch(url, data, config).then((res) => res.data)
  }

  // DELETE 请求
  delete<T = unknown>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return this.instance.delete(url, config).then((res) => res.data)
  }

  // 获取原始 axios 实例（用于特殊场景）
  getRawInstance(): AxiosInstance {
    return this.instance
  }
}

// 导出默认实例
export const http = new HttpClient()

// 导出创建新实例的方法（用于多 API 场景）
export const createHttp = (baseURL: string): HttpClient => {
  return new HttpClient(baseURL)
}

export default http