<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { message } from 'ant-design-vue'
import { teamApi } from '@/api/team'
import { agentApi } from '@/api/agent'
import type { CreateTeamRequest, MemberRequest } from '@/types/team'
import type { Agent } from '@/types/agent'

const props = defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'success', id: string): void
}>()

// 表单数据
const form = ref<CreateTeamRequest>({
  name: '',
  description: '',
  leader_id: '',
  members: [],
})

const loading = ref(false)
const agents = ref<Agent[]>([])
const agentLoading = ref(false)

// 搜索智能体
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
  if (!form.value.leader_id) return ''
  const agent = agents.value.find((a) => a.id === form.value.leader_id)
  return agent?.name || ''
})

// 已使用的成员名称（用于检查重复）
const usedMemberNames = computed(() => {
  const names = new Set<string>()
  // Leader 使用 Agent 名称作为成员名称
  if (getLeaderName.value) {
    names.add(getLeaderName.value)
  }
  form.value.members.forEach((m) => names.add(m.member_name))
  return names
})

// 添加成员
const newMember = ref<MemberRequest>({
  agent_id: '',
  member_name: '',
  description: '',
})

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
    // 自动填充智能体名称作为成员名称
    newMember.value.member_name = agent.name
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!form.value.name) {
    message.warning('请输入团队名称')
    return
  }
  if (!form.value.leader_id) {
    message.warning('请选择 Leader 智能体')
    return
  }

  loading.value = true
  try {
    const res = await teamApi.create(form.value)
    message.success('创建成功')
    emit('success', res.data.id)
    handleClose()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '创建失败')
  } finally {
    loading.value = false
  }
}

// 关闭弹窗
const handleClose = () => {
  emit('update:visible', false)
  resetForm()
}

// 重置表单
const resetForm = () => {
  form.value = {
    name: '',
    description: '',
    leader_id: '',
    members: [],
  }
  newMember.value = { agent_id: '', member_name: '', description: '' }
}

// 监听弹窗打开
watch(
  () => props.visible,
  (val) => {
    if (val) {
      fetchAgents()
    }
  }
)
</script>

<template>
  <a-modal
    :open="visible"
    title="新建团队"
    :confirm-loading="loading"
    @ok="handleSubmit"
    @cancel="handleClose"
    width="600px"
  >
    <a-form :model="form" layout="vertical">
      <a-form-item label="团队名称" required>
        <a-input v-model:value="form.name" placeholder="输入团队名称" />
      </a-form-item>

      <a-form-item label="团队描述">
        <a-textarea
          v-model:value="form.description"
          placeholder="输入团队描述"
          :rows="2"
        />
      </a-form-item>

      <a-form-item label="Leader 智能体" required>
        <a-select
          v-model:value="form.leader_id"
          placeholder="选择 Leader 智能体"
          :loading="agentLoading"
          show-search
          :filter-option="
            (input: string, option: { label: string }) =>
              option.label.toLowerCase().includes(input.toLowerCase())
          "
        >
          <a-select-option
            v-for="agent in agents"
            :key="agent.id"
            :value="agent.id"
            :label="agent.name"
          >
            {{ agent.name }}
            <span v-if="agent.description" class="agent-desc">
              - {{ agent.description.slice(0, 30) }}{{ agent.description.length > 30 ? '...' : '' }}
            </span>
          </a-select-option>
        </a-select>
        <div class="form-hint" v-if="getLeaderName">
          Leader 将使用「{{ getLeaderName }}」作为成员名称
        </div>
      </a-form-item>

      <a-form-item label="队员">
        <div class="member-list" v-if="form.members.length > 0">
          <div
            v-for="(member, index) in form.members"
            :key="`${member.agent_id}-${member.member_name}`"
            class="member-item"
          >
            <div class="member-info">
              <span class="member-name">{{ member.member_name }}</span>
              <span class="member-agent" v-if="agents.find(a => a.id === member.agent_id)?.name">
                ({{ agents.find(a => a.id === member.agent_id)?.name }})
              </span>
            </div>
            <a-button type="link" danger size="small" @click="handleRemoveMember(index)">
              移除
            </a-button>
          </div>
        </div>

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
          <a-button type="dashed" @click="handleAddMember">添加</a-button>
        </div>
        <div class="form-hint error" v-if="newMember.member_name && usedMemberNames.has(newMember.member_name)">
          成员名称已存在
        </div>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<style scoped>
.member-list {
  margin-bottom: 12px;
}

.member-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: var(--hover-bg);
  border-radius: 4px;
  margin-bottom: 8px;
}

.member-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.member-name {
  font-weight: 500;
}

.member-agent {
  color: var(--text-secondary);
  font-size: 13px;
}

.add-member-form {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.agent-desc {
  color: var(--text-secondary);
  font-size: 12px;
  margin-left: 8px;
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