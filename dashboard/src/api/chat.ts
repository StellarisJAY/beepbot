import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type { ChatSession, ChatEvent, MessagesResponse } from '@/types/session'

// Chat API 请求参数
export interface ChatRequest {
  agent_id: string
  message: string
  session_id?: string
}

export interface ChatSessionsRequest {
  agent_id: string
  page?: number
  size?: number
}

export interface ChatMessagesRequest {
  session_id: string
  limit?: number
  before_id?: string
}

/**
 * SSE 聊天回调接口
 */
export interface ChatCallbacks {
  onSessionId?: (sessionId: string) => void
  onMessage?: (content: string) => void
  onThinking?: (content: string) => void
  onToolCall?: (toolInfo: string) => void
  onToolResult?: (result: string) => void
  onToolError?: (error: string) => void
  onError?: (error: string) => void
  onDone?: () => void
}

/**
 * 聊天 API 服务
 */
export const chatApi = {
  /**
   * 流式聊天
   * 使用 fetch + SSE 实现流式响应
   */
  async chat(
    agentId: string,
    message: string,
    sessionId: string | null,
    callbacks: ChatCallbacks,
    token: string,
  ): Promise<void> {
    const baseUrl = import.meta.env.VITE_API_BASE_URL
    const url = `${baseUrl}/chat`

    const body: ChatRequest = {
      agent_id: agentId,
      message,
    }
    if (sessionId) {
      body.session_id = sessionId
    }

    try {
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(body),
      })

      if (!response.ok) {
        const errorText = await response.text()
        callbacks.onError?.(errorText || `HTTP ${response.status}`)
        return
      }

      const reader = response.body?.getReader()
      if (!reader) {
        callbacks.onError?.('No response body')
        return
      }

      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        // 解析 SSE 事件
        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const data = line.slice(6)
              const event: ChatEvent = JSON.parse(data)

              switch (event.type) {
                case 'session_id':
                  callbacks.onSessionId?.(event.content)
                  break
                case 'message':
                  callbacks.onMessage?.(event.content)
                  break
                case 'thinking':
                  callbacks.onThinking?.(event.content)
                  break
                case 'tool_call':
                  callbacks.onToolCall?.(event.content)
                  break
                case 'tool_result':
                  callbacks.onToolResult?.(event.content)
                  break
                case 'tool_error':
                  callbacks.onToolError?.(event.content)
                  break
                case 'error':
                  callbacks.onError?.(event.content)
                  break
                case 'done':
                  callbacks.onDone?.()
                  break
              }
            } catch {
              // 忽略解析错误
            }
          }
        }
      }

      callbacks.onDone?.()
    } catch (error) {
      callbacks.onError?.(error instanceof Error ? error.message : 'Unknown error')
    }
  },

  /**
   * 获取前端聊天会话列表
   */
  async getSessions(agentId: string, page = 1, size = 10): Promise<PageResponse<ChatSession>> {
    const response = await http.post<ChatSession[]>('/chat/sessions', {
      agent_id: agentId,
      page,
      size,
    })
    return response as PageResponse<ChatSession>
  },

  /**
   * 获取会话消息
   */
  async getMessages(sessionId: string, limit = 20, beforeId?: string): Promise<MessagesResponse> {
    const body: ChatMessagesRequest = {
      session_id: sessionId,
      limit,
    }
    if (beforeId) {
      body.before_id = beforeId
    }
    const response = await http.post<MessagesResponse>('/chat/messages', body)
    return response.data
  },
}

export default chatApi