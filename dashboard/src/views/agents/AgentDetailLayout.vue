<script setup lang="ts">
import { ref, onMounted, computed, provide } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  RobotOutlined,
  EditOutlined,
  FileTextOutlined,
  DashboardOutlined,
  SettingOutlined,
  SaveOutlined,
  MessageOutlined,
  PlusOutlined,
  MenuFoldOutlined,
} from '@ant-design/icons-vue'
import { agentApi } from '@/api/agent'
import type { Agent } from '@/types/agent'

const route = useRoute()
const router = useRouter()

const agentId = computed(() => route.params.id as string)

// 数据
const agent = ref<Agent | null>(null)
const loading = ref(false)

// 子组件引用
const componentRef = ref<{ handleSave?: () => Promise<void>; saving?: boolean; newChat?: () => void; toggleSessionDrawer?: () => void } | null>(null)

// 编辑基本信息弹窗
const editInfoModalVisible = ref(false)
const editForm = ref({
  name: '',
  description: '',
})

// 当前路由名称
const currentRoute = computed(() => {
  const path = route.path
  if (path.endsWith('/edit')) return 'edit'
  if (path.endsWith('/logs')) return 'logs'
  if (path.endsWith('/monitor')) return 'monitor'
  if (path.endsWith('/chat')) return 'chat'
  return 'edit'
})

// 获取智能体详情
const fetchAgent = async () => {
  loading.value = true
  try {
    const res = await agentApi.get(agentId.value)
    agent.value = res.data
    editForm.value = {
      name: res.data.name,
      description: res.data.description,
    }
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取智能体信息失败')
    router.push('/agents')
  } finally {
    loading.value = false
  }
}

// 打开编辑基本信息弹窗
const openEditInfoModal = () => {
  editForm.value = {
    name: agent.value?.name || '',
    description: agent.value?.description || '',
  }
  editInfoModalVisible.value = true
}

// 保存基本信息
const saveBasicInfo = async () => {
  if (!editForm.value.name?.trim()) {
    message.warning('名称不能为空')
    return
  }
  try {
    await agentApi.update(agentId.value, {
      name: editForm.value.name,
      description: editForm.value.description,
    })
    message.success('保存成功')
    await fetchAgent()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    editInfoModalVisible.value = false
  }
}

// 返回列表
const handleBack = () => {
  router.push('/agents')
}

// 导航
const navigateTo = (nav: string) => {
  router.push(`/agents/${agentId.value}/${nav}`)
}

// 保存配置
const handleSave = () => {
  if (componentRef.value?.handleSave) {
    componentRef.value.handleSave()
  }
}

// 聊天页面相关状态
const sessionDrawerVisible = ref(false)

// 新对话
const handleNewChat = () => {
  if (componentRef.value?.newChat) {
    componentRef.value.newChat()
  }
}

// 切换会话列表抽屉
const toggleSessionDrawer = () => {
  sessionDrawerVisible.value = !sessionDrawerVisible.value
  if (componentRef.value?.toggleSessionDrawer) {
    componentRef.value.toggleSessionDrawer()
  }
}

// 提供给子组件的数据
provide('agent', agent)
provide('agentId', agentId)
provide('fetchAgent', fetchAgent)

onMounted(() => {
  fetchAgent()
})
</script>

<template>
  <div class="detail-layout">
    <!-- 左侧导航栏 - 固定不变 -->
    <div class="side-nav">
      <!-- 上部分：基本信息 -->
      <div class="nav-header">
        <div class="agent-avatar">
          <RobotOutlined class="avatar-icon" />
        </div>
        <div class="agent-info">
          <div class="agent-name-row">
            <span class="agent-name">{{ agent?.name || '加载中...' }}</span>
            <a-button type="text" size="small" @click="openEditInfoModal">
              <template #icon><EditOutlined /></template>
            </a-button>
          </div>
          <p class="agent-desc">{{ agent?.description || '暂无描述' }}</p>
        </div>
      </div>

      <!-- 下部分：导航菜单 -->
      <div class="nav-menu">
        <div class="nav-divider" />
        <div
          class="nav-item"
          :class="{ active: currentRoute === 'chat' }"
          @click="navigateTo('chat')"
        >
          <MessageOutlined class="nav-icon" />
          <span>聊天</span>
        </div>
        <div
          class="nav-item"
          :class="{ active: currentRoute === 'edit' }"
          @click="navigateTo('edit')"
        >
          <SettingOutlined class="nav-icon" />
          <span>配置</span>
        </div>
        <div
          class="nav-item"
          :class="{ active: currentRoute === 'logs' }"
          @click="navigateTo('logs')"
        >
          <FileTextOutlined class="nav-icon" />
          <span>日志</span>
        </div>
        <div
          class="nav-item"
          :class="{ active: currentRoute === 'monitor' }"
          @click="navigateTo('monitor')"
        >
          <DashboardOutlined class="nav-icon" />
          <span>使用监测</span>
        </div>
      </div>
    </div>

    <!-- 中间区域 -->
    <div class="main-area">
      <!-- Header - 固定 -->
      <div class="content-header">
        <div class="header-left">
          <a-button @click="handleBack">
            返回列表
          </a-button>
        </div>
        <div class="header-right">
          <!-- 聊天页面显示新对话和会话列表按钮 -->
          <template v-if="currentRoute === 'chat'">
            <a-button type="primary" @click="handleNewChat">
              <template #icon><PlusOutlined /></template>
              新对话
            </a-button>
            <a-button @click="toggleSessionDrawer">
              <template #icon><MenuFoldOutlined /></template>
              会话列表
            </a-button>
          </template>
          <!-- 配置页面显示保存按钮 -->
          <a-button
            v-else-if="currentRoute === 'edit'"
            type="primary"
            :loading="componentRef?.saving"
            @click="handleSave"
          >
            <template #icon><SaveOutlined /></template>
            保存
          </a-button>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="content-body" :class="{ 'no-scroll': currentRoute === 'chat' }">
        <a-spin :spinning="loading">
          <router-view v-slot="{ Component }">
            <component :is="Component" ref="componentRef" />
          </router-view>
        </a-spin>
      </div>
    </div>

    <!-- 编辑基本信息弹窗 -->
    <a-modal
      v-model:open="editInfoModalVisible"
      title="编辑基本信息"
      @ok="saveBasicInfo"
    >
      <a-form :label-col="{ span: 4 }" :wrapper-col="{ span: 20 }">
        <a-form-item label="名称" required>
          <a-input
            v-model:value="editForm.name"
            placeholder="请输入名称"
            :maxlength="128"
          />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea
            v-model:value="editForm.description"
            placeholder="请输入描述"
            :rows="3"
            :maxlength="500"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.detail-layout {
  height: 100%;
  display: flex;
  overflow: hidden;
  background: var(--bg-color);
}

/* 左侧导航栏 - 固定不变 */
.side-nav {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--card-bg);
  border-right: 1px solid var(--border-color);
}

.nav-header {
  padding: 24px 16px;
  border-bottom: 1px solid var(--border-color);
}

.agent-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.avatar-icon {
  font-size: 32px;
  color: #fff;
}

.agent-info {
  text-align: center;
}

.agent-name-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.agent-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color);
}

.agent-desc {
  margin-top: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.nav-menu {
  flex: 1;
  padding: 8px;
}

.nav-divider {
  height: 1px;
  background: var(--border-color);
  margin: 8px 0 16px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  cursor: pointer;
  color: var(--text-color);
  transition: all 0.2s ease;
  margin-bottom: 4px;
}

.nav-item:hover {
  background: var(--hover-bg);
}

.nav-item.active {
  background: var(--color-primary);
  color: #fff;
}

.nav-item.active .nav-icon {
  color: #fff;
}

.nav-icon {
  font-size: 16px;
}

/* 中间区域 */
.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Header - 固定 */
.content-header {
  flex-shrink: 0;
  height: 56px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  background: var(--card-bg);
  border-bottom: 1px solid var(--border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* 内容区域 - 可滚动 */
.content-body {
  flex: 1;
  overflow-y: auto;
}

.content-body.no-scroll {
  overflow: hidden;
}
</style>