/**
 * 智能体状态枚举
 */
export enum AgentStatus {
  Active = 'active',
  Inactive = 'inactive',
}

/**
 * 智能体状态显示名称映射
 */
export const AgentStatusLabels: Record<AgentStatus, string> = {
  [AgentStatus.Active]: '活跃',
  [AgentStatus.Inactive]: '停用',
}

/**
 * 智能体状态选项（用于下拉选择）
 */
export const AgentStatusOptions = [
  { value: AgentStatus.Active, label: AgentStatusLabels[AgentStatus.Active] },
  { value: AgentStatus.Inactive, label: AgentStatusLabels[AgentStatus.Inactive] },
]

/**
 * 供应商简要信息（Agent 关联）
 */
export interface ProviderBrief {
  id: string
  name: string
  provider_type: string
  base_url: string
}

/**
 * 智能体配置
 */
export interface Agent {
  id: string
  name: string
  description: string
  provider_id: string
  provider?: ProviderBrief
  model: string
  system_prompt: string
  temperature: number
  max_iterations: number
  max_output_tokens: number
  working_dir: string
  context_window_size: number
  window_size: number
  compression_ratio: number
  context_max_tokens: number
  status: AgentStatus
  created_at: string
  updated_at: string
}

/**
 * 智能体默认配置
 */
export interface AgentDefaults {
  system_prompt: string
  temperature: number
  max_iterations: number
  max_output_tokens: number
  context_window_size: number
  window_size: number
  compression_ratio: number
  context_max_tokens: number
}

/**
 * 校验结果
 */
export interface ValidationResult {
  valid: boolean
  errors: string[]
}

/**
 * 创建智能体请求（简化版，只需名称和描述）
 */
export interface CreateAgentRequest {
  name: string
  description?: string
  // 以下字段可选，创建时可留空
  provider_id?: string
  model?: string
  system_prompt?: string
  working_dir?: string
  temperature?: number
  max_iterations?: number
  max_output_tokens?: number
  context_window_size?: number
  window_size?: number
  compression_ratio?: number
  context_max_tokens?: number
}

/**
 * 更新智能体请求
 */
export interface UpdateAgentRequest {
  name?: string
  description?: string
  provider_id?: string
  model?: string
  system_prompt?: string
  temperature?: number
  max_iterations?: number
  max_output_tokens?: number
  working_dir?: string
  context_window_size?: number
  window_size?: number
  compression_ratio?: number
  context_max_tokens?: number
  status?: AgentStatus
}