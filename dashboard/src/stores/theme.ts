import { computed, watch } from 'vue'
import { defineStore } from 'pinia'
import type { ThemeConfig } from 'ant-design-vue/es/config-provider/context'

export type ThemeMode = 'light' | 'dark'

export const useThemeStore = defineStore('theme', () => {
  // 强制使用 light 主题
  const themeMode = computed(() => 'light' as ThemeMode)

  // 是否为深色主题
  const isDark = computed(() => false)

  // Ant Design Vue 主题配置
  const themeConfig = computed(() => ({
    token: {
      colorPrimary: '#1890ff',
      borderRadius: 8,
    },
    components: {
      Layout: {
        headerBg: '#ffffff',
        siderBg: '#ffffff',
        bodyBg: '#f5f5f5',
      },
      Menu: {
        darkItemBg: '#ffffff',
        darkItemSelectedBg: '#e6f7ff',
        itemBg: '#ffffff',
        itemSelectedBg: '#e6f7ff',
        itemHoverBg: '#f0f0f0',
      },
      Card: {
        colorBgContainer: '#ffffff',
      },
    },
  }))

  // 初始化时设置 light 主题
  watch(
    themeMode,
    (newTheme) => {
      localStorage.setItem('beepbot-theme', newTheme)
      document.documentElement.setAttribute('data-theme', newTheme)
      document.documentElement.classList.remove('light', 'dark')
      document.documentElement.classList.add(newTheme)
    },
    { immediate: true }
  )

  return {
    themeMode,
    isDark,
    themeConfig,
  }
})