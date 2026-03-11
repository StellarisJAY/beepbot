<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import { Card, Table, Tag, Empty, Button, InputNumber, type TableProps } from 'ant-design-vue'
import { ReloadOutlined, CompressOutlined } from '@ant-design/icons-vue'
import { sessionApi } from '@/api/session'
import type { SessionListItem } from '@/types/session'

const route = useRoute()
const router = useRouter()
const agentId = route.params.id as string

// 数据
const sessions = ref<SessionListItem[]>([])
const loading = ref(false)
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})

// 表格列定义
const columns: TableProps['columns'] = [
  {
    title: '会话标识',
    dataIndex: 'key',
    key: 'key',
    ellipsis: true,
  },
  {
    title: '机器人',
    dataIndex: 'bot_name',
    key: 'bot_name',
  },
  {
    title: '平台',
    dataIndex: 'bot_platform',
    key: 'bot_platform',
  },
  {
    title: '消息数',
    dataIndex: 'message_count',
    key: 'message_count',
    width: 80,
    align: 'center',
  },
  {
    title: '上下文Token',
    dataIndex: 'last_context_tokens',
    key: 'last_context_tokens',
    width: 110,
    align: 'right',
  },
  {
    title: 'Token用量',
    dataIndex: 'total_tokens',
    key: 'total_tokens',
    width: 100,
    align: 'right',
  },
  {
    title: '摘要',
    dataIndex: 'summary',
    key: 'summary',
    ellipsis: true,
  },
  {
    title: '更新时间',
    dataIndex: 'updated_at',
    key: 'updated_at',
    width: 160,
  },
]

// 获取会话列表
const fetchSessions = async () => {
  loading.value = true
  try {
    const res = await sessionApi.getAgentSessions(
      agentId,
      pagination.value.current,
      pagination.value.pageSize,
    )
    sessions.value = res.data
    pagination.value.total = res.total
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取会话列表失败')
  } finally {
    loading.value = false
  }
}

// 分页变化
const handleTableChange: TableProps['onChange'] = (pag) => {
  pagination.value.current = pag.current || 1
  pagination.value.pageSize = pag.pageSize || 10
  fetchSessions()
}

// 刷新
const handleRefresh = () => {
  pagination.value.current = 1
  fetchSessions()
}

// 平台颜色映射
const getPlatformColor = (platform: string): string => {
  const colors: Record<string, string> = {
    qq: 'blue',
    feishu: 'green',
  }
  return colors[platform] || 'default'
}

// 格式化时间
const formatTime = (time: string): string => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

// 格式化 Token 数量（大数字使用 K 单位）
const formatTokenCount = (count: number): string => {
  if (count >= 1000) {
    return `${(count / 1000).toFixed(1)}K`
  }
  return count.toString()
}

// 点击行查看消息
const handleRowClick = (record: SessionListItem) => {
  router.push({
    name: 'session-messages',
    params: {
      id: agentId,
      sessionId: record.id,
    },
  })
}

onMounted(() => {
  fetchSessions()
})
</script>

<template>
  <div class="agent-logs">
    <Card title="会话记录" :bordered="false">
      <template #extra>
        <Button type="primary" :loading="loading" @click="handleRefresh">
          <template #icon><ReloadOutlined /></template>
          刷新
        </Button>
      </template>

      <Table
        :columns="columns"
        :data-source="sessions"
        :loading="loading"
        :pagination="{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total: number) => `共 ${total} 条`,
        }"
        row-key="id"
        @change="handleTableChange"
        :custom-row="(record: SessionListItem) => ({
          style: { cursor: 'pointer' },
          onClick: () => handleRowClick(record),
        })"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'bot_platform'">
            <Tag :color="getPlatformColor(record.bot_platform)">
              {{ record.bot_platform?.toUpperCase() || '-' }}
            </Tag>
          </template>
          <template v-else-if="column.key === 'message_count'">
            {{ record.message_count ?? 0 }}
          </template>
          <template v-else-if="column.key === 'last_context_tokens'">
            <span :style="{ color: record.last_context_tokens > 3000 ? '#ff4d4f' : 'inherit' }">
              {{ formatTokenCount(record.last_context_tokens ?? 0) }}
            </span>
          </template>
          <template v-else-if="column.key === 'total_tokens'">
            {{ formatTokenCount(record.total_tokens ?? 0) }}
          </template>
          <template v-else-if="column.key === 'summary'">
            {{ record.summary || '-' }}
          </template>
          <template v-else-if="column.key === 'updated_at'">
            {{ formatTime(record.updated_at) }}
          </template>
        </template>
        <template #emptyText>
          <Empty description="暂无会话记录" />
        </template>
      </Table>
    </Card>
  </div>
</template>

<style scoped>
.agent-logs {
  padding: 0;
}
</style>