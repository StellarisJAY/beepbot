<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ClockCircleOutlined,
  SearchOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import { cronApi, type CronFilter } from '@/api/cron'
import { agentApi } from '@/api/agent'
import type { CronJob } from '@/types/cron'
import { CronJobStatusOptions } from '@/types/cron'
import type { Agent } from '@/types/agent'

// 快捷 Cron 表达式选项（6字段格式：秒 分 时 日 月 周）
const cronPresets = [
  { label: '每分钟', expression: '0 * * * * *', description: '每分钟执行一次' },
  { label: '每小时', expression: '0 0 * * * *', description: '每小时整点执行' },
  { label: '每天', expression: '0 0 9 * * *', description: '每天9点执行' },
  { label: '每天中午', expression: '0 0 12 * * *', description: '每天中午12点执行' },
  { label: '每天晚上', expression: '0 0 18 * * *', description: '每天晚上6点执行' },
  { label: '每周一', expression: '0 0 9 * * 1', description: '每周一9点执行' },
  { label: '每月1号', expression: '0 0 9 1 * *', description: '每月1号9点执行' },
]

// 应用快捷表达式
const applyPreset = (expression: string) => {
  formData.value.cron_expression = expression
}

// 数据
const cronJobs = ref<CronJob[]>([])
const agents = ref<Agent[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 筛选条件
const filters = ref<CronFilter>({
  name: '',
  enabled: undefined,
})

// 弹窗
const modalVisible = ref(false)
const modalLoading = ref(false)
const isEdit = ref(false)
const editingId = ref('')

// 表单数据
const formData = ref({
  name: '',
  cron_expression: '',
  agent_id: '',
  message: '',
  enabled: true,
})

// 获取定时任务列表
const fetchCronJobs = async () => {
  loading.value = true
  try {
    // 构建筛选参数，过滤掉空值
    const filterParams: CronFilter = {}
    if (filters.value.name) filterParams.name = filters.value.name
    if (filters.value.enabled !== undefined) filterParams.enabled = filters.value.enabled

    const res = await cronApi.list(page.value, size.value, filterParams)
    cronJobs.value = res.data
    total.value = res.total
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取定时任务列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchCronJobs()
}

// 重置筛选
const handleReset = () => {
  filters.value = { name: '', enabled: undefined }
  page.value = 1
  fetchCronJobs()
}

// 获取智能体列表（用于下拉选择）
const fetchAgents = async () => {
  try {
    const res = await agentApi.list(1, 100)
    agents.value = res.data
  } catch (error: unknown) {
    console.error('获取智能体列表失败', error)
  }
}

// 打开创建弹窗
const handleCreate = () => {
  isEdit.value = false
  editingId.value = ''
  formData.value = {
    name: '',
    cron_expression: '',
    agent_id: '',
    message: '',
    enabled: true,
  }
  modalVisible.value = true
}

// 打开编辑弹窗
const handleEdit = (job: CronJob) => {
  isEdit.value = true
  editingId.value = job.id
  formData.value = {
    name: job.name,
    cron_expression: job.cron_expression,
    agent_id: job.agent_id,
    message: job.message,
    enabled: job.enabled,
  }
  modalVisible.value = true
}

// 提交表单
const handleSubmit = async () => {
  // 表单验证
  if (!formData.value.name.trim()) {
    message.warning('请输入任务名称')
    return
  }
  if (!formData.value.cron_expression.trim()) {
    message.warning('请输入 Cron 表达式')
    return
  }
  if (!formData.value.agent_id) {
    message.warning('请选择智能体')
    return
  }
  if (!formData.value.message.trim()) {
    message.warning('请输入发送消息')
    return
  }

  modalLoading.value = true
  try {
    if (isEdit.value) {
      await cronApi.update(editingId.value, {
        name: formData.value.name,
        cron_expression: formData.value.cron_expression,
        agent_id: formData.value.agent_id,
        message: formData.value.message,
        enabled: formData.value.enabled,
      })
      message.success('更新成功')
    } else {
      await cronApi.create({
        name: formData.value.name,
        cron_expression: formData.value.cron_expression,
        agent_id: formData.value.agent_id,
        message: formData.value.message,
        enabled: formData.value.enabled,
      })
      message.success('创建成功')
    }
    modalVisible.value = false
    await fetchCronJobs()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || (isEdit.value ? '更新失败' : '创建失败'))
  } finally {
    modalLoading.value = false
  }
}

// 删除定时任务
const handleDelete = (job: CronJob) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除定时任务「${job.name}」吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await cronApi.delete(job.id)
        message.success('删除成功')
        await fetchCronJobs()
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || '删除失败')
      }
    },
  })
}

// 切换状态
const handleToggleStatus = async (job: CronJob) => {
  const newStatus = !job.enabled
  try {
    await cronApi.updateStatus(job.id, newStatus)
    message.success('状态更新成功')
    await fetchCronJobs()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '状态更新失败')
  }
}

// 分页变化
const handlePageChange = (p: number, s: number) => {
  page.value = p
  size.value = s
  fetchCronJobs()
}

// 获取智能体名称
const getAgentName = (agentId: string) => {
  const agent = agents.value.find((a) => a.id === agentId)
  return agent?.name || agentId
}

onMounted(() => {
  fetchCronJobs()
  fetchAgents()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">定时任务</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建任务
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
        v-model:value="filters.enabled"
        placeholder="状态"
        style="width: 120px"
        allow-clear
        :options="CronJobStatusOptions"
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
      <div class="card-grid" v-if="cronJobs.length > 0">
        <a-card v-for="job in cronJobs" :key="job.id" class="cron-card" hoverable>
          <template #title>
            <div class="card-title">
              <ClockCircleOutlined class="card-icon" />
              <span>{{ job.name }}</span>
            </div>
          </template>
          <template #extra>
            <a-switch
              :checked="job.enabled"
              checked-children="启用"
              un-checked-children="禁用"
              @change="handleToggleStatus(job)"
            />
          </template>
          <div class="card-content">
            <div class="card-meta">
              <div class="meta-item">
                <span class="meta-label">Cron 表达式:</span>
                <span class="meta-value code">{{ job.cron_expression }}</span>
              </div>
            </div>
            <div class="card-meta">
              <div class="meta-item">
                <span class="meta-label">绑定智能体:</span>
                <span class="meta-value">{{ job.agent?.name || getAgentName(job.agent_id) }}</span>
              </div>
            </div>
            <div class="card-meta">
              <div class="meta-item">
                <span class="meta-label">发送消息:</span>
                <span class="meta-value message-text">{{ job.message }}</span>
              </div>
            </div>
          </div>
          <template #actions>
            <a-tooltip title="编辑">
              <EditOutlined @click="handleEdit(job)" />
            </a-tooltip>
            <a-tooltip title="删除">
              <DeleteOutlined @click="handleDelete(job)" />
            </a-tooltip>
          </template>
        </a-card>
      </div>

      <a-empty v-else description="暂无定时任务，点击上方按钮创建" />

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

    <!-- 创建/编辑弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="isEdit ? '编辑定时任务' : '新建定时任务'"
      :confirm-loading="modalLoading"
      @ok="handleSubmit"
    >
      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="任务名称" required>
          <a-input v-model:value="formData.name" placeholder="请输入任务名称" />
        </a-form-item>
        <a-form-item label="Cron 表达式" required>
          <a-input
            v-model:value="formData.cron_expression"
            placeholder="例如: 0 9 * * * (每天9点)"
          />
          <div class="form-hint">格式: 分 时 日 月 周 (例如: 0 9 * * * 表示每天9点执行)</div>
          <div class="cron-presets">
            <span class="presets-label">快捷选择：</span>
            <a-tag
              v-for="preset in cronPresets"
              :key="preset.expression"
              class="preset-tag"
              @click="applyPreset(preset.expression)"
            >
              {{ preset.label }}
            </a-tag>
          </div>
        </a-form-item>
        <a-form-item label="绑定智能体" required>
          <a-select v-model:value="formData.agent_id" placeholder="请选择智能体">
            <a-select-option v-for="agent in agents" :key="agent.id" :value="agent.id">
              {{ agent.name }}
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="发送消息" required>
          <a-textarea
            v-model:value="formData.message"
            placeholder="请输入向智能体发送的消息"
            :rows="4"
          />
        </a-form-item>
        <a-form-item label="启用状态">
          <a-switch
            v-model:checked="formData.enabled"
            checked-children="启用"
            un-checked-children="禁用"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.page-container {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 600;
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
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.cron-card {
  border-radius: 8px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-icon {
  font-size: 18px;
  color: #1890ff;
}

.card-content {
  min-height: 80px;
}

.card-description {
  color: #666;
  margin-bottom: 12px;
}

.card-meta {
  margin-bottom: 8px;
}

.meta-item {
  display: flex;
  align-items: flex-start;
}

.meta-label {
  color: #999;
  margin-right: 8px;
  white-space: nowrap;
}

.meta-value {
  color: #333;
}

.meta-value.code {
  font-family: monospace;
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
}

.meta-value.message-text {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}

.form-hint {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.cron-presets {
  margin-top: 8px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}

.presets-label {
  font-size: 12px;
  color: #666;
  margin-right: 4px;
}

.preset-tag {
  cursor: pointer;
  margin: 0;
  transition: all 0.2s;
}

.preset-tag:hover {
  background-color: #1890ff;
  color: #fff;
  border-color: #1890ff;
}
</style>