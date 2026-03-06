<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { message } from 'ant-design-vue'
import { botApi } from '@/api/bot'
import type { Bot, CreateBotRequest, UpdateBotRequest, QQBotConfig } from '@/types/bot'
import { BotPlatform, BotPlatformOptions } from '@/types/bot'
import type { Agent } from '@/types/agent'

// Props
const props = defineProps<{
  visible: boolean
  bot: Bot | null
  agents: Agent[]
}>()

// Emits
const emit = defineEmits<{
  'update:visible': [value: boolean]
  success: []
}>()

// 表单数据
const formData = ref({
  name: '',
  description: '',
  platform: BotPlatform.QQ,
  identifier: '',
  agent_id: undefined as string | undefined,
  // QQ 平台配置
  app_id: '',
  app_secret: '',
})

const loading = ref(false)

// 计算属性
const isEdit = computed(() => !!props.bot)
const modalTitle = computed(() => (isEdit.value ? '编辑机器人' : '新建机器人'))

// 监听 visible 变化，初始化表单
watch(
  () => props.visible,
  (val) => {
    if (val) {
      if (props.bot) {
        // 编辑模式，填充数据
        const botConfig = props.bot.config as unknown as QQBotConfig
        formData.value = {
          name: props.bot.name,
          description: props.bot.description || '',
          platform: props.bot.platform,
          identifier: props.bot.identifier || '',
          agent_id: props.bot.agent_id || undefined,
          app_id: botConfig?.app_id || '',
          app_secret: botConfig?.app_secret || '',
        }
      } else {
        // 新建模式，重置表单
        resetForm()
      }
    }
  }
)

// 重置表单
const resetForm = () => {
  formData.value = {
    name: '',
    description: '',
    platform: BotPlatform.QQ,
    identifier: '',
    agent_id: undefined,
    app_id: '',
    app_secret: '',
  }
}

// 关闭弹窗
const handleCancel = () => {
  emit('update:visible', false)
  resetForm()
}

// 提交表单
const handleSubmit = async () => {
  // 验证必填字段
  if (!formData.value.name.trim()) {
    message.error('请输入机器人名称')
    return
  }
  if (!formData.value.platform) {
    message.error('请选择平台')
    return
  }
  // QQ 平台验证
  if (formData.value.platform === BotPlatform.QQ) {
    if (!formData.value.app_id.trim()) {
      message.error('请输入 App ID')
      return
    }
    if (!formData.value.app_secret.trim()) {
      message.error('请输入 App Secret')
      return
    }
  }

  loading.value = true
  try {
    // 构建配置
    const config: Record<string, unknown> = {
      app_id: formData.value.app_id,
      app_secret: formData.value.app_secret,
    }

    if (isEdit.value && props.bot) {
      // 更新
      const data: UpdateBotRequest = {
        name: formData.value.name,
        description: formData.value.description,
        identifier: formData.value.identifier,
        config,
      }
      await botApi.update(props.bot.id, data)
      message.success('更新成功')
    } else {
      // 创建
      const data: CreateBotRequest = {
        name: formData.value.name,
        description: formData.value.description,
        platform: formData.value.platform,
        identifier: formData.value.identifier,
        config,
        agent_id: formData.value.agent_id,
      }
      await botApi.create(data)
      message.success('创建成功')
    }

    emit('success')
    handleCancel()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || (isEdit.value ? '更新失败' : '创建失败'))
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <a-modal
    :open="visible"
    :title="modalTitle"
    :confirm-loading="loading"
    ok-text="确定"
    cancel-text="取消"
    @ok="handleSubmit"
    @cancel="handleCancel"
    width="520px"
  >
    <a-form layout="vertical">
      <a-form-item label="机器人名称" required>
        <a-input
          v-model:value="formData.name"
          placeholder="请输入机器人名称"
          :maxlength="128"
        />
      </a-form-item>

      <a-form-item label="描述">
        <a-textarea
          v-model:value="formData.description"
          placeholder="请输入描述"
          :rows="3"
          :maxlength="500"
        />
      </a-form-item>

      <a-form-item label="平台" required>
        <a-select
          v-model:value="formData.platform"
          placeholder="请选择平台"
          :disabled="isEdit"
        >
          <a-select-option
            v-for="option in BotPlatformOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </a-select-option>
        </a-select>
      </a-form-item>

      <a-form-item label="标识符">
        <a-input
          v-model:value="formData.identifier"
          placeholder="请输入标识符（可选）"
          :maxlength="128"
        />
      </a-form-item>

      <!-- QQ 平台配置 -->
      <template v-if="formData.platform === BotPlatform.QQ">
        <a-divider>QQ 平台配置</a-divider>

        <a-form-item label="App ID" required>
          <a-input
            v-model:value="formData.app_id"
            placeholder="请输入 QQ 机器人 App ID"
            :maxlength="128"
          />
        </a-form-item>

        <a-form-item label="App Secret" required>
          <a-input-password
            v-model:value="formData.app_secret"
            placeholder="请输入 QQ 机器人 App Secret"
            :maxlength="256"
          />
        </a-form-item>
      </template>

      <a-form-item label="绑定智能体">
        <a-select
          v-model:value="formData.agent_id"
          placeholder="请选择要绑定的智能体（可选）"
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
</template>

<style scoped>
/* 可以添加自定义样式 */
</style>