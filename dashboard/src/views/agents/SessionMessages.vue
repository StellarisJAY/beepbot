<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Card, Button, Spin, Empty, Tag, Typography, Tooltip } from 'ant-design-vue'
import { ArrowLeftOutlined, ReloadOutlined, ToolOutlined, UserOutlined, RobotOutlined, CodeOutlined } from '@ant-design/icons-vue'
import { sessionApi } from '@/api/session'
import type { MessageListItem, ToolCall } from '@/types/session'
import MarkdownRenderer from '@/components/MarkdownRenderer.vue'

const route = useRoute()
const router = useRouter()

// 状态
const messages = ref<MessageListItem[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const hasMore = ref(false)
const total = ref(0)
const sessionId = computed(() => route.params.sessionId as string)
const agentId = computed(() => route.params.id as string)

// 消息容器引用
const messagesContainer = ref<HTMLElement | null>(null)

// 加载消息
async function loadMessages(beforeId?: string) {
  if (!sessionId.value) return

  try {
    if (beforeId) {
      loadingMore.value = true
    } else {
      loading.value = true
    }

    const response = await sessionApi.getSessionMessages(sessionId.value, beforeId, 20)
    
    if (beforeId) {
      // 加载更多，将新消息插入到前面
      messages.value = [...response.messages, ...messages.value]
    } else {
      // 首次加载
      messages.value = response.messages
    }
    
    hasMore.value = response.has_more
    total.value = response.total
  } catch (error) {
    console.error('Failed to load messages:', error)
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

// 加载更多消息
async function loadMore() {
  if (loadingMore.value || !hasMore.value || messages.value.length === 0) return
  
  // 获取最早的消息ID
  const firstMessage = messages.value[0]
  if (firstMessage) {
    await loadMessages(firstMessage.id)
  }
}

// 刷新消息
async function refresh() {
  messages.value = []
  await loadMessages()
  scrollToBottom()
}

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 返回会话列表
function goBack() {
  router.push({ name: 'agent-logs', params: { id: agentId.value } })
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

// 获取角色标签
function getRoleLabel(role: string) {
  const labels: Record<string, string> = {
    user: '用户',
    assistant: '助手',
    tool: '工具',
    system: '系统',
  }
  return labels[role] || role
}

// 获取角色颜色
function getRoleColor(role: string) {
  const colors: Record<string, string> = {
    user: 'blue',
    assistant: 'green',
    tool: 'orange',
    system: 'default',
  }
  return colors[role] || 'default'
}

// 滚动处理
function handleScroll(event: Event) {
  const target = event.target as HTMLElement
  // 当滚动到顶部时加载更多
  if (target.scrollTop < 50 && hasMore.value && !loadingMore.value) {
    // 保存当前滚动位置
    const oldScrollHeight = target.scrollHeight
    loadMore().then(() => {
      // 恢复滚动位置
      nextTick(() => {
        const newScrollHeight = target.scrollHeight
        target.scrollTop = newScrollHeight - oldScrollHeight
      })
    })
  }
}

// 初始化
onMounted(async () => {
  await loadMessages()
  scrollToBottom()
})
</script>

<template>
  <div class="session-messages">
    <Card :bordered="false" class="messages-card">
      <template #title>
        <div class="card-header">
          <div class="header-left">
            <Button type="text" @click="goBack">
              <template #icon><ArrowLeftOutlined /></template>
            </Button>
            <span class="title">会话消息</span>
            <Typography.Text type="secondary" class="total-count">
              共 {{ total }} 条消息
            </Typography.Text>
          </div>
          <div class="header-right">
            <Button type="text" @click="refresh" :loading="loading">
              <template #icon><ReloadOutlined /></template>
              刷新
            </Button>
          </div>
        </div>
      </template>

      <Spin :spinning="loading">
        <div
          ref="messagesContainer"
          class="messages-container"
          @scroll="handleScroll"
        >
          <!-- 加载更多提示 -->
          <div v-if="hasMore" class="load-more">
            <Button type="link" @click="loadMore" :loading="loadingMore">
              加载更多消息
            </Button>
          </div>

          <!-- 消息列表 -->
          <div v-if="messages.length > 0" class="message-list">
            <div
              v-for="message in messages"
              :key="message.id"
              class="message-item"
              :class="[`message-${message.role}`]"
            >
              <!-- 消息头部 -->
              <div class="message-header">
                <Tag :color="getRoleColor(message.role)" class="role-tag">
                  <template #icon>
                    <UserOutlined v-if="message.role === 'user'" />
                    <RobotOutlined v-else-if="message.role === 'assistant'" />
                    <ToolOutlined v-else-if="message.role === 'tool'" />
                    <CodeOutlined v-else />
                  </template>
                  {{ getRoleLabel(message.role) }}
                </Tag>
                <span class="message-time">{{ formatTime(message.created_at) }}</span>
                <span v-if="message.total_tokens" class="message-tokens">
                  {{ message.total_tokens }} tokens
                </span>
              </div>

              <!-- 消息内容 -->
              <div class="message-content">
                <!-- 工具调用展示 -->
                <div v-if="message.tool_calls && message.tool_calls.length > 0" class="tool-calls">
                  <div v-for="toolCall in message.tool_calls" :key="toolCall.id" class="tool-call-item">
                    <Tooltip :title="toolCall.function?.arguments || ''">
                      <Tag color="processing">
                        <ToolOutlined />
                        {{ toolCall.function?.name || toolCall.name || '未知工具' }}
                      </Tag>
                    </Tooltip>
                  </div>
                </div>

                <!-- 文本内容 -->
                <div v-if="message.content" class="message-text">
                  <MarkdownRenderer :content="message.content" />
                </div>

                <!-- 工具响应 -->
                <div v-if="message.role === 'tool' && message.tool_call_id" class="tool-response">
                  <Typography.Text type="secondary">
                    工具响应 ID: {{ message.tool_call_id }}
                  </Typography.Text>
                </div>
              </div>
            </div>
          </div>

          <!-- 空状态 -->
          <Empty v-else-if="!loading" description="暂无消息" />
        </div>
      </Spin>
    </Card>
  </div>
</template>

<style scoped>
.session-messages {
  height: 100%;
}

.messages-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.messages-card :deep(.ant-card-body) {
  flex: 1;
  overflow: hidden;
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title {
  font-size: 16px;
  font-weight: 500;
}

.total-count {
  font-size: 12px;
}

.messages-container {
  height: calc(100vh - 280px);
  overflow-y: auto;
  padding: 16px;
}

.load-more {
  text-align: center;
  padding: 8px 0;
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-item {
  padding: 12px;
  border-radius: 8px;
  background-color: var(--component-bg, #fafafa);
  border: 1px solid var(--border-color, #f0f0f0);
}

.message-user {
  background-color: var(--user-message-bg, #e6f7ff);
  border-color: var(--user-message-border, #91d5ff);
}

.message-assistant {
  background-color: var(--assistant-message-bg, #f6ffed);
  border-color: var(--assistant-message-border, #b7eb8f);
}

.message-tool {
  background-color: var(--tool-message-bg, #fff7e6);
  border-color: var(--tool-message-border, #ffd591);
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.role-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.message-time {
  font-size: 12px;
  color: var(--text-color-secondary, #999);
}

.message-tokens {
  font-size: 12px;
  color: var(--text-color-secondary, #999);
  margin-left: auto;
}

.message-content {
  overflow-x: auto;
}

.tool-calls {
  margin-bottom: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.tool-call-item {
  display: inline-flex;
}

.message-text {
  line-height: 1.6;
}

.tool-response {
  margin-top: 4px;
}

/* 深色模式适配 */
:global(.dark) .message-item {
  background-color: var(--component-bg, #1f1f1f);
  border-color: var(--border-color, #303030);
}

:global(.dark) .message-user {
  background-color: var(--user-message-bg, #111d2c);
  border-color: var(--user-message-border, #15395b);
}

:global(.dark) .message-assistant {
  background-color: var(--assistant-message-bg, #162312);
  border-color: var(--assistant-message-border, #274916);
}

:global(.dark) .message-tool {
  background-color: var(--tool-message-bg, #2b2111);
  border-color: var(--tool-message-border, #594214);
}
</style>