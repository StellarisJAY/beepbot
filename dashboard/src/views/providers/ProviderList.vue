<script setup lang="ts">
import { ref } from 'vue'
import { PlusOutlined, EditOutlined, DeleteOutlined, ApiOutlined } from '@ant-design/icons-vue'

// 模拟数据
const providers = ref([
  {
    id: '1',
    name: 'DashScope',
    type: 'openai-compatible',
    baseUrl: 'https://dashscope.aliyuncs.com/api/v2/apps/protocols/compatible-mode/v1',
    models: ['qwen3.5-plus', 'qwen3.5-turbo', 'qwen-max'],
    status: 'active',
  },
  {
    id: '2',
    name: 'OpenAI',
    type: 'openai',
    baseUrl: 'https://api.openai.com/v1',
    models: ['gpt-4o', 'gpt-4o-mini', 'gpt-3.5-turbo'],
    status: 'active',
  },
  {
    id: '3',
    name: 'DeepSeek',
    type: 'openai-compatible',
    baseUrl: 'https://api.deepseek.com/v1',
    models: ['deepseek-chat', 'deepseek-coder'],
    status: 'inactive',
  },
])

// 新建供应商
const handleCreate = () => {
  console.log('创建供应商')
}

// 编辑供应商
const handleEdit = (id: string) => {
  console.log('编辑供应商:', id)
}

// 删除供应商
const handleDelete = (id: string) => {
  console.log('删除供应商:', id)
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">模型供应商</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建供应商
      </a-button>
    </div>

    <div class="card-grid">
      <a-card v-for="provider in providers" :key="provider.id" class="provider-card" hoverable>
        <template #title>
          <div class="card-title">
            <ApiOutlined class="card-icon" />
            <span>{{ provider.name }}</span>
          </div>
        </template>
        <template #extra>
          <a-tag :color="provider.status === 'active' ? 'green' : 'default'">
            {{ provider.status === 'active' ? '已连接' : '未连接' }}
          </a-tag>
        </template>
        <div class="card-info">
          <div class="info-item">
            <span class="info-label">类型:</span>
            <span class="info-value">{{ provider.type }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">地址:</span>
            <span class="info-value url">{{ provider.baseUrl }}</span>
          </div>
        </div>
        <div class="model-tags">
          <a-tag v-for="model in provider.models.slice(0, 3)" :key="model" color="blue">
            {{ model }}
          </a-tag>
          <a-tag v-if="provider.models.length > 3">+{{ provider.models.length - 3 }}</a-tag>
        </div>
        <template #actions>
          <a-tooltip title="编辑">
            <EditOutlined @click="handleEdit(provider.id)" />
          </a-tooltip>
          <a-tooltip title="删除">
            <DeleteOutlined @click="handleDelete(provider.id)" />
          </a-tooltip>
        </template>
      </a-card>
    </div>

    <a-empty v-if="providers.length === 0" description="暂无供应商，点击上方按钮创建" />
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

.provider-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  transition: all 0.3s ease;
}

.provider-card:hover {
  box-shadow: var(--card-hover-shadow);
  transform: translateY(-2px);
}

/* 卡片操作按钮样式 */
.provider-card :deep(.ant-card-actions) {
  border-top: 1px solid var(--border-color);
}

.provider-card :deep(.ant-card-actions > li) {
  color: var(--card-action-color);
  transition: color 0.3s ease;
}

.provider-card :deep(.ant-card-actions > li:hover) {
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

.card-info {
  margin-bottom: 12px;
}

.info-item {
  display: flex;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 13px;
}

.info-label {
  color: var(--text-tertiary);
  flex-shrink: 0;
}

.info-value {
  color: var(--text-secondary);
  word-break: break-all;
}

.info-value.url {
  font-family: monospace;
  font-size: 12px;
}

.model-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
</style>