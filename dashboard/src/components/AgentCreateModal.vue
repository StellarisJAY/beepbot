<script setup lang="ts">
import { ref, watch } from 'vue'
import { message } from 'ant-design-vue'
import { agentApi } from '@/api/agent'
import type { CreateAgentRequest } from '@/types/agent'
import { ExternalType, ExternalTypeOptions } from '@/types/agent'

const emit = defineEmits<{
  success: [id: string]
  cancel: []
}>()

const visible = defineModel<boolean>('visible', { default: false })

const loading = ref(false)
const form = ref<CreateAgentRequest>({
  name: '',
  description: '',
  external: false,
  external_type: ExternalType.Dify,
  external_config: {
    base_url: '',
    api_key: '',
  },
})

// 监听 external 开关，重置外部配置
watch(
  () => form.value.external,
  (newVal) => {
    if (newVal) {
      form.value.external_type = ExternalType.Dify
      form.value.external_config = {
        base_url: '',
        api_key: '',
      }
    } else {
      form.value.external_type = undefined
      form.value.external_config = undefined
    }
  }
)

const handleSubmit = async () => {
  if (!form.value.name.trim()) {
    message.warning('请输入智能体名称')
    return
  }

  // 外部智能体验证
  if (form.value.external) {
    if (!form.value.external_config?.base_url?.trim()) {
      message.warning('请输入 API 地址')
      return
    }
    if (!form.value.external_config?.api_key?.trim()) {
      message.warning('请输入 API Key')
      return
    }
  }

  loading.value = true
  try {
    const requestData: CreateAgentRequest = {
      name: form.value.name.trim(),
      description: form.value.description?.trim() || undefined,
    }

    if (form.value.external) {
      requestData.external = true
      requestData.external_type = form.value.external_type
      requestData.external_config = {
        base_url: form.value.external_config?.base_url?.trim() || '',
        api_key: form.value.external_config?.api_key?.trim() || '',
      }
    }

    const res = await agentApi.create(requestData)
    message.success('创建成功')
    visible.value = false
    resetForm()
    emit('success', res.data.id)
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '创建失败')
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  visible.value = false
  resetForm()
  emit('cancel')
}

const resetForm = () => {
  form.value = {
    name: '',
    description: '',
    external: false,
    external_type: ExternalType.Dify,
    external_config: {
      base_url: '',
      api_key: '',
    },
  }
}
</script>

<template>
  <a-modal
    v-model:open="visible"
    title="新建智能体"
    :confirm-loading="loading"
    @ok="handleSubmit"
    @cancel="handleCancel"
  >
    <a-form :label-col="{ span: 4 }" :wrapper-col="{ span: 20 }">
      <a-form-item label="名称" required>
        <a-input
          v-model:value="form.name"
          placeholder="请输入智能体名称"
          :maxlength="128"
        />
      </a-form-item>
      <a-form-item label="描述">
        <a-textarea
          v-model:value="form.description"
          placeholder="请输入智能体描述（可选）"
          :rows="3"
          :maxlength="500"
        />
      </a-form-item>

      <!-- 外部智能体开关 -->
      <a-form-item label="外部智能体">
        <a-switch v-model:checked="form.external" />
        <span style="margin-left: 8px; color: #999; font-size: 12px">
          接入 Dify 等外部智能体平台
        </span>
      </a-form-item>

      <!-- 外部智能体配置 -->
      <template v-if="form.external">
        <a-form-item label="平台类型" required>
          <a-select v-model:value="form.external_type" :options="ExternalTypeOptions" />
        </a-form-item>

        <!-- Dify 配置 -->
        <template v-if="form.external_type === 'dify'">
          <a-form-item label="API 地址" required>
            <a-input
              v-model:value="form.external_config!.base_url"
              placeholder="https://api.dify.ai/v1"
            />
          </a-form-item>
          <a-form-item label="API Key" required>
            <a-input-password
              v-model:value="form.external_config!.api_key"
              placeholder="app-xxx"
            />
          </a-form-item>
        </template>
      </template>
    </a-form>

    <div class="create-tip">
      <a-typography-text type="secondary">
        {{ form.external
          ? '外部智能体将使用 Dify 等平台的对话能力。'
          : '创建后将使用系统默认配置，您可以在编辑页面进行详细配置。'
        }}
      </a-typography-text>
    </div>
  </a-modal>
</template>

<style scoped>
.create-tip {
  margin-top: 16px;
  padding: 12px;
  background: var(--card-bg);
  border-radius: 4px;
}
</style>