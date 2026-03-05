import { ref, computed, watch } from 'vue'
import { defineStore } from 'pinia'

export const useSidebarStore = defineStore('sidebar', () => {
  // 从 localStorage 读取折叠状态
  const getStoredCollapsed = (): boolean => {
    const stored = localStorage.getItem('beepbot-sidebar-collapsed')
    return stored === 'true'
  }

  const collapsed = ref<boolean>(getStoredCollapsed())

  // 侧边栏宽度
  const collapsedWidth = 64
  const expandedWidth = 200

  // 当前宽度
  const width = computed(() => (collapsed.value ? collapsedWidth : expandedWidth))

  // 切换折叠状态
  const toggleCollapsed = () => {
    collapsed.value = !collapsed.value
  }

  // 设置折叠状态
  const setCollapsed = (value: boolean) => {
    collapsed.value = value
  }

  // 监听变化，保存到 localStorage
  watch(collapsed, (newValue) => {
    localStorage.setItem('beepbot-sidebar-collapsed', String(newValue))
  })

  return {
    collapsed,
    collapsedWidth,
    expandedWidth,
    width,
    toggleCollapsed,
    setCollapsed,
  }
})