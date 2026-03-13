<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useSidebarStore } from '@/stores/sidebar'
import { useAuthStore } from '@/stores/auth'
import {
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  LogoutOutlined,
  KeyOutlined,
} from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import ChangePasswordModal from '@/components/ChangePasswordModal.vue'

const router = useRouter()
const sidebarStore = useSidebarStore()
const authStore = useAuthStore()

const showDropdown = ref(false)
const showChangePassword = ref(false)

// 获取用户信息
onMounted(async () => {
  if (authStore.isAuthenticated) {
    await authStore.fetchUser()
  }
})

// 获取用户头像显示的首字母
const avatarText = () => {
  if (authStore.user?.username) {
    return authStore.user.username.charAt(0).toUpperCase()
  }
  return 'A'
}

// 处理菜单点击
async function handleMenuClick(key: string) {
  showDropdown.value = false
  if (key === 'logout') {
    authStore.logout()
    message.success('已退出登录')
    router.push('/login')
  } else if (key === 'password') {
    showChangePassword.value = true
  }
}

// 密码修改成功后
function onPasswordChanged() {
  router.push('/login')
}
</script>

<template>
  <a-layout-header class="app-header">
    <div class="header-left">
      <span class="collapse-btn" @click="sidebarStore.toggleCollapsed">
        <MenuUnfoldOutlined v-if="sidebarStore.collapsed" />
        <MenuFoldOutlined v-else />
      </span>
      <span class="logo">BeepBot</span>
    </div>
    <div class="header-right">
      <a-dropdown v-model:open="showDropdown" placement="bottomRight">
        <div class="user-avatar">
          <a-avatar :style="{ backgroundColor: '#667eea', cursor: 'pointer' }">
            {{ avatarText() }}
          </a-avatar>
          <span class="username">{{ authStore.user?.username || 'Admin' }}</span>
        </div>
        <template #overlay>
          <a-menu @click="(e: { key: string }) => handleMenuClick(e.key)">
            <a-menu-item key="password">
              <KeyOutlined />
              <span style="margin-left: 8px">修改密码</span>
            </a-menu-item>
            <a-menu-divider />
            <a-menu-item key="logout">
              <LogoutOutlined />
              <span style="margin-left: 8px">退出登录</span>
            </a-menu-item>
          </a-menu>
        </template>
      </a-dropdown>
    </div>

    <!-- 修改密码对话框 -->
    <ChangePasswordModal v-model:visible="showChangePassword" @success="onPasswordChanged" />
  </a-layout-header>
</template>

<style scoped>
.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  background: var(--header-bg);
  border-bottom: 1px solid var(--border-color);
  height: var(--header-height);
  line-height: var(--header-height);
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-right {
  display: flex;
  align-items: center;
}

.collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  font-size: 18px;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.3s ease;
}

.collapse-btn:hover {
  background-color: var(--nav-item-hover-bg);
}

.logo {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-color);
}

.user-avatar {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.3s ease;
}

.user-avatar:hover {
  background-color: var(--nav-item-hover-bg);
}

.username {
  color: var(--text-color);
  font-size: 14px;
}
</style>