<template>
  <div>
    <div class="page-header">
      <div>
        <div class="page-title">机房网络监控</div>
        <div class="page-subtitle">监控机房内 Leaf 节点间连通性</div>
      </div>
<div class="page-actions">
        <select v-model="currentMetric" class="metric-select" @change="switchMetric(currentMetric)">
          <option v-for="m in metrics" :key="m.id" :value="m.id">{{ m.label }}</option>
        </select>
        <select v-model="carouselMode" class="btn btn-outline" style="padding: 9px 14px;" @change="restartAutoPlay">
          <option value="idc">按机房</option>
          <option value="time">按时间</option>
        </select>
        <select v-model="selectedMinute" class="btn btn-outline" style="padding: 9px 14px;" @change="updateAll">
          <option v-for="(t, idx) in timeOptions" :key="idx" :value="idx">{{ t }}</option>
        </select>
        <select v-model="updateInterval" class="btn btn-outline" style="padding: 9px 14px;" @change="restartAutoPlay">
          <option :value="5000">5秒</option>
          <option :value="10000">10秒</option>
          <option :value="60000">1分钟</option>
        </select>
        <button class="btn btn-outline" @click="toggleAutoPlay">
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

    <div class="heatmap-section">
      <div class="heatmap-card">
        <div class="card-header">
          <span class="card-title">{{ selectedIDC }} - Leaf 连通性热力图</span>
        </div>
        <div class="card-body">
          <div ref="heatmapRef" class="chart-container" style="height: 450px;"></div>
        </div>
        <div class="timeline-control">
          <button class="timeline-btn" @click="toggleAutoPlay">
            {{ isPlaying ? '⏸️' : '▶️' }}
          </button>
          <div class="timeline-slider">
            <input type="range" v-model.number="selectedMinute" :max="timeOptions.length - 1" min="0" step="1" @input="updateAll">
          </div>
          <span class="timeline-time">{{ currentTimeDisplay }}</span>
        </div>
      </div>

      <div class="leaf-list-card">
        <div class="card-header">
          <span class="card-title">最大耗时链路 TOP</span>
        </div>
        <div class="leaf-list">
          <div v-for="link in topLatencyLinks" :key="link.source + link.target" class="leaf-item">
            <div class="leaf-path">{{ link.source }} → {{ link.target }}</div>
            <div class="leaf-stats">
              <div class="stat-row">
                <span class="stat-label">平均</span>
                <span class="stat-value" :style="{ color: getTopLatencyColor(link.avg) }">{{ link.avg.toFixed(1) }}ms</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">P999</span>
                <span class="stat-value" :style="{ color: getTopLatencyColor(link.p99) }">
                  {{ link.p99.toFixed(1) }}ms
                  <span v-if="link.p99 > 200" class="warning-icon">⚠️</span>
                </span>
              </div>
              <div class="stat-row">
                <span class="stat-label">最大</span>
                <span class="stat-value" :style="{ color: getTopLatencyColor(link.max) }">
                  {{ link.max.toFixed(1) }}ms
                  <span v-if="link.max > 200" class="warning-icon">⚠️</span>
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'

const heatmapRef = ref(null)
let heatmapChart = null
let autoPlayInterval = null

const isPlaying = ref(true)
const currentMetric = ref('avg')
const updateInterval = ref(5000)
const carouselMode = ref('idc') // 轮播模式：idc-按机房, time-按时间
const lastUpdate = ref('--')

// 格式化时间
function fmtTime() {
  const now = new Date()
  return `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
}

const metrics = [
  { id: 'avg', label: '平均' },
  { id: 'p99', label: 'P999' },
  { id: 'max', label: '最大' }
]

const idcs = ['北京IDC-1', '北京IDC-2', '杭州IDC-1', '杭州IDC-2', '广州IDC-1', '广州IDC-2']
const selectedIDC = ref('北京IDC-1')

const leafs = Array.from({ length: 15 }, (_, i) => `Leaf-${String(i + 1).padStart(2, '0')}`)

const timeOptions = ref([])
const selectedMinute = ref(0)
const topLatencyLinks = ref([])

const currentTimeDisplay = computed(() => {
  return timeOptions.value[selectedMinute.value] || '--:--'
})

function generateTimeOptions() {
  const options = []
  const now = new Date()
  for (let i = 0; i < 60; i++) {
    const time = new Date(now.getTime() - i * 60000)
    options.unshift(`${String(time.getHours()).padStart(2, '0')}:${String(time.getMinutes()).padStart(2, '0')}`)
  }
  timeOptions.value = options
}

function getLatencyColor(latency) {
  if (latency < 50) return '#10b981'
  if (latency < 250) return '#3b82f6'
  return '#f97316'
}

function getTopLatencyColor(latency) {
  if (latency > 200) return '#ef4444'    // 红色
  if (latency >= 50) return '#3b82f6'    // 蓝色
  return '#10b981'                        // 绿色
}

function rand(min, max) {
  return Math.random() * (max - min) + min
}

function generateLeafData() {
  const data = []
  for (let i = 0; i < leafs.length; i++) {
    for (let j = 0; j < leafs.length; j++) {
      const latency = i === j ? 0 : rand(0.1, 8.5)
      data.push([j, i, parseFloat(latency.toFixed(2))])
    }
  }
  
  // 随机选择2条记录，P999设置为50-60
  const p99Indices = []
  while (p99Indices.length < 2) {
    const idx = Math.floor(Math.random() * data.length)
    const val = data[idx][2]
    // 排除对角线（自身）和已选中的
    if (val !== 0 && !p99Indices.includes(idx)) {
      p99Indices.push(idx)
      data[idx][2] = rand(50, 60)
    }
  }
  
  // 随机选择2条记录，最大值设置为200-210
  const maxIndices = []
  while (maxIndices.length < 2) {
    const idx = Math.floor(Math.random() * data.length)
    const val = data[idx][2]
    // 排除对角线（自身）和已选中的
    if (val !== 0 && !maxIndices.includes(idx) && !p99Indices.includes(idx)) {
      maxIndices.push(idx)
      data[idx][2] = rand(200, 210)
    }
  }
  
  return data
}

function generateTopLatencyLinks() {
  const links = []
  const totalPairs = (leafs.length * (leafs.length - 1)) / 2
  
  for (let i = 0; i < leafs.length; i++) {
    for (let j = i + 1; j < leafs.length; j++) {
      const avg = rand(0.1, 8.5)
      links.push({
        source: leafs[i],
        target: leafs[j],
        avg: avg,
        p99: avg * rand(1.5, 2.5),
        max: avg * rand(2, 4)
      })
    }
  }
  
  // 随机选择2条记录，P999设置为50-60
  const p99Indices = []
  while (p99Indices.length < 2) {
    const idx = Math.floor(Math.random() * links.length)
    if (!p99Indices.includes(idx)) {
      p99Indices.push(idx)
      links[idx].p99 = rand(50, 60)
    }
  }
  
  // 随机选择2条记录，最大值设置为200-210
  const maxIndices = []
  while (maxIndices.length < 2) {
    const idx = Math.floor(Math.random() * links.length)
    if (!maxIndices.includes(idx)) {
      maxIndices.push(idx)
      links[idx].max = rand(200, 210)
    }
  }
  
  links.sort((a, b) => b.max - a.max)
  topLatencyLinks.value = links.slice(0, 8)
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

function initHeatmap() {
  if (!heatmapRef.value) return
  heatmapChart = echarts.init(heatmapRef.value)
  updateHeatmap()
}

function updateHeatmap() {
  const colors = getThemeColors()
  const data = generateLeafData()

  const option = {
    backgroundColor: colors.bg,
    tooltip: {
      position: 'top',
      formatter: (params) => {
        const value = params.value[2]
        if (value === 0) return `${leafs[params.value[1]]} → ${leafs[params.value[0]]}<br/>自身`
        return `${leafs[params.value[1]]} → ${leafs[params.value[0]]}<br/>延迟: ${value.toFixed(2)} ms`
      }
    },
    grid: {
      top: 40,
      left: 80,
      right: 60,
      bottom: 20
    },
    xAxis: {
      type: 'category',
      data: leafs,
      splitArea: { show: true },
      axisLabel: { color: colors.textMuted, fontSize: 10, interval: 2, rotate: 45 },
      axisLine: { lineStyle: { color: colors.grid } }
    },
    yAxis: {
      type: 'category',
      data: leafs,
      splitArea: { show: true },
      axisLabel: { color: colors.textMuted, fontSize: 10, interval: 2 },
      axisLine: { lineStyle: { color: colors.grid } }
    },
    visualMap: {
      min: 0,
      max: 300,
      calculable: true,
      orient: 'vertical',
      right: 10,
      top: 'center',
      inRange: {
        color: ['#065f46', '#059669', '#10b981', '#34d399', '#6ee7b7', '#3b82f6', '#8b5cf6', '#f97316', '#ea580c', '#dc2626']
      },
      textStyle: { color: colors.textMuted },
      formatter: (value) => `${value} ms`
    },
    series: [{
      type: 'heatmap',
      data: data,
      label: { show: true, color: colors.text, fontSize: 8, formatter: (p) => p.value[2] === 0 ? '-' : p.value[2].toFixed(1) },
      emphasis: { itemStyle: { shadowBlur: 10, shadowColor: 'rgba(0, 0, 0, 0.5)' } },
      itemStyle: { borderColor: colors.grid, borderWidth: 1 }
    }]
  }

  heatmapChart.setOption(option)
}

function updateAll() {
  lastUpdate.value = fmtTime()
  updateHeatmap()
  generateTopLatencyLinks()
}

function selectIDC(idc) {
  selectedIDC.value = idc
  updateAll()
}

function toggleAutoPlay() {
  isPlaying.value = !isPlaying.value
  if (isPlaying.value) {
    startAutoPlay()
  } else {
    clearInterval(autoPlayInterval)
  }
}

function startAutoPlay() {
  clearInterval(autoPlayInterval)
  autoPlayInterval = setInterval(() => {
    // 按机房轮播时，每分钟更新一次时间选项
    if (carouselMode.value === 'idc') {
      generateTimeOptions()
      selectedMinute.value = 0 // 始终选择最新时间
    }
    
    if (carouselMode.value === 'time') {
      // 按时间轮播：只更新时间，机房固定
      selectedMinute.value = (selectedMinute.value + 1) % timeOptions.value.length
    } else {
      // 按机房轮播：只轮播机房
      const currentIdx = idcs.indexOf(selectedIDC.value)
      selectedIDC.value = idcs[(currentIdx + 1) % idcs.length]
    }
    updateAll()
    generateTimeOptions()
  }, updateInterval.value)
}

function restartAutoPlay() {
  if (isPlaying.value) {
    startAutoPlay()
  }
}

function switchMetric(metric) {
  currentMetric.value = metric
  updateHeatmap()
}

function refreshData() {
  generateTimeOptions()
  updateAll()
}

function handleResize() {
  heatmapChart?.resize()
}

watch(() => document.documentElement.getAttribute('data-theme'), updateHeatmap)

onMounted(() => {
  generateTimeOptions()
  generateTopLatencyLinks()
  initHeatmap()
  window.addEventListener('resize', handleResize)
  startAutoPlay()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  clearInterval(autoPlayInterval)
  heatmapChart?.dispose()
})
</script>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; background: var(--bg-card); border-bottom: 1px solid var(--border-color); }
.page-title { font-size: 20px; font-weight: 600; color: var(--text-primary); }
.page-subtitle { font-size: 13px; color: var(--text-muted); margin-top: 4px; }
.page-actions { display: flex; align-items: center; gap: 12px; }
.metric-select { background: transparent; border: 1px solid var(--border-color); border-radius: var(--radius-sm); padding: 6px 10px; font-size: 13px; color: var(--text-primary); cursor: pointer; }

.time-control { display: flex; align-items: center; }
.update-frequency { display: flex; align-items: center; }
.time-select { background: transparent; border: 1px solid var(--border-color); border-radius: var(--radius-sm); padding: 6px 10px; font-size: 12px; color: var(--text-primary); cursor: pointer; }

.idc-selector {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.idc-btn {
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.idc-btn:hover {
  border-color: var(--accent-blue);
  color: var(--text-primary);
}

.idc-btn.active {
  background: var(--accent-blue);
  border-color: var(--accent-blue);
  color: white;
}

.heatmap-section {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 20px;
}

.heatmap-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  overflow: hidden;
}

.leaf-list-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  overflow: hidden;
}

.leaf-list {
  max-height: 520px;
  overflow-y: auto;
}

.leaf-item {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}

.leaf-item:last-child {
  border-bottom: none;
}

.leaf-path {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.leaf-stats {
  display: flex;
  gap: 12px;
}

.stat-row {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-row .stat-label {
  font-size: 10px;
  color: var(--text-muted);
}

.stat-row .stat-value {
  font-size: 12px;
  font-weight: 600;
}

.timeline-control {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 22px;
  border-top: 1px solid var(--border-color);
  background: var(--bg-card);
}

.timeline-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--accent-blue);
  color: white;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.timeline-btn:hover {
  background: var(--accent-cyan);
  transform: scale(1.1);
}

.timeline-slider {
  flex: 1;
}

.timeline-slider input[type="range"] {
  width: 100%;
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  outline: none;
  -webkit-appearance: none;
}

.timeline-slider input[type="range"]::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 16px;
  height: 16px;
  background: var(--accent-blue);
  border-radius: 50%;
  cursor: pointer;
}

.timeline-slider input[type="range"]::-webkit-slider-thumb:hover {
  background: var(--accent-cyan);
}

.timeline-time {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  min-width: 50px;
}

.warning-icon {
  margin-left: 4px;
  font-size: 12px;
}
</style>
