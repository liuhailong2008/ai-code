<template>
  <div>
    <div class="page-header">
      <div>
        <div class="page-title">机房耗时监控</div>
        <div class="page-subtitle">监控节点耗时与丢包率分布</div>
      </div>
      <div class="page-actions">
        <select v-model="currentStat" class="metric-select" @change="switchStat(currentStat)">
          <option v-for="m in statTypes" :key="m.id" :value="m.id">{{ m.label }}</option>
        </select>
        <select v-model="carouselMode" class="btn btn-outline" style="padding: 9px 14px;" @change="restartAutoPlay">
          <option value="idc">按机房</option>
          <option value="time">按时间</option>
        </select>
        <select v-model="updateInterval" class="btn btn-outline" style="padding: 9px 14px;" @change="restartAutoPlay">
          <option :value="5000">5秒</option>
          <option :value="10000">10秒</option>
          <option :value="60000">1分钟</option>
        </select>
        <button class="btn btn-outline" @click="togglePlay">
          {{ isPlaying ? '⏸️ 暂停' : '▶️ 播放' }}
        </button>
        <button class="btn btn-outline" @click="refreshData">↻ 刷新</button>
      </div>
    </div>

    <div class="idc-selector">
      <button
        v-for="idc in idcs"
        :key="idc"
        class="idc-btn"
        :class="{ active: selectedIDC === idc }"
        @click="selectIDC(idc)"
      >
        {{ idc }}
      </button>
    </div>

    <div class="scatter-section">
      <div class="scatter-card">
        <div class="card-header">
          <span class="card-title">耗时-丢包率散点图</span>
        </div>
        <div class="card-body">
          <div ref="scatterRef" class="chart-container" style="height: 400px;"></div>
        </div>
      </div>
      <div class="scatter-sidebar">
        <div class="sidebar-card">
          <div class="sidebar-stat total">
            <div class="stat-num">{{ stats.total }}</div>
            <div class="stat-label">总设备数</div>
          </div>
          <div class="sidebar-stat attention">
            <div class="stat-num">{{ stats.warning }}</div>
            <div class="stat-label">需关注</div>
          </div>
          <div class="sidebar-stat alarm">
            <div class="stat-num">{{ stats.critical }}</div>
            <div class="stat-label">告警设备</div>
          </div>
        </div>
        <div class="alarm-list-card">
          <div class="card-header">
            <span class="card-title">异常设备清单</span>
          </div>
          <div class="alarm-list">
            <div v-for="device in abnormalDevices" :key="device.name" class="alarm-item">
              <div class="alarm-device-row">
                <span class="alarm-device">{{ device.name }}</span>
                <span class="device-tag" :class="device.status">{{ device.status === 'critical' ? '告警设备' : '需关注' }}</span>
              </div>
              <div class="alarm-stats">
                <div class="stat-row">
                  <span class="stat-label">丢包率</span>
                  <span class="stat-value">{{ device.loss.toFixed(2) }}%</span>
                </div>
                <div class="stat-row">
                  <span class="stat-label">平均</span>
                  <span class="stat-value">{{ device.avg.toFixed(1) }}ms</span>
                </div>
                <div class="stat-row">
                  <span class="stat-label">P999</span>
                  <span class="stat-value">{{ device.p99.toFixed(1) }}ms</span>
                </div>
                <div class="stat-row">
                  <span class="stat-label">最大</span>
                  <span class="stat-value">{{ device.max.toFixed(1) }}ms</span>
                </div>
              </div>
            </div>
            <div v-if="abnormalDevices.length === 0" class="no-alarm">暂无异常设备</div>
          </div>
        </div>
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
const isPlaying = ref(true)
const currentStat = ref('avg')
const selectedIDC = ref('IDC1-北京')
const carouselMode = ref('idc') // 轮播模式：idc-按机房, time-按时间
const updateInterval = ref(5000)

const idcs = ['IDC1-北京', 'IDC2-上海', 'IDC3-深圳', 'IDC4-杭州', 'IDC5-广州', 'IDC6-成都', 'IDC7-西安']

const statTypes = [
  { id: 'avg', label: '平均' },
  { id: 'p99', label: 'P999' },
  { id: 'max', label: '最大' }
]

const stats = ref({ normal: 0, warning: 0, critical: 0, total: 0 })
const alertDevices = ref([])
const abnormalDevices = ref([])

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
  const abnormal = []
  let normal = 0, warning = 0, critical = 0

  for (let i = 0; i < 25; i++) {
    // 随机生成延迟和丢包率，范围更大以覆盖告警条件
    const avg = rand(0.1, 200)
    const p99 = avg * rand(1.5, 2.5)
    const max = avg * rand(2, 4)
    const loss = rand(0, 3)
    let status = 'normal'

    // 判断规则：
    // 正常：横轴<150 且 纵轴<2.5（绿色）
    // 需关注：横轴>150 且 纵轴<2.5 或 横轴<150 且 纵轴>2.5（黄色）
    // 告警：横轴>150 且 纵轴>2.5（红色）
    if (avg > 150 && loss > 2.5) {
      status = 'critical'
      critical++
      abnormal.push({
        name: `Leaf-${String(i + 1).padStart(2, '0')}`,
        loss: loss,
        avg: avg,
        p99: p99,
        max: max,
        status: 'critical'
      })
    } else if (avg > 150 || loss > 2.5) {
      status = 'warning'
      warning++
      abnormal.push({
        name: `Leaf-${String(i + 1).padStart(2, '0')}`,
        loss: loss,
        avg: avg,
        p99: p99,
        max: max,
        status: 'warning'
      })
    } else {
      normal++
    }

    points.push({
      value: [parseFloat(avg.toFixed(2)), parseFloat(loss.toFixed(3)), `Leaf-${String(i + 1).padStart(2, '0')}`, status],
      itemStyle: {
        color: status === 'critical' ? '#ef4444' : status === 'warning' ? '#f59e0b' : '#10b981'
      }
    })
  }

  stats.value = { normal, warning, critical, total: 25 }
  abnormalDevices.value = abnormal
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
      splitLine: { lineStyle: { color: colors.grid, type: 'dashed' } },
      min: 0,
      max: 300
    },
    yAxis: {
      type: 'value',
      name: '丢包率 (%)',
      nameLocation: 'middle',
      nameGap: 45,
      nameTextStyle: { color: colors.textMuted },
      axisLabel: { color: colors.textMuted, formatter: v => v.toFixed(1) },
      axisLine: { lineStyle: { color: colors.grid } },
      splitLine: { lineStyle: { color: colors.grid, type: 'dashed' } },
      min: 0,
      max: 5
    },
    series: [{
      type: 'scatter',
      data: data,
      symbolSize: (val, params) => params.data.itemStyle.color === '#ef4444' ? 18 : params.data.itemStyle.color === '#f59e0b' ? 14 : 10,
      emphasis: { scale: 1.5 },
      markLine: {
        silent: true,
        symbol: 'none',
        lineStyle: { type: 'dashed', width: 1.5 },
        label: { show: true, position: 'end', color: colors.textMuted },
        data: [
          { xAxis: 150, name: '150ms' },
          { yAxis: 2.5, name: '2.5%' }
        ]
      }
    }]
  }

  scatterChart.setOption(option)
  lastUpdate.value = fmtTime()
}

function togglePlay() {
  isPlaying.value = !isPlaying.value
  if (isPlaying.value) {
    startAutoPlay()
  } else {
    clearInterval(playInterval)
  }
}

function startAutoPlay() {
  clearInterval(playInterval)
  playInterval = setInterval(() => {
    if (carouselMode.value === 'time') {
      // 按时间轮播：暂不需要
    } else {
      // 按机房轮播
      const currentIdx = idcs.indexOf(selectedIDC.value)
      selectedIDC.value = idcs[(currentIdx + 1) % idcs.length]
    }
    updateScatter()
  }, updateInterval.value)
}

function restartAutoPlay() {
  if (isPlaying.value) {
    startAutoPlay()
  }
}

function switchStat(stat) {
  currentStat.value = stat
  updateScatter()
}

function selectIDC(idc) {
  selectedIDC.value = idc
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
  startAutoPlay()
})

onUnmounted(() => {
  clearInterval(playInterval)
  window.removeEventListener('resize', handleResize)
  scatterChart?.dispose()
})
</script>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; background: var(--bg-card); border-bottom: 1px solid var(--border-color); }
.page-title { font-size: 20px; font-weight: 600; color: var(--text-primary); }
.page-subtitle { font-size: 13px; color: var(--text-muted); margin-top: 4px; }
.page-actions { display: flex; align-items: center; gap: 12px; }
.metric-select { background: transparent; border: 1px solid var(--border-color); border-radius: var(--radius-sm); padding: 6px 10px; font-size: 13px; color: var(--text-primary); cursor: pointer; }

.idc-selector { display: flex; gap: 8px; flex-wrap: wrap; }
.idc-btn { padding: 8px 16px; border-radius: var(--radius-sm); background: var(--bg-card); border: 1px solid var(--border-color); color: var(--text-secondary); font-size: 13px; font-weight: 500; cursor: pointer; transition: all 0.2s; }
.idc-btn:hover { border-color: var(--accent-blue); color: var(--text-primary); }
.idc-btn.active { background: var(--accent-blue); border-color: var(--accent-blue); color: white; }

.scatter-section { display: grid; grid-template-columns: 1fr 320px; gap: 20px; margin-bottom: 20px; }
.scatter-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius); overflow: hidden; }
.scatter-sidebar { display: flex; flex-direction: column; gap: 16px; }

.sidebar-card { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius); padding: 16px; }
.sidebar-stat { text-align: center; }
.sidebar-stat .stat-num { font-size: 24px; font-weight: 700; }
.sidebar-stat .stat-label { font-size: 12px; color: var(--text-muted); margin-top: 4px; }
.sidebar-stat.total .stat-num { color: var(--accent-cyan); }
.sidebar-stat.attention .stat-num { color: #f59e0b; }
.sidebar-stat.alarm .stat-num { color: #ef4444; }

.alarm-list-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius); overflow: hidden; flex: 1; display: flex; flex-direction: column; }
.alarm-list-card .card-header { padding: 12px 16px; border-bottom: 1px solid var(--border-color); }
.alarm-list-card .card-title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.alarm-list { max-height: 320px; overflow-y: auto; flex: 1; }
.alarm-item { padding: 12px 16px; border-bottom: 1px solid var(--border-color); }
.alarm-item:last-child { border-bottom: none; }
.alarm-device { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.alarm-device-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.device-tag { padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 500; }
.device-tag.warning { background: #fef3c7; color: #d97706; }
.device-tag.critical { background: #fee2e2; color: #dc2626; }
.alarm-stats { display: flex; gap: 12px; }
.stat-row { display: flex; flex-direction: column; gap: 2px; }
.stat-row .stat-label { font-size: 10px; color: var(--text-muted); }
.stat-row .stat-value { font-size: 12px; font-weight: 600; }
.no-alarm { padding: 20px; text-align: center; color: var(--text-muted); font-size: 13px; }

.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; }
.stat-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius); padding: 16px; }
.stat-card.green { border-left: 3px solid #10b981; }
.stat-card.yellow { border-left: 3px solid #f59e0b; }
.stat-card.red { border-left: 3px solid #ef4444; }
.stat-card.cyan { border-left: 3px solid #06b6d4; }
.stat-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.stat-label { font-size: 12px; color: var(--text-muted); }
.stat-icon { font-size: 16px; }
.stat-value { font-size: 28px; font-weight: 700; color: var(--text-primary); }
.stat-footer { font-size: 11px; color: var(--text-muted); margin-top: 4px; }

.card-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; border-bottom: 1px solid var(--border-color); }
.card-title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.card-body { padding: 16px; }
.chart-container { width: 100%; }
</style>
