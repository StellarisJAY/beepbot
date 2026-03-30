<script setup lang="ts">
import { ref, onMounted, nextTick, inject, type ComputedRef } from 'vue'
import {
  UserOutlined,
  RobotOutlined,
  SendOutlined,
  LoadingOutlined,
  ToolOutlined,
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import { chatApi } from '@/api/chat'
import type { ChatSession, ChatMessage } from '@/types/session'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

// 从父组件注入（agentId 是 ComputedRef）
const agentIdRef = inject<ComputedRef<string>>('agentId')

// 获取 agentId 值
function getAgentId(): string {
  return agentIdRef?.value || ''
}

// 状态
const sessions = ref<ChatSession[]>([])
const currentSessionId = ref<string | null>(null)
const messages = ref<ChatMessage[]>([])
const inputValue = ref('')
const loading = ref(false)
const streaming = ref(false)
const streamingContent = ref('')
const sessionDrawerVisible = ref(false)

// 消息容器引用
const messagesContainer = ref<HTMLElement | null>(null)

// 获取 token
function getToken(): string {
  return localStorage.getItem('beepbot_token') || ''
}

// 发送消息
async function sendMessage(e?: KeyboardEvent) {
  if (e) {
    e.preventDefault() // 阻止默认换行行为
  }
  if (!inputValue.value.trim() || streaming.value) return
  const id = getAgentId()
  if (!id) {
    message.error('智能体 ID 不存在')
    return
  }

  const userMessage = inputValue.value.trim()
  inputValue.value = ''
  await nextTick()

  // 添加用户消息
  const userMsg: ChatMessage = {
    id: `temp-${Date.now()}`,
    role: 'user',
    content: userMessage,
    created_at: new Date().toISOString(),
  }
  messages.value.push(userMsg)
  scrollToBottom()

  streaming.value = true
  streamingContent.value = ''

  try {
    await chatApi.chat(
      id,
      userMessage,
      currentSessionId.value,
      {
        onSessionId: (sessionId) => {
          currentSessionId.value = sessionId
        },
        onMessage: (content) => {
          streamingContent.value += content
          scrollToBottom()
        },
        onThinking: (content) => {
          // 可以显示思考过程
          console.log('Thinking:', content)
        },
        onToolCall: (toolInfo) => {
          // 显示工具调用
          const toolMsg: ChatMessage = {
            id: `tool-${Date.now()}`,
            role: 'tool',
            content: `🔧 ${toolInfo}`,
            created_at: new Date().toISOString(),
          }
          messages.value.push(toolMsg)
          scrollToBottom()
        },
        onError: (error) => {
          message.error(error)
        },
        onDone: () => {
          // 保存最终消息
          if (streamingContent.value) {
            const assistantMsg: ChatMessage = {
              id: `assistant-${Date.now()}`,
              role: 'assistant',
              content: streamingContent.value,
              created_at: new Date().toISOString(),
            }
            messages.value.push(assistantMsg)
          }
          streamingContent.value = ''
          streaming.value = false
          scrollToBottom()
        },
      },
      getToken(),
    )
  } catch (error) {
    streaming.value = false
    message.error('发送消息失败')
  }
}

// 新对话
function newChat() {
  currentSessionId.value = null
  messages.value = []
}

// 切换会话
async function selectSession(session: ChatSession) {
  currentSessionId.value = session.id
  sessionDrawerVisible.value = false
  await loadSessionMessages()
}

// 加载会话消息
async function loadSessionMessages() {
  if (!currentSessionId.value) return

  loading.value = true
  try {
    const response = await chatApi.getMessages(currentSessionId.value, 50)
    messages.value = response.messages.map((msg) => ({
      id: msg.id,
      role: msg.role,
      content: msg.content,
      tool_calls: msg.tool_calls,
      tool_call_id: msg.tool_call_id,
      created_at: msg.created_at,
    }))
    scrollToBottom()
  } catch (error) {
    message.error('加载消息失败')
  } finally {
    loading.value = false
  }
}

// 加载会话列表
async function loadSessions() {
  const id = getAgentId()
  if (!id) return

  try {
    const response = await chatApi.getSessions(id, 1, 20)
    sessions.value = response.data || []
  } catch (error) {
    console.error('加载会话列表失败:', error)
  }
}

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 格式化时间
function formatTime(dateStr: string) {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// 切换会话列表抽屉
function toggleSessionDrawer() {
  sessionDrawerVisible.value = !sessionDrawerVisible.value
}

// 暴露给父组件
defineExpose({
  newChat,
  toggleSessionDrawer,
})

// 初始化
onMounted(async () => {
  await loadSessions()
})
</script>

<template>
  <div class="chat-container">
    <!-- 消息区域 -->
    <div class="messages-area">
      <div ref="messagesContainer" class="messages-wrapper">
        <!-- 欢迎消息 -->
        <div v-if="messages.length === 0 && !streaming" class="welcome-message">
          <RobotOutlined class="welcome-icon" />
          <p>开始与智能体对话</p>
        </div>

        <!-- 消息列表 -->
        <div v-for="msg in messages" :key="msg.id" class="message-item" :class="msg.role">
          <div class="message-avatar">
            <UserOutlined v-if="msg.role === 'user'" />
            <RobotOutlined v-else-if="msg.role === 'assistant'" />
            <ToolOutlined v-else />
          </div>
          <div class="message-content">
            <div class="message-header">
              <span class="message-role">
                {{ msg.role === 'user' ? '我' : msg.role === 'assistant' ? '智能体' : '工具' }}
              </span>
              <span class="message-time">{{ formatTime(msg.created_at) }}</span>
            </div>
            <div class="message-body">
              <MarkdownRenderer v-if="msg.role === 'assistant'" :content="msg.content" />
              <template v-else>{{ msg.content }}</template>
            </div>
          </div>
        </div>

        <!-- 流式输出中的消息 -->
        <div v-if="streaming && streamingContent" class="message-item assistant streaming">
          <div class="message-avatar">
            <RobotOutlined />
          </div>
          <div class="message-content">
            <div class="message-header">
              <span class="message-role">智能体</span>
              <LoadingOutlined class="loading-icon" />
            </div>
            <div class="message-body">
              <MarkdownRenderer :content="streamingContent" />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 输入区域 -->
    <div class="input-area">
      <div class="input-container">
        <div class="input-box">
          <a-textarea
            v-model:value="inputValue"
            placeholder="输入消息..."
            :auto-size="{ minRows: 1, maxRows: 4 }"
            :disabled="streaming"
            @press-enter="sendMessage"
          />
          <a-button
            type="primary"
            class="send-btn"
            :loading="streaming"
            :disabled="!inputValue.trim()"
            @click="sendMessage"
          >
            <template #icon><SendOutlined /></template>
          </a-button>
        </div>
        <div class="input-hint">按 Enter 发送，Shift + Enter 换行</div>
      </div>
    </div>

    <!-- 会话列表抽屉 -->
    <a-drawer
      v-model:open="sessionDrawerVisible"
      title="历史会话"
      placement="right"
      :width="320"
    >
      <div v-if="sessions.length === 0" class="empty-sessions">暂无历史会话</div>
      <div
        v-for="session in sessions"
        :key="session.id"
        class="session-item"
        :class="{ active: session.id === currentSessionId }"
        @click="selectSession(session)"
      >
        <div class="session-info">
          <div class="session-title">
            {{ session.summary || '新对话' }}
          </div>
          <div class="session-time">{{ formatTime(session.updated_at) }}</div>
        </div>
      </div>
    </a-drawer>
  </div>
</template>

<style scoped>
.chat-container {
  height: 100%;
  display: flex;
  overflow: hidden;
  flex-direction: column;
}

.messages-area {
  flex: 1;
  height: 85%;
  overflow-y: auto;
  padding: 20px;
  padding-bottom: 120px;
}

.messages-wrapper {
  max-width: 800px;
  margin: 0 auto;
}

.welcome-message {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-secondary);
}

.welcome-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.message-item {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}

.message-item.user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 16px;
}

.message-item.user .message-avatar {
  background: var(--color-primary);
  color: #fff;
}

.message-item.assistant .message-avatar {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.message-item.tool .message-avatar {
  background: #faad14;
  color: #fff;
}

.message-content {
  max-width: 70%;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.message-item.user .message-header {
  flex-direction: row-reverse;
}

.message-role {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-color);
}

.message-time {
  font-size: 12px;
  color: var(--text-secondary);
}

.message-body {
  padding: 12px 16px;
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
}

.message-item.user .message-body {
  background: var(--color-primary);
  color: #fff;
  border-bottom-right-radius: 4px;
}

.message-item.assistant .message-body {
  background: var(--card-bg);
  color: var(--text-color);
  border-bottom-left-radius: 4px;
  border: 1px solid var(--border-color);
}

.message-item.tool .message-body {
  background: #fffbe6;
  color: #595959;
  border-bottom-left-radius: 4px;
  border: 1px solid #ffe58f;
  font-size: 13px;
}

.loading-icon {
  color: var(--color-primary);
}

.input-area {
  display: flex;
  justify-content: center;
  padding: 0 20px;
  z-index: 100;
  height: 200px;
}

.input-container {
  width: 100%;
  max-width: 800px;
  pointer-events: auto;
}

.input-box {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 16px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08), 0 2px 6px rgba(0, 0, 0, 0.04);
  border: 1px solid var(--border-color);
  transition: box-shadow 0.2s, border-color 0.2s;
}

.input-box:focus-within {
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.12), 0 3px 8px rgba(0, 0, 0, 0.06);
  border-color: var(--color-primary);
}

.input-box :deep(.ant-input) {
  resize: none;
  border: none;
  box-shadow: none !important;
  padding: 0;
  font-size: 15px;
}

.input-box :deep(.ant-input:focus) {
  box-shadow: none !important;
}

/* 禁用状态下保持透明白色背景 */
.input-box :deep(.ant-input-disabled) {
  background: transparent !important;
  color: var(--text-color) !important;
}

.input-box :deep(.ant-input[disabled]) {
  background: transparent !important;
  color: var(--text-color) !important;
}

.send-btn {
  flex-shrink: 0;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.input-hint {
  margin-top: 8px;
  font-size: 12px;
  color: var(--text-secondary);
  text-align: center;
}

/* 会话列表样式 */
.empty-sessions {
  text-align: center;
  padding: 40px 20px;
  color: var(--text-secondary);
}

.session-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 4px;
}

.session-item:hover {
  background: var(--hover-bg);
}

.session-item.active {
  background: rgba(24, 144, 255, 0.1);
  border: 1px solid var(--color-primary);
}

.session-info {
  flex: 1;
  overflow: hidden;
}

.session-title {
  font-size: 14px;
  color: var(--text-color);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-bottom: 4px;
}

.session-time {
  font-size: 12px;
  color: var(--text-secondary);
}

/* 深色模式适配 */
:global(.dark) .message-item.assistant .message-body {
  background: var(--component-bg, #1f1f1f);
  border-color: var(--border-color, #303030);
}

:global(.dark) .message-item.tool .message-body {
  background: #2b2111;
  border-color: #594214;
}

:global(.dark) .input-box {
  background: rgba(30, 30, 30, 0.95);
}
</style>