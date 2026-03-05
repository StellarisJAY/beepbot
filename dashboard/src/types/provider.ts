/**
 * 供应商类型枚举
 */
export enum ProviderType {
  OpenAI = 'openai',
  DashScope = 'dashscope',
  Ollama = 'ollama',
}

/**
 * 供应商类型显示名称映射
 */
export const ProviderTypeLabels: Record<ProviderType, string> = {
  [ProviderType.OpenAI]: 'OpenAI',
  [ProviderType.DashScope]: 'DashScope (阿里云)',
  [ProviderType.Ollama]: 'Ollama',
}

/**
 * 供应商类型选项（用于下拉选择）
 */
export const ProviderTypeOptions = [
  { value: ProviderType.OpenAI, label: ProviderTypeLabels[ProviderType.OpenAI] },
  { value: ProviderType.DashScope, label: ProviderTypeLabels[ProviderType.DashScope] },
  { value: ProviderType.Ollama, label: ProviderTypeLabels[ProviderType.Ollama] },
]

/**
 * 供应商默认 Base URL
 */
export const ProviderDefaultBaseURL: Record<ProviderType, string> = {
  [ProviderType.OpenAI]: 'https://api.openai.com/v1',
  [ProviderType.DashScope]: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
  [ProviderType.Ollama]: 'http://localhost:11434/v1',
};

/**
 * 供应商配置
 */
export interface Provider {
  id: string
  name: string
  provider_type: ProviderType
  api_key_masked: string
  base_url: string
  extra_config?: Record<string, unknown>
  is_default: boolean
  created_at: string
  updated_at: string
}

/**
 * 创建供应商请求
 */
export interface CreateProviderRequest {
  name: string
  provider_type: ProviderType
  api_key: string
  base_url?: string
  extra_config?: Record<string, unknown>
  is_default?: boolean
}

/**
 * 更新供应商请求
 */
export interface UpdateProviderRequest {
  name?: string
  api_key?: string
  base_url?: string
  extra_config?: Record<string, unknown>
  is_default?: boolean
}