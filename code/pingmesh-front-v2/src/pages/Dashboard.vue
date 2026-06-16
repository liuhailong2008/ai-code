<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { nodes, events, protos } from '@/composables/useData'
import { useTopo } from '@/composables/useTopo'
import Chart from 'chart.js/auto'

const router = useRouter()

// ── KPI 数据 ──
const kpiData = ref({
  bw: '38.5',
  pkt: '0.32',
  lat: '11',
  sess: '14,230',
  alerts: '3',
  nodesOnline: '8',
})

let kpiTimer = null

function updateKpi() {
  kpiData.value.bw = (36 + Math.random() * 5).toFixed(1)
  kpiData.value.pkt = (0.2 + Math.random() * 0.3).toFixed(2)
  kpiData.value.lat = Math.round(9 + Math.random() * 5)
  kpiData.value.sess = Math.round(13500 + Math.random() * 1500).toLocaleString()
}

// ── Event 过滤 ──
const eventFilter = ref('all')
const filteredEvents = ref([...events])
const badgeMap = { err: '严重', warn: '警告', ok: '恢复' }

function filterEvents(type) {
  eventFilter.value = type
  filteredEvents.value = type === 'all' ? [...events] : events.filter(e => e.type === type)
}

// ── Mini Topo Canvas ──
const miniTopoCanvas = ref(null)
const { init: initMiniTopo } = useTopo(miniTopoCanvas, { headerH: 38 })

// ── Live Traffic Chart ──
const trafficCanvas = ref(null)
let trafficChart = null
const trafficIn = Array.from({ length: 60 }, () => +(20 + Math.random() * 20).toFixed(1))
const trafficOut = trafficIn.map(v => +(v * 0.65).toFixed(1))
let trafficTimer = null

function initTrafficChart() {
  if (!trafficCanvas.value || trafficChart) return
  trafficChart = new Chart(trafficCanvas.value, {
    type: 'line',
    data: {
      labels: Array.from({ length: 60 }, (_, i) => i === 59 ? 'now' : ''),
      datasets: [
        {
          label: '入站 Gbps', data: [...trafficIn],
          borderColor: '#00f5c4', borderWidth: 1.5, pointRadius: 0,
          fill: true, backgroundColor: 'rgba(0,245,196,0.06)', tension: 0.35,
        },
        {
          label: '出站 Gbps', data: [...trafficOut],
          borderColor: '#39ff7e', borderWidth: 1.5, pointRadius: 0,
          fill: true, backgroundColor: 'rgba(57,255,126,0.05)', tension: 0.35,
        },
      ],
    },
    options: {
      responsive: true, maintainAspectRatio: false, animation: { duration: 0 },
      plugins: { legend: { display: false }, tooltip: { enabled: false } },
      scales: {
        x: { display: false },
        y: {
          display: true, min: 0,
          ticks: { color: '#3a6070', font: { size: 9, family: 'JetBrains Mono' }, maxTicksLimit: 4, callback: v => v + 'G' },
          grid: { color: 'rgba(0,245,196,0.04)' }, border: { display: false },
        },
      },
    },
  })

  trafficTimer = setInterval(() => {
    if (!trafficChart) return
    const newIn = +(18 + Math.random() * 25).toFixed(1)
    const newOut = +(newIn * 0.65).toFixed(1)
    trafficIn.push(newIn); trafficIn.shift()
    trafficOut.push(newOut); trafficOut.shift()
    trafficChart.data.datasets[0].data = [...trafficIn]
    trafficChart.data.datasets[1].data = [...trafficOut]
    trafficChart.update('none')
  }, 2000)
}

// ── Protocol Doughnut Chart ──
const protoCanvas = ref(null)
let protoChart = null

function initProtoChart() {
  if (!protoCanvas.value || protoChart) return
  protoChart = new Chart(protoCanvas.value, {
    type: 'doughnut',
    data: {
      labels: protos.map(p => p.name),
      datasets: [{ data: protos.map(p => p.pct), backgroundColor: protos.map(p => p.color), borderWidth: 0, hoverOffset: 4 }],
    },
    options: {
      responsive: false, cutout: '68%',
      plugins: { legend: { display: false }, tooltip: { enabled: true } },
      animation: { animateRotate: true, duration: 800 },
    },
  })
}

// ── Heatmap ──
const heatmapWrap = ref(null)

function renderHeatmap() {
  if (!heatmapWrap.value) return
  const el = heatmapWrap.value
  const nodeNames = ['RTR-01','RTR-02','SW-01','SW-02','FW-01','BJ-01','GZ-01','WH-01']
  const hours = Array.from({ length: 24 }, (_, i) => i)
  const vals = nodeNames.map(() => hours.map(() => Math.random()))

  const cellW = Math.floor((el.clientWidth - 48) / 24)
  const cellH = Math.floor((el.clientHeight - 20) / nodeNames.length - 1) || 14

  el.innerHTML = ''
  let html = `<div style="display:flex;gap:2px;margin-bottom:4px;margin-left:48px;">`
  html += hours.filter((_, i) => i % 4 === 0).map(h => `<div style="width:${cellW * 4}px;font-size:9px;color:#3a6070">${h}:00</div>`).join('')
  html += '</div>'

  nodeNames.forEach((name, ni) => {
    html += `<div style="display:flex;align-items:center;gap:2px;margin-bottom:2px;">
      <div style="width:46px;font-size:9px;color:#5a8898;text-align:right;padding-right:4px;white-space:nowrap;overflow:hidden">${name}</div>`
    hours.forEach(h => {
      const v = vals[ni][h]
      const a = v > 0.8 ? 0.9 : v > 0.6 ? 0.7 : v > 0.4 ? 0.5 : v > 0.2 ? 0.3 : 0.1
      const c = v > 0.8 ? '#ff4560' : v > 0.6 ? '#ffc233' : '#00f5c4'
      html += `<div style="width:${cellW}px;height:${cellH}px;background:${c};opacity:${a};border-radius:1px" title="${name} ${h}:00 ${Math.round(v*100)}%"></div>`
    })
    html += '</div>'
  })
  el.innerHTML = html
}

// ── SparkLines ──
function drawSparkLines() {
  const sparkData = [8,12,9,14,11,16,13,18,15,20]
  const containers = [
    { id: 'sparkBw',   data: sparkData, color: '#00f5c4' },
    { id: 'sparkPkt',  data: [3,5,4,6,5,7,6,8,7,9], color: '#ffc233' },
    { id: 'sparkLat',  data: [12,10,11,9,10,8,9,11,10,11], color: '#00f5c4' },
    { id: 'sparkSess', data: [11,13,12,14,13,15,14,16,14,17], color: '#39ff7e' },
    { id: 'sparkAlrt', data: [1,2,1,2,3,2,3,2,3,3], color: '#ff4560' },
    { id: 'sparkNode', data: [10,10,9,10,10,9,10,10,9,8], color: '#ffc233' },
  ]
  containers.forEach(c => {
    const el = document.getElementById(c.id)
    if (!el) return
    const W = 60, H = 30
    const maxVal = Math.max(...c.data)
    const bars = c.data.map(v => (v / maxVal) * H)
    el.innerHTML = `<svg width="${W}" height="${H}" style="display:block">
      ${bars.map((h, i) => `<rect x="${i * 7}" y="${H - h}" width="5" height="${h}" rx="1" fill="${c.color}" opacity="${0.3 + (i / bars.length) * 0.7}"/>`).join('')}
    </svg>`
  })
}

// ── Lifecycle ──
onMounted(async () => {
  await nextTick()
  initMiniTopo()
  initTrafficChart()
  initProtoChart()
  drawSparkLines()
  setTimeout(renderHeatmap, 200)
  kpiTimer = setInterval(updateKpi, 3000)
})

onBeforeUnmount(() => {
  if (trafficChart) { trafficChart.destroy(); trafficChart = null }
  if (protoChart) { protoChart.destroy(); protoChart = null }
  if (trafficTimer) clearInterval(trafficTimer)
  if (kpiTimer) clearInterval(kpiTimer)
})
</script>

<template>
  <div class="page-container">
    <!-- KPI Strip -->
    <div class="kpi-strip">
      <div class="kpi-card" data-status="ok">
        <div class="kpi-icon">◈</div>
        <div class="kpi-body">
          <div class="kpi-label">总带宽</div>
          <div class="kpi-value">{{ kpiData.bw }} <small>Gbps</small></div>
          <div class="kpi-trend up">↑ 12% vs 昨日</div>
        </div>
        <div class="kpi-spark" id="sparkBw"></div>
      </div>
      <div class="kpi-card" data-status="warn">
        <div class="kpi-icon">◎</div>
        <div class="kpi-body">
          <div class="kpi-label">丢包率</div>
          <div class="kpi-value warn">{{ kpiData.pkt }} <small>%</small></div>
          <div class="kpi-trend warn">▲ 超阈值 0.30%</div>
        </div>
        <div class="kpi-spark" id="sparkPkt"></div>
      </div>
      <div class="kpi-card" data-status="ok">
        <div class="kpi-icon">⧖</div>
        <div class="kpi-body">
          <div class="kpi-label">平均时延</div>
          <div class="kpi-value">{{ kpiData.lat }} <small>ms</small></div>
          <div class="kpi-trend up">↓ 2ms vs 昨日</div>
        </div>
        <div class="kpi-spark" id="sparkLat"></div>
      </div>
      <div class="kpi-card" data-status="ok">
        <div class="kpi-icon">⬡</div>
        <div class="kpi-body">
          <div class="kpi-label">活跃会话</div>
          <div class="kpi-value">{{ kpiData.sess }}</div>
          <div class="kpi-trend up">↑ 850 并发</div>
        </div>
        <div class="kpi-spark" id="sparkSess"></div>
      </div>
      <div class="kpi-card" data-status="err">
        <div class="kpi-icon">⚠</div>
        <div class="kpi-body">
          <div class="kpi-label">活跃告警</div>
          <div class="kpi-value err">3</div>
          <div class="kpi-trend err">2 严重 · 1 警告</div>
        </div>
        <div class="kpi-spark" id="sparkAlrt"></div>
      </div>
      <div class="kpi-card" data-status="ok">
        <div class="kpi-icon">▣</div>
        <div class="kpi-body">
          <div class="kpi-label">在线节点</div>
          <div class="kpi-value">{{ kpiData.nodesOnline }} <small>/ 10</small></div>
          <div class="kpi-trend warn">2 节点异常</div>
        </div>
        <div class="kpi-spark" id="sparkNode"></div>
      </div>
    </div>

    <!-- Main Grid -->
    <div class="dash-grid">
      <!-- Topo Mini Map -->
      <div class="dash-card topo-card">
        <div class="card-header">
          <span class="card-title">网络拓扑</span>
          <div class="card-actions">
            <button class="card-btn" @click="router.push({name:'topology'})">展开</button>
          </div>
        </div>
        <canvas ref="miniTopoCanvas" class="mini-topo-canvas"></canvas>
      </div>

      <!-- Live Traffic -->
      <div class="dash-card traffic-card">
        <div class="card-header">
          <span class="card-title">实时流量 — 60s</span>
          <div class="legend-row">
            <span class="leg-item cyan">■ 入站</span>
            <span class="leg-item green">■ 出站</span>
          </div>
        </div>
        <canvas ref="trafficCanvas" class="traffic-canvas"></canvas>
      </div>

      <!-- Node Status -->
      <div class="dash-card nodes-card">
        <div class="card-header">
          <span class="card-title">节点状态</span>
          <span class="card-subtitle">实时</span>
        </div>
        <div class="node-list">
          <div v-for="n in nodes" :key="n.id" class="node-item">
            <div class="ni-dot" :class="n.status"></div>
            <div class="ni-name">{{ n.id }}</div>
            <div class="ni-lat" :style="{color: n.status==='err'?'#ff4560':n.status==='warn'?'#ffc233':'#39ff7e'}">
              {{ n.status === 'err' ? '—' : n.lat + 'ms' }}
            </div>
            <div class="ni-bar">
              <div class="ni-bar-fill" :style="{width:(n.cpu||0)+'%',background:n.cpu>70?'#ff4560':n.cpu>50?'#ffc233':'#39ff7e'}"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Protocol Pie -->
      <div class="dash-card proto-card">
        <div class="card-header">
          <span class="card-title">协议分布</span>
          <span class="card-subtitle">今日</span>
        </div>
        <div class="proto-chart-wrap">
          <canvas ref="protoCanvas"></canvas>
          <div class="proto-legend">
            <div v-for="p in protos" :key="p.name" class="pl-item">
              <div class="pl-dot" :style="{background:p.color}"></div>
              <span class="pl-name">{{ p.name }}</span>
              <span class="pl-val">{{ p.pct }}%</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Event Stream -->
      <div class="dash-card events-card">
        <div class="card-header">
          <span class="card-title">事件流</span>
          <div class="event-filters">
            <button class="ef-btn" :class="{active:eventFilter==='all'}" @click="filterEvents('all')">全部</button>
            <button class="ef-btn" :class="{active:eventFilter==='err'}" @click="filterEvents('err')">严重</button>
            <button class="ef-btn" :class="{active:eventFilter==='warn'}" @click="filterEvents('warn')">警告</button>
          </div>
        </div>
        <div class="event-list">
          <div v-for="(e, i) in filteredEvents" :key="i" class="event-item" :class="e.type">
            <div class="ev-time">{{ e.time }}</div>
            <div class="ev-body">
              <div class="ev-host">{{ e.host }}</div>
              <div class="ev-msg">{{ e.msg }}</div>
            </div>
            <div class="ev-badge" :class="e.type">{{ badgeMap[e.type] }}</div>
          </div>
        </div>
      </div>

      <!-- Heatmap -->
      <div class="dash-card heatmap-card">
        <div class="card-header">
          <span class="card-title">带宽热力图</span>
          <span class="card-subtitle">24h × 节点</span>
        </div>
        <div ref="heatmapWrap" class="heatmap-wrap"></div>
      </div>
    </div>
  </div>
</template>
