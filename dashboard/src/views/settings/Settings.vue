<script setup lang="ts">
import { ref } from 'vue'
import { SaveOutlined } from '@ant-design/icons-vue'

// 表单数据
const formData = ref({
  workingDir: 'D:/data/beepbot',
  dataDir: './beepbot',
  logLevel: 'info',
  logFormat: 'json',
  heartbeatEnabled: true,
  heartbeatInterval: '60s',
})

// 保存设置
const handleSave = () => {
  console.log('保存设置:', formData.value)
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">全局设置</h1>
    </div>

    <a-card class="settings-card">
      <a-form :model="formData" layout="vertical">
        <a-divider orientation="left">基础配置</a-divider>

        <a-form-item label="工作目录" name="workingDir">
          <a-input v-model:value="formData.workingDir" placeholder="请输入工作目录路径" />
          <template #extra>
            <span class="form-hint">智能体文件操作的根目录，所有文件操作将限制在此目录内</span>
          </template>
        </a-form-item>

        <a-form-item label="数据目录" name="dataDir">
          <a-input v-model:value="formData.dataDir" placeholder="请输入数据目录路径" />
          <template #extra>
            <span class="form-hint">BeepBot 公共数据目录，用于存储全局技能、共享数据等</span>
          </template>
        </a-form-item>

        <a-divider orientation="left">日志配置</a-divider>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="日志级别" name="logLevel">
              <a-select v-model:value="formData.logLevel">
                <a-select-option value="debug">Debug</a-select-option>
                <a-select-option value="info">Info</a-select-option>
                <a-select-option value="warn">Warn</a-select-option>
                <a-select-option value="error">Error</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="日志格式" name="logFormat">
              <a-select v-model:value="formData.logFormat">
                <a-select-option value="json">JSON</a-select-option>
                <a-select-option value="text">Text</a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <a-divider orientation="left">心跳配置</a-divider>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item label="启用心跳" name="heartbeatEnabled">
              <a-switch v-model:checked="formData.heartbeatEnabled" />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item label="心跳间隔" name="heartbeatInterval">
              <a-input v-model:value="formData.heartbeatInterval" placeholder="例如: 60s, 5m, 1h" />
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item>
          <a-button type="primary" @click="handleSave">
            <template #icon><SaveOutlined /></template>
            保存设置
          </a-button>
        </a-form-item>
      </a-form>
    </a-card>
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

.settings-card {
  border-radius: 8px;
  border: 1px solid var(--card-border);
  background: var(--card-bg);
  max-width: 800px;
}

.form-hint {
  color: var(--text-tertiary);
  font-size: 12px;
}

:deep(.ant-divider) {
  margin: 16px 0 24px;
}

:deep(.ant-divider-inner-text) {
  font-weight: 600;
  color: var(--text-color);
}
</style>