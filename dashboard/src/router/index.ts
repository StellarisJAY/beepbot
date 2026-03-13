import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { title: '登录', public: true },
  },
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
        path: 'chat',
        name: 'AgentChat',
        component: () => import('@/views/agents/AgentChat.vue'),
        meta: { title: '调试聊天' },
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
    path: '/mcp',
    name: 'MCP',
    component: () => import('@/views/mcp/MCPList.vue'),
    meta: { title: 'MCP 服务器' },
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

// 路由守卫
router.beforeEach((to, _from, next) => {
  // 更新页面标题
  const title = to.meta.title as string
  if (title) {
    document.title = `${title} - BeepBot`
  } else {
    document.title = 'BeepBot'
  }

  // 检查是否需要认证
  const isPublic = to.meta.public as boolean
  const token = localStorage.getItem('beepbot_token')

  if (!isPublic && !token) {
    // 需要认证但没有 token，跳转到登录页
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && token) {
    // 已登录访问登录页，跳转到首页
    next({ path: '/' })
  } else {
    next()
  }
})

export default router