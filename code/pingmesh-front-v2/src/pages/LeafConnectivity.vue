<template>
  <div class="lc-root">
    <div class="lc-page-hdr">
      <div>
        <div class="lc-page-title">机房网络监控</div>
        <div class="lc-page-subtitle">监控机房内 Leaf 节点间连通性</div>
      </div>
      <div class="lc-page-actions">
        <div class="lc-seg-group">
          <button
            v-for="m in metrics"
            :key="m.id"
            class="lc-seg-btn"
            :class="{ active: currentMetric === m.id }"
            @click="selectMetric(m.id)"
          >{{ m.label }}</button>
        </div>
        <select v-model="updateInterval" class="lc-interval-select" @change="onIntervalChange">
          <option v-for="iv in intervalOptions" :key="iv.value" :value="iv.value">{{ iv.label }}</option>
        </select>
        <button class="lc-ctrl-btn" @click="toggleAutoPlay">
          {{ isPlaying ? '⏸ 暂停' : '▶ 播放' }}
        </button>
        <button class="lc-ctrl-btn" @click="refreshData">↻ 刷新</button>
      </div>
    </div>

    <div class="lc-idc-selector">
      <button
        v-for="idc in idcs"
        :key="idc.code"
        class="lc-idc-btn"
        :class="{ active: selectedIDC === idc.code }"
        @click="selectIDC(idc.code)"
      >
        {{ idc.name }}（{{ idc.code }}）
      </button>
    </div>

    <div class="lc-heatmap-section">
      <div class="lc-heatmap-card">
        <div class="lc-card-header">
          <span class="lc-card-title">{{ selectedIDCName }} - Leaf 连通性热力图</span>
          <span v-if="heatmapLoading" class="lc-loading-dot">●</span>
        </div>
        <div class="lc-card-body">
          <div ref="heatmapRef" class="lc-chart-container"></div>
        </div>
      </div>

      <div class="lc-leaf-list-card">
        <div class="lc-card-header">
          <span class="lc-card-title">最大耗时链路 TOP</span>
          <span class="lc-card-count">{{ top10Links.length }} 条</span>
        </div>
        <div class="lc-leaf-list">
          <div
            v-for="(link, idx) in top10Links"
            :key="link.sleaf + link.tleaf + idx"
            class="lc-leaf-item"
          >
            <div class="lc-item-rank" :class="'rank-' + (idx + 1)">{{ idx + 1 }}</div>
            <div class="lc-item-body">
              <div class="lc-item-path">
                <span class="lc-path-leaf">{{ link.sleaf }}</span>
                <span class="lc-path-arrow">→</span>
                <span class="lc-path-leaf">{{ link.tleaf }}</span>
              </div>
              <div class="lc-item-meta">
                <span class="lc-idc-badge">{{ link.idc }}</span>
                <div class="lc-item-bar-wrap">
                  <div
                    class="lc-item-bar"
                    :style="{ width: barWidth(link.max) + '%', background: barGradient(link.max) }"
                  ></div>
                </div>
              </div>
            </div>
            <div class="lc-item-value" :style="{ color: latencyColor(link.max) }">
              {{ link.max.toFixed(1) }}
              <span class="lc-item-unit">ms</span>
            </div>
          </div>
          <div v-if="top10Links.length === 0" class="lc-empty">暂无数据</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { apiFetch } from '../utils/api.js'
import { useAppStore } from '../stores/useApp.js'

const heatmapRef = ref(null)
let heatmapChart = null
let autoPlayInterval = null
let visibilityHandler = null

const isPlaying = ref(true)
const currentMetric = ref('avg')
const updateInterval = ref(5000)
const heatmapLoading = ref(false)

const metrics = [
  { id: 'avg', label: '平均' },
  { id: 'p99', label: '99分位' },
  { id: 'max', label: '最大' }
]

const intervalOptions = [
  { value: 5000, label: '5秒' },
  { value: 10000, label: '10秒' },
  { value: 60000, label: '1分钟' }
]

const idcs = ref([])
const selectedIDC = ref('')
const selectedIDCName = computed(() => {
  const found = idcs.value.find(i => i.code === selectedIDC.value)
  return found ? `${found.name}（${found.code}）` : selectedIDC.value
})

const leafs = ref([])
const maxAllLinks = ref([])
const top10Links = computed(() => maxAllLinks.value.slice(0, 10))

const barWidth = (val) => {
  if (maxAllLinks.value.length === 0) return 0
  const maxVal = maxAllLinks.value[0]?.max || 1
  return Math.max((val / maxVal) * 100, 2)
}

const barGradient = (val) => {
  if (val > 200) return 'linear-gradient(90deg, rgba(255,69,96,0.25), #ff4560)'
  if (val >= 50) return 'linear-gradient(90deg, rgba(59,130,246,0.25), #3b82f6)'
  return 'linear-gradient(90deg, rgba(57,255,126,0.25), #39ff7e)'
}

function latencyColor(v) {
  if (v > 200) return '#ff4560'
  if (v >= 50) return '#3b82f6'
  return '#39ff7e'
}

// ── 热力图 ──

function createChart() {
  if (!heatmapRef.value || heatmapChart) return
  heatmapChart = echarts.init(heatmapRef.value)
}

async function fetchHeatmap() {
  if (!selectedIDC.value || !heatmapChart) return
  heatmapLoading.value = true
  try {
    const params = new URLSearchParams({ idc: selectedIDC.value, metric: currentMetric.value })
    const res = await apiFetch(`/api/leaf-connectivity/heatmap?${params}`)
    const json = await res.json()
    if (json.data && json.data.length > 0) {
      if (json.leafs) leafs.value = json.leafs
      renderHeatmap(json.data, json.leafs || [])
    }
  } catch (e) {
    console.error('Failed to fetch heatmap data:', e)
  } finally {
    heatmapLoading.value = false
  }
}

function renderHeatmap(data, lf) {
  if (!heatmapChart) return
  const isDark = true

  heatmapChart.setOption({
    backgroundColor: 'transparent',
    tooltip: {
      position: 'top',
      formatter: (params) => {
        const v = params.value[2]
        const src = lf[params.value[1]] || `Leaf-${params.value[1] + 1}`
        const dst = lf[params.value[0]] || `Leaf-${params.value[0] + 1}`
        // 对角线（sleaf 与 tleaf 相同）
        if (params.value[0] === params.value[1]) {
          return `${src} → ${dst}<br/>自身`
        }
        return v === 0 ? `${src} → ${dst}<br/>暂无数据` : `${src} → ${dst}<br/>延迟: ${v.toFixed(2)} ms`
      }
    },
    grid: { top: 40, left: 70, right: 60, bottom: 20 },
    xAxis: {
      type: 'category',
      data: lf,
      splitArea: { show: true },
      axisLabel: { color: isDark ? '#5a8898' : '#6B7280', fontSize: 10, interval: 2, rotate: 45 },
      axisLine: { lineStyle: { color: 'rgba(0,245,196,0.08)' } }
    },
    yAxis: {
      type: 'category',
      data: lf,
      splitArea: { show: true },
      axisLabel: { color: isDark ? '#5a8898' : '#6B7280', fontSize: 10, interval: 2 },
      axisLine: { lineStyle: { color: 'rgba(0,245,196,0.08)' } }
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
      textStyle: { color: '#5a8898' },
      formatter: (v) => `${v} ms`
    },
    series: [{
      type: 'heatmap',
      data,
      label: { show: true, color: '#b8dde8', fontSize: 8, formatter: (p) => {
        return p.value[2] === 0 ? '-' : p.value[2].toFixed(1)
      }},
      emphasis: { itemStyle: { shadowBlur: 10, shadowColor: 'rgba(0,0,0,0.5)' } },
      itemStyle: { borderColor: 'rgba(0,245,196,0.08)', borderWidth: 1 }
    }]
  })
}

// ── TOP 链路 ──

async function fetchMaxAllLinks() {
  if (!selectedIDC.value) return
  try {
    const res = await apiFetch(`/api/leaf-connectivity/top-links?idc=${selectedIDC.value}`)
    const json = await res.json()
    if (json.links) {
      maxAllLinks.value = json.links
        .map(l => ({
          sleaf: l.source,
          tleaf: l.target,
          max: l.max,
          avg: l.avg,
          p99: l.p99,
          idc: selectedIDC.value
        }))
        .filter(l => l.max > 0)
    }
  } catch { /* 降级 */ }
}

// ── 交互 ──

async function selectIDC(idc) {
  selectedIDC.value = idc
  await fetchHeatmap()
  fetchMaxAllLinks()
}

function toggleAutoPlay() {
  isPlaying.value = !isPlaying.value
  isPlaying.value ? startAutoPlay() : clearInterval(autoPlayInterval)
}

function startAutoPlay() {
  clearInterval(autoPlayInterval)
  autoPlayInterval = setInterval(() => {
    const idx = idcs.value.findIndex(i => i.code === selectedIDC.value)
    if (idx >= 0 && idcs.value.length > 1) {
      selectIDC(idcs.value[(idx + 1) % idcs.value.length].code)
    }
  }, updateInterval.value)
}

function onIntervalChange() {
  if (isPlaying.value) startAutoPlay()
}

function selectMetric(id) {
  currentMetric.value = id
  fetchHeatmap()
}
async function refreshData() { await fetchHeatmap(); fetchMaxAllLinks() }

function handleResize() { heatmapChart?.resize() }

// ── 生命周期 ──

onMounted(async () => {
  // 1. 先创建 ECharts 实例
  await nextTick()
  createChart()

  // 2. 获取 IDC 列表（使用缓存）
  const appStore = useAppStore()
  try {
    const data = await appStore.fetchIdcs()
    idcs.value = data
    if (data.length > 0) selectedIDC.value = data[0].code
  } catch (e) {
    console.error('Failed to fetch IDCs:', e)
  }

  // 3. 加载数据并渲染
  await fetchHeatmap()
  fetchMaxAllLinks()

  window.addEventListener('resize', handleResize)
  startAutoPlay()

  // 页签可见性检测
  visibilityHandler = () => {
    if (document.hidden) {
      clearInterval(autoPlayInterval)
      autoPlayInterval = null
    } else {
      fetchHeatmap()
      fetchMaxAllLinks()
      if (isPlaying.value) startAutoPlay()
    }
  }
  document.addEventListener('visibilitychange', visibilityHandler)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  clearInterval(autoPlayInterval)
  if (visibilityHandler) document.removeEventListener('visibilitychange', visibilityHandler)
  heatmapChart?.dispose()
  heatmapChart = null
})
</script>

<style scoped>
.lc-root {
  flex: 1;
  overflow-y: auto;
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.lc-page-hdr {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.lc-page-title {
  font-family: 'Orbitron', monospace;
  font-size: 18px;
  font-weight: 700;
  color: var(--cyan);
  letter-spacing: 3px;
}
.lc-page-subtitle {
  font-size: 12px;
  color: var(--text-dim);
  margin-top: 4px;
}
.lc-page-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.lc-ctrl-btn {
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
.lc-ctrl-btn:hover {
  border-color: var(--cyan-dim);
  color: var(--text);
}

/* 间隔下拉框 */
.lc-interval-select {
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
.lc-interval-select:hover {
  border-color: var(--cyan-dim);
  color: var(--text);
}
.lc-interval-select:focus {
  outline: none;
  border-color: var(--cyan-dim);
}

/* 指标分段按钮组 */
.lc-seg-group {
  display: flex;
  gap: 0;
  border: 1px solid var(--bg-border);
  border-radius: 4px;
  overflow: hidden;
}
.lc-seg-btn {
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
.lc-seg-btn:last-child {
  border-right: none;
}
.lc-seg-btn:hover {
  color: var(--text);
  background: rgba(0, 245, 196, 0.04);
}
.lc-seg-btn.active {
  background: var(--cyan-ghost);
  color: var(--cyan);
}

.lc-idc-selector {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.lc-idc-btn {
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
.lc-idc-btn:hover {
  border-color: var(--cyan-dim);
  color: var(--text);
}
.lc-idc-btn.active {
  background: var(--cyan-ghost);
  border-color: var(--cyan);
  color: var(--cyan);
}

.lc-heatmap-section {
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 12px;
  flex: 1;
  min-height: 0;
}

.lc-heatmap-card,
.lc-leaf-list-card {
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.lc-heatmap-card::before,
.lc-leaf-list-card::before {
  content: '';
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--cyan-ghost), transparent);
}

.lc-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--bg-border2);
  flex-shrink: 0;
}
.lc-card-title {
  font-family: 'Orbitron', monospace;
  font-size: 11px;
  font-weight: 700;
  color: var(--cyan);
  letter-spacing: 2px;
  text-transform: uppercase;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
  min-width: 0;
}
.lc-card-count {
  font-size: 10px;
  color: var(--text-dim);
  font-family: var(--font-mono);
  flex-shrink: 0;
  margin-left: 8px;
}

.lc-loading-dot {
  color: var(--cyan);
  font-size: 8px;
  animation: lc-pulse 1s ease-in-out infinite;
}
@keyframes lc-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.lc-card-body {
  padding: 10px;
  flex: 1;
  min-height: 0;
}
.lc-chart-container {
  width: 100%;
  height: 420px;
}

/* ── 最大耗时链路 TOP ── */
.lc-leaf-list {
  overflow-y: auto;
  flex: 1;
}

.lc-leaf-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 22px 14px;
  border-bottom: 1px solid var(--bg-border2);
  transition: background 0.2s;
}
.lc-leaf-item:hover {
  background: rgba(0, 245, 196, 0.03);
}
.lc-leaf-item:last-child {
  border-bottom: none;
}

/* 排名序号 */
.lc-item-rank {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  font-family: var(--font-mono);
  flex-shrink: 0;
  background: var(--bg-border);
  color: var(--text-dim);
}
.lc-item-rank.rank-1 { background: rgba(255, 215, 0, 0.15); color: #ffd700; }
.lc-item-rank.rank-2 { background: rgba(192, 192, 192, 0.15); color: #c0c0c0; }
.lc-item-rank.rank-3 { background: rgba(205, 127, 50, 0.15); color: #cd7f32; }

/* 中间内容区 */
.lc-item-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.lc-item-path {
  display: flex;
  align-items: center;
  gap: 6px;
  font-family: var(--font-mono);
  font-size: 12px;
}
.lc-path-leaf {
  color: var(--text);
  font-weight: 600;
}
.lc-path-arrow {
  color: var(--text-dim);
  font-size: 10px;
  margin: 0 2px;
}

.lc-item-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}
.lc-idc-badge {
  display: inline-block;
  background: var(--cyan-ghost);
  border: 1px solid var(--cyan-dim);
  color: var(--cyan);
  font-size: 9px;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: var(--font-mono);
  flex-shrink: 0;
  line-height: 1;
}

.lc-item-bar-wrap {
  flex: 1;
  height: 4px;
  background: var(--bg-border);
  border-radius: 2px;
  overflow: hidden;
}
.lc-item-bar {
  height: 100%;
  border-radius: 2px;
  transition: width 0.4s ease;
}

/* 右侧数值 */
.lc-item-value {
  font-size: 15px;
  font-weight: 700;
  font-family: var(--font-mono);
  text-align: right;
  flex-shrink: 0;
  min-width: 52px;
  line-height: 1;
  position: relative;
  top: 10px;
}
.lc-item-unit {
  font-size: 10px;
  font-weight: 400;
  opacity: 0.6;
  margin-left: 2px;
}

.lc-empty {
  padding: 40px 14px;
  text-align: center;
  color: var(--text-dim);
  font-size: 12px;
}
</style>
