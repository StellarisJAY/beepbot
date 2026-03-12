<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
} from 'echarts/components'
import { sessionApi } from '@/api/session'
import { PeriodOptions, type UsageStatsPoint } from '@/types/session'

// 注册 ECharts 组件
use([
  CanvasRenderer,
  LineChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent,
])

const route = useRoute()
const agentId = computed(() => route.params.id as string)

// 数据
const loading = ref(false)
const statsData = ref<UsageStatsPoint[]>([])
const selectedPeriod = ref('7d')

// 时间格式化
const formatTime = (time: string, isHourly: boolean): string => {
  const date = new Date(time)
  if (isHourly) {
    return `${date.getMonth() + 1}/${date.getDate()} ${date.getHours()}:00`
  }
  return `${date.getMonth() + 1}/${date.getDate()}`
}

// 是否按小时聚合
const isHourly = computed(() => {
  return selectedPeriod.value === '1d' || selectedPeriod.value === '3d'
})

// 会话数量图表配置
const sessionChartOption = computed(() => {
  const times = statsData.value.map((p) => formatTime(p.time, isHourly.value))
  const values = statsData.value.map((p) => p.session_count)

  return {
    title: {
      text: '会话数量趋势',
      left: 'center',
      textStyle: { fontSize: 14 },
    },
    tooltip: {
      trigger: 'axis',
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: times,
      axisLabel: {
        rotate: isHourly.value ? 45 : 0,
        fontSize: 10,
      },
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
    },
    series: [
      {
        name: '会话数',
        type: 'line',
        data: values,
        smooth: true,
        areaStyle: { opacity: 0.3 },
        itemStyle: { color: '#1890ff' },
      },
    ],
  }
})

// 消息数量图表配置
const messageChartOption = computed(() => {
  const times = statsData.value.map((p) => formatTime(p.time, isHourly.value))
  const values = statsData.value.map((p) => p.message_count)

  return {
    title: {
      text: '消息数量趋势',
      left: 'center',
      textStyle: { fontSize: 14 },
    },
    tooltip: {
      trigger: 'axis',
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: times,
      axisLabel: {
        rotate: isHourly.value ? 45 : 0,
        fontSize: 10,
      },
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
    },
    series: [
      {
        name: '消息数',
        type: 'line',
        data: values,
        smooth: true,
        areaStyle: { opacity: 0.3 },
        itemStyle: { color: '#52c41a' },
      },
    ],
  }
})

// Token 用量图表配置（堆叠面积图）
const tokenChartOption = computed(() => {
  const times = statsData.value.map((p) => formatTime(p.time, isHourly.value))
  const inputTokens = statsData.value.map((p) => p.input_tokens)
  const outputTokens = statsData.value.map((p) => p.output_tokens)

  return {
    title: {
      text: 'Token 用量趋势',
      left: 'center',
      textStyle: { fontSize: 14 },
    },
    tooltip: {
      trigger: 'axis',
    },
    legend: {
      data: ['输入 Token', '输出 Token'],
      top: 30,
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLabel: {
        rotate: isHourly.value ? 45 : 0,
        fontSize: 10,
      },
    },
    yAxis: {
      type: 'value',
    },
    series: [
      {
        name: '输入 Token',
        type: 'line',
        stack: 'Total',
        areaStyle: {},
        emphasis: { focus: 'series' },
        data: inputTokens,
        itemStyle: { color: '#1890ff' },
      },
      {
        name: '输出 Token',
        type: 'line',
        stack: 'Total',
        areaStyle: {},
        emphasis: { focus: 'series' },
        data: outputTokens,
        itemStyle: { color: '#faad14' },
      },
    ],
  }
})

// 获取统计数据
const fetchStats = async () => {
  loading.value = true
  try {
    const res = await sessionApi.getAgentUsageStats(agentId.value, selectedPeriod.value)
    statsData.value = res.points || []
  } catch (error: unknown) {
    const err = error as { message?: string }
    message.error(err.message || '获取统计数据失败')
  } finally {
    loading.value = false
  }
}

// 时间范围变化
const handlePeriodChange = () => {
  fetchStats()
}

// 监听路由变化
watch(
  () => route.params.id,
  () => {
    if (agentId.value) {
      fetchStats()
    }
  },
)

onMounted(() => {
  if (agentId.value) {
    fetchStats()
  }
})
</script>

<template>
  <div class="monitor-container">
    <!-- 头部控制栏 -->
    <div class="monitor-header">
      <a-select
        v-model:value="selectedPeriod"
        style="width: 150px"
        @change="handlePeriodChange"
      >
        <a-select-option v-for="opt in PeriodOptions" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </a-select-option>
      </a-select>
    </div>

    <!-- 图表区域 -->
    <a-spin :spinning="loading">
      <div class="charts-grid">
        <!-- 会话数量图表 -->
        <div class="chart-card">
          <v-chart
            :option="sessionChartOption"
            autoresize
            style="height: 280px"
          />
        </div>

        <!-- 消息数量图表 -->
        <div class="chart-card">
          <v-chart
            :option="messageChartOption"
            autoresize
            style="height: 280px"
          />
        </div>

        <!-- Token 用量图表 -->
        <div class="chart-card full-width">
          <v-chart
            :option="tokenChartOption"
            autoresize
            style="height: 320px"
          />
        </div>
      </div>
    </a-spin>
  </div>
</template>

<style scoped>
.monitor-container {
  padding: 24px;
}

.monitor-header {
  margin-bottom: 24px;
  display: flex;
  align-items: center;
  gap: 16px;
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 24px;
}

.chart-card {
  background: var(--ant-color-bg-container);
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.03);
}

.chart-card.full-width {
  grid-column: span 2;
}

@media (max-width: 1024px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }

  .chart-card.full-width {
    grid-column: span 1;
  }
}
</style>