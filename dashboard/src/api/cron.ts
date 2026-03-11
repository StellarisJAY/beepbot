import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type { CronJob, CreateCronJobRequest, UpdateCronJobRequest } from '@/types/cron'

// Cron 筛选参数
export interface CronFilter {
  name?: string
  enabled?: boolean
}

/**
 * 定时任务 API 服务
 */
export const cronApi = {
  /**
   * 获取定时任务列表（分页，支持筛选）
   */
  list(page: number = 1, size: number = 10, filters?: CronFilter): Promise<PageResponse<CronJob>> {
    const params: Record<string, unknown> = { page, size }
    if (filters) {
      if (filters.name) params.name = filters.name
      if (filters.enabled !== undefined) params.enabled = filters.enabled
    }
    return http.get<CronJob[]>('/crons', { params }) as Promise<PageResponse<CronJob>>
  },

  /**
   * 获取单个定时任务
   */
  get(id: string): Promise<ApiResponse<CronJob>> {
    return http.get<CronJob>(`/crons/${id}`)
  },

  /**
   * 创建定时任务
   */
  create(data: CreateCronJobRequest): Promise<ApiResponse<CronJob>> {
    return http.post<CronJob>('/crons', data)
  },

  /**
   * 更新定时任务
   */
  update(id: string, data: UpdateCronJobRequest): Promise<ApiResponse<CronJob>> {
    return http.put<CronJob>(`/crons/${id}`, data)
  },

  /**
   * 删除定时任务
   */
  delete(id: string): Promise<ApiResponse<void>> {
    return http.delete<void>(`/crons/${id}`)
  },

  /**
   * 更新定时任务状态
   */
  updateStatus(id: string, enabled: boolean): Promise<ApiResponse<void>> {
    return http.put<void>(`/crons/${id}/status`, { enabled })
  },

  /**
   * 获取指定智能体的定时任务
   */
  getByAgent(agentId: string): Promise<ApiResponse<CronJob[]>> {
    return http.get<CronJob[]>(`/crons/agent/${agentId}`)
  },
}

export default cronApi