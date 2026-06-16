<template>
  <div class="lm-root">
    <div class="lm-page-hdr">
      <div>
        <div class="lm-page-title">机房耗时监控</div>
        <div class="lm-page-subtitle">监控节点耗时与丢包率分布</div>
      </div>
      <div class="lm-page-actions">
        <div class="lm-seg-group">
          <button
            v-for="m in statTypes"
            :key="m.id"
            class="lm-seg-btn"
            :class="{ active: currentStat === m.id }"
            @click="selectStat(m.id)"
          >{{ m.label }}</button>
        </div>
        <select v-model="updateInterval" class="lm-interval-select" @change="onIntervalChange">
          <option v-for="iv in intervalOptions" :key="iv.value" :value="iv.value">{{ iv.label }}</option>
        </select>
        <button class="lm-ctrl-btn" @click="togglePlay">
          {{ isPlaying ? '⏸ 暂停' : '▶ 播放' }}
        </button>
        <button class="lm-ctrl-btn" @click="refreshData">↻ 刷新</button>
      </div>
    </div>

    <div class="lm-idc-selector">
      <button
        v-for="idc in idcs"
        :key="idc.code"
        class="lm-idc-btn"
        :class="{ active: selectedIDC === idc.code }"
        @click="selectIDC(idc.code)"
      >
        {{ idc.name }}（{{ idc.code }}）
      </button>
    </div>

    <div class="lm-scatter-section">
      <div class="lm-scatter-card">
        <div class="lm-card-header">
          <span class="lm-card-title">耗时-丢包率散点图</span>
          <span class="lm-card-count">共 {{ scatterData.length }} 个Leaf</span>
        </div>
        <div class="lm-card-body">
          <div ref="scatterRef" class="lm-chart-container"></div>
        </div>
      </div>
      <div class="lm-scatter-sidebar">
        <div class="lm-sidebar-card">
          <div class="lm-stat-item total">
            <div class="lm-stat-num">{{ stats.total }}</div>
            <div class="lm-stat-lab">总设备数</div>
          </div>
          <div class="lm-stat-item attention">
            <div class="lm-stat-num">{{ stats.warning }}</div>
            <div class="lm-stat-lab">需关注</div>
          </div>
          <div class="lm-stat-item alarm">
            <div class="lm-stat-num">{{ stats.critical }}</div>
            <div class="lm-stat-lab">告警设备</div>
          </div>
        </div>
        <div class="lm-alarm-list-card">
          <div class="lm-card-header">
            <span class="lm-card-title">异常设备清单</span>
          </div>
          <div class="lm-alarm-list">
            <div v-for="device in abnormalDevices" :key="device.name" class="lm-alarm-item">
              <div class="lm-alarm-device-row">
                <span class="lm-alarm-device">{{ device.name }}</span>
                <span class="lm-device-tag" :class="device.status">{{ device.status === 'critical' ? '告警设备' : '需关注' }}</span>
              </div>
              <div class="lm-alarm-stats">
                <div class="lm-stat-row">
                  <span class="lm-stat-label">丢包率</span>
                  <span class="lm-stat-val">{{ device.loss.toFixed(2) }}%</span>
                </div>
                <div class="lm-stat-row">
                  <span class="lm-stat-label">平均</span>
                  <span class="lm-stat-val">{{ device.avg.toFixed(1) }}ms</span>
                </div>
                <div class="lm-stat-row">
                  <span class="lm-stat-label">99分位</span>
                  <span class="lm-stat-val">{{ device.p99.toFixed(1) }}ms</span>
                </div>
                <div class="lm-stat-row">
                  <span class="lm-stat-label">最大</span>
                  <span class="lm-stat-val">{{ device.max.toFixed(1) }}ms</span>
                </div>
              </div>
            </div>
            <div v-if="abnormalDevices.length === 0" class="lm-no-alarm">暂无异常设备</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import * as echarts from 'echarts'
import { apiFetch } from '../utils/api.js'
import { useAppStore } from '../stores/useApp.js'

const scatterRef = ref(null)
let playInterval = null
let scatterChart = null
let visibilityHandler = null
let debounceTimer = null

const isPlaying = ref(true)
const currentStat = ref('avg')
const selectedIDC = ref('BJ12')
const updateInterval = ref(5000)

const idcs = ref([])

const statTypes = [
  { id: 'avg', label: '平均' },
  { id: 'p99', label: '99分位' },
  { id: 'max', label: '最大' }
]

const intervalOptions = [
  { value: 5000, label: '5秒' },
  { value: 10000, label: '10秒' },
  { value: 60000, label: '1分钟' }
]

const stats = ref({ normal: 0, warning: 0, critical: 0, total: 0 })
const abnormalDevices = ref([])
const scatterData = ref([])

async function fetchScatterData() {
  try {
    const res = await apiFetch(`/api/latency/scatter?idc=${selectedIDC.value}&stat=${currentStat.value}`)
    const data = await res.json()
    if (data.error) {
      console.error('Scatter API error:', data.error)
      return
    }
    stats.value = {
      total: data.total || 0,
      normal: data.normal || 0,
      warning: data.warning || 0,
      critical: data.critical || 0
    }
    abnormalDevices.value = data.abnormalDevices || []
    scatterData.value = data.points || []
    updateScatter()
  } catch (err) {
    console.error('Fetch scatter data failed:', err)
  }
}

function initScatter() {
  if (!scatterRef.value) return
  scatterChart = echarts.init(scatterRef.value)
}

function updateScatter() {
  if (!scatterChart) return

  const data = scatterData.value
  // 横坐标：取 latency 最大值的 210%
  const xMaxVal = data.length > 0 ? Math.max(...data.map(d => d.value[0])) : 0
  const xMax = Math.max(Math.ceil(xMaxVal * 2.1), 2)
  // 纵坐标：取丢包率最大值的 120%，如果全为 0 则取 0.01
  const yMaxVal = data.length > 0 ? Math.max(...data.map(d => d.value[1])) : 0
  const yMax = yMaxVal > 0 ? Math.ceil(yMaxVal * 1.2 * 100) / 100 : 0.01

  const option = {
    backgroundColor: 'transparent',
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
      nameTextStyle: { color: '#5a8898' },
      axisLabel: { color: '#5a8898' },
      axisLine: { lineStyle: { color: 'rgba(0,245,196,0.08)' } },
      splitLine: { lineStyle: { color: 'rgba(0,245,196,0.06)', type: 'dashed' } },
      min: 0,
      max: xMax
    },
    yAxis: {
      type: 'value',
      name: '丢包率 (%)',
      nameLocation: 'middle',
      nameGap: 45,
      nameTextStyle: { color: '#5a8898' },
      axisLabel: { color: '#5a8898', formatter: v => v.toFixed(1) },
      axisLine: { lineStyle: { color: 'rgba(0,245,196,0.08)' } },
      splitLine: { lineStyle: { color: 'rgba(0,245,196,0.06)', type: 'dashed' } },
      min: 0,
      max: yMax
    },
    series: [{
      type: 'scatter',
      data: scatterData.value,
      symbolSize: (_val, params) => {
        const color = params.data?.itemStyle?.color
        return color === '#ff4560' ? 18 : color === '#ffab00' ? 14 : 10
      },
      emphasis: { scale: 1.5 },
      markLine: {
        silent: true,
        symbol: 'none',
        lineStyle: { type: 'dashed', width: 1.5 },
        label: { show: true, position: 'end', color: '#5a8898' },
        data: [
          { xAxis: 150, name: '150ms' },
          { yAxis: 2.5, name: '2.5%' }
        ]
      }
    }]
  }

  scatterChart.setOption(option, { notMerge: true })
}

function togglePlay() {
  isPlaying.value = !isPlaying.value
  if (isPlaying.value) { startAutoPlay() } else { clearInterval(playInterval) }
}

function startAutoPlay() {
  if (idcs.value.length === 0) return
  clearInterval(playInterval)
  playInterval = setInterval(() => {
    const currentIdx = idcs.value.findIndex(i => i.code === selectedIDC.value)
    if (currentIdx < 0) return
    selectedIDC.value = idcs.value[(currentIdx + 1) % idcs.value.length].code
    fetchScatterData()
  }, updateInterval.value)
}

function onIntervalChange() {
  if (isPlaying.value) startAutoPlay()
}

function selectStat(id) {
  currentStat.value = id
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => fetchScatterData(), 300)
}

function selectIDC(idc) {
  selectedIDC.value = idc
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => fetchScatterData(), 300)
}

function refreshData() {
  fetchScatterData()
}

function handleResize() {
  scatterChart?.resize()
}

onMounted(() => {
  const appStore = useAppStore()
  appStore.fetchIdcs().then(data => {
    idcs.value = data
    startAutoPlay()
  }).catch(() => {})
  initScatter()
  fetchScatterData()
  window.addEventListener('resize', handleResize)

  // 页签可见性检测
  visibilityHandler = () => {
    if (document.hidden) {
      clearInterval(playInterval)
      playInterval = null
    } else {
      fetchScatterData()
      if (isPlaying.value) startAutoPlay()
    }
  }
  document.addEventListener('visibilitychange', visibilityHandler)
})

onUnmounted(() => {
  clearInterval(playInterval)
  if (visibilityHandler) document.removeEventListener('visibilitychange', visibilityHandler)
  window.removeEventListener('resize', handleResize)
  scatterChart?.dispose()
})
</script>

<style scoped>
.lm-root {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.lm-page-hdr {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.lm-page-title {
  font-family: 'Orbitron', monospace;
  font-size: 18px;
  font-weight: 700;
  color: var(--cyan);
  letter-spacing: 3px;
}
.lm-page-subtitle {
  font-size: 12px;
  color: var(--text-dim);
  margin-top: 4px;
}
.lm-page-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.lm-ctrl-btn {
  background: none;
  border: 1px solid var(--bg-border);
  color: var(--text-dim);
  font-family: var(--font-mono);
  font-size: 11px;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
}
.lm-ctrl-btn:hover {
  border-color: var(--cyan-dim);
  color: var(--text);
}

/* 间隔下拉框 */
.lm-interval-select {
  appearance: none;
  -webkit-appearance: none;
  background: var(--bg-card);
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='6'%3E%3Cpath d='M0 0l5 6 5-6z' fill='%235a8898'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
  padding-right: 26px;
  border: 1px solid var(--bg-border);
  color: var(--text-dim);
  font-family: var(--font-mono);
  font-size: 11px;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.15s;
  min-width: 80px;
}
.lm-interval-select:hover {
  border-color: var(--cyan-dim);
  color: var(--text);
}
.lm-interval-select:focus {
  outline: none;
  border-color: var(--cyan-dim);
}

/* 指标分段按钮组 */
.lm-seg-group {
  display: flex;
  gap: 0;
  border: 1px solid var(--bg-border);
  border-radius: 4px;
  overflow: hidden;
}
.lm-seg-btn {
  background: none;
  border: none;
  color: var(--text-dim);
  font-family: var(--font-mono);
  font-size: 11px;
  padding: 6px 12px;
  cursor: pointer;
  transition: all 0.15s;
  border-right: 1px solid var(--bg-border);
}
.lm-seg-btn:last-child {
  border-right: none;
}
.lm-seg-btn:hover {
  color: var(--text);
  background: rgba(0, 245, 196, 0.04);
}
.lm-seg-btn.active {
  background: var(--cyan-ghost);
  color: var(--cyan);
}

.lm-idc-selector {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.lm-idc-btn {
  padding: 7px 14px;
  border-radius: 4px;
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  color: var(--text-dim);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  font-family: var(--font-mono);
}
.lm-idc-btn:hover {
  border-color: var(--cyan-dim);
  color: var(--text);
}
.lm-idc-btn.active {
  background: var(--cyan-ghost);
  border-color: var(--cyan);
  color: var(--cyan);
}

.lm-scatter-section {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 12px;
  flex: 1;
  min-height: 0;
}

.lm-scatter-card {
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.lm-scatter-card::before {
  content: '';
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--cyan-ghost), transparent);
}

.lm-scatter-sidebar {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.lm-sidebar-card {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  padding: 16px;
  position: relative;
}
.lm-sidebar-card::before {
  content: '';
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--cyan-ghost), transparent);
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
}
.lm-stat-item {
  text-align: center;
  position: relative;
}
.lm-stat-num {
  font-size: 22px;
  font-weight: 700;
  font-family: 'Orbitron', monospace;
}
.lm-stat-item.total .lm-stat-num { color: var(--cyan); }
.lm-stat-item.attention .lm-stat-num { color: var(--warn); }
.lm-stat-item.alarm .lm-stat-num { color: var(--err); }
.lm-stat-lab {
  font-size: 11px;
  color: var(--text-dim);
  margin-top: 4px;
}

.lm-alarm-list-card {
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  overflow: hidden;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.lm-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--bg-border2);
  flex-shrink: 0;
}
.lm-card-title {
  font-family: 'Orbitron', monospace;
  font-size: 11px;
  font-weight: 700;
  color: var(--cyan);
  letter-spacing: 2px;
  text-transform: uppercase;
}
.lm-card-count {
  font-size: 10px;
  color: var(--text-dim);
  font-family: var(--font-mono);
}

.lm-card-body {
  padding: 10px;
  flex: 1;
  min-height: 0;
}
.lm-chart-container {
  width: 100%;
  height: 400px;
}

.lm-alarm-list {
  max-height: 300px;
  overflow-y: auto;
  flex: 1;
}
.lm-alarm-item {
  padding: 10px 14px;
  border-bottom: 1px solid var(--bg-border2);
}
.lm-alarm-item:last-child {
  border-bottom: none;
}
.lm-alarm-device {
  font-size: 12px;
  font-weight: 600;
  color: var(--text);
  font-family: var(--font-mono);
}
.lm-alarm-device-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}
.lm-device-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  font-family: var(--font-mono);
}
.lm-device-tag.warning {
  background: rgba(255,171,0,0.15);
  color: var(--warn);
}
.lm-device-tag.critical {
  background: rgba(255,69,96,0.15);
  color: var(--err);
}
.lm-alarm-stats {
  display: flex;
  gap: 12px;
}
.lm-stat-row {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.lm-stat-label {
  font-size: 9px;
  color: var(--text-dim);
}
.lm-stat-val {
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.lm-no-alarm {
  padding: 20px;
  text-align: center;
  color: var(--text-dim);
  font-size: 12px;
}
</style>
