<script setup lang="ts">
import { ref } from 'vue'
import { PlusOutlined, EditOutlined, DeleteOutlined, MessageOutlined } from '@ant-design/icons-vue'

// 模拟数据
const bots = ref([
  {
    id: '1',
    name: 'QQ 官方机器人',
    platform: 'qq',
    appId: '1024xxxxx',
    status: 'active',
    description: '通过 QQ 机器人 API 提供服务',
  },
  {
    id: '2',
    name: '控制台机器人',
    platform: 'console',
    appId: '-',
    status: 'active',
    description: '通过命令行控制台进行交互',
  },
])

// 平台名称映射
const platformNames: Record<string, string> = {
  qq: 'QQ',
  console: '控制台',
  discord: 'Discord',
  telegram: 'Telegram',
}

// 平台颜色映射
const platformColors: Record<string, string> = {
  qq: 'blue',
  console: 'green',
  discord: 'purple',
  telegram: 'cyan',
}

// 新建机器人
const handleCreate = () => {
  console.log('创建机器人')
}

// 编辑机器人
const handleEdit = (id: string) => {
  console.log('编辑机器人:', id)
}

// 删除机器人
const handleDelete = (id: string) => {
  console.log('删除机器人:', id)
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">IM机器人</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建机器人
      </a-button>
    </div>

    <div class="card-grid">
      <a-card v-for="bot in bots" :key="bot.id" class="bot-card" hoverable>
        <template #title>
          <div class="card-title">
            <MessageOutlined class="card-icon" />
            <span>{{ bot.name }}</span>
          </div>
        </template>
        <template #extra>
          <a-tag :color="bot.status === 'active' ? 'green' : 'default'">
            {{ bot.status === 'active' ? '在线' : '离线' }}
          </a-tag>
        </template>
        <p class="card-description">{{ bot.description }}</p>
        <div class="card-meta">
          <div class="meta-item">
            <span class="meta-label">平台:</span>
            <a-tag :color="platformColors[bot.platform] || 'default'">
              {{ platformNames[bot.platform] || bot.platform }}
            </a-tag>
          </div>
          <div class="meta-item" v-if="bot.appId !== '-'">
            <span class="meta-label">App ID:</span>
            <span class="meta-value">{{ bot.appId }}</span>
          </div>
        </div>
        <template #actions>
          <a-tooltip title="编辑">
            <EditOutlined @click="handleEdit(bot.id)" />
          </a-tooltip>
          <a-tooltip title="删除">
            <DeleteOutlined @click="handleDelete(bot.id)" />
          </a-tooltip>
        </template>
      </a-card>
    </div>

    <a-empty v-if="bots.length === 0" description="暂无机器人，点击上方按钮创建" />
  </div>
</template>

<style scoped>
.page-container {
  padding: 24px;
  min-height: 100%;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-color);
  margin: 0;
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.bot-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  transition: all 0.3s ease;
}

.bot-card:hover {
  box-shadow: var(--card-hover-shadow);
  transform: translateY(-2px);
}

/* 卡片操作按钮样式 */
.bot-card :deep(.ant-card-actions) {
  border-top: 1px solid var(--border-color);
}

.bot-card :deep(.ant-card-actions > li) {
  color: var(--card-action-color);
  transition: color 0.3s ease;
}

.bot-card :deep(.ant-card-actions > li:hover) {
  color: var(--card-action-hover-color);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.card-icon {
  font-size: 18px;
  color: var(--color-primary);
}

.card-description {
  color: var(--text-secondary);
  font-size: 14px;
  line-height: 1.6;
  margin-bottom: 12px;
}

.card-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  font-size: 13px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-label {
  color: var(--text-tertiary);
}

.meta-value {
  color: var(--text-secondary);
  font-family: monospace;
}
</style>