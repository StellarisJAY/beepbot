<script setup lang="ts">
import { ref, onMounted, h, computed } from 'vue'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  ApiOutlined,
  StarOutlined,
  StarFilled,
  ExclamationCircleOutlined,
} from '@ant-design/icons-vue'
import { providerApi } from '@/api/provider'
import {
  type Provider,
  type CreateProviderRequest,
  type UpdateProviderRequest,
  ProviderType,
  ProviderTypeLabels,
  ProviderTypeOptions,
  ProviderDefaultBaseURL,
} from '@/types/provider'

// 供应商列表
const providers = ref<Provider[]>([])
const loading = ref(false)

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
    { required: !isEdit.value, message: '请输入 API Key', trigger: 'blur' },
    { min: 8, message: 'API Key 长度不能少于 8 个字符', trigger: 'blur' },
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
    const { data } = await providerApi.list()
    providers.value = data || []
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取供应商列表失败')
  } finally {
    loading.value = false
  }
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

    <a-spin :spinning="loading">
      <div class="card-grid">
        <a-card v-for="provider in providers" :key="provider.id" class="provider-card" hoverable>
          <template #title>
            <div class="card-title">
              <ApiOutlined class="card-icon" />
              <span>{{ provider.name }}</span>
              <StarFilled v-if="provider.is_default" class="default-star" />
            </div>
          </template>
          <template #extra>
            <a-tag :color="provider.is_default ? 'gold' : 'blue'">
              {{ ProviderTypeLabels[provider.provider_type as ProviderType] }}
            </a-tag>
          </template>
          <div class="card-info">
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
            :options="ProviderTypeOptions"
            :disabled="isEdit"
            placeholder="请选择供应商类型"
            @change="handleProviderTypeChange"
          />
        </a-form-item>

        <a-form-item label="API Key" name="api_key">
          <a-input-password
            v-model:value="formData.api_key"
            :placeholder="isEdit ? '留空则不修改 API Key' : '请输入 API Key'"
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

.card-icon {
  font-size: 18px;
  color: var(--color-primary);
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

.provider-form {
  margin-top: 16px;
}
</style>