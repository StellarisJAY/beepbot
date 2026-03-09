import type { Agent } from './agent'

/**
 * 机器人状态枚举
 */
export enum BotStatus {
  Active = 'active',
  Inactive = 'inactive',
}

/**
 * 机器人状态显示名称映射
 */
export const BotStatusLabels: Record<BotStatus, string> = {
  [BotStatus.Active]: '启用',
  [BotStatus.Inactive]: '禁用',
}

/**
 * 机器人状态选项（用于下拉选择）
 */
export const BotStatusOptions = [
  { value: BotStatus.Active, label: BotStatusLabels[BotStatus.Active] },
  { value: BotStatus.Inactive, label: BotStatusLabels[BotStatus.Inactive] },
]

/**
 * 机器人平台类型枚举
 */
export enum BotPlatform {
  QQ = 'qq',
  Feishu = 'feishu',
}

/**
 * 机器人平台显示名称映射
 */
export const BotPlatformLabels: Record<BotPlatform, string> = {
  [BotPlatform.QQ]: 'QQ',
  [BotPlatform.Feishu]: '飞书',
}

/**
 * 机器人平台选项（用于下拉选择）
 */
export const BotPlatformOptions = [
  { value: BotPlatform.QQ, label: BotPlatformLabels[BotPlatform.QQ] },
  { value: BotPlatform.Feishu, label: BotPlatformLabels[BotPlatform.Feishu] },
]

/**
 * 机器人配置
 */
export interface Bot {
  id: string
  name: string
  description: string
  platform: BotPlatform
  identifier: string
  config: Record<string, unknown>
  agent_id: string | null
  agent?: Agent
  status: BotStatus
  created_at: string
  updated_at: string
}

/**
 * 创建机器人请求
 */
export interface CreateBotRequest {
  name: string
  description?: string
  platform: BotPlatform
  identifier?: string
  config?: Record<string, unknown>
  agent_id?: string
}

/**
 * 更新机器人请求
 */
export interface UpdateBotRequest {
  name?: string
  description?: string
  identifier?: string
  config?: Record<string, unknown>
  status?: BotStatus
}

/**
 * 绑定智能体请求
 */
export interface BindAgentRequest {
  agent_id: string | null
}

/**
 * QQ 机器人配置
 */
export interface QQBotConfig {
  app_id: string
  app_secret: string
}

/**
 * 飞书机器人配置
 */
export interface FeishuBotConfig {
  app_id: string
  app_secret: string
  encrypt_key?: string
}