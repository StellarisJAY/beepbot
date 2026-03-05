<script setup lang="ts">
import { useThemeStore } from '@/stores/theme'
import AppHeader from './Header.vue'
import AppSidebar from './Sidebar.vue'

const themeStore = useThemeStore()
</script>

<template>
  <a-config-provider :theme="themeStore.themeConfig">
    <a-layout class="app-layout">
      <AppSidebar />
      <a-layout class="main-layout">
        <AppHeader />
        <a-layout-content class="main-content">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </a-layout-content>
      </a-layout>
    </a-layout>
  </a-config-provider>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
  background-color: var(--bg-color);
  transition: background-color 0.3s ease;
}

.main-layout {
  background-color: var(--bg-color);
  transition: background-color 0.3s ease;
}

.main-content {
  margin: 0;
  min-height: calc(100vh - var(--header-height));
  background-color: var(--bg-color);
  overflow: auto;
  transition: background-color 0.3s ease;
}

/* 页面切换动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>