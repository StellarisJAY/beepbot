<script setup lang="ts">
import { ref } from 'vue'
import { message } from 'ant-design-vue'
import { agentApi } from '@/api/agent'
import type { CreateAgentRequest } from '@/types/agent'

const emit = defineEmits<{
  success: [id: string]
  cancel: []
}>()

const visible = defineModel<boolean>('visible', { default: false })

const loading = ref(false)
const form = ref<CreateAgentRequest>({
  name: '',
  description: '',
})

const handleSubmit = async () => {
  if (!form.value.name.trim()) {
    message.warning('请输入智能体名称')
    return
  }

  loading.value = true
  try {
    const res = await agentApi.create({
      name: form.value.name.trim(),
      description: form.value.description?.trim() || undefined,
    })
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
    </a-form>
    <div class="create-tip">
      <a-typography-text type="secondary">
        创建后将使用系统默认配置，您可以在编辑页面进行详细配置。
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