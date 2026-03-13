<script setup lang="ts">
import { ref, onMounted, h, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  StopOutlined,
  ExclamationCircleOutlined,
  SearchOutlined,
  ReloadOutlined,
  ApiOutlined,
  ToolOutlined,
  LinkOutlined,
} from '@ant-design/icons-vue'
import { mcpApi } from '@/api/mcp'
import {
  type MCPServer,
  type CreateMCPServerRequest,
  type UpdateMCPServerRequest,
  type MCPServerFilter,
  MCPTransportType,
  MCPTransportTypeLabels,
  MCPServerStatus,
  MCPServerStatusLabels,
} from '@/types/mcp'

// MCP 服务器列表
const servers = ref<MCPServer[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 筛选条件
const filters = ref<MCPServerFilter>({
  name: '',
  status: undefined,
})

// 弹窗相关
const modalVisible = ref(false)
const modalLoading = ref(false)
const isEdit = ref(false)
const currentServer = ref<MCPServer | null>(null)

// 工具列表弹窗
const toolsModalVisible = ref(false)
const toolsLoading = ref(false)
const currentTools = ref<{ name: string; description: string }[]>([])

// 表单数据
const formData = ref<CreateMCPServerRequest>({
  name: '',
  description: '',
  transport_type: MCPTransportType.SSE,
  url: '',
  headers: {},
})

// 动态 Headers
const headerKeys = ref<string[]>([])
const headerValues = ref<string[]>([])

// 表单引用
const formRef = ref()

// 表单校验规则
const formRules = computed(() => ({
  name: [
    { required: true, message: '请输入服务器名称', trigger: 'blur' },
    { min: 2, max: 64, message: '名称长度为 2-64 个字符', trigger: 'blur' },
  ],
  transport_type: [{ required: true, message: '请选择传输类型', trigger: 'change' }],
  url: [
    {
      required: formData.value.transport_type === MCPTransportType.SSE,
      message: '请输入服务器 URL',
      trigger: 'blur',
    },
    { type: 'url', message: '请输入有效的 URL', trigger: 'blur' },
  ],
}))

// 获取服务器列表
const fetchServers = async () => {
  loading.value = true
  try {
    const filterParams: MCPServerFilter = {}
    if (filters.value.name) filterParams.name = filters.value.name
    if (filters.value.status) filterParams.status = filters.value.status

    const res = await mcpApi.list(page.value, size.value, filterParams)
    servers.value = res.data
    total.value = res.total
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取服务器列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchServers()
}

// 重置筛选
const handleReset = () => {
  filters.value = { name: '', status: undefined }
  page.value = 1
  fetchServers()
}

// 打开创建弹窗
const handleCreate = () => {
  isEdit.value = false
  currentServer.value = null
  formData.value = {
    name: '',
    description: '',
    transport_type: MCPTransportType.SSE,
    url: '',
    headers: {},
  }
  headerKeys.value = []
  headerValues.value = []
  modalVisible.value = true
}

// 打开编辑弹窗
const handleEdit = (server: MCPServer) => {
  isEdit.value = true
  currentServer.value = server
  formData.value = {
    name: server.name,
    description: server.description,
    transport_type: server.transport_type,
    url: server.url,
    headers: server.headers || {},
  }
  // 设置 Headers
  if (server.headers) {
    headerKeys.value = Object.keys(server.headers)
    headerValues.value = Object.values(server.headers)
  } else {
    headerKeys.value = []
    headerValues.value = []
  }
  modalVisible.value = true
}

// 添加 Header
const addHeader = () => {
  headerKeys.value.push('')
  headerValues.value.push('')
}

// 删除 Header
const removeHeader = (index: number) => {
  headerKeys.value.splice(index, 1)
  headerValues.value.splice(index, 1)
}

// 构建 Headers 对象
const buildHeaders = (): Record<string, string> | undefined => {
  const headers: Record<string, string> = {}
  for (let i = 0; i < headerKeys.value.length; i++) {
    const key = headerKeys.value[i]
    const value = headerValues.value[i]
    if (key && value) {
      headers[key] = value
    }
  }
  return Object.keys(headers).length > 0 ? headers : undefined
}

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  modalLoading.value = true
  try {
    const submitData = {
      ...formData.value,
      headers: buildHeaders(),
    }

    if (isEdit.value && currentServer.value) {
      const updateData: UpdateMCPServerRequest = {
        name: submitData.name,
        description: submitData.description,
        transport_type: submitData.transport_type,
        url: submitData.url,
        headers: submitData.headers,
      }
      await mcpApi.update(currentServer.value.id, updateData)
      message.success('更新成功')
    } else {
      await mcpApi.create(submitData)
      message.success('创建成功')
    }
    modalVisible.value = false
    fetchServers()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '操作失败')
  } finally {
    modalLoading.value = false
  }
}

// 删除服务器
const handleDelete = (server: MCPServer) => {
  Modal.confirm({
    title: '确认删除',
    icon: h(ExclamationCircleOutlined),
    content: `确定要删除 MCP 服务器 "${server.name}" 吗？此操作不可恢复。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await mcpApi.delete(server.id)
        message.success('删除成功')
        fetchServers()
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || '删除失败')
      }
    },
  })
}

// 启动服务器
const handleStart = async (server: MCPServer) => {
  try {
    await mcpApi.start(server.id)
    message.success('服务器已启动')
    fetchServers()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '启动失败')
  }
}

// 停止服务器
const handleStop = async (server: MCPServer) => {
  try {
    await mcpApi.stop(server.id)
    message.success('服务器已停止')
    fetchServers()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '停止失败')
  }
}

// 测试连接
const handleTest = async (server: MCPServer) => {
  try {
    const res = await mcpApi.testConnection(server.id)
    message.success(res.data?.message || '连接测试成功')
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '连接测试失败')
  }
}

// 查看工具列表
const handleViewTools = async (server: MCPServer) => {
  toolsModalVisible.value = true
  toolsLoading.value = true
  currentTools.value = []
  try {
    const res = await mcpApi.getTools(server.id)
    currentTools.value = res.data || []
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取工具列表失败')
  } finally {
    toolsLoading.value = false
  }
}

// 初始化
onMounted(() => {
  fetchServers()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">MCP 服务器</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建服务器
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
        placeholder="连接状态"
        style="width: 150px"
        allow-clear
      >
        <a-select-option :value="MCPServerStatus.Active">已连接</a-select-option>
        <a-select-option :value="MCPServerStatus.Inactive">未连接</a-select-option>
      </a-select>
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
        <a-card v-for="server in servers" :key="server.id" class="mcp-card" hoverable>
          <template #title>
            <div class="card-title">
              <ApiOutlined class="server-icon" />
              <span>{{ server.name }}</span>
              <a-tag :color="server.status === MCPServerStatus.Active ? 'success' : 'default'">
                {{ MCPServerStatusLabels[server.status] }}
              </a-tag>
            </div>
          </template>
          <div class="card-info">
            <div class="info-item">
              <span class="info-label">描述:</span>
              <span class="info-value">{{ server.description || '暂无描述' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">传输:</span>
              <a-tag size="small">{{ MCPTransportTypeLabels[server.transport_type] }}</a-tag>
            </div>
            <div class="info-item" v-if="server.url">
              <span class="info-label">URL:</span>
              <span class="info-value url">{{ server.url }}</span>
            </div>
            <div class="info-item" v-if="server.tools && server.tools.length > 0">
              <span class="info-label">工具:</span>
              <a-tag color="blue">{{ server.tools.length }} 个</a-tag>
            </div>
          </div>
          <template #actions>
            <a-tooltip v-if="server.status === MCPServerStatus.Inactive" title="启动连接">
              <PlayCircleOutlined @click="handleStart(server)" />
            </a-tooltip>
            <a-tooltip v-else title="停止连接">
              <StopOutlined @click="handleStop(server)" />
            </a-tooltip>
            <a-tooltip title="测试连接">
              <LinkOutlined @click="handleTest(server)" />
            </a-tooltip>
            <a-tooltip title="查看工具">
              <ToolOutlined @click="handleViewTools(server)" />
            </a-tooltip>
            <a-tooltip title="编辑">
              <EditOutlined @click="handleEdit(server)" />
            </a-tooltip>
            <a-tooltip title="删除">
              <DeleteOutlined @click="handleDelete(server)" />
            </a-tooltip>
          </template>
        </a-card>
      </div>

      <a-empty v-if="!loading && servers.length === 0" description="暂无 MCP 服务器，点击上方按钮创建" />
    </a-spin>

    <!-- 分页 -->
    <div class="pagination-container" v-if="total > size">
      <a-pagination
        v-model:current="page"
        :total="total"
        :pageSize="size"
        show-less-items
        @change="fetchServers"
      />
    </div>

    <!-- 创建/编辑弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="isEdit ? '编辑服务器' : '新建服务器'"
      :confirm-loading="modalLoading"
      width="600px"
      @ok="handleSubmit"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        layout="vertical"
        class="server-form"
      >
        <a-form-item label="服务器名称" name="name">
          <a-input v-model:value="formData.name" placeholder="请输入服务器名称" />
        </a-form-item>

        <a-form-item label="描述" name="description">
          <a-textarea v-model:value="formData.description" placeholder="请输入服务器描述" :rows="2" />
        </a-form-item>

        <a-form-item label="传输类型" name="transport_type">
          <a-select v-model:value="formData.transport_type" placeholder="请选择传输类型">
            <a-select-option :value="MCPTransportType.SSE">SSE (Server-Sent Events)</a-select-option>
            <a-select-option :value="MCPTransportType.Http">Http (Streamable HTTP)</a-select-option>
            <a-select-option :value="MCPTransportType.Stdio">Stdio</a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="服务器 URL" name="url">
          <a-input v-model:value="formData.url" placeholder="请输入 MCP 服务器 URL，如 http://localhost:8080/sse" />
        </a-form-item>

        <a-form-item label="请求头">
          <div class="headers-container">
            <div v-for="(_, index) in headerKeys" :key="index" class="header-row">
              <a-input v-model:value="headerKeys[index]" placeholder="Header 名称" style="width: 180px" />
              <a-input v-model:value="headerValues[index]" placeholder="Header 值" style="flex: 1" />
              <a-button type="text" danger @click="removeHeader(index)">
                <template #icon><DeleteOutlined /></template>
              </a-button>
            </div>
            <a-button type="dashed" block @click="addHeader">
              <template #icon><PlusOutlined /></template>
              添加请求头
            </a-button>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 工具列表弹窗 -->
    <a-modal
      v-model:open="toolsModalVisible"
      title="可用工具"
      :footer="null"
      width="600px"
    >
      <a-spin :spinning="toolsLoading">
        <div class="tools-list">
          <div v-for="tool in currentTools" :key="tool.name" class="tool-item">
            <div class="tool-name">{{ tool.name }}</div>
            <div class="tool-desc">{{ tool.description }}</div>
          </div>
          <a-empty v-if="!toolsLoading && currentTools.length === 0" description="暂无可用工具" />
        </div>
      </a-spin>
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
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: 16px;
}

.mcp-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  transition: all 0.3s ease;
}

.mcp-card:hover {
  box-shadow: var(--card-hover-shadow);
  transform: translateY(-2px);
}

.mcp-card :deep(.ant-card-actions) {
  border-top: 1px solid var(--border-color);
}

.mcp-card :deep(.ant-card-actions > li) {
  color: var(--card-action-color);
  transition: color 0.3s ease;
}

.mcp-card :deep(.ant-card-actions > li:hover) {
  color: var(--card-action-hover-color);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.server-icon {
  font-size: 18px;
  color: var(--primary-color);
}

.card-info {
  margin-bottom: 12px;
}

.info-item {
  display: flex;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 13px;
}

.info-label {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.info-value {
  color: var(--text-secondary);
  word-break: break-all;
}

.info-value.url {
  font-family: monospace;
  font-size: 12px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

.server-form {
  margin-top: 16px;
}

.headers-container {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.header-row {
  display: flex;
  gap: 8px;
  align-items: center;
}

.tools-list {
  max-height: 400px;
  overflow-y: auto;
}

.tool-item {
  padding: 12px;
  border-bottom: 1px solid var(--border-color);
}

.tool-item:last-child {
  border-bottom: none;
}

.tool-name {
  font-weight: 500;
  margin-bottom: 4px;
}

.tool-desc {
  font-size: 13px;
  color: var(--text-secondary);
}
</style>