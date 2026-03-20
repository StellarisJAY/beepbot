<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  PlusOutlined,
  TeamOutlined,
  SearchOutlined,
  ReloadOutlined,
  MoreOutlined,
  DeleteOutlined,
  UserOutlined,
} from '@ant-design/icons-vue'
import { teamApi, type TeamFilter } from '@/api/team'
import type { Team } from '@/types/team'
import { TeamStatus, TeamStatusOptions } from '@/types/team'
import TeamCreateModal from '@/components/TeamCreateModal.vue'

const router = useRouter()

// 数据
const teams = ref<Team[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 筛选条件
const filters = ref<TeamFilter>({
  name: '',
  status: undefined,
})

// 弹窗
const createModalVisible = ref(false)

// 获取团队列表
const fetchTeams = async () => {
  loading.value = true
  try {
    // 构建筛选参数，过滤掉空值
    const filterParams: TeamFilter = {}
    if (filters.value.name) filterParams.name = filters.value.name
    if (filters.value.status) filterParams.status = filters.value.status

    const res = await teamApi.list(page.value, size.value, filterParams)
    teams.value = res.data
    total.value = res.total
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取团队列表失败')
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  page.value = 1
  fetchTeams()
}

// 重置筛选
const handleReset = () => {
  filters.value = { name: '', status: undefined }
  page.value = 1
  fetchTeams()
}

// 新建团队
const handleCreate = () => {
  createModalVisible.value = true
}

// 创建成功后跳转编辑页
const handleCreateSuccess = (id: string) => {
  router.push(`/teams/${id}/edit`)
}

// 点击卡片进入编辑页
const handleCardClick = (id: string) => {
  router.push(`/teams/${id}/edit`)
}

// 删除团队
const handleDelete = (team: Team, e?: Event) => {
  e?.stopPropagation()
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除团队「${team.name}」吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await teamApi.delete(team.id)
        message.success('删除成功')
        await fetchTeams()
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || '删除失败')
      }
    },
  })
}

// 切换状态
const handleToggleStatus = async (team: Team, e?: Event) => {
  e?.stopPropagation()

  const newStatus = team.status === TeamStatus.Active ? TeamStatus.Inactive : TeamStatus.Active
  try {
    await teamApi.updateStatus(team.id, newStatus)
    message.success('状态更新成功')
    await fetchTeams()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '状态更新失败')
  }
}

// 分页变化
const handlePageChange = (p: number, s: number) => {
  page.value = p
  size.value = s
  fetchTeams()
}

// 阻止下拉菜单点击冒泡
const handleDropdownClick = (e: Event) => {
  e.stopPropagation()
}

// 获取成员数量
const getMemberCount = (team: Team) => {
  return (team.members?.length || 0) + 1 // +1 for leader
}

onMounted(() => {
  fetchTeams()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">团队管理</h1>
      <a-button type="primary" @click="handleCreate">
        <template #icon><PlusOutlined /></template>
        新建团队
      </a-button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <a-input
        v-model:value="filters.name"
        placeholder="搜索名称"
        style="width: 200px"
        allow-clear
        @pressEnter="handleSearch"
      >
        <template #prefix><SearchOutlined /></template>
      </a-input>
      <a-select
        v-model:value="filters.status"
        placeholder="状态"
        style="width: 120px"
        allow-clear
        :options="TeamStatusOptions"
      />
      <a-button type="primary" @click="handleSearch">
        <template #icon><SearchOutlined /></template>
        查询
      </a-button>
      <a-button @click="handleReset">
        <template #icon><ReloadOutlined /></template>
        重置
      </a-button>
    </div>

    <a-spin :spinning="loading">
      <div class="card-grid" v-if="teams.length > 0">
        <a-card
          v-for="team in teams"
          :key="team.id"
          class="team-card"
          hoverable
          @click="handleCardClick(team.id)"
        >
          <template #title>
            <div class="card-title">
              <TeamOutlined class="card-icon" />
              <span>{{ team.name }}</span>
            </div>
          </template>
          <template #extra>
            <a-switch
              :checked="team.status === TeamStatus.Active"
              checked-children="启用"
              un-checked-children="禁用"
              @change="handleToggleStatus(team, $event)"
            />
          </template>
          <p class="card-description">{{ team.description || '暂无描述' }}</p>
          <div class="team-info">
            <div class="leader-info" v-if="team.leader">
              <UserOutlined class="leader-icon" />
              <span class="leader-label">Leader:</span>
              <span class="leader-name">{{ team.leader.member_name }}</span>
            </div>
            <div class="member-count">
              <TeamOutlined />
              <span>{{ getMemberCount(team) }} 名成员</span>
            </div>
          </div>
          <div class="card-footer">
            <a-dropdown :trigger="['click']" @click.stop="handleDropdownClick">
              <a-button class="more-btn" @click.stop>
                <MoreOutlined />
              </a-button>
              <template #overlay>
                <a-menu>
                  <a-menu-item key="delete" danger @click="handleDelete(team, $event)">
                    <DeleteOutlined />
                    <span>删除</span>
                  </a-menu-item>
                </a-menu>
              </template>
            </a-dropdown>
          </div>
        </a-card>
      </div>

      <a-empty v-else description="暂无团队，点击上方按钮创建" />

      <!-- 分页 -->
      <div class="pagination-container" v-if="total > size">
        <a-pagination
          v-model:current="page"
          v-model:pageSize="size"
          :total="total"
          :show-size-changer="true"
          :show-quick-jumper="true"
          :show-total="(t: number) => `共 ${t} 条`"
          @change="handlePageChange"
        />
      </div>
    </a-spin>

    <!-- 创建弹窗 -->
    <TeamCreateModal
      v-model:visible="createModalVisible"
      @success="handleCreateSuccess"
    />
  </div>
</template>

<style scoped>
.page-container {
  padding: 24px;
  height: 100%;
  overflow-y: auto;
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

.filter-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: var(--card-bg);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
}

.team-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  transition: all 0.3s ease;
  cursor: pointer;
}

.team-card:hover {
  box-shadow: var(--card-hover-shadow);
  transform: translateY(-2px);
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
  min-height: 44px;
}

.team-info {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
  padding: 12px;
  background: var(--hover-bg);
  border-radius: 6px;
}

.leader-info {
  display: flex;
  align-items: center;
  gap: 6px;
}

.leader-icon {
  color: var(--color-primary);
}

.leader-label {
  color: var(--text-secondary);
  font-size: 13px;
}

.leader-name {
  font-weight: 500;
  color: var(--text-color);
}

.member-count {
  display: flex;
  align-items: center;
  gap: 6px;
  color: var(--text-secondary);
  font-size: 13px;
}

.card-footer {
  display: flex;
  justify-content: flex-end;
  padding-top: 12px;
  border-top: 1px solid var(--border-color);
}

.more-btn {
  padding: 4px 8px;
  border: none !important;
  background: transparent !important;
  color: var(--text-secondary);
  border-radius: 4px;
  transition: all 0.2s ease;
}

.more-btn:hover {
  color: var(--color-primary) !important;
  background: var(--hover-bg) !important;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
}
</style>