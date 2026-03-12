// 会话类型
export type SessionType = 'chat' | 'cron'

// 会话类型选项（用于下拉选择）
export const SessionTypeOptions = [
  { value: 'chat', label: '聊天' },
  { value: 'cron', label: '定时任务' },
]

// 会话列表项
export interface SessionListItem {
  id: string
  key: string
  bot_id: string
  bot_name: string
  bot_platform: string
  session_type: SessionType // 会话类型：chat/cron
  summary: string
  message_count: number
  total_tokens: number
  last_context_tokens: number // 当前上下文 token 数量
  created_at: string
  updated_at: string
}

// 工具调用函数
export interface ToolCallFunction {
  name: string
  arguments: string
}

// 工具调用
export interface ToolCall {
  id: string
  type: string
  function?: ToolCallFunction
  name?: string
  arguments?: Record<string, unknown>
}

// 消息列表项
export interface MessageListItem {
  id: string
  role: 'user' | 'assistant' | 'tool' | 'system'
  content: string
  tool_calls?: ToolCall[]
  tool_call_id?: string
  input_tokens?: number
  output_tokens?: number
  total_tokens?: number
  in_window: boolean
  created_at: string
}

// 消息列表响应
export interface MessagesResponse {
  messages: MessageListItem[]
  total: number
  has_more: boolean
}

// 压缩会话请求
export interface CompressSessionRequest {
  session_id: string
}

// 压缩会话响应
export interface CompressSessionResponse {
  success: boolean
  message: string
}

// 用量统计数据点
export interface UsageStatsPoint {
  time: string
  session_count: number
  message_count: number
  input_tokens: number
  output_tokens: number
  total_tokens: number
}

// 用量统计响应
export interface UsageStatsResponse {
  points: UsageStatsPoint[]
}

// 时间范围选项
export const PeriodOptions = [
  { value: '1d', label: '最近1天' },
  { value: '3d', label: '最近3天' },
  { value: '7d', label: '最近7天' },
  { value: '14d', label: '最近14天' },
  { value: '30d', label: '最近30天' },
]