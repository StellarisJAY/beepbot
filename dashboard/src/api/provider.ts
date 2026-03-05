import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type { Provider, CreateProviderRequest, UpdateProviderRequest } from '@/types/provider'

/**
 * 供应商 API 服务
 */
export const providerApi = {
  /**
   * 获取供应商列表
   */
  list(): Promise<ApiResponse<Provider[]>> {
    return http.get<Provider[]>('/providers')
  },

  /**
   * 获取单个供应商
   */
  get(id: string): Promise<ApiResponse<Provider>> {
    return http.get<Provider>(`/providers/${id}`)
  },

  /**
   * 创建供应商
   */
  create(data: CreateProviderRequest): Promise<ApiResponse<Provider>> {
    return http.post<Provider>('/providers', data)
  },

  /**
   * 更新供应商
   */
  update(id: string, data: UpdateProviderRequest): Promise<ApiResponse<Provider>> {
    return http.put<Provider>(`/providers/${id}`, data)
  },

  /**
   * 删除供应商
   */
  delete(id: string): Promise<ApiResponse<void>> {
    return http.delete<void>(`/providers/${id}`)
  },

  /**
   * 设置默认供应商
   */
  setDefault(id: string): Promise<ApiResponse<Provider>> {
    return http.patch<Provider>(`/providers/${id}/default`)
  },

  /**
   * 测试供应商连接
   */
  testConnection(id: string): Promise<ApiResponse<{ success: boolean; message: string }>> {
    return http.post<{ success: boolean; message: string }>(`/providers/${id}/test`)
  },
}

export default providerApi