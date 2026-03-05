import { ref, computed, watch } from 'vue'
import { defineStore } from 'pinia'
import type { ThemeConfig } from 'ant-design-vue/es/config-provider/context'

export type ThemeMode = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  // 从 localStorage 读取主题偏好
  const getStoredTheme = (): ThemeMode => {
    const stored = localStorage.getItem('beepbot-theme')
    if (stored === 'light' || stored === 'dark') {
      return stored
    }
    // 检测系统主题偏好
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      return 'dark'
    }
    return 'light'
  }

  const themeMode = ref<ThemeMode>(getStoredTheme())

  // 是否为深色主题
  const isDark = computed(() => themeMode.value === 'dark')

  // Ant Design Vue 主题配置
  const themeConfig = computed(() => ({
    token: {
      colorPrimary: '#1890ff',
      borderRadius: 8,
    },
    components: {
      Layout: {
        headerBg: isDark.value ? '#1f1f1f' : '#ffffff',
        siderBg: isDark.value ? '#1f1f1f' : '#ffffff',
        bodyBg: isDark.value ? '#141414' : '#f5f5f5',
      },
      Menu: {
        darkItemBg: isDark.value ? '#1f1f1f' : '#ffffff',
        darkItemSelectedBg: isDark.value ? '#1890ff' : '#e6f7ff',
        itemBg: isDark.value ? '#1f1f1f' : '#ffffff',
        itemSelectedBg: isDark.value ? '#1890ff20' : '#e6f7ff',
        itemHoverBg: isDark.value ? '#ffffff10' : '#f0f0f0',
      },
      Card: {
        colorBgContainer: isDark.value ? '#1f1f1f' : '#ffffff',
      },
    },
  }))

  // 切换主题
  const toggleTheme = () => {
    themeMode.value = isDark.value ? 'light' : 'dark'
  }

  // 设置主题
  const setTheme = (mode: ThemeMode) => {
    themeMode.value = mode
  }

  // 监听主题变化，保存到 localStorage 并更新 DOM
  watch(
    themeMode,
    (newTheme) => {
      localStorage.setItem('beepbot-theme', newTheme)
      // 更新 html 元素的 data-theme 属性
      document.documentElement.setAttribute('data-theme', newTheme)
      // 更新 html 元素的 class
      document.documentElement.classList.remove('light', 'dark')
      document.documentElement.classList.add(newTheme)
    },
    { immediate: true }
  )

  return {
    themeMode,
    isDark,
    themeConfig,
    toggleTheme,
    setTheme,
  }
})