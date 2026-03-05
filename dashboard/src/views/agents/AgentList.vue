<script setup lang="ts">
import { ref } from 'vue'
import { PlusOutlined, EditOutlined, DeleteOutlined, RobotOutlined } from '@ant-design/icons-vue'

// 模拟数据
const agents = ref([
  {
    id: '1',
    name: '默认智能体',
    description: '基于 Qwen3.5 的通用智能助手，支持多轮对话和工具调用',
    model: 'qwen3.5-plus',
    provider: 'dashscope',
    status: 'active',
  },
  {
    id: '2',
    name: '代码助手',
    description: '专注于代码编写和调试的智能助手',
    model: 'gpt-4o',
    provider: 'openai',
    status: 'active',
  },
  {
    id: '3',
    name: '文档助手',
    description: '帮助用户处理文档、撰写报告的智能助手',
    model: 'qwen3.5-plus',
    provider: 'dashscope',
    status: 'inactive',
  },
])

// 新建智能体
const handleCreate = () => {
  // TODO: 打开创建智能体弹窗
  console.log('创建智能体')
}

// 编辑智能体
const handleEdit = (id: string) => {
  // TODO: 打开编辑智能体弹窗
  console.log('编辑智能体:', id)
}

// 删除智能体
const handleDelete = (id: string) => {
  // TODO: 确认删除
  console.log('删除智能体:', id)
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">智能体</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建智能体
      </a-button>
    </div>

    <div class="card-grid">
      <a-card v-for="agent in agents" :key="agent.id" class="agent-card" hoverable>
        <template #title>
          <div class="card-title">
            <RobotOutlined class="card-icon" />
            <span>{{ agent.name }}</span>
          </div>
        </template>
        <template #extra>
          <a-tag :color="agent.status === 'active' ? 'green' : 'default'">
            {{ agent.status === 'active' ? '运行中' : '已停止' }}
          </a-tag>
        </template>
        <p class="card-description">{{ agent.description }}</p>
        <div class="card-meta">
          <span class="meta-item">
            <span class="meta-label">模型:</span>
            <span class="meta-value">{{ agent.model }}</span>
          </span>
          <span class="meta-item">
            <span class="meta-label">供应商:</span>
            <span class="meta-value">{{ agent.provider }}</span>
          </span>
        </div>
        <template #actions>
          <a-tooltip title="编辑">
            <EditOutlined @click="handleEdit(agent.id)" />
          </a-tooltip>
          <a-tooltip title="删除">
            <DeleteOutlined @click="handleDelete(agent.id)" />
          </a-tooltip>
        </template>
      </a-card>
    </div>

    <a-empty v-if="agents.length === 0" description="暂无智能体，点击上方按钮创建" />
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

.agent-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  transition: all 0.3s ease;
}

.agent-card:hover {
  box-shadow: var(--card-hover-shadow);
  transform: translateY(-2px);
}

/* 卡片操作按钮样式 */
.agent-card :deep(.ant-card-actions) {
  border-top: 1px solid var(--border-color);
}

.agent-card :deep(.ant-card-actions > li) {
  color: var(--card-action-color);
  transition: color 0.3s ease;
}

.agent-card :deep(.ant-card-actions > li:hover) {
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
  flex-wrap: wrap;
  gap: 12px;
  font-size: 12px;
}

.meta-item {
  display: flex;
  gap: 4px;
}

.meta-label {
  color: var(--text-tertiary);
}

.meta-value {
  color: var(--text-secondary);
}
</style>