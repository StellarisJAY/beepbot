/**
 * 定时任务状态常量
 */
export const CRON_JOB_STATUS = {
  ENABLED: 'enabled',
  DISABLED: 'disabled',
} as const

export type CronJobStatus = (typeof CRON_JOB_STATUS)[keyof typeof CRON_JOB_STATUS]

/**
 * 定时任务状态显示名称映射
 */
export const CronJobStatusLabels: Record<CronJobStatus, string> = {
  [CRON_JOB_STATUS.ENABLED]: '启用',
  [CRON_JOB_STATUS.DISABLED]: '禁用',
}

/**
 * 定时任务状态选项（用于下拉选择）
 */
export const CronJobStatusOptions = [
  { value: true, label: CronJobStatusLabels[CRON_JOB_STATUS.ENABLED] },
  { value: false, label: CronJobStatusLabels[CRON_JOB_STATUS.DISABLED] },
]

/**
 * 智能体简要信息（CronJob 关联）
 */
export interface AgentBrief {
  id: string
  name: string
}

/**
 * 定时任务配置
 */
export interface CronJob {
  id: string
  name: string
  cron_expression: string
  agent_id: string
  agent?: AgentBrief
  message: string
  enabled: boolean
  created_at: string
  updated_at: string
}

/**
 * 创建定时任务请求
 */
export interface CreateCronJobRequest {
  name: string
  cron_expression: string
  agent_id: string
  message: string
  enabled?: boolean
}

/**
 * 更新定时任务请求
 */
export interface UpdateCronJobRequest {
  name: string
  cron_expression: string
  agent_id: string
  message: string
  enabled: boolean
}