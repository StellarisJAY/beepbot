<script setup lang="ts">
import { ref, inject, watch, computed, type Ref, type ComputedRef } from 'vue'
import { message } from 'ant-design-vue'
import { agentApi } from '@/api/agent'
import type { Agent, UpdateAgentRequest } from '@/types/agent'
import { AgentStatus, ExternalTypeLabels } from '@/types/agent'

// 从父组件注入的数据
const agent = inject<Ref<Agent | null>>('agent')
const agentId = inject<ComputedRef<string>>('agentId')
const fetchAgent = inject<() => Promise<void>>('fetchAgent')

// 表单数据
const form = ref<UpdateAgentRequest>({})
const saving = ref(false)

// 外部配置
const externalConfig = ref({
  base_url: '',
  api_key: '',
})

// 初始化表单
const initForm = () => {
  if (agent?.value) {
    form.value = {
      name: agent.value.name,
      description: agent.value.description,
      status: agent.value.status,
    }

    // 初始化外部配置
    if (agent.value.external_config) {
      externalConfig.value = {
        base_url: agent.value.external_config.base_url || '',
        api_key: '', // API Key 不回显，留空表示保持不变
      }
    }
  }
}

// 外部类型显示名称
const externalTypeLabel = computed(() => {
  if (agent?.value?.external_type) {
    return ExternalTypeLabels[agent.value.external_type as keyof typeof ExternalTypeLabels] || agent.value.external_type
  }
  return ''
})

// 保存
const handleSave = async () => {
  // 校验必填项
  if (!externalConfig.value.base_url.trim()) {
    message.warning('请输入 API 地址')
    return
  }

  saving.value = true
  try {
    const externalConfigData: { base_url: string; api_key?: string } = {
      base_url: externalConfig.value.base_url.trim(),
    }

    // 只有在输入了新的 API Key 时才传递
    if (externalConfig.value.api_key.trim()) {
      externalConfigData.api_key = externalConfig.value.api_key.trim()
    }

    const requestData: UpdateAgentRequest = {
      name: form.value.name,
      description: form.value.description,
      status: form.value.status,
      external_config: externalConfigData,
    }

    await agentApi.update(agentId!.value, requestData)
    message.success('保存成功')
    if (fetchAgent) {
      await fetchAgent()
    }
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    saving.value = false
  }
}

// 暴露给父组件的方法
defineExpose({
  handleSave,
  saving,
})

// 监听 agent 变化初始化表单
watch(
  () => agent?.value,
  () => {
    if (agent?.value) {
      initForm()
    }
  },
  { immediate: true }
)
</script>

<template>
  <div class="config-content">
    <!-- 基本信息 -->
    <a-card title="基本信息" size="small" class="config-card">
      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="名称">
          <a-input v-model:value="form.name" placeholder="请输入名称" :maxlength="128" />
        </a-form-item>
        <a-form-item label="描述">
          <a-textarea
            v-model:value="form.description"
            placeholder="请输入描述"
            :rows="3"
            :maxlength="500"
          />
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 外部平台配置 -->
    <a-card title="平台配置" size="small" class="config-card">
      <a-alert message="平台类型创建后不可修改" type="info" show-icon style="margin-bottom: 16px" />

      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="平台类型">
          <a-tag color="blue">{{ externalTypeLabel }}</a-tag>
        </a-form-item>

        <!-- Dify 配置 -->
        <template v-if="agent?.external_type === 'dify'">
          <a-form-item label="API 地址" required>
            <a-input
              v-model:value="externalConfig.base_url"
              placeholder="https://api.dify.ai/v1"
            />
          </a-form-item>
          <a-form-item label="API Key">
            <a-input-password
              v-model:value="externalConfig.api_key"
              :placeholder="agent?.external_config?.api_key_masked || '保持不变'"
            />
            <div style="margin-top: 4px; color: #999; font-size: 12px">
              留空则保持原有密钥不变
            </div>
          </a-form-item>
        </template>
      </a-form>
    </a-card>

    <!-- 状态配置 -->
    <a-card title="状态配置" size="small" class="config-card">
      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="状态">
          <a-switch
            :checked="form.status === AgentStatus.Active"
            checked-children="启用"
            un-checked-children="禁用"
            @change="(checked: boolean) => form.status = checked ? AgentStatus.Active : AgentStatus.Inactive"
          />
        </a-form-item>
      </a-form>
    </a-card>
  </div>
</template>

<style scoped>
.config-content {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.config-card {
  border-radius: 8px;
}

.config-card :deep(.ant-card-head) {
  border-bottom: 1px solid var(--border-color);
}

.config-card :deep(.ant-card-body) {
  padding: 16px 24px;
}
</style>