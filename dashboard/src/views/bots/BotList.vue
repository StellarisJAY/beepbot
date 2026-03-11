<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  MessageOutlined,
  LinkOutlined,
  SearchOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import { botApi, type BotFilter } from '@/api/bot'
import { agentApi } from '@/api/agent'
import type { Bot, CreateBotRequest, UpdateBotRequest, BindAgentRequest } from '@/types/bot'
import { BotStatus, BotStatusLabels, BotPlatformLabels, BotStatusOptions, BotPlatformOptions } from '@/types/bot'
import type { Agent } from '@/types/agent'
import BotFormModal from '@/components/BotFormModal.vue'

// 数据
const bots = ref<Bot[]>([])
const agents = ref<Agent[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 筛选条件
const filters = ref<BotFilter>({
  name: '',
  status: undefined,
  platform: undefined,
})

// 弹窗相关
const formModalVisible = ref(false)
const bindModalVisible = ref(false)
const editingBot = ref<Bot | null>(null)
const selectedAgentId = ref<string | null>(null)
const bindingBotId = ref<string | null>(null)

// 获取机器人列表
const fetchBots = async () => {
  loading.value = true
  try {
    // 构建筛选参数，过滤掉空值
    const filterParams: BotFilter = {}
    if (filters.value.name) filterParams.name = filters.value.name
    if (filters.value.status) filterParams.status = filters.value.status
    if (filters.value.platform) filterParams.platform = filters.value.platform

    const res = await botApi.list(page.value, size.value, filterParams)
    bots.value = res.data
    total.value = res.total
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取机器人列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchBots()
}

// 重置筛选
const handleReset = () => {
  filters.value = { name: '', status: undefined, platform: undefined }
  page.value = 1
  fetchBots()
}

// 获取智能体列表（用于绑定）
const fetchAgents = async () => {
  try {
    const res = await agentApi.list(1, 100)
    agents.value = res.data
  } catch (error: unknown) {
    console.error('获取智能体列表失败:', error)
  }
}

// 新建机器人
const handleCreate = () => {
  editingBot.value = null
  formModalVisible.value = true
}

// 编辑机器人
const handleEdit = (bot: Bot) => {
  editingBot.value = bot
  formModalVisible.value = true
}

// 删除机器人
const handleDelete = (bot: Bot) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除机器人「${bot.name}」吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await botApi.delete(bot.id)
        message.success('删除成功')
        await fetchBots()
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || '删除失败')
      }
    },
  })
}

// 打开绑定智能体弹窗
const handleBindAgent = (bot: Bot) => {
  bindingBotId.value = bot.id
  selectedAgentId.value = bot.agent_id
  bindModalVisible.value = true
}

// 确认绑定智能体
const confirmBindAgent = async () => {
  if (!bindingBotId.value) return

  try {
    const data: BindAgentRequest = {
      agent_id: selectedAgentId.value,
    }
    await botApi.bindAgent(bindingBotId.value, data)
    message.success('绑定成功')
    bindModalVisible.value = false
    await fetchBots()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '绑定失败')
  }
}

// 表单提交成功
const handleFormSuccess = async () => {
  formModalVisible.value = false
  await fetchBots()
}

// 切换状态
const handleToggleStatus = async (bot: Bot) => {
  const newStatus = bot.status === BotStatus.Active ? BotStatus.Inactive : BotStatus.Active
  try {
    await botApi.updateStatus(bot.id, newStatus)
    message.success('状态更新成功')
    await fetchBots()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '状态更新失败')
  }
}

onMounted(() => {
  fetchBots()
  fetchAgents()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">IM机器人</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建机器人
      </a-button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <a-input
        v-model:value="filters.name"
        placeholder="搜索名称"
        style="width: 200px"
        allow-clear
        @pressEnter="handleSearch"
      >
        <template #prefix><SearchOutlined /></template>
      </a-input>
      <a-select
        v-model:value="filters.status"
        placeholder="状态"
        style="width: 120px"
        allow-clear
        :options="BotStatusOptions"
      />
      <a-select
        v-model:value="filters.platform"
        placeholder="平台"
        style="width: 120px"
        allow-clear
        :options="BotPlatformOptions"
      />
      <a-button type="primary" @click="handleSearch">
        <template #icon><SearchOutlined /></template>
        查询
      </a-button>
      <a-button @click="handleReset">
        <template #icon><ReloadOutlined /></template>
        重置
      </a-button>
    </div>

    <a-spin :spinning="loading">
      <div class="card-grid">
        <a-card v-for="bot in bots" :key="bot.id" class="bot-card" hoverable>
          <template #title>
            <div class="card-title">
              <MessageOutlined class="card-icon" />
              <span>{{ bot.name }}</span>
            </div>
          </template>
          <template #extra>
            <a-switch
              :checked="bot.status === BotStatus.Active"
              :checked-children="'启用'"
              :un-checked-children="'禁用'"
              @change="handleToggleStatus(bot)"
            />
          </template>
          <p class="card-description">{{ bot.description || '暂无描述' }}</p>
          <div class="card-meta">
            <div class="meta-item">
              <span class="meta-label">平台:</span>
              <a-tag color="blue">
                {{ BotPlatformLabels[bot.platform] || bot.platform }}
              </a-tag>
            </div>
            <div class="meta-item" v-if="bot.identifier">
              <span class="meta-label">标识符:</span>
              <span class="meta-value">{{ bot.identifier }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">绑定智能体:</span>
              <span v-if="bot.agent" class="meta-value linked">
                <LinkOutlined /> {{ bot.agent.name }}
              </span>
              <span v-else class="meta-value unbound">未绑定</span>
            </div>
          </div>
          <template #actions>
            <a-tooltip title="编辑">
              <EditOutlined @click="handleEdit(bot)" />
            </a-tooltip>
            <a-tooltip title="绑定智能体">
              <LinkOutlined @click="handleBindAgent(bot)" />
            </a-tooltip>
            <a-tooltip title="删除">
              <DeleteOutlined @click="handleDelete(bot)" />
            </a-tooltip>
          </template>
        </a-card>
      </div>
    </a-spin>

    <a-empty v-if="!loading && bots.length === 0" description="暂无机器人，点击上方按钮创建" />

    <!-- 分页 -->
    <div class="pagination-container" v-if="total > size">
      <a-pagination
        v-model:current="page"
        :total="total"
        :pageSize="size"
        show-less-items
        @change="fetchBots"
      />
    </div>

    <!-- 创建/编辑机器人弹窗 -->
    <BotFormModal
      v-model:visible="formModalVisible"
      :bot="editingBot"
      :agents="agents"
      @success="handleFormSuccess"
    />

    <!-- 绑定智能体弹窗 -->
    <a-modal
      v-model:open="bindModalVisible"
      title="绑定智能体"
      ok-text="确定"
      cancel-text="取消"
      @ok="confirmBindAgent"
    >
      <a-form layout="vertical">
        <a-form-item label="选择智能体">
          <a-select
            v-model:value="selectedAgentId"
            placeholder="请选择要绑定的智能体"
            allow-clear
            show-search
            :filter-option="(input: string, option: { label: string }) => option.label.toLowerCase().includes(input.toLowerCase())"
          >
            <a-select-option
              v-for="agent in agents"
              :key="agent.id"
              :value="agent.id"
              :label="agent.name"
            >
              {{ agent.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.page-container {
  padding: 24px;
  height: 100%;
  overflow-y: auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0;
}

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: var(--card-bg);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.bot-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  transition: all 0.3s ease;
}

.bot-card:hover {
  box-shadow: var(--card-hover-shadow);
  transform: translateY(-2px);
}

/* 卡片操作按钮样式 */
.bot-card :deep(.ant-card-actions) {
  border-top: 1px solid var(--border-color);
}

.bot-card :deep(.ant-card-actions > li) {
  color: var(--card-action-color);
  transition: color 0.3s ease;
}

.bot-card :deep(.ant-card-actions > li:hover) {
  color: var(--card-action-hover-color);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.card-icon {
  font-size: 18px;
  color: var(--color-primary);
}

.card-description {
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.6;
  margin-bottom: 12px;
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 13px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-label {
  color: var(--text-tertiary);
}

.meta-value {
  color: var(--text-secondary);
  font-family: monospace;
}

.meta-value.linked {
  color: var(--color-primary);
}

.meta-value.unbound {
  color: var(--text-tertiary);
  font-style: italic;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}
</style>