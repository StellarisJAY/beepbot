<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import ChangePasswordModal from '@/components/ChangePasswordModal.vue'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const showChangePassword = ref(false)

// 判断是否为公开页面（登录页）
const isPublicPage = computed(() => {
  return route.meta.public === true || route.path === '/login'
})

// 获取用户信息并检查是否需要修改密码
onMounted(async () => {
  if (authStore.isAuthenticated) {
    const user = await authStore.fetchUser()
    // 首次登录需要修改密码
    if (user?.require_password_change) {
      showChangePassword.value = true
    }
  }
})

// 密码修改成功后的处理
function onPasswordChanged() {
  // 跳转到登录页
  router.push('/login')
}
</script>

<template>
  <template v-if="!isPublicPage">
    <AppLayout />
    <ChangePasswordModal
      v-model:visible="showChangePassword"
      :is-first-login="authStore.user?.require_password_change"
      @success="onPasswordChanged"
    />
  </template>
  <router-view v-else />
</template>

<style>
/* 引入全局样式 */
@import '@/assets/styles/variables.css';
@import '@/assets/styles/global.css';
</style>