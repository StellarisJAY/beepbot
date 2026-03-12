<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  BookOutlined,
  FileOutlined,
  FolderOutlined,
  DeleteOutlined,
  ArrowLeftOutlined,
  EditOutlined,
} from '@ant-design/icons-vue'
import { skillApi } from '@/api/skill'
import type { SkillWithFiles, SkillFile, SkillFileContent } from '@/types/skill'

const route = useRoute()
const router = useRouter()

// 数据
const skill = ref<SkillWithFiles | null>(null)
const loading = ref(false)
const fileLoading = ref(false)
const selectedFile = ref<SkillFileContent | null>(null)

// 计算属性：按目录组织的文件树
const fileTree = computed(() => {
  if (!skill.value?.files) return []

  const tree: { name: string; path: string; isDir: boolean; children?: SkillFile[] }[] = []
  const dirMap = new Map<string, SkillFile[]>()

  // 按目录分组
  for (const file of skill.value.files) {
    const parts = file.file_path.split('/')
    if (parts.length === 1) {
      // 根目录文件
      tree.push({
        name: file.file_name,
        path: file.file_path,
        isDir: false,
        children: [file],
      })
    } else {
      // 子目录文件
      const dirName = parts[0]
      if (!dirName) continue

      if (!dirMap.has(dirName)) {
        dirMap.set(dirName, [])
        tree.push({
          name: dirName,
          path: dirName,
          isDir: true,
          children: dirMap.get(dirName),
        })
      }
      dirMap.get(dirName)!.push(file)
    }
  }

  return tree
})

// 获取技能详情
const fetchSkill = async () => {
  const id = route.params.id as string
  if (!id) return

  loading.value = true
  try {
    const res = await skillApi.getWithFiles(id)
    skill.value = res.data
    // 默认选中第一个文件
    if (res.data.files?.length > 0) {
      handleViewFile(res.data.files[0])
    }
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取技能详情失败')
  } finally {
    loading.value = false
  }
}

// 查看文件内容
const handleViewFile = async (file: SkillFile) => {
  if (!skill.value) return
  if (selectedFile.value?.id === file.id) return

  fileLoading.value = true
  try {
    const res = await skillApi.getFileContent(skill.value.id, file.id)
    selectedFile.value = res.data
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取文件内容失败')
  } finally {
    fileLoading.value = false
  }
}

// 返回列表
const handleBack = () => {
  router.push('/skills')
}

// 删除技能
const handleDelete = () => {
  if (!skill.value) return

  Modal.confirm({
    title: '确认删除',
    content: `确定要删除技能「${skill.value.name}」吗？此操作将删除技能的所有文件。`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        await skillApi.delete(skill.value!.id)
        message.success('删除成功')
        router.push('/skills')
      } catch (error: unknown) {
        const err = error as { message?: string }
        message.error(err.message || '删除失败')
      }
    },
  })
}

// 切换状态
const handleToggleStatus = async () => {
  if (!skill.value) return

  const newStatus = skill.value.status === 'active' ? 'inactive' : 'active'
  try {
    await skillApi.updateStatus(skill.value.id, newStatus)
    message.success('状态更新成功')
    skill.value.status = newStatus
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '状态更新失败')
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

onMounted(() => {
  fetchSkill()
})
</script>

<template>
  <div class="detail-layout">
    <!-- 左侧导航栏 -->
    <div class="side-nav">
      <!-- 上部分：技能信息 -->
      <div class="nav-header">
        <div class="skill-avatar">
          <BookOutlined class="avatar-icon" />
        </div>
        <div class="skill-info">
          <div class="skill-name-row">
            <span class="skill-name">{{ skill?.name || '加载中...' }}</span>
          </div>
          <p class="skill-desc">{{ skill?.description || '暂无描述' }}</p>
          <div class="skill-meta">
            <span v-if="skill?.version" class="version">v{{ skill.version }}</span>
            <a-tag v-if="skill" :color="skill.status === 'active' ? 'green' : 'default'" size="small">
              {{ skill.status === 'active' ? '启用' : '禁用' }}
            </a-tag>
          </div>
        </div>
        <div class="skill-actions">
          <a-button
            v-if="skill"
            :type="skill.status === 'active' ? 'default' : 'primary'"
            size="small"
            @click="handleToggleStatus"
          >
            {{ skill.status === 'active' ? '禁用' : '启用' }}
          </a-button>
          <a-button danger size="small" @click="handleDelete">
            <DeleteOutlined />
            删除
          </a-button>
        </div>
      </div>

      <!-- 下部分：文件列表 -->
      <div class="file-list-section">
        <div class="section-title">
          <FileOutlined />
          <span>文件列表</span>
          <span v-if="skill?.files" class="file-count">({{ skill.files.length }})</span>
        </div>
        <div class="file-list">
          <a-spin :spinning="fileLoading" size="small">
            <div v-if="skill?.files?.length === 0" class="empty-files">
              暂无文件
            </div>
            <div v-else class="file-tree">
              <template v-for="item in fileTree" :key="item.path">
                <!-- 目录 -->
                <div v-if="item.isDir" class="dir-item">
                  <FolderOutlined class="dir-icon" />
                  <span class="dir-name">{{ item.name }}</span>
                </div>
                <div v-if="item.isDir" class="dir-children">
                  <div
                    v-for="child in item.children"
                    :key="child.id"
                    class="file-item"
                    :class="{ active: selectedFile?.id === child.id }"
                    @click="handleViewFile(child)"
                  >
                    <FileOutlined class="file-icon" />
                    <span class="file-name">{{ child.file_name }}</span>
                    <span class="file-size">{{ formatFileSize(child.file_size) }}</span>
                  </div>
                </div>
                <!-- 根目录文件 -->
                <template v-if="!item.isDir && item.children && item.children.length > 0">
                  <div
                    class="file-item"
                    :class="{ active: selectedFile?.id === item.children[0].id }"
                    @click="item.children[0] && handleViewFile(item.children[0])"
                  >
                    <FileOutlined class="file-icon" />
                    <span class="file-name">{{ item.name }}</span>
                    <span class="file-size">{{ formatFileSize(item.children[0].file_size) }}</span>
                  </div>
                </template>
              </template>
            </div>
          </a-spin>
        </div>
      </div>
    </div>

    <!-- 中间区域 -->
    <div class="main-area">
      <!-- Header -->
      <div class="content-header">
        <div class="header-left">
          <a-button @click="handleBack">
            <ArrowLeftOutlined />
            返回列表
          </a-button>
        </div>
        <div class="header-right">
          <span v-if="selectedFile" class="file-info">
            {{ selectedFile.file_name }}
            <span class="file-type">{{ selectedFile.file_type }}</span>
            <span class="file-size">{{ formatFileSize(selectedFile.file_size) }}</span>
          </span>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="content-body">
        <a-spin :spinning="loading || fileLoading">
          <template v-if="selectedFile">
            <div class="file-content">
              <pre>{{ selectedFile.content }}</pre>
            </div>
          </template>
          <div v-else class="empty-content">
            <FileOutlined style="font-size: 48px; color: #ccc" />
            <p>选择文件查看内容</p>
          </div>
        </a-spin>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail-layout {
  height: 100%;
  display: flex;
  overflow: hidden;
  background: var(--bg-color);
}

/* 左侧导航栏 */
.side-nav {
  width: 280px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: var(--card-bg);
  border-right: 1px solid var(--border-color);
}

/* 上部分：技能信息 */
.nav-header {
  padding: 24px 16px;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.skill-avatar {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  background: linear-gradient(135deg, #52c41a 0%, #389e0d 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.avatar-icon {
  font-size: 28px;
  color: #fff;
}

.skill-info {
  text-align: center;
  margin-bottom: 16px;
}

.skill-name-row {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.skill-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-color);
}

.skill-desc {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.skill-meta {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-top: 8px;
}

.version {
  font-size: 12px;
  color: var(--text-secondary);
}

.skill-actions {
  display: flex;
  justify-content: center;
  gap: 8px;
}

/* 下部分：文件列表 */
.file-list-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

.file-count {
  font-size: 12px;
  color: var(--text-secondary);
}

.file-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.file-tree {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dir-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  font-weight: 500;
  color: var(--text-color);
}

.dir-icon {
  color: #faad14;
  font-size: 14px;
}

.dir-name {
  font-size: 13px;
}

.dir-children {
  padding-left: 16px;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-color);
}

.file-item:hover {
  background: var(--hover-bg);
}

.file-item.active {
  background: var(--color-primary);
  color: #fff;
}

.file-item.active .file-icon,
.file-item.active .file-size {
  color: #fff;
}

.file-icon {
  font-size: 14px;
  color: var(--text-secondary);
}

.file-name {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 11px;
  color: var(--text-secondary);
}

.empty-files {
  text-align: center;
  color: var(--text-secondary);
  padding: 24px;
  font-size: 13px;
}

/* 中间区域 */
.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* Header */
.content-header {
  flex-shrink: 0;
  height: 56px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  background: var(--card-bg);
  border-bottom: 1px solid var(--border-color);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--text-color);
}

.file-type {
  padding: 2px 6px;
  background: var(--bg-color);
  border-radius: 4px;
  font-size: 11px;
  text-transform: uppercase;
}

.file-info .file-size {
  font-size: 12px;
  color: var(--text-secondary);
}

/* 内容区域 */
.content-body {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.content-body :deep(.ant-spin-nested-loading) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.content-body :deep(.ant-spin-container) {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.file-content {
  flex: 1;
  overflow: auto;
  padding: 24px;
  background: var(--card-bg);
}

.file-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  line-height: 1.6;
  color: var(--text-color);
}

.empty-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--text-secondary);
}

.empty-content p {
  margin-top: 16px;
}
</style>
