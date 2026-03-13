<script setup lang="ts">
import { ref } from 'vue'
import {
  UserOutlined,
  RobotOutlined,
  SendOutlined,
  DeleteOutlined,
} from '@ant-design/icons-vue'

// 会话列表抽屉
const sessionDrawerVisible = ref(false)

// 暴露给父组件的方法
const newChat = () => {
  messages.value = []
  messages.value.push({
    id: 'welcome',
    role: 'assistant',
    content: '你好！我是智能体助手，有什么可以帮助你的吗？',
    time: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
  })
}

const toggleSessionDrawer = () => {
  sessionDrawerVisible.value = !sessionDrawerVisible.value
}

defineExpose({
  newChat,
  toggleSessionDrawer,
})

// 模拟会话列表数据
const sessions = ref([
  { id: '1', title: '如何使用 Go 进行并发编程', time: '2024-01-15 14:30', active: true },
  { id: '2', title: '帮我写一个 Python 脚本', time: '2024-01-15 10:20', active: false },
  { id: '3', title: '解释一下 React Hooks', time: '2024-01-14 16:45', active: false },
  { id: '4', title: '数据库查询优化建议', time: '2024-01-13 09:15', active: false },
])

// 模拟消息列表
const messages = ref([
  {
    id: '1',
    role: 'user',
    content: '你好，请帮我解释一下 Go 语言的并发模型。',
    time: '14:30',
  },
  {
    id: '2',
    role: 'assistant',
    content: `Go 语言的并发模型基于 **CSP (Communicating Sequential Processes)** 理论，核心概念包括：

## Goroutine
Goroutine 是 Go 中的轻量级线程，启动一个非常简单：
\`\`\`go
go func() {
    // 并发执行的代码
}()
\`\`\`

## Channel
Channel 是 goroutine 之间通信的管道：
\`\`\`go
ch := make(chan int)
go func() { ch <- 42 }()  // 发送
value := <-ch              // 接收
\`\`\`

## Select
Select 语句用于处理多个 channel 操作：
\`\`\`go
select {
case msg := <-ch1:
    fmt.Println("from ch1:", msg)
case msg := <-ch2:
    fmt.Println("from ch2:", msg)
}
\`\`\`

这种模型提倡"**不要通过共享内存来通信，而要通过通信来共享内存**"的设计理念。`,
    time: '14:31',
  },
  {
    id: '3',
    role: 'user',
    content: '能给我一个完整的示例吗？',
    time: '14:32',
  },
  {
    id: '4',
    role: 'assistant',
    content: `当然！这是一个完整的工作池示例：

\`\`\`go
package main

import (
    "fmt"
    "time"
)

// Worker 处理任务
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
        fmt.Printf("worker %d: started job %d\\n", id, j)
        time.Sleep(time.Second) // 模拟耗时操作
        fmt.Printf("worker %d: finished job %d\\n", id, j)
        results <- j * 2
    }
}

func main() {
    const numJobs = 5
    jobs := make(chan int, numJobs)
    results := make(chan int, numJobs)

    // 启动 3 个 worker
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }

    // 发送 5 个任务
    for j := 1; j <= numJobs; j++ {
        jobs <- j
    }
    close(jobs)

    // 收集结果
    for a := 1; a <= numJobs; a++ {
        <-results
    }
}
\`\`\`

这个示例展示了：
- 启动多个 goroutine 并发处理任务
- 使用 channel 进行任务分发和结果收集
- 优雅地处理并发工作流`,
    time: '14:33',
  },
])

// 输入框内容
const inputValue = ref('')

// 发送消息
const sendMessage = () => {
  if (!inputValue.value.trim()) return

  messages.value.push({
    id: Date.now().toString(),
    role: 'user',
    content: inputValue.value,
    time: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
  })

  inputValue.value = ''

  setTimeout(() => {
    messages.value.push({
      id: (Date.now() + 1).toString(),
      role: 'assistant',
      content: '这是一个模拟的 AI 回复。实际功能需要连接后端服务。',
      time: new Date().toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }),
    })
  }, 500)
}

// 选择会话
const selectSession = (session: { id: string }) => {
  sessions.value.forEach((s) => (s.active = s.id === session.id))
  sessionDrawerVisible.value = false
}

// 删除会话
const deleteSession = (sessionId: string) => {
  sessions.value = sessions.value.filter((s) => s.id !== sessionId)
}
</script>

<template>
  <div class="chat-container">
    <!-- 消息区域 - 整个页面滚动 -->
    <div class="messages-area">
      <div class="messages-wrapper">
        <div
          v-for="message in messages"
          :key="message.id"
          class="message-item"
          :class="message.role"
        >
          <div class="message-avatar">
            <UserOutlined v-if="message.role === 'user'" />
            <RobotOutlined v-else />
          </div>
          <div class="message-content">
            <div class="message-header">
              <span class="message-role">
                {{ message.role === 'user' ? '我' : '智能体' }}
              </span>
              <span class="message-time">{{ message.time }}</span>
            </div>
            <div class="message-body" v-html="message.content"></div>
          </div>
        </div>
      </div>
      <!-- 底部留白，给输入框腾空间 -->
      <div class="bottom-spacer"></div>
    </div>

    <!-- 输入区域 - 固定悬浮在底部 -->
    <div class="input-area">
      <div class="input-container">
        <div class="input-box">
          <a-textarea
            v-model:value="inputValue"
            placeholder="输入消息进行调试..."
            :auto-size="{ minRows: 1, maxRows: 4 }"
            @press-enter="sendMessage"
          />
          <a-button type="primary" class="send-btn" @click="sendMessage">
            <template #icon><SendOutlined /></template>
          </a-button>
        </div>
        <div class="input-hint">
          按 Enter 发送，Shift + Enter 换行
        </div>
      </div>
    </div>

    <!-- 会话列表抽屉 -->
    <div class="session-drawer" :class="{ visible: sessionDrawerVisible }">
      <div class="drawer-header">
        <span>历史会话</span>
      </div>
      <div class="drawer-body">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="session-item"
          :class="{ active: session.active }"
          @click="selectSession(session)"
        >
          <div class="session-info">
            <div class="session-title">{{ session.title }}</div>
            <div class="session-time">{{ session.time }}</div>
          </div>
          <a-button
            type="text"
            size="small"
            danger
            class="delete-btn"
            @click.stop="deleteSession(session.id)"
          >
            <template #icon><DeleteOutlined /></template>
          </a-button>
        </div>
      </div>
    </div>

    <!-- 抽屉遮罩 -->
    <div
      v-if="sessionDrawerVisible"
      class="drawer-mask"
      @click="sessionDrawerVisible = false"
    ></div>
  </div>
</template>

<style scoped>
.chat-container {
  height: 100%;
  display: flex;
  position: relative;
  overflow: hidden;
}

/* 消息区域 - 整个页面滚动 */
.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  padding-bottom: 0;
}

.messages-wrapper {
  max-width: 800px;
  margin: 0 auto;
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

/* 底部留白 */
.bottom-spacer {
  height: 100px;
}

/* 输入区域 - 固定悬浮在底部 */
.input-area {
  position: absolute;
  left: 0;
  right: 320px;
  bottom: 0;
  padding: 16px 20px 20px;
  background: linear-gradient(to top, var(--bg-color) 80%, transparent);
  pointer-events: none;
  transition: right 0.3s ease;
}

.input-container {
  max-width: 800px;
  margin: 0 auto;
  pointer-events: auto;
}

.input-box {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  background: var(--card-bg);
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

/* 会话列表抽屉 */
.session-drawer {
  position: absolute;
  right: -320px;
  top: 0;
  bottom: 0;
  width: 320px;
  background: var(--card-bg);
  border-left: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  transition: right 0.3s ease;
  z-index: 10;
}

.session-drawer.visible {
  right: 0;
}

.drawer-header {
  flex-shrink: 0;
  height: 56px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  border-bottom: 1px solid var(--border-color);
  font-size: 15px;
  font-weight: 500;
  color: var(--text-color);
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
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

.delete-btn {
  opacity: 0;
  transition: opacity 0.2s;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

/* 抽屉遮罩 */
.drawer-mask {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  right: 320px;
  background: rgba(0, 0, 0, 0.3);
  z-index: 5;
}
</style>