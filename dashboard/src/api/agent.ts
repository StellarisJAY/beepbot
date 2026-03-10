import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type {
  Agent,
  AgentDefaults,
  CreateAgentRequest,
  UpdateAgentRequest,
  ValidationResult,
  SkillBrief,
  UpdateAgentSkillsRequest,
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
   * 获取智能体默认配置
   */
  getDefaults(): Promise<ApiResponse<AgentDefaults>> {
    return http.get<AgentDefaults>('/agents/defaults')
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
    return http.put<void>(`/agents/${id}/status`, { status })
  },

  /**
   * 校验智能体配置是否完整
   */
  validate(id: string): Promise<ApiResponse<ValidationResult>> {
    return http.post<ValidationResult>(`/agents/${id}/validate`)
  },

  /**
   * 获取智能体关联的技能列表
   */
  getSkills(id: string): Promise<ApiResponse<SkillBrief[]>> {
    return http.get<SkillBrief[]>(`/agents/${id}/skills`)
  },

  /**
   * 更新智能体技能配置
   */
  updateSkills(id: string, data: UpdateAgentSkillsRequest): Promise<ApiResponse<void>> {
    return http.put<void>(`/agents/${id}/skills`, data)
  },
}

export default agentApi