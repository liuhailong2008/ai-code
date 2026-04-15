<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">耗时监控</div>
        <div class="page-subtitle">监控节点耗时与丢包率分布 · 最后更新: {{ lastUpdate }}</div>
      </div>
      <div class="page-actions">
        <select v-model="selectedIDC" class="btn btn-outline" style="padding: 9px 14px;">
          <option v-for="idc in idcs" :key="idc" :value="idc">{{ idc }}</option>
        </select>
        <button class="btn btn-outline" @click="togglePlay">
          {{ isPlaying ? '⏸️ 暂停' : '▶️ 播放' }}
        </button>
        <button class="btn btn-outline" @click="refreshData">↻ 刷新</button>
      </div>
    </div>

    <div class="charts-grid" style="grid-template-columns: 1fr;">
      <div class="card animate-in delay-1">
        <div class="card-header">
          <span class="card-title">耗时-丢包率散点图</span>
          <div class="card-actions">
            <button
              v-for="m in statTypes"
              :key="m.id"
              class="card-tab"
              :class="{ active: currentStat === m.id }"
              @click="switchStat(m.id)"
            >
              {{ m.label }}
            </button>
          </div>
        </div>
        <div class="card-body">
          <div ref="scatterRef" class="chart-container" style="height: 400px;"></div>
        </div>
      </div>
    </div>

    <div class="stats-row animate-in delay-2">
      <div class="stat-card green">
        <div class="stat-header">
          <span class="stat-label">正常节点</span>
          <div class="stat-icon">✓</div>
        </div>
        <div class="stat-value">{{ stats.normal }}</div>
        <div class="stat-footer trend-up">占比 {{ ((stats.normal / stats.total) * 100).toFixed(1) }}%</div>
      </div>
      <div class="stat-card yellow">
        <div class="stat-header">
          <span class="stat-label">警告节点</span>
          <div class="stat-icon">⚠</div>
        </div>
        <div class="stat-value">{{ stats.warning }}</div>
        <div class="stat-footer trend-neutral">占比 {{ ((stats.warning / stats.total) * 100).toFixed(1) }}%</div>
      </div>
      <div class="stat-card red">
        <div class="stat-header">
          <span class="stat-label">严重节点</span>
          <div class="stat-icon">✗</div>
        </div>
        <div class="stat-value">{{ stats.critical }}</div>
        <div class="stat-footer trend-down">占比 {{ ((stats.critical / stats.total) * 100).toFixed(1) }}%</div>
      </div>
      <div class="stat-card cyan">
        <div class="stat-header">
          <span class="stat-label">总节点数</span>
          <div class="stat-icon">📊</div>
        </div>
        <div class="stat-value">{{ stats.total }}</div>
        <div class="stat-footer trend-neutral">当前机房</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'

const scatterRef = ref(null)
let scatterChart = null
let playInterval = null

const lastUpdate = ref('--')
const isPlaying = ref(false)
const currentStat = ref('avg')
const selectedIDC = ref('IDC1-北京')

const idcs = ['IDC1-北京', 'IDC2-上海', 'IDC3-深圳', 'IDC4-杭州', 'IDC5-广州', 'IDC6-成都', 'IDC7-西安']

const statTypes = [
  { id: 'avg', label: '平均' },
  { id: 'p99', label: 'P99' },
  { id: 'max', label: '最大' }
]

const stats = ref({ normal: 0, warning: 0, critical: 0, total: 0 })

function fmtTime() {
  const d = new Date()
  return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function rand(min, max) {
  return Math.random() * (max - min) + min
}

function getThemeColors() {
  const theme = document.documentElement.getAttribute('data-theme') || 'style1'
  if (theme === 'style2') {
    return { bg: '#FFFFFF', text: '#111827', textMuted: '#9CA3AF', grid: '#E5E7EB' }
  }
  if (theme === 'style3') {
    return { bg: '#FFFFFF', text: '#1F2937', textMuted: '#6B7280', grid: '#E5E7EB' }
  }
  return { bg: '#1a2235', text: '#e8ecf4', textMuted: '#5a6a85', grid: 'rgba(30,42,63,.5)' }
}

function generateScatterData() {
  const points = []
  let normal = 0, warning = 0, critical = 0

  for (let i = 0; i < 25; i++) {
    const latency = rand(0.1, 12)
    const loss = rand(0, 1.5)
    let status = 'normal'
    let size = 10

    if (latency > 8 || loss > 0.8) {
      status = 'critical'
      critical++
      size = 18
    } else if (latency > 3 || loss > 0.2) {
      status = 'warning'
      warning++
      size = 14
    } else {
      normal++
      size = 10
    }

    points.push({
      value: [parseFloat(latency.toFixed(2)), parseFloat(loss.toFixed(3)), `Leaf-${String(i + 1).padStart(2, '0')}`, status],
      itemStyle: {
        color: status === 'critical' ? '#ef4444' : status === 'warning' ? '#f59e0b' : '#10b981'
      }
    })
  }

  stats.value = { normal, warning, critical, total: 25 }
  return points
}

function initScatter() {
  if (!scatterRef.value) return
  scatterChart = echarts.init(scatterRef.value)
  updateScatter()
}

function updateScatter() {
  const colors = getThemeColors()
  const data = generateScatterData()

  const option = {
    backgroundColor: colors.bg,
    tooltip: {
      trigger: 'item',
      formatter: (params) => `${params.value[2]}<br/>延迟: ${params.value[0]} ms<br/>丢包率: ${params.value[1]}%`
    },
    grid: { top: 30, left: 70, right: 30, bottom: 60 },
    xAxis: {
      type: 'value',
      name: '延迟 (ms)',
      nameLocation: 'middle',
      nameGap: 30,
      nameTextStyle: { color: colors.textMuted },
      axisLabel: { color: colors.textMuted },
      axisLine: { lineStyle: { color: colors.grid } },
      splitLine: { lineStyle: { color: colors.grid, type: 'dashed' } }
    },
    yAxis: {
      type: 'value',
      name: '丢包率 (%)',
      nameLocation: 'middle',
      nameGap: 45,
      nameTextStyle: { color: colors.textMuted },
      axisLabel: { color: colors.textMuted, formatter: v => v.toFixed(1) },
      axisLine: { lineStyle: { color: colors.grid } },
      splitLine: { lineStyle: { color: colors.grid, type: 'dashed' } }
    },
    series: [{
      type: 'scatter',
      data: data,
      symbolSize: (val, params) => params.data.itemStyle.color === '#ef4444' ? 18 : params.data.itemStyle.color === '#f59e0b' ? 14 : 10,
      emphasis: { scale: 1.5 }
    }]
  }

  scatterChart.setOption(option)
  lastUpdate.value = fmtTime()
}

function togglePlay() {
  isPlaying.value = !isPlaying.value
  if (isPlaying.value) {
    playInterval = setInterval(updateScatter, 2000)
  } else {
    clearInterval(playInterval)
  }
}

function switchStat(stat) {
  currentStat.value = stat
  updateScatter()
}

function refreshData() {
  updateScatter()
}

function handleResize() {
  scatterChart?.resize()
}

watch(() => document.documentElement.getAttribute('data-theme'), updateScatter)
watch(selectedIDC, updateScatter)

onMounted(() => {
  initScatter()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  clearInterval(playInterval)
  window.removeEventListener('resize', handleResize)
  scatterChart?.dispose()
})
</script>
