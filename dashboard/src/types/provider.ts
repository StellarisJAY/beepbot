import OpenAIIcon from '@/assets/icons/openai.svg'
import DashScopeIcon from '@/assets/icons/dashscope.svg'
import OllamaIcon from '@/assets/icons/ollama.svg'
import AnthropicIcon from '@/assets/icons/anthropic.svg'
import DeepSeekIcon from '@/assets/icons/deepseek.svg'

/**
 * 供应商类型枚举
 */
export enum ProviderType {
  OpenAI = 'openai',
  DashScope = 'dashscope',
  Ollama = 'ollama',
  Anthropic = 'anthropic',
  DeepSeek = 'deepseek',
}

/**
 * 供应商类型图标映射
 */
export const ProviderTypeIcons: Record<ProviderType, string> = {
  [ProviderType.OpenAI]: OpenAIIcon,
  [ProviderType.DashScope]: DashScopeIcon,
  [ProviderType.Ollama]: OllamaIcon,
  [ProviderType.Anthropic]: AnthropicIcon,
  [ProviderType.DeepSeek]: DeepSeekIcon,
}

/**
 * 供应商类型显示名称映射
 */
export const ProviderTypeLabels: Record<ProviderType, string> = {
  [ProviderType.OpenAI]: 'OpenAI',
  [ProviderType.DashScope]: 'DashScope (阿里云)',
  [ProviderType.Ollama]: 'Ollama',
  [ProviderType.Anthropic]: 'Anthropic',
  [ProviderType.DeepSeek]: 'DeepSeek',
}

/**
 * 供应商类型选项（用于下拉选择）
 */
export const ProviderTypeOptions = [
  { value: ProviderType.OpenAI, label: ProviderTypeLabels[ProviderType.OpenAI], icon: ProviderTypeIcons[ProviderType.OpenAI] },
  { value: ProviderType.DashScope, label: ProviderTypeLabels[ProviderType.DashScope], icon: ProviderTypeIcons[ProviderType.DashScope] },
  { value: ProviderType.Ollama, label: ProviderTypeLabels[ProviderType.Ollama], icon: ProviderTypeIcons[ProviderType.Ollama] },
  { value: ProviderType.Anthropic, label: ProviderTypeLabels[ProviderType.Anthropic], icon: ProviderTypeIcons[ProviderType.Anthropic] },
  { value: ProviderType.DeepSeek, label: ProviderTypeLabels[ProviderType.DeepSeek], icon: ProviderTypeIcons[ProviderType.DeepSeek] },
]

/**
 * 供应商默认 Base URL
 */
export const ProviderDefaultBaseURL: Record<ProviderType, string> = {
  [ProviderType.OpenAI]: 'https://api.openai.com/v1',
  [ProviderType.DashScope]: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
  [ProviderType.Ollama]: 'http://localhost:11434/v1',
  [ProviderType.Anthropic]: 'https://api.anthropic.com',
  [ProviderType.DeepSeek]: 'https://api.deepseek.com/v1',
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

/**
 * 供应商筛选参数
 */
export interface ProviderFilter {
  name?: string
  provider_type?: ProviderType
}