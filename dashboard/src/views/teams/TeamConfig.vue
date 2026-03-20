<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import {
  SaveOutlined,
  ArrowLeftOutlined,
  UserOutlined,
  PlusOutlined,
  DeleteOutlined,
} from '@ant-design/icons-vue'
import { teamApi } from '@/api/team'
import { agentApi } from '@/api/agent'
import type { Team, MemberRequest } from '@/types/team'
import type { Agent } from '@/types/agent'
import { TeamStatus, TeamStatusOptions } from '@/types/team'

const route = useRoute()
const router = useRouter()

// 数据
const team = ref<Team | null>(null)
const loading = ref(false)
const saving = ref(false)
const agents = ref<Agent[]>([])
const agentLoading = ref(false)

// 表单数据
const form = ref({
  name: '',
  description: '',
  leader_id: '',
  members: [] as MemberRequest[],
  status: TeamStatus.Inactive,
})

// 获取团队详情
const fetchTeam = async () => {
  const id = route.params.id as string
  if (!id) return

  loading.value = true
  try {
    const res = await teamApi.get(id)
    team.value = res.data
    // 填充表单
    form.value.name = res.data.name
    form.value.description = res.data.description || ''
    form.value.leader_id = res.data.leader_id
    form.value.status = res.data.status
    form.value.members = (res.data.members || []).map((m) => ({
      agent_id: m.agent_id,
      member_name: m.member_name,
      description: m.description || '',
    }))
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取团队详情失败')
  } finally {
    loading.value = false
  }
}

// 获取智能体列表
const fetchAgents = async () => {
  agentLoading.value = true
  try {
    const res = await agentApi.list(1, 100)
    agents.value = res.data
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取智能体列表失败')
  } finally {
    agentLoading.value = false
  }
}

// 获取 Leader 名称（用于名称重复检查）
const getLeaderName = computed(() => {
  return team.value?.leader?.member_name || ''
})

// 已使用的成员名称（用于检查重复）
const usedMemberNames = computed(() => {
  const names = new Set<string>()
  // Leader 名称
  if (getLeaderName.value) {
    names.add(getLeaderName.value)
  }
  form.value.members.forEach((m) => names.add(m.member_name))
  return names
})

// 新成员表单
const newMember = ref<MemberRequest>({
  agent_id: '',
  member_name: '',
  description: '',
})

// 添加成员
const handleAddMember = () => {
  if (!newMember.value.agent_id || !newMember.value.member_name) {
    message.warning('请选择智能体并填写成员名称')
    return
  }

  // 检查成员名称是否重复
  if (usedMemberNames.value.has(newMember.value.member_name)) {
    message.warning('成员名称已存在，请使用不同的名称')
    return
  }

  form.value.members.push({ ...newMember.value })
  newMember.value = { agent_id: '', member_name: '', description: '' }
}

// 移除成员
const handleRemoveMember = (index: number) => {
  form.value.members.splice(index, 1)
}

// 当选择智能体时自动填充成员名称
const handleAgentSelect = (agentId: string) => {
  const agent = agents.value.find((a) => a.id === agentId)
  if (agent) {
    newMember.value.member_name = agent.name
  }
}

// 保存
const handleSave = async () => {
  if (!form.value.name) {
    message.warning('请输入团队名称')
    return
  }
  if (!form.value.leader_id) {
    message.warning('请选择 Leader 智能体')
    return
  }

  saving.value = true
  try {
    await teamApi.update(route.params.id as string, {
      name: form.value.name,
      description: form.value.description,
      leader_id: form.value.leader_id,
      members: form.value.members,
      status: form.value.status,
    })
    message.success('保存成功')
    await fetchTeam()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '保存失败')
  } finally {
    saving.value = false
  }
}

// 返回列表
const handleBack = () => {
  router.push('/teams')
}

// 获取智能体名称
const getAgentName = (agentId: string) => {
  return agents.value.find((a) => a.id === agentId)?.name || ''
}

onMounted(() => {
  fetchTeam()
  fetchAgents()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <a-button type="text" @click="handleBack">
        <template #icon><ArrowLeftOutlined /></template>
        返回
      </a-button>
      <h1 class="page-title">{{ team?.name || '团队配置' }}</h1>
      <a-button type="primary" :loading="saving" @click="handleSave">
        <template #icon><SaveOutlined /></template>
        保存
      </a-button>
    </div>

    <a-spin :spinning="loading">
      <div class="form-container">
        <a-form :model="form" layout="vertical">
          <a-card title="基本信息" class="form-card">
            <a-row :gutter="24">
              <a-col :span="12">
                <a-form-item label="团队名称" required>
                  <a-input v-model:value="form.name" placeholder="输入团队名称" />
                </a-form-item>
              </a-col>
              <a-col :span="12">
                <a-form-item label="状态">
                  <a-select v-model:value="form.status" :options="TeamStatusOptions" />
                </a-form-item>
              </a-col>
            </a-row>

            <a-form-item label="团队描述">
              <a-textarea
                v-model:value="form.description"
                placeholder="输入团队描述"
                :rows="2"
              />
            </a-form-item>
          </a-card>

          <a-card title="团队成员" class="form-card">
            <template #extra>
              <span class="member-count">
                共 {{ form.members.length + 1 }} 名成员（含 Leader）
              </span>
            </template>

            <!-- Leader 展示 -->
            <div class="leader-section">
              <div class="section-title">
                <UserOutlined class="leader-icon" />
                <span>Leader</span>
              </div>
              <div class="leader-card" v-if="team?.leader">
                <div class="leader-info">
                  <span class="leader-name">{{ team.leader.member_name }}</span>
                  <span class="leader-agent">{{ team.leader.agent_name }}</span>
                </div>
              </div>
            </div>

            <!-- 队员列表 -->
            <div class="members-section">
              <div class="section-title">
                <span>队员</span>
              </div>

              <div class="member-list" v-if="form.members.length > 0">
                <div
                  v-for="(member, index) in form.members"
                  :key="`${member.agent_id}-${member.member_name}`"
                  class="member-item"
                >
                  <div class="member-info">
                    <div class="member-main">
                      <span class="member-name">{{ member.member_name }}</span>
                      <span class="member-agent-name">({{ getAgentName(member.agent_id) }})</span>
                    </div>
                    <span class="member-desc" v-if="member.description">{{ member.description }}</span>
                  </div>
                  <a-button type="text" danger size="small" @click="handleRemoveMember(index)">
                    <template #icon><DeleteOutlined /></template>
                  </a-button>
                </div>
              </div>

              <!-- 添加队员 -->
              <div class="add-member-form">
                <a-select
                  v-model:value="newMember.agent_id"
                  placeholder="选择队员智能体"
                  :loading="agentLoading"
                  style="width: 200px"
                  show-search
                  :filter-option="
                    (input: string, option: { label: string }) =>
                      option.label.toLowerCase().includes(input.toLowerCase())
                  "
                  @change="(val: string) => handleAgentSelect(val)"
                >
                  <a-select-option
                    v-for="agent in agents"
                    :key="agent.id"
                    :value="agent.id"
                    :label="agent.name"
                  >
                    {{ agent.name }}
                  </a-select-option>
                </a-select>
                <a-input
                  v-model:value="newMember.member_name"
                  placeholder="成员名称（必填）"
                  style="width: 150px"
                  :status="newMember.member_name && usedMemberNames.has(newMember.member_name) ? 'error' : ''"
                />
                <a-input
                  v-model:value="newMember.description"
                  placeholder="角色描述（可选）"
                  style="flex: 1"
                />
                <a-button type="dashed" @click="handleAddMember">
                  <template #icon><PlusOutlined /></template>
                  添加
                </a-button>
              </div>
              <div class="form-hint error" v-if="newMember.member_name && usedMemberNames.has(newMember.member_name)">
                成员名称已存在
              </div>
            </div>
          </a-card>
        </a-form>
      </div>
    </a-spin>
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
  flex: 1;
  text-align: center;
}

.form-container {
  max-width: 800px;
  margin: 0 auto;
}

.form-card {
  margin-bottom: 16px;
  border-radius: 8px;
}

.member-count {
  color: var(--text-secondary);
  font-size: 14px;
}

.leader-section,
.members-section {
  margin-bottom: 24px;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 15px;
  margin-bottom: 12px;
  color: var(--text-color);
}

.leader-icon {
  color: var(--color-primary);
}

.leader-card {
  padding: 16px;
  background: var(--hover-bg);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.leader-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.leader-name {
  font-weight: 600;
  font-size: 16px;
}

.leader-agent {
  color: var(--text-secondary);
  font-size: 13px;
}

.member-list {
  margin-bottom: 16px;
}

.member-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--hover-bg);
  border-radius: 6px;
  margin-bottom: 8px;
  border: 1px solid var(--border-color);
}

.member-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.member-main {
  display: flex;
  align-items: center;
  gap: 8px;
}

.member-name {
  font-weight: 500;
}

.member-agent-name {
  color: var(--text-secondary);
  font-size: 13px;
}

.member-desc {
  color: var(--text-secondary);
  font-size: 13px;
}

.add-member-form {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  padding: 16px;
  background: var(--hover-bg);
  border-radius: 8px;
  border: 1px dashed var(--border-color);
}

.form-hint {
  color: var(--text-secondary);
  font-size: 12px;
  margin-top: 4px;
}

.form-hint.error {
  color: var(--color-error);
}
</style>