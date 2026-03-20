<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useSidebarStore } from '@/stores/sidebar'
import {
  RobotOutlined,
  ApiOutlined,
  MessageOutlined,
  ClockCircleOutlined,
  BookOutlined,
  CloudServerOutlined,
  TeamOutlined,
} from '@ant-design/icons-vue'

const route = useRoute()
const router = useRouter()
const sidebarStore = useSidebarStore()

// 导航菜单项
const menuItems = [
  {
    key: '/agents',
    icon: RobotOutlined,
    title: '智能体',
  },
  {
    key: '/providers',
    icon: ApiOutlined,
    title: '模型供应商',
  },
  {
    key: '/bots',
    icon: MessageOutlined,
    title: 'IM机器人',
  },
  {
    key: '/teams',
    icon: TeamOutlined,
    title: '团队管理',
  },
  {
    key: '/crons',
    icon: ClockCircleOutlined,
    title: '定时任务',
  },
  {
    key: '/skills',
    icon: BookOutlined,
    title: '技能管理',
  },
  {
    key: '/mcp',
    icon: CloudServerOutlined,
    title: 'MCP 服务器',
  },
]

// 当前选中的菜单
const selectedKeys = computed(() => {
  const path = route.path
  // 匹配当前路由
  const matchedItem = menuItems.find((item) => path.startsWith(item.key))
  return matchedItem ? [matchedItem.key] : ['/agents']
})

// 点击菜单项
const handleMenuClick = ({ key }: { key: string }) => {
  router.push(key)
}
</script>

<template>
  <a-layout-sider
    :collapsed="sidebarStore.collapsed"
    :collapsed-width="sidebarStore.collapsedWidth"
    :width="sidebarStore.expandedWidth"
    :trigger="null"
    collapsible
    class="app-sidebar"
  >
    <a-menu
      mode="inline"
      :selected-keys="selectedKeys"
      :style="{ borderRight: 0 }"
      @click="handleMenuClick"
    >
      <a-menu-item v-for="item in menuItems" :key="item.key" :title="item.title">
        <component :is="item.icon" />
        <span class="menu-title">{{ item.title }}</span>
      </a-menu-item>
    </a-menu>
  </a-layout-sider>
</template>

<style scoped>
.app-sidebar {
  background: var(--sidebar-bg);
  border-right: 1px solid var(--border-color);
  transition: all 0.2s ease;
  overflow: hidden;
}

.app-sidebar :deep(.ant-menu) {
  background: transparent;
  border: none;
}

.app-sidebar :deep(.ant-menu-item) {
  margin: 4px 8px;
  border-radius: 6px;
  height: 44px;
  line-height: 44px;
  color: var(--text-color);
  transition: all 0.3s ease;
}

.app-sidebar :deep(.ant-menu-item:hover) {
  background-color: var(--nav-item-hover-bg);
}

.app-sidebar :deep(.ant-menu-item-selected) {
  background-color: var(--nav-item-active-bg);
  color: var(--nav-item-active-color);
}

.app-sidebar :deep(.ant-menu-item-selected::after) {
  display: none;
}

.menu-title {
  margin-left: 10px;
}

/* 折叠状态下的菜单项样式 */
.app-sidebar.ant-layout-sider-collapsed :deep(.ant-menu-item) {
  padding: 0 !important;
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-sidebar.ant-layout-sider-collapsed :deep(.ant-menu-item .ant-tooltip) {
  display: flex;
  align-items: center;
  justify-content: center;
}

.app-sidebar.ant-layout-sider-collapsed :deep(.ant-menu-item .anticon) {
  margin-right: 0;
}

.app-sidebar.ant-layout-sider-collapsed :deep(.menu-title) {
  display: none;
}
</style>