import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type {
  Bot,
  CreateBotRequest,
  UpdateBotRequest,
  BindAgentRequest,
} from '@/types/bot'

// Bot 筛选参数
export interface BotFilter {
  name?: string
  status?: string
  platform?: string
}

/**
 * 机器人 API 服务
 */
export const botApi = {
  /**
   * 获取机器人列表（分页，支持筛选）
   */
  list(page: number = 1, size: number = 10, filters?: BotFilter): Promise<PageResponse<Bot>> {
    const params: Record<string, unknown> = { page, size }
    if (filters) {
      if (filters.name) params.name = filters.name
      if (filters.status) params.status = filters.status
      if (filters.platform) params.platform = filters.platform
    }
    return http.get<Bot[]>('/bots', { params }) as Promise<PageResponse<Bot>>
  },

  /**
   * 获取单个机器人
   */
  get(id: string): Promise<ApiResponse<Bot>> {
    return http.get<Bot>(`/bots/${id}`)
  },

  /**
   * 创建机器人
   */
  create(data: CreateBotRequest): Promise<ApiResponse<Bot>> {
    return http.post<Bot>('/bots', data)
  },

  /**
   * 更新机器人
   */
  update(id: string, data: UpdateBotRequest): Promise<ApiResponse<Bot>> {
    return http.put<Bot>(`/bots/${id}`, data)
  },

  /**
   * 删除机器人
   */
  delete(id: string): Promise<ApiResponse<void>> {
    return http.delete<void>(`/bots/${id}`)
  },

  /**
   * 更新机器人状态
   */
  updateStatus(id: string, status: string): Promise<ApiResponse<void>> {
    return http.put<void>(`/bots/${id}/status`, { status })
  },

  /**
   * 绑定智能体
   */
  bindAgent(id: string, data: BindAgentRequest): Promise<ApiResponse<void>> {
    return http.put<void>(`/bots/${id}/agent`, data)
  },

  /**
   * 获取未绑定智能体的机器人
   */
  getUnbound(): Promise<ApiResponse<Bot[]>> {
    return http.get<Bot[]>('/bots/unbound')
  },

  /**
   * 按平台获取机器人
   */
  getByPlatform(platform: string): Promise<ApiResponse<Bot[]>> {
    return http.get<Bot[]>(`/bots/platform/${platform}`)
  },
}

export default botApi