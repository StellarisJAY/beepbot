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
  // 会话推送信息（通过智能体对话创建的定时任务才有）
  push_channel?: string // 推送渠道类型：qq/feishu
  push_bot_id?: string // 推送机器人ID
  push_user_id?: string // 推送目标用户ID
  push_group_id?: string // 推送目标群ID（群聊时）
  push_chat_id?: string // 会话ID（飞书 chat_id）
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