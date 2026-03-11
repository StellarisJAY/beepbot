<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  RobotOutlined,
  FileTextOutlined,
  LineChartOutlined,
  SearchOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import { agentApi, type AgentFilter } from '@/api/agent'
import type { Agent } from '@/types/agent'
import { AgentStatus, AgentStatusOptions } from '@/types/agent'
import AgentCreateModal from '@/components/AgentCreateModal.vue'

const router = useRouter()

// 数据
const agents = ref<Agent[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 筛选条件
const filters = ref<AgentFilter>({
  name: '',
  status: undefined,
})

// 弹窗
const createModalVisible = ref(false)

// 获取智能体列表
const fetchAgents = async () => {
  loading.value = true
  try {
    // 构建筛选参数，过滤掉空值
    const filterParams: AgentFilter = {}
    if (filters.value.name) filterParams.name = filters.value.name
    if (filters.value.status) filterParams.status = filters.value.status

    const res = await agentApi.list(page.value, size.value, filterParams)
    agents.value = res.data
    total.value = res.total
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取智能体列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchAgents()
}

// 重置筛选
const handleReset = () => {
  filters.value = { name: '', status: undefined }
  page.value = 1
  fetchAgents()
}

// 新建智能体
const handleCreate = () => {
  createModalVisible.value = true
}

// 创建成功后跳转编辑页
const handleCreateSuccess = (id: string) => {
  router.push(`/agents/${id}/edit`)
}

// 编辑智能体
const handleEdit = (id: string) => {
  router.push(`/agents/${id}/edit`)
}

// 查看日志
const handleLogs = (id: string) => {
  router.push(`/agents/${id}/logs`)
}

// 查看监测
const handleMonitor = (id: string) => {
  router.push(`/agents/${id}/monitor`)
}

// 删除智能体
const handleDelete = (agent: Agent) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除智能体「${agent.name}」吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await agentApi.delete(agent.id)
        message.success('删除成功')
        await fetchAgents()
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || '删除失败')
      }
    },
  })
}

// 切换状态
const handleToggleStatus = async (agent: Agent) => {
  // 如果要启用，先校验配置是否完整
  if (agent.status !== AgentStatus.Active) {
    try {
      const validateRes = await agentApi.validate(agent.id)
      if (!validateRes.data.valid) {
        message.warning(`配置不完整: ${validateRes.data.errors.join(', ')}`)
        return
      }
    } catch (error: unknown) {
      const err = error as { message?: string }
      message.error(err.message || '校验失败')
      return
    }
  }

  const newStatus = agent.status === AgentStatus.Active ? AgentStatus.Inactive : AgentStatus.Active
  try {
    await agentApi.updateStatus(agent.id, newStatus)
    message.success('状态更新成功')
    await fetchAgents()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '状态更新失败')
  }
}

// 分页变化
const handlePageChange = (p: number, s: number) => {
  page.value = p
  size.value = s
  fetchAgents()
}

onMounted(() => {
  fetchAgents()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">智能体</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建智能体
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
        :options="AgentStatusOptions"
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
      <div class="card-grid" v-if="agents.length > 0">
        <a-card v-for="agent in agents" :key="agent.id" class="agent-card" hoverable>
          <template #title>
            <div class="card-title">
              <RobotOutlined class="card-icon" />
              <span>{{ agent.name }}</span>
            </div>
          </template>
          <template #extra>
            <a-switch
              :checked="agent.status === AgentStatus.Active"
              checked-children="启用"
              un-checked-children="禁用"
              @change="handleToggleStatus(agent)"
            />
          </template>
          <p class="card-description">{{ agent.description || '暂无描述' }}</p>
          <div class="card-meta">
            <div class="meta-item">
              <span class="meta-label">模型:</span>
              <span class="meta-value">{{ agent.model || '未配置' }}</span>
            </div>
          </div>
          <template #actions>
            <a-tooltip title="编辑">
              <EditOutlined @click="handleEdit(agent.id)" />
            </a-tooltip>
            <a-tooltip title="日志">
              <FileTextOutlined @click="handleLogs(agent.id)" />
            </a-tooltip>
            <a-tooltip title="监测">
              <LineChartOutlined @click="handleMonitor(agent.id)" />
            </a-tooltip>
            <a-tooltip title="删除">
              <DeleteOutlined @click="handleDelete(agent)" />
            </a-tooltip>
          </template>
        </a-card>
      </div>

      <a-empty v-else description="暂无智能体，点击上方按钮创建" />

      <!-- 分页 -->
      <div class="pagination-container" v-if="total > size">
        <a-pagination
          v-model:current="page"
          v-model:pageSize="size"
          :total="total"
          :show-size-changer="true"
          :show-quick-jumper="true"
          :show-total="(t: number) => `共 ${t} 条`"
          @change="handlePageChange"
        />
      </div>
    </a-spin>

    <!-- 创建弹窗 -->
    <AgentCreateModal
      v-model:visible="createModalVisible"
      @success="handleCreateSuccess"
    />
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

.agent-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  transition: all 0.3s ease;
}

.agent-card:hover {
  box-shadow: var(--card-hover-shadow);
  transform: translateY(-2px);
}

/* 卡片操作按钮样式 */
.agent-card :deep(.ant-card-actions) {
  border-top: 1px solid var(--border-color);
}

.agent-card :deep(.ant-card-actions > li) {
  color: var(--card-action-color);
  transition: color 0.3s ease;
}

.agent-card :deep(.ant-card-actions > li:hover) {
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
  min-height: 44px;
}

.card-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 12px;
}

.meta-item {
  display: flex;
  gap: 4px;
}

.meta-label {
  color: var(--text-tertiary);
}

.meta-value {
  color: var(--text-secondary);
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}
</style>