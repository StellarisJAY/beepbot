<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { message, Modal } from 'ant-design-vue'
import {
  ArrowLeftOutlined,
  FileOutlined,
  FolderOutlined,
  DeleteOutlined,
  DownloadOutlined,
} from '@ant-design/icons-vue'
import { skillApi } from '@/api/skill'
import type { SkillWithFiles, SkillFile, SkillFileContent } from '@/types/skill'

const route = useRoute()
const router = useRouter()

// 数据
const skill = ref<SkillWithFiles | null>(null)
const loading = ref(false)
const selectedFile = ref<SkillFileContent | null>(null)
const fileLoading = ref(false)

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
      if (!dirName) continue // 跳过空目录名

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

// 获取文件图标
const getFileIcon = (fileType: string) => {
  const iconMap: Record<string, string> = {
    md: 'markdown',
    txt: 'file-text',
    json: 'code',
    yaml: 'code',
    yml: 'code',
    js: 'code',
    ts: 'code',
    py: 'code',
    go: 'code',
  }
  return iconMap[fileType] || 'file'
}

onMounted(() => {
  fetchSkill()
})
</script>

<template>
  <div class="skill-detail">
    <a-spin :spinning="loading" style="height:100%">
      <template v-if="skill">
        <!-- 头部 -->
        <div class="detail-header">
          <a-button type="text" @click="handleBack">
            <ArrowLeftOutlined />
            返回列表
          </a-button>
        </div>

        <!-- 技能信息 -->
        <a-card class="info-card">
          <div class="skill-info">
            <div class="skill-title">
              <h2>{{ skill.name }}</h2>
              <a-tag :color="skill.status === 'active' ? 'green' : 'default'">
                {{ skill.status === 'active' ? '启用' : '禁用' }}
              </a-tag>
              <span v-if="skill.version" class="version">v{{ skill.version }}</span>
            </div>

            <div class="skill-meta">
              <span v-if="skill.author">作者: {{ skill.author }}</span>
              <span>安装时间: {{ formatDate(skill.installed_at) }}</span>
              <span>更新时间: {{ formatDate(skill.updated_at) }}</span>
            </div>

            <p class="skill-description">{{ skill.description }}</p>

            <div class="skill-actions">
              <a-button
                :type="skill.status === 'active' ? 'default' : 'primary'"
                @click="handleToggleStatus"
              >
                {{ skill.status === 'active' ? '禁用' : '启用' }}
              </a-button>
              <a-button danger @click="handleDelete">
                <DeleteOutlined />
                删除技能
              </a-button>
            </div>
          </div>
        </a-card>

        <!-- 文件列表和内容预览 -->
        <div class="content-section">
          <a-row :gutter="16">
            <!-- 文件列表 -->
            <a-col :span="8">
              <a-card title="文件列表" class="file-list-card">
                <a-spin :spinning="fileLoading">
                  <div v-if="skill.files?.length === 0" class="empty-files">
                    暂无文件
                  </div>
                  <div v-else class="file-tree">
                    <template v-for="item in fileTree" :key="item.path">
                      <!-- 目录 -->
                      <div v-if="item.isDir" class="file-item dir">
                        <FolderOutlined />
                        <span>{{ item.name }}</span>
                      </div>
                      <div v-if="item.isDir" class="dir-children">
                        <div
                          v-for="child in item.children"
                          :key="child.id"
                          class="file-item"
                          :class="{ active: selectedFile?.id === child.id }"
                          @click="handleViewFile(child)"
                        >
                          <FileOutlined />
                          <span>{{ child.file_name }}</span>
                          <span class="file-size">{{ formatFileSize(child.file_size) }}</span>
                        </div>
                      </div>
                      <!-- 根目录文件 -->
                      <template v-if="!item.isDir && item.children && item.children.length > 0">
                        <div
                          class="file-item"
                          :class="{ active: selectedFile?.id === item.children[0]!.id }"
                          @click="item.children && item.children[0] && handleViewFile(item.children[0])"
                        >
                          <FileOutlined />
                          <span>{{ item.name }}</span>
                          <span class="file-size">{{ formatFileSize(item.children[0]!.file_size) }}</span>
                        </div>
                      </template>
                    </template>
                  </div>
                </a-spin>
              </a-card>
            </a-col>

            <!-- 文件内容预览 -->
            <a-col :span="16">
              <a-card title="文件内容" class="content-card">
                <template v-if="selectedFile">
                  <div class="file-header">
                    <span class="file-name">{{ selectedFile.file_name }}</span>
                    <span class="file-type">{{ selectedFile.file_type }}</span>
                    <span class="file-size">{{ formatFileSize(selectedFile.file_size) }}</span>
                  </div>
                  <div class="file-content">
                    <pre>{{ selectedFile.content }}</pre>
                  </div>
                </template>
                <div v-else class="empty-content">
                  <FileOutlined style="font-size: 48px; color: #ccc" />
                  <p>选择文件查看内容</p>
                </div>
              </a-card>
            </a-col>
          </a-row>
        </div>
      </template>
    </a-spin>
  </div>
</template>

<style scoped>
.skill-detail {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 24px;
  overflow: hidden;
}

.detail-header {
  margin-bottom: 16px;
  flex-shrink: 0;
  height: 20px;
}

.info-card {
  margin-bottom: 24px;
  flex-shrink: 0;
}

.skill-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}

.skill-title {
  display: flex;
  align-items: center;
  gap: 12px;
}

.skill-title h2 {
  margin: 0;
  font-size: 24px;
}

.skill-title .version {
  color: #999;
  font-size: 14px;
}

.skill-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  font-size: 14px;
  color: #666;
}

.skill-description {
  color: #333;
  line-height: 1.6;
  margin: 0;
}

.skill-actions {
  display: flex;
  gap: 12px;
  margin-top: 8px;
}

.content-section {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  height: 100%;
}

.content-section :deep(.ant-row) {
  height: 100%;
}

.content-section :deep(.ant-col) {
  height: 100%;
}

.file-list-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.file-list-card :deep(.ant-card-body) {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.file-list-card :deep(.ant-spin-nested-loading) {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.file-list-card :deep(.ant-spin-container) {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.file-tree {
  flex: 1;
  overflow-y: auto;
  min-height: 0;
  max-height: 100%;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.file-item:hover {
  background-color: #f5f5f5;
}

.file-item.active {
  background-color: #e6f7ff;
}

.file-item.dir {
  font-weight: 500;
  cursor: default;
}

.file-item.dir:hover {
  background-color: transparent;
}

.dir-children {
  padding-left: 24px;
}

.file-size {
  margin-left: auto;
  font-size: 12px;
  color: #999;
}

.content-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.content-card :deep(.ant-card-body) {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.file-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 12px;
  flex-shrink: 0;
}

.file-name {
  font-weight: 500;
}

.file-type {
  padding: 2px 8px;
  background: #f0f0f0;
  border-radius: 4px;
  font-size: 12px;
  text-transform: uppercase;
}

.file-content {
  background: #f8f8f8;
  border-radius: 4px;
  padding: 16px;
  flex: 1;
  overflow: auto;
  min-height: 0;
}

.file-content pre {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
}

.empty-files {
  text-align: center;
  color: #999;
  padding: 24px;
}

.empty-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 300px;
  color: #999;
}

.empty-content p {
  margin-top: 16px;
}
</style>