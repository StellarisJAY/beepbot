<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { message, Modal, Upload } from 'ant-design-vue'
import {
  PlusOutlined,
  DeleteOutlined,
  EyeOutlined,
  UploadOutlined,
  BookOutlined,
} from '@ant-design/icons-vue'
import type { UploadProps } from 'ant-design-vue'
import { skillApi } from '@/api/skill'
import type { Skill } from '@/types/skill'

const router = useRouter()

// 数据
const skills = ref<Skill[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const size = ref(10)

// 上传相关
const uploadLoading = ref(false)
const uploadVisible = ref(false)

// 获取技能列表
const fetchSkills = async () => {
  loading.value = true
  try {
    const res = await skillApi.list(page.value, size.value)
    skills.value = res.data
    total.value = res.total
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取技能列表失败')
  } finally {
    loading.value = false
  }
}

// 查看技能详情
const handleView = (id: string) => {
  router.push(`/skills/${id}`)
}

// 删除技能
const handleDelete = (skill: Skill) => {
  Modal.confirm({
    title: '确认删除',
    content: `确定要删除技能「${skill.name}」吗？此操作将删除技能的所有文件。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await skillApi.delete(skill.id)
        message.success('删除成功')
        await fetchSkills()
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || '删除失败')
      }
    },
  })
}

// 切换状态
const handleToggleStatus = async (skill: Skill) => {
  const newStatus = skill.status === 'active' ? 'inactive' : 'active'
  try {
    await skillApi.updateStatus(skill.id, newStatus)
    message.success('状态更新成功')
    await fetchSkills()
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '状态更新失败')
  }
}

// 上传前校验
const beforeUpload: UploadProps['beforeUpload'] = (file) => {
  const isZip = file.name.endsWith('.zip')
  if (!isZip) {
    message.error('只能上传 ZIP 文件')
    return false
  }
  const isLt10M = file.size / 1024 / 1024 < 10
  if (!isLt10M) {
    message.error('文件大小不能超过 10MB')
    return false
  }
  return true
}

// 自定义上传
const customUpload: UploadProps['customRequest'] = async (options) => {
  const { file, onSuccess, onError } = options
  uploadLoading.value = true

  try {
    const res = await skillApi.upload(file as File)
    message.success(`技能「${res.data.name}」安装成功`)
    uploadVisible.value = false
    await fetchSkills()
    onSuccess?.(res)
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '上传失败')
    onError?.(error as Error)
  } finally {
    uploadLoading.value = false
  }
}

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

// 格式化日期
const formatDate = (dateStr: string): string => {
  return new Date(dateStr).toLocaleString('zh-CN')
}

// 分页改变
const handlePageChange = (p: number) => {
  page.value = p
  fetchSkills()
}

onMounted(() => {
  fetchSkills()
})
</script>

<template>
  <div class="skill-list">
    <div class="page-header">
      <h1>技能管理</h1>
      <a-button type="primary" @click="uploadVisible = true">
        <UploadOutlined />
        安装技能
      </a-button>
    </div>

    <!-- 技能列表 -->
    <a-spin :spinning="loading">
      <div v-if="skills.length === 0 && !loading" class="empty-state">
        <BookOutlined style="font-size: 48px; color: #ccc" />
        <p>暂无技能，点击上方按钮安装新技能</p>
      </div>

      <div v-else class="skill-cards">
        <a-card v-for="skill in skills" :key="skill.id" class="skill-card" hoverable>
          <template #title>
            <div class="card-title">
              <BookOutlined />
              <span>{{ skill.name }}</span>
              <a-tag :color="skill.status === 'active' ? 'green' : 'default'">
                {{ skill.status === 'active' ? '启用' : '禁用' }}
              </a-tag>
            </div>
          </template>
          <template #extra>
            <a-space>
              <a-tooltip title="查看详情">
                <a-button type="text" size="small" @click="handleView(skill.id)">
                  <EyeOutlined />
                </a-button>
              </a-tooltip>
              <a-popconfirm
                :title="skill.status === 'active' ? '确定要禁用此技能吗？' : '确定要启用此技能吗？'"
                @confirm="handleToggleStatus(skill)"
              >
                <a-button type="text" size="small">
                  {{ skill.status === 'active' ? '禁用' : '启用' }}
                </a-button>
              </a-popconfirm>
              <a-tooltip title="删除">
                <a-button type="text" size="small" danger @click="handleDelete(skill)">
                  <DeleteOutlined />
                </a-button>
              </a-tooltip>
            </a-space>
          </template>

          <p class="skill-description">{{ skill.description }}</p>
          <div class="skill-meta">
            <span v-if="skill.version">v{{ skill.version }}</span>
            <span v-if="skill.author">作者: {{ skill.author }}</span>
            <span>安装时间: {{ formatDate(skill.installed_at) }}</span>
          </div>
        </a-card>
      </div>
    </a-spin>

    <!-- 分页 -->
    <div v-if="total > size" class="pagination">
      <a-pagination
        :current="page"
        :page-size="size"
        :total="total"
        show-less-items
        @change="handlePageChange"
      />
    </div>

    <!-- 上传弹窗 -->
    <a-modal
      v-model:open="uploadVisible"
      title="安装技能"
      :footer="null"
      :confirm-loading="uploadLoading"
    >
      <a-upload-dragger
        :before-upload="beforeUpload"
        :custom-request="customUpload"
        :show-upload-list="false"
        accept=".zip"
      >
        <p class="ant-upload-drag-icon">
          <UploadOutlined />
        </p>
        <p class="ant-upload-text">点击或拖拽文件到此区域上传</p>
        <p class="ant-upload-hint">
          仅支持 .zip 格式的技能包，文件大小不超过 10MB
        </p>
      </a-upload-dragger>

      <div class="upload-tips">
        <h4>技能包格式要求：</h4>
        <ul>
          <li>ZIP 文件必须包含 SKILL.md 文件</li>
          <li>SKILL.md 第一行为技能名称（# 名称）</li>
          <li>可选包含 skill.json 元数据文件</li>
        </ul>
      </div>
    </a-modal>
  </div>
</template>

<style scoped>
.skill-list {
  padding: 24px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  font-size: 24px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  color: #999;
}

.empty-state p {
  margin-top: 16px;
}

.skill-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 16px;
}

.skill-card {
  transition: all 0.3s;
}

.skill-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.card-title span {
  font-weight: 600;
}

.skill-description {
  color: #666;
  margin-bottom: 12px;
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  max-height: 4.5em;
}

.skill-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  font-size: 12px;
  color: #999;
}

.pagination {
  display: flex;
  justify-content: center;
  margin-top: 24px;
}

.upload-tips {
  margin-top: 16px;
  padding: 12px;
  background: #f5f5f5;
  border-radius: 4px;
}

.upload-tips h4 {
  margin: 0 0 8px 0;
  font-size: 14px;
}

.upload-tips ul {
  margin: 0;
  padding-left: 20px;
  font-size: 12px;
  color: #666;
}

.upload-tips li {
  margin-bottom: 4px;
}
</style>