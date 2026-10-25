<template>
  <div class="dashboard-container">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>欢迎使用 blog_apis</span>
          <el-button :icon="Refresh" :loading="loading" @click="load">刷新</el-button>
        </div>
      </template>
      <div class="welcome">
        <div class="welcome-title">
          <el-icon class="welcome-icon"><component :is="greeting.icon" /></el-icon>
          <span>{{ greeting.period }}好，{{ username }}</span>
        </div>
        <el-text type="info">{{ dateText }} · {{ line }}</el-text>
      </div>
      <el-divider />
      <el-row :gutter="20">
        <el-col v-for="tile in tiles" :key="tile.label" :xs="12" :sm="12" :md="6" class="stat-col">
          <el-statistic :value="0" :formatter="() => tile.value">
            <template #title>
              <div class="stat-title">
                <el-icon><component :is="tile.icon" /></el-icon>
                <span>{{ tile.label }}</span>
              </div>
            </template>
          </el-statistic>
        </el-col>
      </el-row>
    </el-card>

    <el-row :gutter="12">
      <el-col :xs="24" :md="10">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <div class="card-header">友链存活状态</div>
          </template>
          <div ref="pieChart" class="chart"></div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="14">
        <el-card shadow="never" class="chart-card">
          <template #header>
            <div class="card-header">每月 Feed Post 数量</div>
          </template>
          <div ref="lineChart" class="chart"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import {
  Box,
  Coin,
  Collection,
  Document,
  FolderOpened,
  HomeFilled,
  Link,
  Reading,
  Refresh,
  Timer
} from '@element-plus/icons-vue'
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import dayjs from 'dayjs'
import { getSystemStatus, type SystemStatus } from '@/api/status'
import { findGreetingPeriod, pickGreetingLine } from '@/utils/greetings'
import * as echarts from 'echarts/core'
import { LineChart, PieChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

definePage({ meta: { title: '仪表板', icon: HomeFilled, order: 1 } })

echarts.use([LineChart, PieChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

const username = localStorage.getItem('username') || '管理员'
const pieChart = ref<HTMLElement | null>(null)
const lineChart = ref<HTMLElement | null>(null)
const now = ref(new Date())
const loading = ref(false)

const greeting = computed(() => findGreetingPeriod(now.value.getHours()))

const WEEKDAYS = ['日', '一', '二', '三', '四', '五', '六']
const dateText = computed(() => {
  const d = dayjs(now.value)
  return `${d.format('YYYY 年 M 月 D 日')} · 星期${WEEKDAYS[d.day()]}`
})

// 一次只露一句，进面板随机抽，时段变了再重抽
const line = ref(pickGreetingLine(greeting.value))

watch(greeting, (period) => {
  line.value = pickGreetingLine(period)
})

const stats = ref<SystemStatus>({
  uptime: '-',
  database_size_bytes: 0,
  data_folder_size_bytes: 0,
  kv_cache: { entry_count: 0, payload_bytes: 0 },
  status_data: {
    friend_link_count: 0,
    rss_count: 0,
    rss_post_count: 0,
    friend_link_status_pie: [],
    rss_post_count_monthly: []
  },
  time: 0
})

const formatBytes = (bytes: number): string => {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / 1024 ** i
  const digits = i === 0 || value >= 100 ? 0 : value >= 10 ? 1 : 2
  return `${value.toFixed(digits)} ${units[i]}`
}

const formatCount = (n: number) => n.toLocaleString('zh-CN')

const tiles = computed(() => {
  const s = stats.value
  return [
    { label: '友链', icon: Link, value: formatCount(s.status_data.friend_link_count) },
    { label: 'RSS 订阅', icon: Reading, value: formatCount(s.status_data.rss_count) },
    { label: 'RSS 文章', icon: Document, value: formatCount(s.status_data.rss_post_count) },
    { label: '在线时长', icon: Timer, value: s.uptime },
    { label: '数据库', icon: Coin, value: formatBytes(s.database_size_bytes) },
    { label: '数据目录', icon: FolderOpened, value: formatBytes(s.data_folder_size_bytes) },
    { label: 'KV 条目', icon: Collection, value: formatCount(s.kv_cache.entry_count) },
    { label: 'KV 占用', icon: Box, value: formatBytes(s.kv_cache.payload_bytes) }
  ]
})

const renderCharts = () => {
  const data = stats.value.status_data

  if (pieChart.value) {
    const pie = echarts.getInstanceByDom(pieChart.value) || echarts.init(pieChart.value)
    pie.setOption({
      tooltip: { trigger: 'item' },
      legend: { bottom: 0 },
      series: [
        {
          name: '友链状态',
          type: 'pie',
          radius: ['45%', '70%'],
          center: ['50%', '45%'],
          label: { show: false },
          data: data.friend_link_status_pie.map((item) => ({ value: item.count, name: item.status }))
        }
      ]
    })
  }

  if (lineChart.value) {
    const chart = echarts.getInstanceByDom(lineChart.value) || echarts.init(lineChart.value)
    chart.setOption({
      tooltip: { trigger: 'axis' },
      grid: { left: 8, right: 16, top: 16, bottom: 8, containLabel: true },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: data.rss_post_count_monthly.map((item) => item.month)
      },
      yAxis: { type: 'value' },
      series: [
        {
          name: '文章数量',
          type: 'line',
          smooth: true,
          areaStyle: { opacity: 0.15 },
          data: data.rss_post_count_monthly.map((item) => item.count)
        }
      ]
    })
  }
}

const load = async () => {
  if (loading.value) return
  loading.value = true
  try {
    const res = await getSystemStatus()
    stats.value = res.data
    renderCharts()
  } catch {
    // 请求错误由拦截器统一提示
  } finally {
    loading.value = false
  }
}

const resize = () => {
  for (const el of [pieChart.value, lineChart.value]) {
    if (el) echarts.getInstanceByDom(el)?.resize()
  }
}

const timer = setInterval(() => {
  now.value = new Date()
}, 60_000)

onMounted(() => {
  load()
  window.addEventListener('resize', resize)
})

onUnmounted(() => {
  clearInterval(timer)
  window.removeEventListener('resize', resize)
  for (const el of [pieChart.value, lineChart.value]) {
    if (el) echarts.getInstanceByDom(el)?.dispose()
  }
})
</script>

<style scoped>
.dashboard-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.card-header {
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.welcome {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 0;
}

.welcome-title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: var(--el-text-color-primary);
  font-size: 22px;
  font-weight: 500;
}

.welcome-icon {
  color: var(--el-color-primary);
}

.stat-col {
  margin-bottom: 20px;
}

.stat-title {
  display: flex;
  align-items: center;
  gap: 6px;
}

.chart-card {
  margin-bottom: 12px;
}

.chart {
  width: 100%;
  height: 320px;
}
</style>
