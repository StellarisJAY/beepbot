<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { message } from 'ant-design-vue'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{
  visible: boolean
  isFirstLogin?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'success'): void
}>()

const authStore = useAuthStore()

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const loading = ref(false)

// 重置表单
watch(
  () => props.visible,
  (val) => {
    if (!val) {
      form.old_password = ''
      form.new_password = ''
      form.confirm_password = ''
    }
  }
)

async function handleSubmit() {
  if (!form.old_password || !form.new_password || !form.confirm_password) {
    message.error('请填写所有字段')
    return
  }

  if (form.new_password !== form.confirm_password) {
    message.error('两次输入的新密码不一致')
    return
  }

  if (form.new_password.length < 6) {
    message.error('新密码长度至少为 6 个字符')
    return
  }

  loading.value = true
  try {
    await authStore.changePassword(form.old_password, form.new_password)
    message.success('密码修改成功，请重新登录')
    emit('update:visible', false)
    emit('success')
    // 退出登录
    authStore.logout()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '密码修改失败')
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  // 如果是首次登录强制修改密码，不允许取消
  if (props.isFirstLogin) {
    message.warning('首次登录需要修改密码')
    return
  }
  emit('update:visible', false)
}
</script>

<template>
  <a-modal
    :open="visible"
    :title="isFirstLogin ? '首次登录 - 修改密码' : '修改密码'"
    :closable="!isFirstLogin"
    :maskClosable="false"
    :keyboard="false"
    @cancel="handleCancel"
  >
    <a-alert
      v-if="isFirstLogin"
      message="首次登录需要修改密码后才能继续使用"
      type="warning"
      show-icon
      style="margin-bottom: 16px"
    />

    <a-form :model="form" layout="vertical">
      <a-form-item label="旧密码" required>
        <a-input-password
          v-model:value="form.old_password"
          placeholder="请输入旧密码"
          autocomplete="current-password"
        />
      </a-form-item>

      <a-form-item label="新密码" required>
        <a-input-password
          v-model:value="form.new_password"
          placeholder="请输入新密码（至少6位）"
          autocomplete="new-password"
        />
      </a-form-item>

      <a-form-item label="确认新密码" required>
        <a-input-password
          v-model:value="form.confirm_password"
          placeholder="请再次输入新密码"
          autocomplete="new-password"
        />
      </a-form-item>
    </a-form>

    <template #footer>
      <a-button v-if="!isFirstLogin" @click="handleCancel">取消</a-button>
      <a-button type="primary" :loading="loading" @click="handleSubmit"> 确认修改 </a-button>
    </template>
  </a-modal>
</template>