import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as authApi from '@/api/auth'
import type { AdminUser } from '@/types/auth'

const TOKEN_KEY = 'beepbot_token'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const user = ref<AdminUser | null>(null)

  const isAuthenticated = computed(() => !!token.value)

  // 登录
  async function login(username: string, password: string) {
    const response = await authApi.login({ username, password })
    token.value = response.token
    localStorage.setItem(TOKEN_KEY, response.token)
    return response
  }

  // 登出
  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem(TOKEN_KEY)
  }

  // 获取当前用户信息
  async function fetchUser() {
    if (!token.value) return null
    try {
      const userInfo = await authApi.getMe()
      user.value = userInfo
      return userInfo
    } catch {
      logout()
      return null
    }
  }

  // 修改密码
  async function changePassword(oldPassword: string, newPassword: string) {
    await authApi.changePassword({ old_password: oldPassword, new_password: newPassword })
  }

  // 修改用户名
  async function changeUsername(newUsername: string) {
    await authApi.changeUsername({ new_username: newUsername })
    if (user.value) {
      user.value.username = newUsername
    }
  }

  // 获取 Token
  function getToken() {
    return token.value
  }

  return {
    token,
    user,
    isAuthenticated,
    login,
    logout,
    fetchUser,
    changePassword,
    changeUsername,
    getToken,
  }
})