<script setup lang="ts">
import { ref, onMounted, h, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  StarOutlined,
  StarFilled,
  ExclamationCircleOutlined,
  SearchOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import { providerApi } from '@/api/provider'
import {
  type Provider,
  type CreateProviderRequest,
  type UpdateProviderRequest,
  type ProviderFilter,
  ProviderType,
  ProviderTypeLabels,
  ProviderTypeOptions,
  ProviderTypeIcons,
  ProviderDefaultBaseURL,
} from '@/types/provider'

// 供应商列表
const providers = ref<Provider[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 筛选条件
const filters = ref<ProviderFilter>({
  name: '',
  provider_type: undefined,
})

// 弹窗相关
const modalVisible = ref(false)
const modalLoading = ref(false)
const isEdit = ref(false)
const currentProvider = ref<Provider | null>(null)

// 表单数据
const formData = ref<CreateProviderRequest>({
  name: '',
  provider_type: ProviderType.OpenAI,
  api_key: '',
  base_url: '',
  is_default: false,
})

// 表单引用
const formRef = ref()

// 表单校验规则
const formRules = computed(() => ({
  name: [
    { required: true, message: '请输入供应商名称', trigger: 'blur' },
    { min: 2, max: 64, message: '名称长度为 2-64 个字符', trigger: 'blur' },
  ],
  provider_type: [{ required: true, message: '请选择供应商类型', trigger: 'change' }],
  api_key: [
    {
      required: !isEdit.value && formData.value.provider_type !== ProviderType.Ollama,
      message: '请输入 API Key',
      trigger: 'blur',
    },
    {
      min: 8,
      message: 'API Key 长度不能少于 8 个字符',
      trigger: 'blur',
      validator: (_rule: unknown, value: string) => {
        // Ollama 不需要 API Key，跳过长度验证
        if (formData.value.provider_type === ProviderType.Ollama) {
          return Promise.resolve()
        }
        if (isEdit.value && !value) {
          return Promise.resolve() // 编辑时留空表示不修改
        }
        if (value && value.length < 8) {
          return Promise.reject('API Key 长度不能少于 8 个字符')
        }
        return Promise.resolve()
      },
    },
  ],
  base_url: [
    { required: true, message: '请输入 Base URL', trigger: 'blur' },
    { type: 'url', message: '请输入有效的 URL', trigger: 'blur' },
  ],
}))

// 获取供应商列表
const fetchProviders = async () => {
  loading.value = true
  try {
    // 构建筛选参数，过滤掉空值
    const filterParams: ProviderFilter = {}
    if (filters.value.name) filterParams.name = filters.value.name
    if (filters.value.provider_type) filterParams.provider_type = filters.value.provider_type

    const res = await providerApi.list(page.value, size.value, filterParams)
    providers.value = res.data
    total.value = res.total
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取供应商列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchProviders()
}

// 重置筛选
const handleReset = () => {
  filters.value = { name: '', provider_type: undefined }
  page.value = 1
  fetchProviders()
}

// 打开创建弹窗
const handleCreate = () => {
  isEdit.value = false
  currentProvider.value = null
  formData.value = {
    name: '',
    provider_type: ProviderType.OpenAI,
    api_key: '',
    base_url: ProviderDefaultBaseURL[ProviderType.OpenAI],
    is_default: false,
  }
  modalVisible.value = true
}

// 打开编辑弹窗
const handleEdit = (provider: Provider) => {
  isEdit.value = true
  currentProvider.value = provider
  formData.value = {
    name: provider.name,
    provider_type: provider.provider_type,
    api_key: '', // 编辑时不回显 API Key
    base_url: provider.base_url,
    is_default: provider.is_default,
  }
  modalVisible.value = true
}

// 供应商类型变更时更新默认 URL
const handleProviderTypeChange = (type: ProviderType) => {
  if (!isEdit.value) {
    formData.value.base_url = ProviderDefaultBaseURL[type]
  }
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
    if (isEdit.value && currentProvider.value) {
      // 更新
      const updateData: UpdateProviderRequest = {
        name: formData.value.name,
        base_url: formData.value.base_url,
        is_default: formData.value.is_default,
      }
      // 只有输入了新的 API Key 才更新
      if (formData.value.api_key) {
        updateData.api_key = formData.value.api_key
      }
      await providerApi.update(currentProvider.value.id, updateData)
      message.success('更新成功')
    } else {
      // 创建
      await providerApi.create(formData.value)
      message.success('创建成功')
    }
    modalVisible.value = false
    fetchProviders()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '操作失败')
  } finally {
    modalLoading.value = false
  }
}

// 删除供应商
const handleDelete = (provider: Provider) => {
  Modal.confirm({
    title: '确认删除',
    icon: h(ExclamationCircleOutlined),
    content: `确定要删除供应商 "${provider.name}" 吗？此操作不可恢复。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await providerApi.delete(provider.id)
        message.success('删除成功')
        fetchProviders()
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || '删除失败')
      }
    },
  })
}

// 设置默认供应商
const handleSetDefault = async (provider: Provider) => {
  try {
    await providerApi.setDefault(provider.id)
    message.success('已设置为默认供应商')
    fetchProviders()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '设置失败')
  }
}

// 初始化
onMounted(() => {
  fetchProviders()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">模型供应商</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建供应商
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
        v-model:value="filters.provider_type"
        placeholder="供应商类型"
        style="width: 180px"
        allow-clear
      >
        <a-select-option
          v-for="option in ProviderTypeOptions"
          :key="option.value"
          :value="option.value"
        >
          <div class="provider-option">
            <img :src="option.icon" class="provider-option-icon" :alt="option.value" />
            <span>{{ option.label }}</span>
          </div>
        </a-select-option>
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
        <a-card v-for="provider in providers" :key="provider.id" class="provider-card" hoverable>
          <template #title>
            <div class="card-title">
              <img
                :src="ProviderTypeIcons[provider.provider_type as ProviderType]"
                class="provider-icon"
                :alt="provider.provider_type"
              />
              <span>{{ provider.name }}</span>
              <StarFilled v-if="provider.is_default" class="default-star" />
            </div>
          </template>
          <div class="card-info">
            <div class="info-item">
              <span class="info-label">类型:</span>
              <a-tag :color="provider.is_default ? 'gold' : 'blue'">
                {{ ProviderTypeLabels[provider.provider_type as ProviderType] }}
              </a-tag>
            </div>
            <div class="info-item">
              <span class="info-label">API Key:</span>
              <span class="info-value masked">{{ provider.api_key_masked || '未设置' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">地址:</span>
              <span class="info-value url">{{ provider.base_url }}</span>
            </div>
          </div>
          <template #actions>
            <a-tooltip title="设为默认">
              <StarOutlined
                :class="{ 'action-default': true, 'is-default': provider.is_default }"
                @click="!provider.is_default && handleSetDefault(provider)"
              />
            </a-tooltip>
            <a-tooltip title="编辑">
              <EditOutlined @click="handleEdit(provider)" />
            </a-tooltip>
            <a-tooltip title="删除">
              <DeleteOutlined @click="handleDelete(provider)" />
            </a-tooltip>
          </template>
        </a-card>
      </div>

      <a-empty v-if="!loading && providers.length === 0" description="暂无供应商，点击上方按钮创建" />
    </a-spin>

    <!-- 分页 -->
    <div class="pagination-container" v-if="total > size">
      <a-pagination
        v-model:current="page"
        :total="total"
        :pageSize="size"
        show-less-items
        @change="fetchProviders"
      />
    </div>

    <!-- 创建/编辑弹窗 -->
    <a-modal
      v-model:open="modalVisible"
      :title="isEdit ? '编辑供应商' : '新建供应商'"
      :confirm-loading="modalLoading"
      @ok="handleSubmit"
    >
      <a-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        layout="vertical"
        class="provider-form"
      >
        <a-form-item label="供应商名称" name="name">
          <a-input v-model:value="formData.name" placeholder="请输入供应商名称" />
        </a-form-item>

        <a-form-item label="供应商类型" name="provider_type">
          <a-select
            v-model:value="formData.provider_type"
            :disabled="isEdit"
            placeholder="请选择供应商类型"
            @change="handleProviderTypeChange"
          >
            <a-select-option
              v-for="option in ProviderTypeOptions"
              :key="option.value"
              :value="option.value"
            >
              <div class="provider-option">
                <img :src="option.icon" class="provider-option-icon" :alt="option.value" />
                <span>{{ option.label }}</span>
              </div>
            </a-select-option>
          </a-select>
        </a-form-item>

        <a-form-item label="API Key" name="api_key">
          <a-input-password
            v-model:value="formData.api_key"
            :placeholder="formData.provider_type === ProviderType.Ollama
              ? 'Ollama 不需要 API Key，可留空'
              : (isEdit ? '留空则不修改 API Key' : '请输入 API Key')"
          />
        </a-form-item>

        <a-form-item label="Base URL" name="base_url">
          <a-input v-model:value="formData.base_url" placeholder="请输入 API Base URL" />
        </a-form-item>

        <a-form-item name="is_default">
          <a-checkbox v-model:checked="formData.is_default">设为默认供应商</a-checkbox>
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

.provider-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  transition: all 0.3s ease;
}

.provider-card:hover {
  box-shadow: var(--card-hover-shadow);
  transform: translateY(-2px);
}

/* 卡片操作按钮样式 */
.provider-card :deep(.ant-card-actions) {
  border-top: 1px solid var(--border-color);
}

.provider-card :deep(.ant-card-actions > li) {
  color: var(--card-action-color);
  transition: color 0.3s ease;
}

.provider-card :deep(.ant-card-actions > li:hover) {
  color: var(--card-action-hover-color);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.provider-icon {
  width: 20px;
  height: 20px;
  object-fit: contain;
}

.default-star {
  color: #faad14;
  font-size: 14px;
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

.info-value.masked {
  font-family: monospace;
}

.action-default {
  cursor: pointer;
}

.action-default.is-default {
  color: #faad14;
  cursor: default;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

.provider-form {
  margin-top: 16px;
}

.provider-option {
  display: flex;
  align-items: center;
  gap: 8px;
}

.provider-option-icon {
  width: 18px;
  height: 18px;
  object-fit: contain;
}
</style>