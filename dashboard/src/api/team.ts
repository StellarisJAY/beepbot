import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type { Team, CreateTeamRequest, UpdateTeamRequest } from '@/types/team'

/**
 * 团队 API 服务
 */
// Team 筛选参数
export interface TeamFilter {
  name?: string
  status?: string
}

export const teamApi = {
  /**
   * 获取团队列表（分页，支持筛选）
   */
  list(page: number = 1, size: number = 10, filters?: TeamFilter): Promise<PageResponse<Team>> {
    const params: Record<string, unknown> = { page, size }
    if (filters) {
      if (filters.name) params.name = filters.name
      if (filters.status) params.status = filters.status
    }
    return http.get<Team[]>('/teams', { params }) as Promise<PageResponse<Team>>
  },

  /**
   * 获取单个团队
   */
  get(id: string): Promise<ApiResponse<Team>> {
    return http.get<Team>(`/teams/${id}`)
  },

  /**
   * 创建团队
   */
  create(data: CreateTeamRequest): Promise<ApiResponse<Team>> {
    return http.post<Team>('/teams', data)
  },

  /**
   * 更新团队
   */
  update(id: string, data: UpdateTeamRequest): Promise<ApiResponse<Team>> {
    return http.put<Team>(`/teams/${id}`, data)
  },

  /**
   * 删除团队
   */
  delete(id: string): Promise<ApiResponse<void>> {
    return http.delete<void>(`/teams/${id}`)
  },

  /**
   * 更新团队状态
   */
  updateStatus(id: string, status: string): Promise<ApiResponse<void>> {
    return http.put<void>(`/teams/${id}/status`, { status })
  },

  /**
   * 获取智能体所属的团队列表
   */
  getAgentTeams(agentId: string): Promise<ApiResponse<Team[]>> {
    return http.get<Team[]>(`/teams/agent/${agentId}`)
  },
}

export default teamApi