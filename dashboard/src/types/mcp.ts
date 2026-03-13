/**
 * MCP 传输类型枚举
 */
export enum MCPTransportType {
  SSE = 'sse',
  Http = 'http',
  Stdio = 'stdio',
}

/**
 * MCP 服务器状态枚举
 */
export enum MCPServerStatus {
  Active = 'active',
  Inactive = 'inactive',
}

/**
 * MCP 传输类型显示名称映射
 */
export const MCPTransportTypeLabels: Record<MCPTransportType, string> = {
  [MCPTransportType.SSE]: 'SSE',
  [MCPTransportType.Http]: 'HTTP',
  [MCPTransportType.Stdio]: 'Stdio',
}

/**
 * MCP 服务器状态显示名称映射
 */
export const MCPServerStatusLabels: Record<MCPServerStatus, string> = {
  [MCPServerStatus.Active]: '已连接',
  [MCPServerStatus.Inactive]: '未连接',
}

/**
 * MCP 工具信息
 */
export interface MCPTool {
  name: string
  description: string
  inputSchema: Record<string, unknown>
}

/**
 * MCP 服务器配置
 */
export interface MCPServer {
  id: string
  name: string
  description: string
  transport_type: MCPTransportType
  url?: string
  command?: string
  args?: string[]
  env?: Record<string, string>
  headers?: Record<string, string>
  status: MCPServerStatus
  tools?: MCPTool[]
  created_at: string
  updated_at: string
}

/**
 * 创建 MCP 服务器请求
 */
export interface CreateMCPServerRequest {
  name: string
  description?: string
  transport_type: MCPTransportType
  url?: string
  command?: string
  args?: string[]
  env?: Record<string, string>
  headers?: Record<string, string>
}

/**
 * 更新 MCP 服务器请求
 */
export interface UpdateMCPServerRequest {
  name?: string
  description?: string
  transport_type?: MCPTransportType
  url?: string
  command?: string
  args?: string[]
  env?: Record<string, string>
  headers?: Record<string, string>
}

/**
 * MCP 服务器筛选参数
 */
export interface MCPServerFilter {
  name?: string
  status?: MCPServerStatus
}