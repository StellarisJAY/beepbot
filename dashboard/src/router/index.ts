import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/agents',
  },
  {
    path: '/agents',
    name: 'Agents',
    component: () => import('@/views/agents/AgentList.vue'),
    meta: { title: '智能体' },
  },
  {
    path: '/providers',
    name: 'Providers',
    component: () => import('@/views/providers/ProviderList.vue'),
    meta: { title: '模型供应商' },
  },
  {
    path: '/bots',
    name: 'Bots',
    component: () => import('@/views/bots/BotList.vue'),
    meta: { title: 'IM机器人' },
  },
  {
    path: '/settings',
    name: 'Settings',
    component: () => import('@/views/settings/Settings.vue'),
    meta: { title: '全局设置' },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/agents',
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
})

// 路由守卫 - 更新页面标题
router.beforeEach((to, _from, next) => {
  const title = to.meta.title as string
  if (title) {
    document.title = `${title} - BeepBot`
  } else {
    document.title = 'BeepBot'
  }
  next()
})

export default router
