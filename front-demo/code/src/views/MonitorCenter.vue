<template>
  <div>
    <!-- 全局指标切换 - 页面最上方 -->
    <div class="page-header">
      <div>
        <div class="page-title">整体网络监控</div>
        <div class="page-subtitle">实时监控跨机房网络质量</div>
      </div>
      <div class="page-actions-right">
        <select v-model="currentGlobalMetric" class="metric-select" @change="switchGlobalMetric(currentGlobalMetric)">
          <option v-for="m in allMetrics" :key="m.id" :value="m.id">{{ m.label }}</option>
        </select>
        <div class="time-control">
          <select v-model="selectedMinute" class="time-select" @change="updateAll">
            <option v-for="(t, idx) in timeOptions" :key="idx" :value="idx">{{ t }}</option>
          </select>
        </div>
        <div class="update-frequency">
          <select v-model="updateInterval" class="time-select" @change="restartAutoPlay">
            <option :value="5000">5秒</option>
            <option :value="10000">10秒</option>
            <option :value="60000">1分钟</option>
          </select>
        </div>
        <button class="btn btn-outline" @click="toggleAutoPlay">
          {{ isPlaying ? '⏸️' : '▶️' }}
        </button>
        <button class="btn btn-outline" @click="refreshData">↻</button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card blue">
        <div class="stat-header">
          <span class="stat-label">跨机房延迟</span>
          <div class="stat-icon">📡</div>
        </div>
        <div class="stat-value-row">
          <div class="stat-value">{{ crossDCData.avgLatency.toFixed(1) }}<span style="font-size:16px;color:var(--text-muted)">ms</span></div>
          <div ref="crossDCChartRef" class="stat-chart"></div>
        </div>
        <div class="stat-footer">范围: {{ crossDCData.minLatency.toFixed(0) }}-{{ crossDCData.maxLatency.toFixed(0) }}ms</div>
      </div>
      <div class="stat-card cyan">
        <div class="stat-header">
          <span class="stat-label">机房内延迟</span>
          <div class="stat-icon">🏢</div>
        </div>
        <div class="stat-value-row">
          <div class="stat-value">{{ intraDCData.avgLatency.toFixed(1) }}<span style="font-size:16px;color:var(--text-muted)">ms</span></div>
          <div ref="intraDCChartRef" class="stat-chart"></div>
        </div>
        <div class="stat-footer">正常</div>
      </div>
      <div class="stat-card green">
        <div class="stat-header">
          <span class="stat-label">丢包率</span>
          <div class="stat-icon">📉</div>
        </div>
        <div class="stat-value-row">
          <div class="stat-value">{{ lossRate }}<span style="font-size:16px;color:var(--text-muted)">%</span></div>
          <div ref="lossChartRef" class="stat-chart"></div>
        </div>
        <div class="stat-footer">较上周 -0.01%</div>
      </div>
      <div class="stat-card red">
        <div class="stat-header">
          <span class="stat-label">活跃告警</span>
          <div class="stat-icon">⚠️</div>
        </div>
        <div class="stat-value-row">
          <div class="stat-value">{{ alerts.length }}</div>
          <div ref="alertChartRef" class="stat-chart"></div>
        </div>
        <div class="stat-footer" style="color:var(--accent-green)">已处理 12 条</div>
      </div>
    </div>

    <!-- 机房间网络监控 + 活跃告警 -->
    <div class="section-row">
      <div class="card">
        <div class="card-header">
          <span class="card-title">机房间网络监控</span>
        </div>
        <div class="card-body">
          <div ref="topologyRef" class="topology-container" style="height: 400px;"></div>
        </div>
      </div>
      <div class="side-card">
        <div class="card-header">
          <span class="card-title">活跃告警</span>
          <button class="btn btn-outline" style="padding:4px 8px;font-size:11px;">全部</button>
        </div>
        <div class="alert-list" style="max-height:400px;overflow-y:auto;">
          <div v-for="alert in alerts" :key="alert.id" class="alert-item">
            <div class="alert-severity" :class="alert.severity"></div>
            <div class="alert-content">
              <div class="alert-title">{{ alert.title }}</div>
              <div class="alert-desc">{{ alert.desc }}</div>
              <div class="alert-time">{{ alert.time }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 机房间网络热力图 + 延迟大的链路 TOP -->
    <div class="section-row">
      <div class="card">
        <div class="card-header">
          <span class="card-title">机房间网络热力图</span>
        </div>
        <div class="card-body">
          <div ref="heatmapRef" class="chart-container"></div>
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
      <div class="side-card">
        <div class="card-header">
          <span class="card-title">延迟大的链路 TOP</span>
        </div>
        <div class="leaf-list">
          <div v-for="leaf in topLatencyLinks" :key="leaf.source + leaf.target" class="leaf-item">
            <div class="leaf-path">{{ leaf.source }} → {{ leaf.target }}</div>
            <div class="leaf-stats">
              <div class="stat-row">
                <span class="stat-label">平均</span>
                <span class="stat-value" :style="{ color: getLatencyColor(leaf.avg) }">{{ leaf.avg.toFixed(1) }}ms</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">P999</span>
                <span class="stat-value">{{ leaf.p99.toFixed(1) }}ms</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">最大</span>
                <span class="stat-value">{{ leaf.max.toFixed(1) }}ms</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 网络设备状态 + 延迟大的设备 TOP -->
    <div class="section-row">
      <div class="card">
        <div class="card-header">
          <span class="card-title">网络设备状态</span>
          <div class="card-actions">
            <button class="card-tab" :class="{ active: nodeFilter === 'all' }" @click="nodeFilter = 'all'">全部</button>
            <button class="card-tab" :class="{ active: nodeFilter === 'normal' }" @click="nodeFilter = 'normal'">正常</button>
            <button class="card-tab" :class="{ active: nodeFilter === 'high' }" @click="nodeFilter = 'high'">耗时增加</button>
            <button class="card-tab" :class="{ active: nodeFilter === 'error' }" @click="nodeFilter = 'error'">故障超时</button>
          </div>
        </div>
        <div class="card-header" style="border-top:1px solid var(--border-color);padding-top:12px;">
          <div class="card-actions" style="gap:12px;">
            <select v-model="filterIDC" class="filter-select">
              <option value="">全部机房</option>
              <option v-for="idc in idcs" :key="idc" :value="idc">{{ idc }}</option>
            </select>
            <select v-model="filterDeviceType" class="filter-select">
              <option value="">全部设备类型</option>
              <option value="核心交换机">核心交换机</option>
              <option value="汇聚交换机">汇聚交换机</option>
              <option value="接入交换机">接入交换机</option>
            </select>
          </div>
        </div>
        <div class="card-body" style="padding:0;overflow-x:auto;max-height:420px;overflow-y:auto;">
          <table class="data-table">
            <thead>
              <tr>
                <th>机房</th>
                <th>设备类型</th>
                <th>设备名称</th>
                <th style="text-align:center;">状态</th>
                <th>平均延迟</th>
                <th>P999延迟</th>
                <th>最大延迟</th>
                <th>丢包率</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="node in filteredNodes" :key="node.name">
                <td><span class="region-tag">{{ node.idc }}</span></td>
                <td><span class="type-tag">{{ node.deviceType }}</span></td>
                <td><span class="node-name">{{ node.name }}</span></td>
                <td style="text-align:center;">
                  <span class="status-badge" :class="node.status">
                    {{ statusLabels[node.status] }}
                  </span>
                </td>
                <td>
                  <div class="latency-bar">
                    <span>{{ node.avgLatency }} ms</span>
                    <div class="latency-bar-track">
                      <div
                        class="latency-bar-fill"
                        :style="{
                          width: Math.min(node.avgLatency / 50 * 100, 100) + '%',
                          background: node.avgLatency > 30 ? (node.avgLatency > 45 ? '#ef4444' : '#f59e0b') : '#10b981'
                        }"
                      ></div>
                    </div>
                  </div>
                </td>
                <td>{{ node.p99Latency }} ms</td>
                <td>{{ node.maxLatency }} ms</td>
                <td>{{ Number(node.loss).toFixed(2) }}%</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div class="side-card">
        <div class="card-header">
          <span class="card-title">延迟大的设备 TOP</span>
        </div>
        <div class="leaf-list">
          <div v-for="device in topLatencyDevices" :key="device.name" class="leaf-item">
            <div class="leaf-path">{{ device.name }}</div>
            <div class="leaf-stats">
              <div class="stat-row">
                <span class="stat-label">平均</span>
                <span class="stat-value" :style="{ color: getLatencyColor(device.avg) }">{{ device.avg.toFixed(1) }}ms</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">P999</span>
                <span class="stat-value">{{ device.p99.toFixed(1) }}ms</span>
              </div>
              <div class="stat-row">
                <span class="stat-label">最大</span>
                <span class="stat-value" :style="{ color: getLatencyColor(device.max) }">{{ device.max.toFixed(1) }}ms</span>
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
const topologyRef = ref(null)
const crossDCChartRef = ref(null)
const intraDCChartRef = ref(null)
const lossChartRef = ref(null)
const alertChartRef = ref(null)
let heatmapChart = null
let topologyChart = null
let crossDCChart = null
let intraDCChart = null
let lossChart = null
let alertChart = null
let autoPlayInterval = null

const isPlaying = ref(true)
const currentMetric = ref('avg')
const updateInterval = ref(5000)

// 全局指标
const allMetrics = [
  { id: 'avg', label: '平均' },
  { id: 'p99', label: 'P999' },
  { id: 'max', label: '最大' }
]
const currentGlobalMetric = ref('avg')

const metrics = [
  { id: 'avg', label: '平均' },
  { id: 'p99', label: 'P999' },
  { id: 'max', label: '最大' }
]

const idcs = ['北京IDC-1', '北京IDC-2', '杭州IDC-1', '杭州IDC-2', '广州IDC-1', '广州IDC-2']

// 城市颜色配置
const cityColors = {
  '北京IDC-1': '#ef4444',
  '北京IDC-2': '#ef4444',
  '杭州IDC-1': '#f59e0b',
  '杭州IDC-2': '#f59e0b',
  '广州IDC-1': '#06b6d4',
  '广州IDC-2': '#06b6d4'
}

const defaultIcon = 'circle'

// 节点筛选
const nodeFilter = ref('all')
const filterIDC = ref('')
const filterDeviceType = ref('')

function getCityColor(idc) {
  return cityColors[idc] || '#3b82f6'
}

const crossDCData = ref({
  avgLatency: 28.5,
  minLatency: 12,
  maxLatency: 52
})

const intraDCData = ref({
  avgLatency: 15.2
})

const lossRate = ref('0.02')

const nodeList = ref([])
const alerts = ref([
  { id: 1, severity: 'critical', title: '北京IDC-1→广州IDC-1链路延迟超标', desc: '当前延迟 52ms，阈值 50ms，持续 5 分钟', time: '2分钟前' },
  { id: 2, severity: 'warning', title: '杭州IDC-1丢包率升高', desc: '丢包率 0.45%，超过警告阈值 0.3%', time: '5分钟前' },
  { id: 3, severity: 'info', title: '新增探针 gz1-srv-01 上线', desc: '探针已成功注册并开始采集数据', time: '15分钟前' }
])

const timeOptions = ref([])
const selectedMinute = ref(0)

const topLatencyLinks = ref([])
const topLatencyDevices = ref([])

const statusLabels = {
  normal: '正常',
  high: '耗时增加',
  error: '故障超时'
}

const latencyMatrixCache = []

const currentTimeDisplay = computed(() => {
  return timeOptions.value[selectedMinute.value] || '--:--'
})

const filteredNodes = computed(() => {
  return nodeList.value.filter(node => {
    if (nodeFilter.value !== 'all' && node.status !== nodeFilter.value) return false
    if (filterIDC.value && node.idc !== filterIDC.value) return false
    if (filterDeviceType.value && node.deviceType !== filterDeviceType.value) return false
    return true
  })
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

function getLossClass(loss) {
  if (loss < 0.1) return 'loss-low'
  if (loss < 0.3) return 'loss-med'
  return 'loss-high'
}

function getLatencyColor(latency) {
  // 0-50: 绿色，数值越小越浅
  if (latency < 10) return '#d1fae5'
  if (latency < 20) return '#6ee7b7'
  if (latency < 30) return '#34d399'
  if (latency < 40) return '#10b981'
  if (latency < 50) return '#059669'
  // 50-200: 蓝色，数值越小越浅
  if (latency < 75) return '#bfdbfe'
  if (latency < 100) return '#93c5fd'
  if (latency < 125) return '#60a5fa'
  if (latency < 150) return '#3b82f6'
  if (latency < 175) return '#2563eb'
  if (latency < 200) return '#1d4ed8'
  // 200以上: 橙色，数值越小越浅
  if (latency < 250) return '#fed7aa'
  if (latency < 300) return '#fdba74'
  if (latency < 350) return '#fb923c'
  if (latency < 400) return '#f97316'
  return '#ea580c'
}

// 获取渐变起始颜色（用于双向链路的起始端）
function getLatencyGradientStart(latency) {
  return getLatencyColor(latency)
}

// 获取渐变结束颜色（用于双向链路的结束端）
function getLatencyGradientEnd(latency) {
  return getLatencyColor(latency)
}

function rand(min, max) {
  return Math.random() * (max - min) + min
}

function randInt(min, max) {
  return Math.floor(rand(min, max + 1))
}

function pick(arr) {
  return arr[randInt(0, arr.length - 1)]
}

function isSameCity(i, j) {
  return Math.floor(i / 2) === Math.floor(j / 2)
}

function generateNodeList() {
  const deviceTypes = ['核心交换机', '汇聚交换机', '接入交换机']
  const statuses = ['normal', 'normal', 'normal', 'high']
  const result = []
  
  for (let i = 0; i < 16; i++) {
    const idc = idcs[i % idcs.length]
    const type = deviceTypes[randInt(0, 2)]
    const num = i + 1
    
    // 前2条为故障超时，其余随机
    const status = i < 2 ? 'error' : pick(statuses)
    
    result.push({
      name: `${idc.replace('IDC-', '')}-${type.replace('交换机', 'SW')}-${String(num).padStart(2, '0')}`,
      idc: idc,
      deviceType: type,
      status: status,
      avgLatency: rand(5, 50).toFixed(1),
      p99Latency: rand(20, 80).toFixed(1),
      maxLatency: rand(40, 120).toFixed(1),
      loss: rand(0, 0.5),
      lastHeartbeat: randInt(0, 5) === 0 ? '刚刚' : randInt(1, 5) + '分钟前'
    })
  }
  
  return result
}

function generateTopLatencyLinks() {
  const links = []
  for (let i = 0; i < idcs.length; i++) {
    for (let j = i + 1; j < idcs.length; j++) {
      const isSame = isSameCity(i, j)
      const avg = isSame ? rand(12, 18) : rand(25, 45)
      links.push({
        source: idcs[i],
        target: idcs[j],
        avg: avg,
        p99: avg * rand(1.5, 2.5),
        max: avg * rand(2, 4)
      })
    }
  }
  links.sort((a, b) => b.avg - a.avg)
  topLatencyLinks.value = links.slice(0, 8)
}

function generateTopLatencyDevices() {
  const deviceTypes = ['核心交换机', '汇聚交换机', '接入交换机']
  const devices = []
  
  for (let i = 0; i < 8; i++) {
    const idc = idcs[randInt(0, idcs.length - 1)]
    const type = deviceTypes[randInt(0, deviceTypes.length - 1)]
    const avg = rand(30, 90)
    
    devices.push({
      idc: idc,
      type: type,
      name: `${idc}-${type.replace('交换机', 'SW')}-${String(i+1).padStart(2, '0')}`,
      avg: avg,
      p99: avg * rand(1.3, 1.8),
      max: avg * rand(1.5, 2.2)
    })
  }
  
  devices.sort((a, b) => b.avg - a.avg)
  topLatencyDevices.value = devices
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

function generateLatencyMatrix(metric = currentGlobalMetric.value) {
  const matrix = []
  for (let i = 0; i < idcs.length; i++) {
    matrix[i] = []
    for (let j = 0; j < idcs.length; j++) {
      if (i === j) {
        matrix[i][j] = 0
      } else if (isSameCity(i, j)) {
        // 同城机房
        if (metric === 'p99') {
          matrix[i][j] = rand(20, 45)
        } else if (metric === 'max') {
          // 约1/3在60-70
          if (Math.random() < 0.33) {
            matrix[i][j] = rand(60, 70)
          } else {
            matrix[i][j] = rand(20, 50)
          }
        } else {
          matrix[i][j] = rand(12, 18)
        }
      } else {
        // 异地机房
        if (metric === 'p99') {
          // 大部分在50以下，个别60-70
          if (Math.random() < 0.2) {
            matrix[i][j] = rand(60, 70)
          } else {
            matrix[i][j] = rand(20, 48)
          }
        } else if (metric === 'max') {
          // 大部分在50以下，约1/3在60-70，个别210左右
          const randVal = Math.random()
          if (randVal < 0.1) {
            matrix[i][j] = rand(200, 220)
          } else if (randVal < 0.4) {
            matrix[i][j] = rand(60, 70)
          } else {
            matrix[i][j] = rand(20, 50)
          }
        } else {
          matrix[i][j] = rand(25, 35)
        }
      }
    }
  }
  for (let i = 0; i < idcs.length; i++) {
    for (let j = i + 1; j < idcs.length; j++) {
      matrix[i][j] = matrix[j][i] = (matrix[i][j] + matrix[j][i]) / 2
    }
  }
  return matrix
}

function initHeatmap() {
  if (!heatmapRef.value) return
  heatmapChart = echarts.init(heatmapRef.value)
  updateHeatmap()
}

function updateHeatmap() {
  const colors = getThemeColors()
  const matrix = generateLatencyMatrix()
  
  const data = []
  const highlightData = []
  for (let i = 0; i < idcs.length; i++) {
    for (let j = 0; j < idcs.length; j++) {
      const latency = matrix[i][j]
      data.push([j, i, latency])
      // 蓝色(>=50)或橙色(>=200)时添加到高亮列表
      if (latency >= 50) {
        highlightData.push({ source: idcs[i], target: idcs[j], latency })
      }
    }
  }

  const option = {
    backgroundColor: colors.bg,
    tooltip: {
      position: 'top',
      formatter: function(params) {
        const source = idcs[params.value[1]]
        const target = idcs[params.value[0]]
        const latency = params.value[2]
        if (latency === 0) return `${source} → ${target}<br/>-`
        return `${source} → ${target}<br/><strong>${latency.toFixed(0)} ms</strong>`
      }
    },
    grid: { top: 40, left: 80, right: 70, bottom: 60 },
    xAxis: {
      type: 'category',
      data: idcs,
      splitArea: { show: true },
      axisLabel: { color: colors.textMuted, fontSize: 12, interval: 0, rotate: 45 },
      axisLine: { lineStyle: { color: colors.grid } }
    },
    yAxis: {
      type: 'category',
      data: idcs,
      splitArea: { show: true },
      axisLabel: { color: colors.textMuted, fontSize: 12 },
      axisLine: { lineStyle: { color: colors.grid } }
    },
    visualMap: {
      min: 0, max: 250, calculable: true, orient: 'vertical', right: 10, top: 10,
      inRange: { color: ['#d1fae5', '#6ee7b7', '#34d399', '#10b981', '#059669', '#bfdbfe', '#93c5fd', '#60a5fa', '#3b82f6', '#2563eb', '#1d4ed8', '#fed7aa', '#fdba74', '#fb923c', '#f97316', '#ea580c'] },
      textStyle: { color: colors.textMuted }
    },
    series: [{
      type: 'heatmap',
      data: data,
      label: { show: true, color: colors.text, fontSize: 11, formatter: (p) => p.value[2] === 0 ? '-' : p.value[2].toFixed(0) },
      itemStyle: { borderColor: colors.grid, borderWidth: 1, borderRadius: 2 }
    }]
  }
  heatmapChart.setOption(option)

  // 自动显示高亮单元格的tooltip（轮流展示）
  if (highlightData.length > 0) {
    let idx = 0
    const showNextTooltip = () => {
      if (!heatmapChart) return
      const item = highlightData[idx]
      const sourceIdx = idcs.indexOf(item.source)
      const targetIdx = idcs.indexOf(item.target)
      heatmapChart.dispatchAction({
        type: 'showTip',
        seriesIndex: 0,
        dataIndex: [sourceIdx, targetIdx]
      })
      idx = (idx + 1) % highlightData.length
    }
    // 每3秒切换一次
    setInterval(showNextTooltip, 3000)
  }
}

function initTopology() {
  if (!topologyRef.value) return
  topologyChart = echarts.init(topologyRef.value)
  updateTopology()
}

function updateTopology() {
  const colors = getThemeColors()
  const matrix = generateLatencyMatrix(currentGlobalMetric.value)
  
  const getPosition = (idx, total) => {
    const angle = (idx / total) * Math.PI * 2 - Math.PI / 2
    const radius = 35
    const centerX = 50
    const centerY = 50
    return [centerX + radius * Math.cos(angle), centerY + radius * Math.sin(angle)]
  }

  const nodeData = idcs.map((idc, idx) => {
    const pos = getPosition(idx, idcs.length)
    return {
      name: idc, x: pos[0], y: pos[1],
      itemStyle: { color: getCityColor(idc), shadowBlur: 15, shadowColor: getCityColor(idc) + '80' },
      symbol: defaultIcon, symbolSize: 30
    }
  })

  const linesData = []
  const labelData = []
  for (let i = 0; i < idcs.length; i++) {
    for (let j = i + 1; j < idcs.length; j++) {
      const posI = getPosition(i, idcs.length)
      const posJ = getPosition(j, idcs.length)
      const latency = matrix[i][j]
      const color = getLatencyColor(latency)
      const startColor = getLatencyGradientStart(latency)
      const endColor = getLatencyGradientEnd(latency)
      
      // 正向链路：从 i 到 j，使用渐变色
      linesData.push({
        coords: [posI, posJ],
        latency,
        lineStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 1, y2: 0,
            colorStops: [
              { offset: 0, color: startColor },
              { offset: 1, color: endColor }
            ]
          },
          width: 1.3,
          opacity: 0.9,
          curveness: 0
        }
      })
      // 反向链路：从 j 到 i，使用渐变色（反向渐变）
      linesData.push({
        coords: [posJ, posI],
        latency,
        lineStyle: {
          color: {
            type: 'linear',
            x: 0, y: 0, x2: 1, y2: 0,
            colorStops: [
              { offset: 0, color: endColor },
              { offset: 1, color: startColor }
            ]
          },
          width: 1.3,
          opacity: 0.9,
          curveness: 0
        }
      })
      
      const midX = (posI[0] + posJ[0]) / 2
      const midY = (posI[1] + posJ[1]) / 2
      labelData.push({ value: [midX, midY], latency, color })
    }
  }

  const option = {
    backgroundColor: colors.bg,
    xAxis: { min: 0, max: 100, show: false, silent: true },
    yAxis: { min: 0, max: 100, show: false, silent: true, inverse: false },
    series: [
      {
        type: 'scatter', zlevel: 2, symbolSize: 30,
        data: nodeData.map(n => ({ name: n.name, value: [n.x, n.y], itemStyle: n.itemStyle })),
        label: { show: true, position: 'bottom', formatter: (p) => p.name, fontSize: 12, color: colors.text, fontWeight: 600, distance: 8 }
      },
      {
        type: 'lines', zlevel: 1, coordinateSystem: 'cartesian2d',
        effect: { show: true, period: 2, trailLength: 0.4, symbol: 'circle', symbolSize: 5, color: '#ffffff', loop: true },
        lineStyle: { width: 1.3, opacity: 0.9, curveness: 0 },
        data: linesData
      },
      {
        type: 'scatter', coordinateSystem: 'cartesian2d', zlevel: 3, symbol: 'none', data: labelData,
        label: { show: true, position: 'center', formatter: (p) => p.data.latency.toFixed(0) + 'ms', fontSize: 10, color: '#fff', fontWeight: 700, backgroundColor: (p) => p.data.color, padding: [2, 5], borderRadius: 3 }
      }
    ]
  }
  topologyChart.setOption(option)
}

function switchGlobalMetric(metric) {
  currentGlobalMetric.value = metric
  currentMetric.value = metric
  updateAll()
}

function switchMetric(metric) {
  currentMetric.value = metric
  updateHeatmap()
}

function updateAll() {
  updateHeatmap()
  updateTopology()
  generateTopLatencyLinks()
  generateTopLatencyDevices()
}

function toggleAutoPlay() {
  isPlaying.value = !isPlaying.value
  if (isPlaying.value) startAutoPlay()
  else clearInterval(autoPlayInterval)
}

function startAutoPlay() {
  clearInterval(autoPlayInterval)
  autoPlayInterval = setInterval(() => {
    selectedMinute.value = (selectedMinute.value + 1) % timeOptions.value.length
    updateAll()
    generateTimeOptions()
    // 轮播节点筛选状态
    const filters = ['all', 'normal', 'high', 'error']
    const currentIdx = filters.indexOf(nodeFilter.value)
    nodeFilter.value = filters[(currentIdx + 1) % filters.length]
  }, updateInterval.value)
}

function restartAutoPlay() {
  if (isPlaying.value) startAutoPlay()
}

function refreshData() {
  nodeList.value = generateNodeList()
  crossDCData.value = { avgLatency: rand(25, 35), minLatency: rand(8, 15), maxLatency: rand(40, 55) }
  intraDCData.value = { avgLatency: rand(12, 18) }
  lossRate.value = rand(0.005, 0.05).toFixed(2)
  generateTimeOptions()
  updateAll()
}

function handleResize() {
  heatmapChart?.resize()
  topologyChart?.resize()
  crossDCChart?.resize()
  intraDCChart?.resize()
  lossChart?.resize()
  alertChart?.resize()
}

function handleThemeChange() {
  updateHeatmap()
  updateTopology()
  updateStatCharts()
}

// 生成3小时历史数据（18个点，每10分钟一个）
function generateHistoryData(baseValue, variance, sparseSpike = false) {
  const data = []
  for (let i = 0; i < 18; i++) {
    if (sparseSpike && Math.random() < 0.15) {
      // 偶发升高
      data.push(baseValue + rand(variance * 3, variance * 8))
    } else {
      data.push(rand(0, variance))
    }
  }
  return data
}

function initStatCharts() {
  const colors = getThemeColors()
  
  // 跨机房延迟图表
  if (crossDCChartRef.value) {
    crossDCChart = echarts.init(crossDCChartRef.value)
    const historyData = generateHistoryData(28, 10)
    crossDCChart.setOption({
      backgroundColor: 'transparent',
      grid: { top: 5, left: 5, right: 5, bottom: 5 },
      xAxis: { type: 'category', show: false, data: Array(18).fill('') },
      yAxis: { type: 'value', show: false, min: 'dataMin', max: 'dataMax' },
      series: [{
        type: 'line',
        data: historyData,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#3b82f6', width: 1.5 },
        areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(59,130,246,0.3)' }, { offset: 1, color: 'rgba(59,130,246,0.05)' }] } }
      }]
    })
  }
  
  // 机房内延迟图表
  if (intraDCChartRef.value) {
    intraDCChart = echarts.init(intraDCChartRef.value)
    const historyData = generateHistoryData(15, 5)
    intraDCChart.setOption({
      backgroundColor: 'transparent',
      grid: { top: 5, left: 5, right: 5, bottom: 5 },
      xAxis: { type: 'category', show: false, data: Array(18).fill('') },
      yAxis: { type: 'value', show: false, min: 'dataMin', max: 'dataMax' },
      series: [{
        type: 'line',
        data: historyData,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#06b6d4', width: 1.5 },
        areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(6,182,212,0.3)' }, { offset: 1, color: 'rgba(6,182,212,0.05)' }] } }
      }]
    })
  }
  
  // 丢包率图表 - 大部分时间为0，偶发升高
  if (lossChartRef.value) {
    lossChart = echarts.init(lossChartRef.value)
    const historyData = generateHistoryData(0, 0.05, true)
    lossChart.setOption({
      backgroundColor: 'transparent',
      grid: { top: 5, left: 5, right: 5, bottom: 5 },
      xAxis: { type: 'category', show: false, data: Array(18).fill('') },
      yAxis: { type: 'value', show: false, min: 0, max: 0.5 },
      series: [{
        type: 'line',
        data: historyData,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#10b981', width: 1.5 },
        areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(16,185,129,0.3)' }, { offset: 1, color: 'rgba(16,185,129,0.05)' }] } }
      }]
    })
  }
  
  // 告警图表 - 横线为主，偶发尖刺
  if (alertChartRef.value) {
    alertChart = echarts.init(alertChartRef.value)
    const baseValue = 2
    const data = []
    for (let i = 0; i < 18; i++) {
      if (Math.random() < 0.15) {
        // 偶发尖刺
        data.push(baseValue + rand(3, 8))
      } else {
        data.push(baseValue)
      }
    }
    alertChart.setOption({
      backgroundColor: 'transparent',
      grid: { top: 5, left: 5, right: 5, bottom: 5 },
      xAxis: { type: 'category', show: false, data: Array(18).fill('') },
      yAxis: { type: 'value', show: false, min: 0, max: 12 },
      series: [{
        type: 'line',
        data: data,
        smooth: false,
        symbol: 'none',
        lineStyle: { color: '#ef4444', width: 1.5 },
        areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(239,68,68,0.3)' }, { offset: 1, color: 'rgba(239,68,68,0.05)' }] } }
      }]
    })
  }
}

function updateStatCharts() {
  const colors = getThemeColors()
  // 刷新图表主题
  if (crossDCChart) crossDCChart.setOption({ backgroundColor: 'transparent' })
  if (intraDCChart) intraDCChart.setOption({ backgroundColor: 'transparent' })
  if (lossChart) lossChart.setOption({ backgroundColor: 'transparent' })
  if (alertChart) alertChart.setOption({ backgroundColor: 'transparent' })
}

watch(() => document.documentElement.getAttribute('data-theme'), handleThemeChange)

onMounted(() => {
  // 先初始化数据
  nodeList.value = generateNodeList()
  alerts.value = [
    { id: 1, severity: 'critical', title: '北京IDC-1→广州IDC-1链路延迟超标', desc: '当前延迟 52ms，阈值 50ms，持续 5 分钟', time: '2分钟前' },
    { id: 2, severity: 'warning', title: '杭州IDC-1丢包率升高', desc: '丢包率 0.45%，超过警告阈值 0.3%', time: '5分钟前' },
    { id: 3, severity: 'info', title: '新增探针 gz1-srv-01 上线', desc: '探针已成功注册并开始采集数据', time: '15分钟前' }
  ]
  crossDCData.value = { avgLatency: 28.5, minLatency: 12, maxLatency: 52 }
  intraDCData.value = { avgLatency: 15.2 }
  lossRate.value = '0.02'
  
  generateTimeOptions()
  generateTopLatencyLinks()
  generateTopLatencyDevices()
  initHeatmap()
  initTopology()
  initStatCharts()
  window.addEventListener('resize', handleResize)
  startAutoPlay()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  clearInterval(autoPlayInterval)
  heatmapChart?.dispose()
  topologyChart?.dispose()
  crossDCChart?.dispose()
  intraDCChart?.dispose()
  lossChart?.dispose()
  alertChart?.dispose()
})
</script>

<style scoped>
.page-header { display: flex; align-items: center; justify-content: space-between; padding: 16px 20px; background: var(--bg-card); border-bottom: 1px solid var(--border-color); }
.page-title { font-size: 20px; font-weight: 600; color: var(--text-primary); }
.page-subtitle { font-size: 13px; color: var(--text-muted); margin-top: 4px; }
.page-actions { display: flex; align-items: center; gap: 12px; }
.page-actions-right { display: flex; align-items: flex-end; gap: 8px; }
.global-metric-tabs { display: flex; gap: 4px; background: var(--bg-page); padding: 4px; border-radius: 8px; }
.metric-tab { padding: 8px 16px; border: none; background: transparent; color: var(--text-muted); cursor: pointer; border-radius: 6px; font-size: 13px; font-weight: 500; transition: all 0.2s; }
.metric-tab.active { background: var(--accent-blue); color: white; }
.metric-tab:hover:not(.active) { background: var(--border-color); }
.metric-select { background: transparent; border: 1px solid var(--border-color); border-radius: var(--radius-sm); padding: 6px 10px; font-size: 13px; color: var(--text-primary); cursor: pointer; }

.time-control { display: flex; align-items: center; }
.update-frequency { display: flex; align-items: center; }
.time-select { background: transparent; border: 1px solid var(--border-color); border-radius: var(--radius-sm); padding: 6px 10px; font-size: 12px; color: var(--text-primary); cursor: pointer; }

.timeline-control { display: flex; align-items: center; gap: 12px; padding: 12px 16px; background: var(--bg-card); border-top: 1px solid var(--border-color); }
.timeline-btn { background: var(--accent-blue); border: none; border-radius: 50%; width: 32px; height: 32px; display: flex; align-items: center; justify-content: center; cursor: pointer; font-size: 12px; }
.timeline-slider { flex: 1; height: 6px; }
.timeline-slider input { width: 100%; height: 6px; -webkit-appearance: none; background: var(--border-color); border-radius: 3px; outline: none; }
.timeline-slider input::-webkit-slider-thumb { -webkit-appearance: none; width: 14px; height: 14px; background: var(--accent-blue); border-radius: 50%; cursor: pointer; }
.timeline-time { font-size: 12px; color: var(--text-muted); white-space: nowrap; }

.stats-row { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin: 20px 0; }
.stat-card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius); padding: 12px 16px; position: relative; }
.stat-card.blue { border-left: 2px solid #3b82f6; }
.stat-card.cyan { border-left: 2px solid #06b6d4; }
.stat-card.green { border-left: 2px solid #10b981; }
.stat-card.red { border-left: 2px solid #ef4444; }
.stat-header { display: flex; justify-content: space-between; align-items: center; }
.stat-label { font-size: 12px; color: var(--text-muted); }

.leaf-list { max-height: 420px; overflow-y: auto; }
.leaf-item { padding: 12px 16px; border-bottom: 1px solid var(--border-color); }
.leaf-item:last-child { border-bottom: none; }
.leaf-path { font-size: 12px; color: var(--text-primary); margin-bottom: 8px; font-weight: 500; }
.leaf-stats { display: flex; gap: 12px; }
.stat-row { display: flex; flex-direction: column; gap: 2px; }
.stat-row .stat-label { font-size: 11px; }
.stat-row .stat-value { font-size: 13px; font-weight: 600; }
.stat-value-row { display: flex; justify-content: space-between; align-items: center; margin: 4px 0; }
.stat-value { font-size: 24px; font-weight: 700; color: var(--text-primary); }
.stat-footer { font-size: 11px; color: var(--text-muted); }
.stat-chart { width: 80px; height: 36px; opacity: 0.8; flex-shrink: 0; }

.topology-section { display: grid; grid-template-columns: 1fr 300px; gap: 20px; }
.topology-main { min-width: 0; display: flex; flex-direction: column; }
.topology-side { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius); overflow: hidden; height: fit-content; max-height: 560px; }
.topology-container { flex: 1; height: 400px; display: flex; align-items: flex-start; justify-content: flex-start; }

.section-row { display: flex; gap: 20px; margin-bottom: 20px; align-items: stretch; }
.section-row .card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius); flex: 1; height: 400px; display: flex; flex-direction: column; overflow: hidden; }
.section-row .card-header { display: flex; justify-content: space-between; align-items: center; padding: 12px 16px; border-bottom: 1px solid var(--border-color); flex-shrink: 0; }
.section-row .card-title { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.section-row .card-actions { display: flex; gap: 4px; }
.section-row .card-tab { padding: 6px 12px; border: none; background: transparent; color: var(--text-muted); cursor: pointer; border-radius: 4px; font-size: 12px; }
.section-row .card-tab.active { background: var(--accent-blue); color: white; }
.section-row .card-body { padding: 16px; flex: 1; min-height: 0; overflow: hidden; }

.filter-select { background: transparent; border: 1px solid var(--border-color); border-radius: var(--radius-sm); padding: 6px 10px; font-size: 12px; color: var(--text-primary); cursor: pointer; }

.side-card { width: 300px; background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius); overflow: hidden; flex-shrink: 0; height: 400px; display: flex; flex-direction: column; }

.alert-list { max-height: 400px; overflow-y: auto; }
.alert-item { display: flex; gap: 12px; padding: 12px 16px; border-bottom: 1px solid var(--border-color); }
.alert-item:last-child { border-bottom: none; }
.alert-severity { width: 4px; border-radius: 2px; flex-shrink: 0; }
.alert-severity.critical { background: #ef4444; }
.alert-severity.warning { background: #f59e0b; }
.alert-severity.info { background: #3b82f6; }
.alert-content { flex: 1; }
.alert-title { font-size: 13px; font-weight: 600; color: var(--text-primary); }
.alert-desc { font-size: 12px; color: var(--text-muted); margin-top: 4px; }
.alert-time { font-size: 11px; color: var(--text-muted); margin-top: 4px; }

.status-badge { display: inline-block; padding: 4px 10px; border-radius: 4px; font-size: 12px; font-weight: 500; }
.status-badge.normal { background: #d1fae5; color: #059669; }
.status-badge.high { background: #fef3c7; color: #d97706; }
.status-badge.error { background: #fee2e2; color: #dc2626; }
</style>