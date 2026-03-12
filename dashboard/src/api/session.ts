import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type {
  SessionListItem,
  MessagesResponse,
  CompressSessionRequest,
  CompressSessionResponse,
  UsageStatsResponse,
} from '@/types/session'

// Session 筛选参数
export interface SessionFilter {
  session_type?: string
  platform?: string
}

/**
 * 会话 API 服务
 */
export const sessionApi = {
  /**
   * 获取智能体的会话列表（支持筛选）
   */
  getAgentSessions(
    agentId: string,
    page: number = 1,
    size: number = 10,
    filters?: SessionFilter,
  ): Promise<PageResponse<SessionListItem>> {
    const params: Record<string, unknown> = { page, size }
    if (filters) {
      if (filters.session_type) params.session_type = filters.session_type
      if (filters.platform) params.platform = filters.platform
    }
    return http.get<SessionListItem[]>(`/agents/${agentId}/sessions`, {
      params,
    }) as Promise<PageResponse<SessionListItem>>
  },

  /**
   * 获取会话消息列表
   * @param sessionId 会话ID
   * @param beforeId 加载此ID之前的消息，为空则加载最新消息
   * @param limit 每次加载的消息数量
   */
  async getSessionMessages(
    sessionId: string,
    beforeId?: string,
    limit: number = 20,
  ): Promise<MessagesResponse> {
    const params: Record<string, unknown> = { limit }
    if (beforeId) {
      params.before_id = beforeId
    }
    const response = await http.get<MessagesResponse>(`/sessions/${sessionId}/messages`, {
      params,
    })
    return response.data
  },

  /**
   * 压缩会话上下文
   * @param sessionId 会话ID
   * @param request 压缩请求参数
   */
  async compressSession(
    sessionId: string,
    request?: CompressSessionRequest,
  ): Promise<CompressSessionResponse> {
    const response = await http.post<CompressSessionResponse>(
      `/sessions/${sessionId}/compress`,
      request || {},
    )
    return response.data
  },

  /**
   * 获取智能体用量统计
   * @param agentId 智能体ID
   * @param period 时间范围: 1d, 3d, 7d, 14d, 30d
   */
  async getAgentUsageStats(agentId: string, period: string = '7d'): Promise<UsageStatsResponse> {
    const response = await http.get<UsageStatsResponse>(`/agents/${agentId}/usage`, {
      params: { period },
    })
    return response.data
  },
}

export default sessionApi