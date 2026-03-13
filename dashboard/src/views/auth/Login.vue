<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { message } from 'ant-design-vue'
import { UserOutlined, LockOutlined, RobotOutlined } from '@ant-design/icons-vue'

// 导入 SVG 图标
import openaiIcon from '@/assets/icons/openai.svg'
import anthropicIcon from '@/assets/icons/anthropic.svg'
import dashscopeIcon from '@/assets/icons/dashscope.svg'
import ollamaIcon from '@/assets/icons/ollama.svg'
import deepseekIcon from '@/assets/icons/deepseek.svg'
import qqIcon from '@/assets/icons/qq.svg'
import feishuIcon from '@/assets/icons/feishu.svg'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({
  username: '',
  password: '',
})

const loading = ref(false)

// 背景图标配置
// 从左边缘出发的图标：left 为负值，top 分布在不同高度
// 从下边缘出发的图标：top 大于 100%，left 分布在不同宽度
const bgIcons = [
  // 左边缘出发
  { src: openaiIcon, size: 48, duration: 28, delay: 0, top: '20%', left: '-60px' },
  { src: anthropicIcon, size: 42, duration: 32, delay: 4, top: '50%', left: '-50px' },
  { src: dashscopeIcon, size: 36, duration: 25, delay: 8, top: '80%', left: '-40px' },
  { src: ollamaIcon, size: 44, duration: 30, delay: 12, top: '35%', left: '-55px' },
  { src: deepseekIcon, size: 40, duration: 26, delay: 16, top: '65%', left: '-45px' },
  // 下边缘出发
  { src: qqIcon, size: 38, duration: 24, delay: 2, top: 'calc(100% + 50px)', left: '10%' },
  { src: feishuIcon, size: 34, duration: 29, delay: 6, top: 'calc(100% + 40px)', left: '40%' },
  { src: openaiIcon, size: 32, duration: 27, delay: 10, top: 'calc(100% + 55px)', left: '70%' },
  { src: anthropicIcon, size: 46, duration: 31, delay: 14, top: 'calc(100% + 45px)', left: '25%' },
  { src: dashscopeIcon, size: 38, duration: 23, delay: 18, top: 'calc(100% + 60px)', left: '55%' },
  { src: ollamaIcon, size: 36, duration: 33, delay: 20, top: 'calc(100% + 35px)', left: '85%' },
  { src: deepseekIcon, size: 44, duration: 28, delay: 22, top: '5%', left: '-70px' },
]

async function handleLogin() {
  if (!form.value.username || !form.value.password) {
    message.error('请输入用户名和密码')
    return
  }

  loading.value = true
  try {
    await authStore.login(form.value.username, form.value.password)
    message.success('登录成功')
    router.push('/')
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <!-- 动态背景图标 -->
    <div class="bg-icons">
      <img
        v-for="(item, index) in bgIcons"
        :key="index"
        :src="item.src"
        class="bg-icon"
        :style="{
          top: item.top,
          left: item.left,
          width: `${item.size}px`,
          height: `${item.size}px`,
          animationDuration: `${item.duration}s`,
          animationDelay: `${item.delay}s`,
        }"
        alt=""
      />
    </div>

    <!-- 登录卡片 -->
    <div class="login-card">
      <div class="logo-section">
        <RobotOutlined class="logo-icon" />
        <h1 class="title">BeepBot</h1>
      </div>
      <p class="subtitle">AI 智能代理机器人管理平台</p>

      <a-form :model="form" @finish="handleLogin">
        <a-form-item name="username" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input
            v-model:value="form.username"
            size="large"
            placeholder="用户名"
            autocomplete="username"
          >
            <template #prefix>
              <UserOutlined class="input-icon" />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item name="password" :rules="[{ required: true, message: '请输入密码' }]">
          <a-input-password
            v-model:value="form.password"
            size="large"
            placeholder="密码"
            autocomplete="current-password"
            @press-enter="handleLogin"
          >
            <template #prefix>
              <LockOutlined class="input-icon" />
            </template>
          </a-input-password>
        </a-form-item>

        <a-form-item>
          <a-button type="primary" size="large" html-type="submit" block :loading="loading">
            登录
          </a-button>
        </a-form-item>
      </a-form>

      <p class="hint">默认用户名和密码：admin / admin</p>
    </div>
  </div>
</template>

<style scoped>
.login-container {
  position: relative;
  min-height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 50%, #d5dbe3 100%);
  overflow: hidden;
}

/* 背景图标容器 */
.bg-icons {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  overflow: hidden;
}

/* 单个背景图标 */
.bg-icon {
  position: absolute;
  opacity: 0.2;
  animation: slide-diagonal linear infinite;
  will-change: transform;
}

/* 斜向滑动动画 - 从左下到右上 */
@keyframes slide-diagonal {
  0% {
    transform: translate(0, 0) rotate(0deg);
    opacity: 0;
  }
  5% {
    opacity: 0.2;
  }
  95% {
    opacity: 0.2;
  }
  100% {
    transform: translate(calc(100vw + 100px), calc(-100vh - 100px)) rotate(360deg);
    opacity: 0;
  }
}

/* 登录卡片 */
.login-card {
  position: relative;
  z-index: 10;
  width: 400px;
  padding: 48px 40px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(4px);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
}

/* Logo 区域 */
.logo-section {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 8px;
}

.logo-icon {
  font-size: 36px;
  color: #667eea;
}

.title {
  text-align: center;
  font-size: 28px;
  font-weight: 700;
  color: #1a1a2e;
  margin: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.subtitle {
  text-align: center;
  color: #666;
  margin-bottom: 32px;
  font-size: 14px;
}

/* 输入框图标 */
.input-icon {
  color: #999;
}

/* 提示文字 */
.hint {
  text-align: center;
  color: #999;
  font-size: 12px;
  margin-top: 16px;
}
</style>