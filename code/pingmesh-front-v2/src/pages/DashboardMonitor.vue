<template>
  <div class="dm-root">
    <!-- Background layers -->
    <div class="dm-bg">
      <div class="dm-bg-grid"></div>
      <div class="dm-bg-scan"></div>
      <div class="dm-bg-beam"></div>
      <div class="dm-china-map">
        <svg id="china-map-svg" viewBox="0 0 1000 800" xmlns="http://www.w3.org/2000/svg"></svg>
      </div>
    </div>

    <!-- Main layout -->
    <div class="dm-layout">
      <header class="dm-hdr">
        <div class="dm-hdr-live"><div class="dm-ldot"></div>LIVE</div>
        <div class="dm-hdr-title">
          <div class="dm-hdr-glow"></div>
          <div class="dm-hdr-cr dm-hdr-cr-tl"></div>
          <div class="dm-hdr-cr dm-hdr-cr-tr"></div>
          <div class="dm-hdr-cr dm-hdr-cr-bl"></div>
          <div class="dm-hdr-cr dm-hdr-cr-br"></div>
          数据中心网络监控
        </div>
        <div class="dm-hdr-clk">
          <div class="dm-hdr-time" id="hdr-time">--:--:--</div>
          <div class="dm-hdr-date" id="hdr-date"></div>
        </div>
      </header>
      <div class="dm-side-l" id="sideL"></div>
      <div class="dm-center">
        <canvas id="tc"></canvas>
        <div class="dm-coords-panel" id="coords"></div>
      </div>
      <div class="dm-side-r" id="sideR"></div>
      <div class="dm-bottom">
        <div class="dm-ip">
          <div class="dm-ip-title">机房间网络</div>
          <div class="dm-ig">
            <div>
              <div class="dm-il-head">丢包</div>
              <div id="idc-loss-list" class="dm-idc-mgrid"></div>
            </div>
            <div>
              <div class="dm-il-head">时延</div>
              <div id="idc-latency-list" class="dm-idc-mgrid"></div>
            </div>
          </div>
        </div>
        <div class="dm-ip">
          <div class="dm-ip-title">本周告警统计</div>
          <div class="dm-alert-stats-split">
            <div class="dm-stats-col">
              <div class="dm-stats-col-title">按类型</div>
              <div v-if="typeBars.length" class="dm-bars-wrap">
                <div v-for="b in typeBars" :key="b.type" class="dm-bar-row">
                  <span class="dm-bar-label">{{ b.type }}</span>
                  <div class="dm-bar-track">
                    <div class="dm-bar-fill" :style="{ width: b.pct + '%', background: b.color }"></div>
                  </div>
                  <span class="dm-bar-count">{{ b.count }}</span>
                </div>
              </div>
              <div v-else class="dm-ir"><span class="dm-il">--</span></div>
            </div>
            <div class="dm-stats-col">
              <div class="dm-stats-col-title">按状态</div>
              <div class="dm-pie-wrap" v-if="pieSegments.length">
                <svg viewBox="0 0 72 72" class="dm-pie-svg">
                  <path v-for="seg in pieSegments" :key="seg.status" :d="seg.path" :fill="seg.color" stroke="#0a1628" stroke-width="1.5"/>
                </svg>
                <div class="dm-pie-legend">
                  <div v-for="seg in pieSegments" :key="seg.status" class="dm-pl-item">
                    <span class="dm-pl-dot" :style="{ background: seg.color }"></span>
                    <span class="dm-pl-text">{{ seg.status }}</span>
                    <span class="dm-pl-num">{{ seg.count }}</span>
                  </div>
                </div>
              </div>
              <div v-else class="dm-ir"><span class="dm-il">--</span></div>
            </div>
          </div>
        </div>
        <div class="dm-ip" id="alert-info-panel" @mouseenter="pauseAlertCarousel" @mouseleave="resumeAlertCarousel">
          <div class="dm-ip-title">告警信息</div>
          <div class="dm-alert-page" id="alert-page-container">
            <div v-for="(a, i) in currentAlertPage" :key="i" class="dm-amrow">
              <span class="dm-atime">{{ formatAlertTime(a.create_time) }}</span>
              <span class="dm-atxt"><strong>{{ a.alert_idc }}</strong> {{ a.alert_leaf || '' }} <span class="dm-abadge">{{ a.alert_type || '' }}</span></span>
              <span class="dm-astatus" :class="a.alert_status === '告警中' ? 'active' : 'recovered'">{{ a.alert_status || '' }}</span>
            </div>
            <div v-if="currentAlertPage.length === 0" class="dm-amrow">
              <span class="dm-atime">--/-- --:--</span>
              <span class="dm-atxt">暂无告警</span>
            </div>
            <div class="dm-alert-nav">
              <button class="dm-alert-nav-btn" :disabled="!hasPrevPage" @click="prevAlertPage" @mouseenter.stop @mouseleave.stop>&#8249;</button>
              <button class="dm-alert-nav-btn" :disabled="!hasNextPage" @click="nextAlertPage" @mouseenter.stop @mouseleave.stop>&#8250;</button>
            </div>
          </div>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { apiFetch } from '../utils/api.js'

const DATA = reactive({
  nodes: {},
  links: [],
  panels: {},
  treeNodes: {}
})

const alertPageList = ref([])
const currentAlertPage = computed(() => {
  const list = alertPageList.value
  if (!list.length) return []
  const totalPages = Math.ceil(list.length / ALERT_PAGE_SIZE)
  const idx = alertPageIdx % totalPages
  const start = idx * ALERT_PAGE_SIZE
  return list.slice(start, start + ALERT_PAGE_SIZE)
})

const alertTotalPages = computed(() => Math.max(1, Math.ceil(alertPageList.value.length / ALERT_PAGE_SIZE)))
const hasPrevPage = computed(() => alertTotalPages.value > 1 && alertPageIdx > 0)
const hasNextPage = computed(() => alertTotalPages.value > 1 && alertPageIdx < alertTotalPages.value - 1)

let summaryStatus = {}
let thresholds = {}

const alertTypeStats = ref([])
const alertStatusStats = ref([])

// 饼图颜色映射
const PIE_COLORS = { '告警中': '#c96e6e', '处置中': '#c9a050', '已处置': '#5a9a6a' }
const PIE_CX = 36, PIE_CY = 36, PIE_R = 30

function polarToCartesian(cx, cy, r, angleDeg) {
  const rad = (angleDeg - 90) * Math.PI / 180
  return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) }
}

function arcPath(startDeg, endDeg, r, cx, cy) {
  const s = polarToCartesian(cx, cy, r, endDeg)
  const e = polarToCartesian(cx, cy, r, startDeg)
  const large = endDeg - startDeg > 180 ? 1 : 0
  return `M ${s.x} ${s.y} A ${r} ${r} 0 ${large} 0 ${e.x} ${e.y} L ${cx} ${cy} Z`
}

const pieSegments = computed(() => {
  const segs = []
  const total = alertStatusStats.value.reduce((s, i) => s + i.count, 0)
  if (total === 0) return segs
  // 固定顺序：告警中 → 处置中 → 已处置
  const ORDER = ['告警中', '处置中', '已处置']
  const sorted = [...alertStatusStats.value].sort((a, b) => ORDER.indexOf(a.alert_status) - ORDER.indexOf(b.alert_status))
  let cumDeg = 0
  for (const item of sorted) {
    const sweep = (item.count / total) * 360
    const seg = {
      status: item.alert_status,
      count: item.count,
      color: PIE_COLORS[item.alert_status] || '#666',
      path: arcPath(cumDeg, cumDeg + sweep, PIE_R, PIE_CX, PIE_CY),
      percent: Math.round((item.count / total) * 100)
    }
    segs.push(seg)
    cumDeg += sweep
  }
  return segs
})

const TYPE_COLORS = ['#00e5ff', '#448aff', '#ff9800', '#4caf50', '#e040fb', '#ff5252']
const typeBars = computed(() => {
  const items = alertTypeStats.value || []
  const max = items.reduce((m, i) => Math.max(m, i.count), 0)
  return items.map((item, idx) => ({
    type: item.alert_type,
    count: item.count,
    pct: max > 0 ? (item.count / max) * 100 : 0,
    color: TYPE_COLORS[idx % TYPE_COLORS.length]
  }))
})

let alertPage = { total: 0, list: [] }
const ALERT_PAGE_SIZE = 3
let alertPageIdx = 0
let carouselTimer = null
let carouselPaused = false
let visibilityHandler = null

async function loadAllData() {
  try {
    const [nodesRes, linksRes, panelsRes, treeRes, summaryRes, typeStatsRes, statusStatsRes, alertPageRes, thresholdsRes] = await Promise.all([
      apiFetch('/api/dashboard-monitor/graph-idc-status').then(r => r.json()),
      apiFetch('/api/dashboard-monitor/idc-link-status').then(r => r.json()),
      apiFetch('/api/dashboard-monitor/idc-status').then(r => r.json()),
      apiFetch('/api/dashboard-monitor/idc-nodes').then(r => r.json()),
      apiFetch('/api/dashboard-monitor/summary-status').then(r => r.json()),
      apiFetch('/api/dashboard-monitor/alert-type-stats').then(r => r.json()),
      apiFetch('/api/dashboard-monitor/alert-status-stats').then(r => r.json()),
      apiFetch('/api/dashboard-monitor/unresolved-alert-page?page=1&pageSize=4').then(r => r.json()),
      apiFetch('/api/dashboard-monitor/thresholds').then(r => r.json())
    ])
    Object.assign(DATA.nodes, nodesRes)
    DATA.links.length = 0
    linksRes.forEach(l => DATA.links.push(l))
    Object.assign(DATA.panels, panelsRes)
    Object.assign(DATA.treeNodes, treeRes)
    summaryStatus = summaryRes
    alertTypeStats.value = Array.isArray(typeStatsRes) ? typeStatsRes : []
    alertStatusStats.value = Array.isArray(statusStatsRes) ? statusStatsRes : []
    alertPage = alertPageRes || { total: 0, list: [] }
    alertPageList.value = alertPage.list || []
    thresholds = thresholdsRes || {}
    updateBottomPanel()
  } catch (e) {
    console.error('加载 Monitor 数据失败:', e)
  }
}

function updateBottomPanel() {
  const order = ['BJ12','BJ13','SH22','SH23','SZ32','SZ33']

  // 各机房时延
  if (summaryStatus.idcLatency) {
    const listEl = document.getElementById('idc-latency-list')
    if (listEl) {
      const items = []
      for (const code of order) {
        const val = summaryStatus.idcLatency[code]
        const cls = val != null ? ' ' + latencyBetweenClass(val) : ''
        const dim = val == null ? ' dm-dim' : ''
        items.push(`<div class="dm-idc-cell"><span class="dm-idc-label">${code}</span><span class="dm-idc-val${cls}${dim}">${val != null ? val + 'ms' : '--'}</span></div>`)
      }
      listEl.innerHTML = items.join('')
    }
  }
  // 各机房丢包率
  if (summaryStatus.idcLoss) {
    const listEl = document.getElementById('idc-loss-list')
    if (listEl) {
      const items = []
      for (const code of order) {
        const val = summaryStatus.idcLoss[code]
        const cls = val != null ? ' ' + lossBetweenClass(val) : ''
        const dim = val == null ? ' dm-dim' : ''
        items.push(`<div class="dm-idc-cell"><span class="dm-idc-label">${code}</span><span class="dm-idc-val${cls}${dim}">${val != null ? val + '%' : '--'}</span></div>`)
      }
      listEl.innerHTML = items.join('')
    }
  }

  startAlertCarousel()
}

function startAlertCarousel() {
  if (carouselTimer) clearInterval(carouselTimer)
  const list = alertPageList.value
  if (list.length === 0) return

  carouselTimer = setInterval(() => {
    if (carouselPaused) return
    const totalPages = Math.ceil(list.length / ALERT_PAGE_SIZE)
    alertPageIdx = (alertPageIdx + 1) % totalPages
  }, 4000)
}

function pauseAlertCarousel() {
  carouselPaused = true
}

function resumeAlertCarousel() {
  carouselPaused = false
}

function prevAlertPage() {
  if (hasPrevPage.value) alertPageIdx--
}

function nextAlertPage() {
  if (hasNextPage.value) alertPageIdx++
}

function formatAlertTime(timeStr) {
  if (!timeStr) return '--/-- --:--'
  const d = new Date(timeStr)
  return d.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }) + ' ' +
    d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function stateToNum(s) {
  if (s === 'alarm') return 3
  if (s === 'warn') return 2
  return 1
}
function stateColor(s) {
  const n = typeof s === 'string' ? stateToNum(s) : s
  if (n === 3) return { stroke:'#ff1744', fill:'rgba(60,0,10,0.95)', glow:'rgba(255,23,68,0.6)' }
  if (n === 2) return { stroke:'#ffab00', fill:'rgba(50,25,0,0.95)', glow:'rgba(255,171,0,0.6)' }
  return             { stroke:'#00e5ff', fill:'rgba(0,40,80,0.95)', glow:'rgba(0,229,255,0.5)' }
}
function nodeColor(s) {
  const n = typeof s === 'string' ? stateToNum(s) : s
  if (n === 3) return '#ff1744'
  if (n === 2) return '#ffab00'
  return '#00ff88'
}
function linkColor(s) {
  const n = typeof s === 'string' ? stateToNum(s) : s
  if (n === 3) return '#ff1744'
  if (n === 2) return '#ffab00'
  return '#00ff88'
}
function stateClass(s) {
  const n = typeof s === 'string' ? stateToNum(s) : s
  return n === 3 ? 's-alarm' : n === 2 ? 's-warn' : 's-normal'
}

function latencyClass(valStr) {
  const v = parseFloat(valStr) || 0
  const t = thresholds.latency_levels_for_idc_inner || {}
  const warn = t.warn != null ? t.warn : 3
  const alarm = t.alarm != null ? t.alarm : 10
  if (v >= alarm) return 's-alarm'
  if (v >= warn) return 's-warn'
  return 's-normal'
}
function lossClass(valStr) {
  const v = parseFloat(valStr) || 0
  const t = thresholds.lost_levels_for_idc_inner || {}
  const warn = t.warn != null ? t.warn : 0.01
  const alarm = t.alarm != null ? t.alarm : 0.1
  if (v >= alarm) return 's-alarm'
  if (v >= warn) return 's-warn'
  return 's-normal'
}
function latencyBetweenClass(valStr) {
  const v = parseFloat(valStr) || 0
  const t = thresholds.latency_levels_for_idc_between || {}
  const warn = t.warn != null ? t.warn : 15
  const alarm = t.alarm != null ? t.alarm : 100
  if (v >= alarm) return 's-alarm'
  if (v >= warn) return 's-warn'
  return 's-normal'
}
function lossBetweenClass(valStr) {
  const v = parseFloat(valStr) || 0
  const t = thresholds.lost_levels_for_idc_between || {}
  const warn = t.warn != null ? t.warn : 5
  const alarm = t.alarm != null ? t.alarm : 10
  if (v >= alarm) return 's-alarm'
  if (v >= warn) return 's-warn'
  return 's-normal'
}

function flattenTreeNodes(nodeStates) {
  if (!nodeStates) return [1,1,1,1,1,1,1,1,1,1,1,1]
  return [].concat(
    nodeStates.core || [],
    nodeStates.spine1 || [],
    nodeStates.spine2 || [],
    nodeStates.leaf1 || [],
    nodeStates.leaf2 || []
  )
}

// ── Clock ──
const DAYS = ['周日','周一','周二','周三','周四','周五','周六']
let clockTimer = null
function tick() {
  const n = new Date()
  const hdrTime = document.getElementById('hdr-time')
  if (hdrTime) hdrTime.textContent = n.toLocaleTimeString('zh-CN',{hour12:false})
  const hdrDate = document.getElementById('hdr-date')
  if (hdrDate) hdrDate.textContent = `${n.getFullYear()}年${n.getMonth()+1}月${n.getDate()}日 ${DAYS[n.getDay()]}`
}
// ── Side Panel SVG ──
function buildPanelSVG(panelId, nodeStates) {
  const W = 150, H = 78, r = 5.5
  const cores = [{x:W*0.35,y:10},{x:W*0.65,y:10}]
  const spines = [{x:W*0.18,y:32},{x:W*0.40,y:32},{x:W*0.60,y:32},{x:W*0.82,y:32}]
  const leaves = [{x:W*0.08,y:60},{x:W*0.24,y:60},{x:W*0.40,y:60},{x:W*0.60,y:60},{x:W*0.76,y:60},{x:W*0.92,y:60}]
  const statusColor = (s) => s === 3 ? '#ff4560' : s === 2 ? '#ffc233' : '#00ff88'
  const links = []
  for (let c=0;c<2;c++) for (let s=0;s<4;s++) links.push([cores[c],spines[s]])
  links.push([spines[0],leaves[0]],[spines[0],leaves[1]],[spines[0],leaves[2]])
  links.push([spines[1],leaves[0]],[spines[1],leaves[1]],[spines[1],leaves[2]])
  links.push([spines[2],leaves[3]],[spines[2],leaves[4]],[spines[2],leaves[5]])
  links.push([spines[3],leaves[3]],[spines[3],leaves[4]],[spines[3],leaves[5]])
  function nodeCircle(pos, color) {
    const c = color
    return `<circle cx="${pos.x}" cy="${pos.y}" r="${r}" fill="#0c1a2e" stroke="${c}" stroke-width="1.3"/><circle cx="${pos.x}" cy="${pos.y}" r="${r*0.35}" fill="${c}"/>`
  }
  let svg = `<svg width="${W}" height="${H}" viewBox="0 0 ${W} ${H}" xmlns="http://www.w3.org/2000/svg">`
  links.forEach(([a,b])=>{svg+=`<line x1="${a.x}" y1="${a.y}" x2="${b.x}" y2="${b.y}" stroke="#00e5ff" stroke-width="1" stroke-opacity="0.35"/>`})
  cores.forEach((n,i)=>svg+=nodeCircle(n, statusColor(nodeStates.core[i])))
  spines.forEach((n,i)=>svg+=nodeCircle(n, statusColor((i<2 ? nodeStates.spine1 : nodeStates.spine2)[i%2])))
  leaves.forEach((n,i)=>svg+=nodeCircle(n, statusColor((i<3 ? nodeStates.leaf1 : nodeStates.leaf2)[i%3])))
  svg += '</svg>'
  return svg
}
function buildPanel(id, panelData, treeStates) {
  const s = panelData.state
  const sc = stateClass(s)
  const disabled = panelData.enable === false
  const div = document.createElement('div')
  div.className = `dm-np ${sc}` + (disabled ? ' dm-np-disabled' : '')
  div.innerHTML = '<div class="dm-ctl"></div><div class="dm-cbr"></div>'
  const treeDiv = document.createElement('div')
  treeDiv.className = 'dm-np-tree'
  treeDiv.innerHTML = `
    <div class="dm-np-id ${stateClass(s)}">${id}<span class="dm-np-id-sub"> 机房内网络</span></div>
    ${buildPanelSVG(id, treeStates)}`
  div.appendChild(treeDiv)
  const dataDiv = document.createElement('div')
  dataDiv.className = 'dm-np-data'
  dataDiv.innerHTML = `
    <div class="dm-pm"><div class="dm-pml">时延</div>
      <div class="dm-pmr"><span class="dm-pmk">平均:</span><span class="dm-pmv ${latencyClass(panelData.latency.avg)}">${panelData.latency.avg}ms</span></div>
      <div class="dm-pmr"><span class="dm-pmk">99分位:</span><span class="dm-pmv ${latencyClass(panelData.latency.p99)}">${panelData.latency.p99}ms</span></div>
      <div class="dm-pmr"><span class="dm-pmk">最大:</span><span class="dm-pmv ${latencyClass(panelData.latency.max)}">${panelData.latency.max}ms</span></div>
    </div>
    <div class="dm-pm"><div class="dm-pml">丢包</div>
      <div class="dm-pmr"><span class="dm-pmk">丢包率:</span><span class="dm-pmv ${lossClass(panelData["package-lost"].max)}">${panelData["package-lost"].max}%</span></div>
    </div>`
  div.appendChild(dataDiv)
  if (disabled) {
    const badge = document.createElement('div')
    badge.className = 'dm-np-disabled-badge'
    badge.textContent = '接入中'
    div.appendChild(badge)
  }
  return div
}

// ── Topo Canvas ──
const nodePositions = {
  'BJ12': { nx: 0.3843, ny: 0.1741 },
  'BJ13': { nx: 0.6016, ny: 0.1764 },
  'SH22': { nx: 0.1190, ny: 0.8261 },
  'SH23': { nx: 0.4082, ny: 0.6357 },
  'SZ32': { nx: 0.6992, ny: 0.8387 },
  'SZ33': { nx: 0.8569, ny: 0.6393 },
}
let W=0, H=0, af=0
let packets = []
let coordsVisible = false
let draggedNode = null

function gp(id) {
  const pos = nodePositions[id]
  return { x: pos.nx * W, y: pos.ny * H }
}
function lerp(a,b,t){return a+(b-a)*t}

function drawArrow(ctx, x1,y1,x2,y2,col,alpha) {
  const dx=x2-x1,dy=y2-y1,len=Math.hypot(dx,dy)
  if(len<1)return
  const ux=dx/len,uy=dy/len,rn=38
  const sx=x1+ux*rn,sy=y1+uy*rn,ex=x2-ux*rn,ey=y2-uy*rn
  if(Math.hypot(ex-sx,ey-sy)<1)return
  const g=ctx.createLinearGradient(sx,sy,ex,ey)
  g.addColorStop(0,col+'18');g.addColorStop(0.45,col+'99');g.addColorStop(1,col+'33')
  ctx.beginPath();ctx.moveTo(sx,sy);ctx.lineTo(ex,ey)
  ctx.strokeStyle=g;ctx.lineWidth=1.8;ctx.globalAlpha=alpha;ctx.stroke();ctx.globalAlpha=1
  const hl=9,ang=Math.atan2(ey-sy,ex-sx)
  ctx.beginPath()
  ctx.moveTo(ex,ey)
  ctx.lineTo(ex-hl*Math.cos(ang-0.38),ey-hl*Math.sin(ang-0.38))
  ctx.lineTo(ex-hl*Math.cos(ang+0.38),ey-hl*Math.sin(ang+0.38))
  ctx.closePath()
  ctx.fillStyle=col;ctx.globalAlpha=alpha*0.85;ctx.fill();ctx.globalAlpha=1
}

function drawNode(ctx, id) {
  // 从 DATA.nodes 读取机房状态，查询失败默认为 1（正常）
  const state = (DATA.nodes[id] && DATA.nodes[id].state) || 1
  const c = stateColor(state)
  const pos = gp(id)
  const R = 34
  const gr=ctx.createRadialGradient(pos.x,pos.y,R*0.4,pos.x,pos.y,R*2.4)
  gr.addColorStop(0,c.stroke+'24');gr.addColorStop(0.55,c.stroke+'0a');gr.addColorStop(1,'transparent')
  ctx.beginPath();ctx.arc(pos.x,pos.y,R*2.4,0,Math.PI*2);ctx.fillStyle=gr;ctx.fill()
  const pulse=0.5+0.5*Math.sin(af*0.045+(pos.x*0.01))
  ctx.beginPath();ctx.arc(pos.x,pos.y,R+5+pulse*5,0,Math.PI*2)
  ctx.strokeStyle=c.stroke+'2a';ctx.lineWidth=1.5;ctx.stroke()
  ctx.beginPath();ctx.arc(pos.x,pos.y,R+1.5,0,Math.PI*2)
  ctx.strokeStyle=c.stroke+'44';ctx.lineWidth=1;ctx.stroke()
  const fg=ctx.createRadialGradient(pos.x-R*0.22,pos.y-R*0.22,0,pos.x,pos.y,R)
  fg.addColorStop(0,c.fill);fg.addColorStop(1,c.fill.replace('0.95','0.98'))
  ctx.beginPath();ctx.arc(pos.x,pos.y,R,0,Math.PI*2)
  ctx.fillStyle=fg;ctx.fill()
  ctx.strokeStyle=c.stroke;ctx.lineWidth=2;ctx.stroke()
  ctx.fillStyle=c.stroke
  ctx.font="bold 12px 'Orbitron',monospace"
  ctx.textAlign='center';ctx.textBaseline='middle'
  ctx.shadowColor=c.stroke;ctx.shadowBlur=14
  ctx.fillText(id,pos.x,pos.y)
  ctx.shadowBlur=0
}

function seedPackets() {
  DATA.links.forEach(([from,to,state])=>{
    const count=1+Math.floor(Math.random()*2)
    for(let i=0;i<count;i++){packets.push({from,to,state,t:Math.random(),speed:0.0028+Math.random()*0.003})}
  })
}

let animFrameId = null
function frame(ctx) {
  if (W === 0 || H === 0) { animFrameId = requestAnimationFrame(()=>frame(ctx)); return }
  ctx.clearRect(0,0,W,H)
  DATA.links.forEach(([from,to,state])=>{
      if (!nodePositions[from] || !nodePositions[to]) return
      const fp=gp(from),tp=gp(to)
      drawArrow(ctx,fp.x,fp.y,tp.x,tp.y,linkColor(state),0.6)
    })
  packets=packets.filter(p=>p.t<1)
  packets.forEach(p=>{
    p.t+=p.speed
    const fp=gp(p.from),tp=gp(p.to)
    const dx=tp.x-fp.x,dy=tp.y-fp.y,len=Math.hypot(dx,dy)
    if(len<1)return
    const ux=dx/len,uy=dy/len,rn=40
    const sx=fp.x+ux*rn,sy=fp.y+uy*rn,ex=tp.x-ux*rn,ey=tp.y-uy*rn
    const px=lerp(sx,ex,p.t),py=lerp(sy,ey,p.t)
    const col=linkColor(p.state)
    const gr=ctx.createRadialGradient(px,py,0,px,py,6)
    gr.addColorStop(0,col+'ff');gr.addColorStop(0.4,col+'88');gr.addColorStop(1,'transparent')
    ctx.beginPath();ctx.arc(px,py,6,0,Math.PI*2);ctx.fillStyle=gr;ctx.fill()
    ctx.beginPath();ctx.arc(px,py,2.2,0,Math.PI*2);ctx.fillStyle=col;ctx.fill()
  })
  Object.keys(nodePositions).forEach(id=>drawNode(ctx,id))
  af++
  animFrameId = requestAnimationFrame(()=>frame(ctx))
}

function updateCoordDisplay() {
  const coordEl = document.getElementById('coords')
  if (!coordEl) return
  let html = '<div style="font-family:Orbitron,monospace;font-size:10px;color:#00e5ff;margin-bottom:8px;letter-spacing:2px">NODE COORDINATES</div>'
  html += '<pre style="font-family:\'Share Tech Mono\',monospace;font-size:11px;color:#b0d8e8;line-height:1.6;margin:0">const nodePositions = {\n'
  for (const id of ['BJ12','BJ13','SH22','SH23','SZ32','SZ33']) {
    const pos = nodePositions[id]
    html += `  '${id}': { nx: ${pos.nx.toFixed(4)}, ny: ${pos.ny.toFixed(4)} },\n`
  }
  html += '};\n</pre>'
  coordEl.innerHTML = html
}

function toggleCoords() {
  const coordEl = document.getElementById('coords')
  if (!coordEl) return
  coordsVisible = !coordsVisible
  if (coordsVisible) { coordEl.classList.add('show'); updateCoordDisplay() }
  else { coordEl.classList.remove('show') }
}

function getNodeAtPosition(x,y) {
  for (const id of Object.keys(nodePositions)) {
    const pos = gp(id)
    if (Math.hypot(x-pos.x,y-pos.y)<34) return id
  }
  return null
}

function resize(canvas, wrap) {
  W = canvas.width = wrap.clientWidth
  H = canvas.height = wrap.clientHeight
}

let topoTimer = null, refreshTimer = null

// ── China Map ──
async function renderChinaMap() {
  try {
    const response = await fetch('/assets/map.json')
    const chinaData = await response.json()
    const svg = document.getElementById('china-map-svg')
    if (!svg) return
    const bounds = { minLon: 73, maxLon: 135, minLat: 18, maxLat: 54 }
    function project(lon, lat) {
      const x = ((lon - bounds.minLon) / (bounds.maxLon - bounds.minLon)) * 900 + 50
      const y = ((bounds.maxLat - lat) / (bounds.maxLat - bounds.minLat)) * 700 + 50
      return [Math.max(0,Math.min(1000,x)), Math.max(0,Math.min(800,y))]
    }
    let svgContent = ''
    ;[80,90,100,110,120,130].forEach(lon=>{
      const x = ((lon-bounds.minLon)/(bounds.maxLon-bounds.minLon))*900+50
      svgContent += `<line x1="${x}" y1="50" x2="${x}" y2="750" class="dm-grid-line"/><text x="${x}" y="770" class="dm-grid-label" text-anchor="middle">${lon}°E</text>`
    })
    ;[20,30,40,50].forEach(lat=>{
      const y = ((bounds.maxLat-lat)/(bounds.maxLat-bounds.minLat))*700+50
      svgContent += `<line x1="50" y1="${y}" x2="950" y2="${y}" class="dm-grid-line"/><text x="40" y="${y+4}" class="dm-grid-label">${lat}°N</text>`
    })
    if (chinaData.features && chinaData.features.length>0) {
      const feature = chinaData.features[0]
      if (feature.geometry && feature.geometry.coordinates) {
        const coords = feature.geometry.coordinates
        let pathIndex = 0
        coords.forEach(polygon=>{
          polygon.forEach((ring,ringIdx)=>{
            if (ringIdx===0 && ring.length>2) {
              const points = ring.filter(c=>c[1]>0).map(c=>{const [x,y]=project(c[0],c[1]);return `${x},${y}`}).join(' ')
              if (points.length>0) {
                const sw = pathIndex===0?3:1.5
                const so = pathIndex===0?0.9:0.5
                svgContent += `<polygon points="${points}" fill="none" stroke="rgba(0,229,255,${so})" stroke-width="${sw}"/>`
                pathIndex++
              }
            }
          })
        })
      }
    }
    const cities = [
      {name:'北京',lon:116.4,lat:39.9,color:'#00e5ff'},
      {name:'上海',lon:121.5,lat:31.2,color:'#00ff88'},
      {name:'深圳',lon:114.1,lat:22.5,color:'#ffab00'}
    ]
    cities.forEach((city,idx)=>{
      const [x,y]=project(city.lon,city.lat)
      svgContent+=`<g class="dm-city-marker"><circle class="dm-city-outer-ring" cx="${x}" cy="${y}" r="7" fill="none" stroke="${city.color}cc" stroke-width="2.5" style="animation-delay:${idx*0.6}s"/><circle class="dm-city-outer-ring" cx="${x}" cy="${y}" r="7" fill="none" stroke="${city.color}99" stroke-width="2" style="animation-delay:${idx*0.6+0.8}s"/><circle class="dm-city-inner-circle" cx="${x}" cy="${y}" r="6" fill="${city.color}e6" style="color:${city.color}"/><circle cx="${x}" cy="${y}" r="3" fill="${city.color}"/><text x="${x}" y="${y+28}" class="dm-city-label" text-anchor="middle" style="fill:${city.color}">${city.name}</text></g>`
    })
    const [bjx,bjy]=project(116.4,39.9)
    const [shx,shy]=project(121.5,31.2)
    const [szx,szy]=project(114.1,22.5)
    svgContent+=`<path class="dm-flow-line" d="M${bjx},${bjy} Q${(bjx+shx)/2+50},${(bjy+shy)/2-30} ${shx},${shy}" fill="none" stroke="rgba(0,229,255,0.6)" stroke-width="2.5"/><path class="dm-flow-line" d="M${shx},${shy} Q${(shx+szx)/2-30},${(shy+szy)/2+40} ${szx},${szy}" fill="none" stroke="rgba(0,255,136,0.6)" stroke-width="2.5" style="animation-delay:1.2s"/><path class="dm-flow-line" d="M${szx},${szy} Q${(szx+bjx)/2-50},${(szy+bjy)/2} ${bjx},${bjy}" fill="none" stroke="rgba(0,229,255,0.6)" stroke-width="2.5" style="animation-delay:2.4s"/>`
    svg.innerHTML = svgContent
  } catch(err) { console.error('加载 assets/map.json 失败:', err) }
}

onMounted(async () => {
  tick(); clockTimer = setInterval(tick, 1000)

  // Load all data from backend APIs
  await loadAllData()

  // Side panels
  ;['BJ12','BJ13','SZ32'].forEach(id=>{
    if (DATA.panels[id] && DATA.treeNodes[id]) {
      document.getElementById('sideL').appendChild(buildPanel(id, DATA.panels[id], DATA.treeNodes[id]))
    }
  })
  ;['SH22','SH23','SZ33'].forEach(id=>{
    if (DATA.panels[id] && DATA.treeNodes[id]) {
      document.getElementById('sideR').appendChild(buildPanel(id, DATA.panels[id], DATA.treeNodes[id]))
    }
  })

  // Topo canvas
  const canvas = document.getElementById('tc')
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  const wrap = canvas.parentElement

  seedPackets()
  topoTimer = setInterval(()=>{
    DATA.links.forEach(([from,to,state])=>{
      if(Math.random()<0.25) packets.push({from,to,state,t:0,speed:0.0028+Math.random()*0.003})
    })
    if(packets.length>80) packets=packets.slice(-50)
  },1000)

  resize(canvas, wrap)
  window.addEventListener('resize', ()=>resize(canvas, wrap))
  animFrameId = requestAnimationFrame(()=>frame(ctx))

  // Drag
  canvas.addEventListener('pointerdown',(e)=>{
    const x=e.offsetX,y=e.offsetY
    draggedNode = getNodeAtPosition(x,y)
  })
  canvas.addEventListener('pointermove',(e)=>{
    if(!draggedNode)return
    const rect = canvas.getBoundingClientRect()
    const x=Math.max(24,Math.min(rect.width-24,e.offsetX))
    const y=Math.max(24,Math.min(rect.height-24,e.offsetY))
    nodePositions[draggedNode]={nx:x/rect.width,ny:y/rect.height}
    updateCoordDisplay()
  })
  canvas.addEventListener('pointerup',()=>draggedNode=null)
  canvas.addEventListener('pointercancel',()=>draggedNode=null)

  document.addEventListener('keydown',(e)=>{
    if(e.ctrlKey && e.key.toLowerCase()==='d'){ e.preventDefault(); toggleCoords() }
  })

  // Periodic refresh of overview data
  refreshTimer = setInterval(() => {
    Promise.all([
      apiFetch('/api/dashboard-monitor/summary-status').then(r => r.json()).catch(() => {})
    ]).then(([ss]) => {
      if (ss) summaryStatus = ss
      updateBottomPanel()
    }).catch(() => {})
  }, 10000)

  // 页签可见性检测：不可见时暂停轮询，节省 HTTP 请求
  visibilityHandler = () => {
    if (document.hidden) {
      if (refreshTimer) { clearInterval(refreshTimer); refreshTimer = null }
      if (carouselTimer) { clearInterval(carouselTimer); carouselTimer = null }
    } else {
      // 恢复时立即刷新一次，再开启轮询
      apiFetch('/api/dashboard-monitor/summary-status').then(r => r.json()).then(ss => {
        if (ss) summaryStatus = ss
        updateBottomPanel()
      }).catch(() => {})
      refreshTimer = setInterval(() => {
        apiFetch('/api/dashboard-monitor/summary-status').then(r => r.json()).then(ss => {
          if (ss) summaryStatus = ss
          updateBottomPanel()
        }).catch(() => {})
      }, 10000)
      startAlertCarousel()
    }
  }
  document.addEventListener('visibilitychange', visibilityHandler)

  // China map
  renderChinaMap()
})

onBeforeUnmount(() => {
  if (animFrameId) cancelAnimationFrame(animFrameId)
  if (clockTimer) clearInterval(clockTimer)
  if (topoTimer) clearInterval(topoTimer)
  if (refreshTimer) clearInterval(refreshTimer)
  if (carouselTimer) clearInterval(carouselTimer)
  if (visibilityHandler) document.removeEventListener('visibilitychange', visibilityHandler)
  window.removeEventListener('resize', ()=>resize)
})
</script>

<style>
/* ══ DASHBOARD MONITOR PAGE STYLES ══ */
.dm-root {
  flex: 1;
  position: relative;
  overflow: hidden;
  background: #03080f;
  color: #b0d8e8;
  font-family: 'Share Tech Mono', monospace;
  font-size: 13px;
}
.dm-bg { position:absolute; inset:0; z-index:0 }
.dm-bg::before {
  content:''; position:absolute; inset:0;
  background: radial-gradient(ellipse 65% 55% at 50% 42%,rgba(0,55,110,0.32) 0%,transparent 70%),
              linear-gradient(180deg,#030a16 0%,#040c1e 55%,#030810 100%);
}
.dm-bg-grid {
  position:absolute; inset:0;
  background-image: linear-gradient(rgba(0,229,255,0.035) 1px,transparent 1px),
                    linear-gradient(90deg,rgba(0,229,255,0.035) 1px,transparent 1px);
  background-size: 48px 48px;
}
.dm-bg-scan {
  position:absolute; inset:0; pointer-events:none;
  background: repeating-linear-gradient(0deg,transparent,transparent 3px,rgba(0,0,0,0.07) 3px,rgba(0,0,0,0.07) 4px);
}
.dm-bg-beam {
  position:absolute; left:0; right:0; height:2px; top:-2px;
  background: linear-gradient(90deg,transparent,rgba(0,229,255,0.35),transparent);
  animation: dm-beam 9s linear infinite;
}
@keyframes dm-beam { from{top:-2px} to{top:100%} }

.dm-china-map {
  position:absolute; inset:0; pointer-events:none; opacity:0.18;
  display:flex; align-items:center; justify-content:center;
}
.dm-china-map svg {
  width:85%; height:auto; max-height:85%;
  filter:drop-shadow(0 0 35px rgba(0,229,255,0.45));
}

/* City markers */
.dm-city-marker circle { transform-origin:center center }
.dm-city-outer-ring { animation: dm-cityRipple 2s ease-out infinite }
.dm-city-inner-circle { animation: dm-cityGlow 2s ease-in-out infinite }
@keyframes dm-cityRipple { 0%{opacity:0.85;r:6} 70%{opacity:0;r:20} 100%{opacity:0;r:20} }
@keyframes dm-cityGlow { 0%,100%{opacity:0.7;filter:drop-shadow(0 0 5px currentColor)} 50%{opacity:1;filter:drop-shadow(0 0 15px currentColor)} }
.dm-city-label {
  font-family: 'Share Tech Mono', monospace; font-size:12px; font-weight:bold;
  fill:#00ffff; opacity:0.95;
}
.dm-flow-line { stroke-dasharray:10,5; animation: dm-flowMove 3.5s linear infinite }
@keyframes dm-flowMove { from{stroke-dashoffset:0} to{stroke-dashoffset:-30} }
.dm-grid-line { stroke:rgba(0,229,255,0.15); stroke-width:1; stroke-dasharray:5,3 }
.dm-grid-label { font-family:'Share Tech Mono',monospace; font-size:8px; fill:rgba(0,229,255,0.4) }

/* Layout */
.dm-layout {
  position:relative; z-index:1;
  width:100%; height:100%;
  display:grid;
  grid-template-columns: 340px 1fr 340px;
  grid-template-rows: 80px 1fr 140px;
  gap:7px;
  padding:8px;
}
/* ══ HEADER ══ */
.dm-hdr {
  grid-column:1/-1; grid-row:1;
  display:flex; align-items:flex-start; justify-content:center; position:relative;
  padding:24px 0 8px 0;
}
.dm-hdr-title {
  font-family:'Orbitron',monospace; font-size:clamp(16px,1.8vw,24px); font-weight:900;
  color:#00e5ff; letter-spacing:6px; text-transform:uppercase;
  text-shadow:0 0 18px rgba(0,229,255,0.55),0 0 50px rgba(0,229,255,0.18);
  animation:dm-glitch 7s ease-in-out infinite;
  position:relative; padding:15px 20px 10px 20px;
}
.dm-hdr-title::before,.dm-hdr-title::after {
  content:''; position:absolute; top:50%; transform:translateY(-50%);
  height:1px; width:70px; background:linear-gradient(90deg,transparent,#00e5ff);
}
.dm-hdr-title::before { right:calc(100% + 14px) }
.dm-hdr-title::after { left:calc(100% + 14px); transform:translateY(-50%) scaleX(-1) }
.dm-hdr-cr { position:absolute; width:12px; height:12px; border:1px solid #00e5ff; opacity:0.7 }
.dm-hdr-cr-tl { top:-6px; left:-6px; border-right:none; border-bottom:none }
.dm-hdr-cr-tr { top:-6px; right:-6px; border-left:none; border-bottom:none }
.dm-hdr-cr-bl { bottom:-6px; left:-6px; border-right:none; border-top:none }
.dm-hdr-cr-br { bottom:-6px; right:-6px; border-left:none; border-top:none }
.dm-hdr-glow {
  position:absolute; top:50%; left:50%; transform:translate(-50%,-50%);
  width:200%; height:200%;
  background:radial-gradient(circle,rgba(0,229,255,0.08) 0%,transparent 70%);
  pointer-events:none; z-index:-1;
}
@keyframes dm-glitch {
  0%,93%,100%{text-shadow:0 0 18px rgba(0,229,255,0.55),0 0 50px rgba(0,229,255,0.18)}
  94%{text-shadow:-2px 0 #ff1744,2px 0 #00e5ff,0 0 18px rgba(0,229,255,0.55)}
  95%{text-shadow:2px 0 #ff1744,-2px 0 #00e5ff,0 0 18px rgba(0,229,255,0.55)}
  96%{text-shadow:0 0 18px rgba(0,229,255,0.55),0 0 50px rgba(0,229,255,0.18)}
}
.dm-hdr-live {
  position:absolute; left:0; display:flex; align-items:center; gap:7px;
  font-size:10px; color:#3a6070; letter-spacing:2px;
}
.dm-ldot {
  width:7px; height:7px; border-radius:50%; background:#00ff88;
  box-shadow:0 0 7px #00ff88; animation:dm-lp 1.6s ease-in-out infinite;
}
@keyframes dm-lp { 0%,100%{opacity:1;transform:scale(1)} 50%{opacity:0.35;transform:scale(0.6)} }
.dm-hdr-clk { position:absolute; right:0; text-align:right }
.dm-hdr-time { font-family:'Orbitron',monospace; font-size:13px; color:#00b8d4; letter-spacing:2px }
.dm-hdr-date { font-size:10px; color:#3a6070; letter-spacing:1px; margin-top:2px }

/* ══ SIDE PANELS ══ */
.dm-side-l { grid-column:1; grid-row:2; display:flex; flex-direction:column; gap:7px; justify-content:flex-end }
.dm-side-r { grid-column:3; grid-row:2; display:flex; flex-direction:column; gap:7px; justify-content:flex-end }
.dm-center { grid-column:2; grid-row:2; position:relative; min-height:0; display:flex; flex-direction:column }
#tc {
  width:100%; height:100%; min-height:0; display:block; cursor:grab; flex:1;
}
#tc:active { cursor:grabbing }

.dm-coords-panel {
  position:absolute; top:8px; right:8px;
  background: rgba(6,18,38,0.88); border: 1px solid rgba(0,229,255,0.22);
  border-radius:5px; padding:10px 12px; z-index:10; display:none;
}
.dm-coords-panel.show { display:block }

.dm-bottom {
  grid-column: 1/-1; grid-row: 3;
  display:grid; grid-template-columns: 1fr 1fr 1fr; gap:7px; min-height:0;
}

/* Info Panels */
.dm-ip {
  background: rgba(6,18,38,0.88); border:1px solid rgba(0,229,255,0.22);
  border-radius:5px; padding:9px 14px; position:relative; overflow:hidden;
}
.dm-ip::before {
  content:''; position:absolute; top:0; left:0; right:0; height:1px;
  background:linear-gradient(90deg,transparent,#00e5ff,transparent); opacity:.35;
}
.dm-ip-title {
  font-family:'Orbitron',monospace; font-size:9px; letter-spacing:3px;
  color:#00b8d4; text-transform:uppercase; margin-bottom:6px;
  border-bottom:1px solid rgba(0,229,255,0.08); padding-bottom:4px;
}
.dm-alert-nav { display:flex; justify-content:center; gap:4px; margin-top:4px; width:100% }
.dm-alert-nav-btn {
  background:none; border:1px solid rgba(0,229,255,0.25); color:#00b8d4;
  font-size:12px; line-height:1; width:16px; height:16px; padding:0;
  border-radius:2px; cursor:pointer; display:flex; align-items:center; justify-content:center;
}
.dm-alert-nav-btn:disabled {
  border-color:rgba(0,229,255,0.08); color:rgba(0,229,255,0.15); cursor:default;
}
.dm-ig { display:grid; grid-template-columns:1fr; gap:3px 0 }
.dm-idc-mgrid { display:grid; grid-template-columns:1fr 1fr 1fr; column-gap:42px; row-gap:2px; padding-left:2ch }
.dm-idc-cell { display:flex; justify-content:space-between; gap:2ch; font-size:9px; line-height:1.4 }
.dm-idc-label { color:#4a9a9a; white-space:nowrap }
.dm-idc-val {
  font-family:'Orbitron',monospace; font-size:8px; letter-spacing:1px; white-space:nowrap;
}
.dm-idc-val.s-normal { color:#00ff88 }
.dm-idc-val.s-warn { color:#ffab00 }
.dm-idc-val.s-alarm { color:#ff1744 }
.dm-idc-val.dm-dim { color:#2a4050 }
.dm-il-head {
  font-size:9px; color:#00ffff; letter-spacing:2px; text-transform:uppercase;
  font-weight:600; margin-bottom:2px;
}
.dm-ir { display:flex; justify-content:space-between; font-size:10px }
.dm-il { color:#3a6070 }
.dm-iv {
  font-family:'Orbitron',monospace; font-size:9px; letter-spacing:1px; color:#e0f4ff;
}
.dm-iv.ok { color:#00ff88 }
.dm-iv.w { color:#ffab00 }
.dm-iv.e { color:#ff1744 }
.dm-alg { display:grid; grid-template-columns:1fr 1fr; gap:3px 10px }
/* 告警统计分列 */
.dm-alert-stats-split {
  display: flex; gap: 24px;
}
.dm-stats-col {
  flex: 1; min-width: 0;
}
.dm-stats-col-title {
  font-size: 9px; color: #00ffff; letter-spacing: 2px; text-transform: uppercase;
  font-weight: 600; margin-bottom: 4px;
  border-bottom: 1px solid rgba(0, 229, 255, 0.1); padding-bottom: 3px;
}
.dm-stats-items {
  display: flex; flex-direction: column; gap: 2px;
}
.dm-bars-wrap {
  display:flex; flex-direction:column; gap:4px;
}
.dm-bar-row {
  display:flex; align-items:center; gap:4px;
}
.dm-bar-label {
  font-size:9px; color:#90a4ae; width:48px; flex-shrink:0; text-align:right;
}
.dm-bar-track {
  flex:1; height:8px; background:rgba(255,255,255,0.06); border-radius:4px; overflow:hidden;
}
.dm-bar-fill {
  height:100%; border-radius:4px; transition: width 0.5s;
}
.dm-bar-count {
  font-size:9px; color:#b0d8e8; font-family:'Orbitron',monospace; width:20px; flex-shrink:0;
}
.dm-pie-wrap {
  display:flex; align-items:center; gap:8px;
}
.dm-pie-svg {
  width:56px; height:56px; flex-shrink:0;
}
.dm-pie-legend {
  display:flex; flex-direction:column; gap:3px;
}
.dm-pl-item {
  display:flex; align-items:center; gap:4px;
}
.dm-pl-dot {
  width:7px; height:7px; border-radius:50%; flex-shrink:0;
}
.dm-pl-text {
  font-size:10px; color:#8098a8;
}
.dm-pl-num {
  font-size:10px; color:#ffab00; font-family:'Orbitron',monospace; margin-left:6px;
}
.dm-alert-page { display:flex; flex-direction:column; gap:5px; margin-top:2px }
.dm-amrow { display:flex; align-items:center; gap:7px; font-size:10px; min-height:18px }
.dm-atime { font-family:'Orbitron',monospace; font-size:9px; color:#00b8d4; white-space:nowrap; min-width:60px }
.dm-atxt { color:#b0d8e8; line-height:1.45; font-size:10px; flex:1 }
.dm-atxt strong { color:#ffab00 }
.dm-astatus {
  display:inline-block; font-size:8px; padding:1px 6px; border-radius:9px; white-space:nowrap;
  font-family:'Orbitron',monospace; letter-spacing:1px;
}
.dm-astatus.active { background:rgba(255,23,68,0.14); color:#ff1744; border:1px solid rgba(255,23,68,0.3) }
.dm-astatus.recovered { background:rgba(0,255,136,0.14); color:#00ff88; border:1px solid rgba(0,255,136,0.3) }
.dm-abadge {
  display:inline-block; font-size:8px; padding:1px 6px; border-radius:9px;
  font-family:'Orbitron',monospace; letter-spacing:1px; margin-left:3px;
  background:rgba(255,171,0,0.14); color:#ffab00; border:1px solid rgba(255,171,0,0.3);
}

/* Node Panels */
.dm-np {
  flex:0 1 auto; max-height:130px;
  background:rgba(6,18,38,0.88); border:1px solid rgba(0,229,255,0.22);
  border-radius:5px; display:flex; gap:0; overflow:hidden; position:relative;
  transition:border-color .25s,box-shadow .25s; cursor:default;
}
.dm-np:hover { border-color:rgba(0,229,255,0.45); box-shadow:0 0 18px rgba(0,229,255,0.1) }
.dm-np.s-warn { border-color:rgba(0,229,255,0.22); box-shadow:0 0 14px rgba(0,229,255,0.1) }
.dm-np.s-alarm { border-color:rgba(0,229,255,0.22); box-shadow:0 0 14px rgba(0,229,255,0.1) }
@keyframes dm-alarm-pulse { 0%,100%{box-shadow:0 0 14px rgba(255,23,68,0.18)} 50%{box-shadow:0 0 28px rgba(255,23,68,0.38)} }
.dm-np::before {
  content:''; position:absolute; top:0; left:0; right:0; height:1px;
  background:linear-gradient(90deg,transparent,#00e5ff,transparent); opacity:.4;
}
.dm-np.s-warn::before { background:linear-gradient(90deg,transparent,#00e5ff,transparent) }
.dm-np.s-alarm::before { background:linear-gradient(90deg,transparent,#00e5ff,transparent) }
.dm-ctl, .dm-cbr { position:absolute; width:9px; height:9px }
.dm-ctl { top:3px; left:3px; border-top:1px solid #00e5ff; border-left:1px solid #00e5ff }
.dm-cbr { bottom:3px; right:3px; border-bottom:1px solid #00e5ff; border-right:1px solid #00e5ff }
.dm-np.s-warn .dm-ctl, .dm-np.s-warn .dm-cbr { border-color:#00e5ff }
.dm-np.s-alarm .dm-ctl, .dm-np.s-alarm .dm-cbr { border-color:#00e5ff }
.dm-np-tree { flex:1.4; flex-shrink:0; padding:8px 6px 8px 8px; display:flex; flex-direction:column; align-items:center; justify-content:center }
.dm-np-tree svg { display:block }
.dm-np-data { flex:1; padding:8px 8px 8px 4px; display:flex; flex-direction:column; justify-content:center; gap:2px }
.dm-np-id { font-family:'Orbitron',monospace; font-size:11px; font-weight:700; letter-spacing:2px; margin-bottom:4px; text-align:center }
.dm-np-id-sub { font-family:inherit; font-size:9px; font-weight:400; letter-spacing:1px; color:#00b8d4 }
.dm-np-id.s-normal { color:#00e5ff; text-shadow:0 0 8px rgba(0,229,255,0.5) }
.dm-np-id.s-warn { color:#00e5ff; text-shadow:0 0 8px rgba(0,229,255,0.5) }
.dm-np-id.s-alarm { color:#00e5ff; text-shadow:0 0 8px rgba(0,229,255,0.5) }
.dm-pm { display:flex; flex-direction:column; gap:0; margin-bottom:2px }
.dm-pml { font-size:9px; color:#00ffff; letter-spacing:2px; text-transform:uppercase; font-weight:600 }
.dm-pmr { display:flex; justify-content:space-between; font-size:9px; line-height:1.4; gap:3px }
.dm-pmk { color:#3a6070; white-space:nowrap }
.dm-pmv { font-family:'Orbitron',monospace; font-size:8px; letter-spacing:1px; white-space:nowrap }
.dm-pmv.s-normal { color:#00ff88 }
.dm-pmv.s-warn { color:#ffab00 }
.dm-pmv.s-alarm { color:#ff1744 }

/* Disabled panel */
.dm-np-disabled {
  opacity: 0.55;
}
.dm-np-disabled:hover {
  border-color: rgba(0,229,255,0.22) !important;
  box-shadow: none !important;
  cursor: default;
}
.dm-np-disabled-badge {
  position: absolute; bottom: 4px; left: 6px; z-index: 2;
  font-size: 9px; color: rgba(176,216,232,0.4); letter-spacing: 1px;
}
</style>
