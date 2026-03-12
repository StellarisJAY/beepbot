<script setup lang="ts">
import { useSidebarStore } from '@/stores/sidebar'
import { MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons-vue'

const sidebarStore = useSidebarStore()
</script>

<template>
  <a-layout-header class="app-header">
    <div class="header-left">
      <span class="collapse-btn" @click="sidebarStore.toggleCollapsed">
        <MenuUnfoldOutlined v-if="sidebarStore.collapsed" />
        <MenuFoldOutlined v-else />
      </span>
      <span class="logo"></span>
    </div>
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
</style>