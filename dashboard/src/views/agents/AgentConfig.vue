<script setup lang="ts">
import { ref, onMounted, inject, watch, type Ref, type ComputedRef } from 'vue'
import { message } from 'ant-design-vue'
import { agentApi } from '@/api/agent'
import { providerApi } from '@/api/provider'
import type { Agent, AgentDefaults, UpdateAgentRequest } from '@/types/agent'
import type { Provider } from '@/types/provider'
import { AgentStatus } from '@/types/agent'

// 从父组件注入的数据
const agent = inject<Ref<Agent | null>>('agent')
const agentId = inject<ComputedRef<string>>('agentId')
const fetchAgent = inject<() => Promise<void>>('fetchAgent')

// 数据
const defaults = ref<AgentDefaults | null>(null)
const providers = ref<Provider[]>([])
const saving = ref(false)

// 表单数据
const form = ref<UpdateAgentRequest>({})

// 获取默认配置
const fetchDefaults = async () => {
  try {
    const res = await agentApi.getDefaults()
    defaults.value = res.data
  } catch (error) {
    console.error('获取默认配置失败:', error)
  }
}

// 获取供应商列表
const fetchProviders = async () => {
  try {
    const res = await providerApi.list()
    providers.value = res.data
  } catch (error) {
    console.error('获取供应商列表失败:', error)
  }
}

// 初始化表单
const initForm = () => {
  if (agent?.value) {
    form.value = {
      name: agent.value.name,
      description: agent.value.description,
      provider_id: agent.value.provider_id,
      model: agent.value.model,
      system_prompt: agent.value.system_prompt,
      temperature: agent.value.temperature,
      max_iterations: agent.value.max_iterations,
      max_output_tokens: agent.value.max_output_tokens,
      working_dir: agent.value.working_dir,
      context_window_size: agent.value.context_window_size,
      window_size: agent.value.window_size,
      compression_ratio: agent.value.compression_ratio,
      context_max_tokens: agent.value.context_max_tokens,
      status: agent.value.status,
    }
  }
}

// 保存
const handleSave = async () => {
  // 校验必填项
  if (!form.value.provider_id) {
    message.warning('请选择供应商')
    return
  }
  if (!form.value.model) {
    message.warning('请选择或输入模型名称')
    return
  }
  if (!form.value.working_dir) {
    message.warning('请输入工作目录')
    return
  }

  saving.value = true
  try {
    await agentApi.update(agentId!.value, form.value)
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

onMounted(() => {
  fetchDefaults()
  fetchProviders()
})
</script>

<template>
  <div class="config-content">
    <!-- 模型配置 -->
    <a-card title="模型配置" size="small" class="config-card">
      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="供应商" required>
          <a-select
            v-model:value="form.provider_id"
            placeholder="请选择供应商"
            allow-clear
          >
            <a-select-option
              v-for="p in providers"
              :key="p.id"
              :value="p.id"
            >
              {{ p.name }} ({{ p.provider_type }})
            </a-select-option>
          </a-select>
        </a-form-item>
        <a-form-item label="模型" required>
          <a-input
            v-model:value="form.model"
            placeholder="请输入模型名称"
            :maxlength="128"
          />
        </a-form-item>
        <a-form-item label="温度">
          <a-row :gutter="16">
            <a-col :span="18">
              <a-slider
                v-model:value="form.temperature"
                :min="0"
                :max="2"
                :step="0.1"
              />
            </a-col>
            <a-col :span="6">
              <a-input-number
                v-model:value="form.temperature"
                :min="0"
                :max="2"
                :step="0.1"
                size="small"
              />
            </a-col>
          </a-row>
        </a-form-item>
        <a-form-item label="最大输出Token">
          <a-input-number
            v-model:value="form.max_output_tokens"
            :min="1"
            :max="1000000"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="最大迭代次数">
          <a-input-number
            v-model:value="form.max_iterations"
            :min="1"
            :max="1000"
            style="width: 100%"
          />
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 提示词配置 -->
    <a-card title="提示词配置" size="small" class="config-card">
      <a-form :label-col="{ span: 4 }" :wrapper-col="{ span: 20 }">
        <a-form-item label="系统提示词">
          <a-textarea
            v-model:value="form.system_prompt"
            placeholder="请输入系统提示词"
            :rows="6"
            show-count
            :maxlength="10000"
          />
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 能力配置 -->
    <a-card title="能力配置" size="small" class="config-card">
      <a-form :label-col="{ span: 6 }" :wrapper-col="{ span: 18 }">
        <a-form-item label="工作目录" required>
          <a-input
            v-model:value="form.working_dir"
            placeholder="请输入工作目录"
            :maxlength="512"
          />
        </a-form-item>
      </a-form>
    </a-card>

    <!-- 上下文配置 -->
    <a-card title="上下文配置" size="small" class="config-card">
      <a-form :label-col="{ span: 8 }" :wrapper-col="{ span: 16 }">
        <a-form-item label="上下文窗口大小">
          <a-input-number
            v-model:value="form.context_window_size"
            :min="1"
            :max="1000"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="消息窗口大小">
          <a-input-number
            v-model:value="form.window_size"
            :min="1"
            :max="1000"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="压缩比例">
          <a-input-number
            v-model:value="form.compression_ratio"
            :min="0"
            :max="1"
            :step="0.1"
            :precision="2"
            style="width: 100%"
          />
        </a-form-item>
        <a-form-item label="最大上下文Token">
          <a-input-number
            v-model:value="form.context_max_tokens"
            :min="1"
            :max="1000000"
            style="width: 100%"
          />
        </a-form-item>
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