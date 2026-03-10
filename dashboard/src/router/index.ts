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
    path: '/agents/:id',
    component: () => import('@/views/agents/AgentDetailLayout.vue'),
    children: [
      {
        path: '',
        redirect: (to) => `/agents/${to.params.id}/edit`,
      },
      {
        path: 'edit',
        name: 'AgentEdit',
        components: {
          default: () => import('@/views/agents/AgentConfig.vue'),
          header: () => import('@/views/agents/AgentConfig.vue'),
        },
        meta: { title: '编辑智能体' },
      },
      {
        path: 'logs',
        name: 'agent-logs',
        component: () => import('@/views/agents/AgentLogs.vue'),
        meta: { title: '智能体日志' },
      },
      {
        path: 'sessions/:sessionId',
        name: 'session-messages',
        component: () => import('@/views/agents/SessionMessages.vue'),
        meta: { title: '会话消息' },
      },
      {
        path: 'monitor',
        name: 'AgentMonitor',
        component: () => import('@/views/agents/AgentMonitor.vue'),
        meta: { title: '智能体使用监测' },
      },
    ],
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
    path: '/crons',
    name: 'Crons',
    component: () => import('@/views/crons/CronList.vue'),
    meta: { title: '定时任务' },
  },
  {
    path: '/skills',
    name: 'Skills',
    component: () => import('@/views/skills/SkillList.vue'),
    meta: { title: '技能管理' },
  },
  {
    path: '/skills/:id',
    name: 'SkillDetail',
    component: () => import('@/views/skills/SkillDetail.vue'),
    meta: { title: '技能详情' },
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
