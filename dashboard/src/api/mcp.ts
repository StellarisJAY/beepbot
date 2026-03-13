import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type { MCPServer, CreateMCPServerRequest, UpdateMCPServerRequest, MCPServerFilter, MCPTool } from '@/types/mcp'

/**
 * MCP 服务器 API 服务
 */
export const mcpApi = {
  /**
   * 获取 MCP 服务器列表（分页，支持筛选）
   */
  list(page: number = 1, size: number = 10, filters?: MCPServerFilter): Promise<PageResponse<MCPServer>> {
    const params: Record<string, unknown> = { page, size }
    if (filters) {
      if (filters.name) params.name = filters.name
      if (filters.status) params.status = filters.status
    }
    return http.get<MCPServer[]>('/mcp', { params }) as Promise<PageResponse<MCPServer>>
  },

  /**
   * 获取单个 MCP 服务器
   */
  get(id: string): Promise<ApiResponse<MCPServer>> {
    return http.get<MCPServer>(`/mcp/${id}`)
  },

  /**
   * 创建 MCP 服务器
   */
  create(data: CreateMCPServerRequest): Promise<ApiResponse<MCPServer>> {
    return http.post<MCPServer>('/mcp', data)
  },

  /**
   * 更新 MCP 服务器
   */
  update(id: string, data: UpdateMCPServerRequest): Promise<ApiResponse<MCPServer>> {
    return http.put<MCPServer>(`/mcp/${id}`, data)
  },

  /**
   * 删除 MCP 服务器
   */
  delete(id: string): Promise<ApiResponse<void>> {
    return http.delete<void>(`/mcp/${id}`)
  },

  /**
   * 启动 MCP 服务器连接
   */
  start(id: string): Promise<ApiResponse<void>> {
    return http.put<void>(`/mcp/${id}/start`)
  },

  /**
   * 停止 MCP 服务器连接
   */
  stop(id: string): Promise<ApiResponse<void>> {
    return http.put<void>(`/mcp/${id}/stop`)
  },

  /**
   * 测试 MCP 服务器连接
   */
  testConnection(id: string): Promise<ApiResponse<{ message: string }>> {
    return http.post<{ message: string }>(`/mcp/${id}/test`)
  },

  /**
   * 获取 MCP 服务器提供的工具列表
   */
  getTools(id: string): Promise<ApiResponse<MCPTool[]>> {
    return http.get<MCPTool[]>(`/mcp/${id}/tools`)
  },

  /**
   * 重连 MCP 服务器
   */
  reconnect(id: string): Promise<ApiResponse<void>> {
    return http.post<void>(`/mcp/${id}/reconnect`)
  },
}

export default mcpApi