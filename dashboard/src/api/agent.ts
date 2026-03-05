import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type {
  Agent,
  CreateAgentRequest,
  UpdateAgentRequest,
} from '@/types/agent'

/**
 * 智能体 API 服务
 */
export const agentApi = {
  /**
   * 获取智能体列表（分页）
   */
  list(page: number = 1, size: number = 10): Promise<PageResponse<Agent>> {
    return http.get<Agent[]>('/agents', { params: { page, size } }) as Promise<PageResponse<Agent>>
  },

  /**
   * 获取单个智能体
   */
  get(id: string): Promise<ApiResponse<Agent>> {
    return http.get<Agent>(`/agents/${id}`)
  },

  /**
   * 创建智能体
   */
  create(data: CreateAgentRequest): Promise<ApiResponse<Agent>> {
    return http.post<Agent>('/agents', data)
  },

  /**
   * 更新智能体
   */
  update(id: string, data: UpdateAgentRequest): Promise<ApiResponse<Agent>> {
    return http.put<Agent>(`/agents/${id}`, data)
  },

  /**
   * 删除智能体
   */
  delete(id: string): Promise<ApiResponse<void>> {
    return http.delete<void>(`/agents/${id}`)
  },

  /**
   * 更新智能体状态
   */
  updateStatus(id: string, status: string): Promise<ApiResponse<void>> {
    return http.patch<void>(`/agents/${id}/status`, { status })
  },
}

export default agentApi