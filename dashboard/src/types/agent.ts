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
 * 外部智能体类型枚举
 */
export enum ExternalType {
  Dify = 'dify',
}

/**
 * 外部智能体类型显示名称映射
 */
export const ExternalTypeLabels: Record<ExternalType, string> = {
  [ExternalType.Dify]: 'Dify',
}

/**
 * 外部智能体类型选项
 */
export const ExternalTypeOptions = [
  { value: ExternalType.Dify, label: ExternalTypeLabels[ExternalType.Dify] },
]

/**
 * Dify 配置
 */
export interface DifyConfig {
  base_url: string
  api_key?: string // 创建/更新时传入
  api_key_masked?: string // 响应时返回
}

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
 * 技能简要信息
 */
export interface SkillBrief {
  id: string
  name: string
  description: string
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
  compression_ratio: number
  compression_keep_size: number
  context_max_tokens: number
  status: AgentStatus
  // 是否使用所有技能
  use_all_skills: boolean
  // 关联的技能列表
  skills?: SkillBrief[]
  // 是否使用所有工具
  use_all_tools: boolean
  // 关联的工具名称列表
  tool_names?: string[]
  // 是否可作为子智能体被调用
  callable: boolean
  // 作为子智能体时的工具描述
  callable_description: string
  // 是否启用 MCP 工具
  enable_mcp: boolean
  // 是否为外部智能体
  external: boolean
  // 外部智能体类型
  external_type?: ExternalType
  // 外部智能体配置
  external_config?: DifyConfig
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
  compression_ratio: number
  compression_keep_size: number
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
  compression_ratio?: number
  compression_keep_size?: number
  context_max_tokens?: number
  // 技能配置
  use_all_skills?: boolean
  skill_ids?: string[]
  // 工具配置
  use_all_tools?: boolean
  tool_names?: string[]
  // 子智能体配置
  callable?: boolean
  callable_description?: string
  // MCP 配置
  enable_mcp?: boolean
  // 外部智能体配置
  external?: boolean
  external_type?: ExternalType
  external_config?: DifyConfig
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
  compression_ratio?: number
  compression_keep_size?: number
  context_max_tokens?: number
  status?: AgentStatus
  // 技能配置
  use_all_skills?: boolean
  skill_ids?: string[]
  // 工具配置
  use_all_tools?: boolean
  tool_names?: string[]
  // 子智能体配置
  callable?: boolean
  callable_description?: string
  // MCP 配置
  enable_mcp?: boolean
  // 外部智能体配置（仅外部智能体可用）
  external_config?: DifyConfig
}

/**
 * 更新智能体技能配置请求
 */
export interface UpdateAgentSkillsRequest {
  use_all_skills: boolean
  skill_ids?: string[]
}

/**
 * 可用工具列表
 */
export const AvailableTools = [
  { name: 'list_dir', label: '列出目录' },
  { name: 'read_file', label: '读取文件' },
  { name: 'write_file', label: '写入文件' },
  { name: 'edit_file', label: '编辑文件' },
  { name: 'shell', label: '执行命令' },
  { name: 'todo_write', label: '任务管理' },
  { name: 'cron', label: '定时任务' },
] as const