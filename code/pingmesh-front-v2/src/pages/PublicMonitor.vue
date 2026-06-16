<template>
  <div class="po-root">
    <canvas id="tc"></canvas>
  </div>
</template>

<script setup>
import { reactive, onMounted, onBeforeUnmount } from 'vue'

const DATA = reactive({
  nodes: {},
  links: [],
})

// ── 数据加载 ──
async function loadData() {
  try {
    const [nodesRes, linksRes] = await Promise.all([
      fetch('/api/public/graph-idc-status').then(r => r.json()),
      fetch('/api/public/idc-link-status').then(r => r.json())
    ])
    Object.assign(DATA.nodes, nodesRes)
    DATA.links.length = 0
    linksRes.forEach(l => DATA.links.push(l))
    seedPackets()
  } catch (e) {
    console.error('加载公共监控数据失败:', e)
  }
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
let W = 0, H = 0, af = 0
let packets = []
let draggedNode = null

function gp(id) { const pos = nodePositions[id]; return { x: pos.nx * W, y: pos.ny * H } }
function lerp(a, b, t) { return a + (b - a) * t }

function stateToNum(s) { if (s === 'alarm') return 3; if (s === 'warn') return 2; return 1 }
function stateColor(s) {
  const n = typeof s === 'string' ? stateToNum(s) : s
  if (n === 3) return { stroke: '#ff1744', fill: 'rgba(60,0,10,0.95)' }
  if (n === 2) return { stroke: '#ffab00', fill: 'rgba(50,25,0,0.95)' }
  return { stroke: '#00e5ff', fill: 'rgba(0,40,80,0.95)' }
}
function linkColor(s) {
  const n = typeof s === 'string' ? stateToNum(s) : s
  if (n === 3) return '#ff1744'
  if (n === 2) return '#ffab00'
  return '#00e5ff'
}

function drawLink(ctx, x1, y1, x2, y2, col) {
  const dx = x2 - x1, dy = y2 - y1, len = Math.hypot(dx, dy)
  if (len < 1) return
  const ux = dx / len, uy = dy / len, rn = 42
  const sx = x1 + ux * rn, sy = y1 + uy * rn, ex = x2 - ux * rn, ey = y2 - uy * rn

  // 连线主体
  ctx.beginPath()
  ctx.moveTo(sx, sy)
  ctx.lineTo(ex, ey)
  ctx.strokeStyle = col
  ctx.lineWidth = 1.5
  ctx.globalAlpha = 0.35
  ctx.stroke()

  // 发光外延
  ctx.beginPath()
  ctx.moveTo(sx, sy)
  ctx.lineTo(ex, ey)
  ctx.strokeStyle = col
  ctx.lineWidth = 4
  ctx.globalAlpha = 0.08
  ctx.stroke()
  ctx.globalAlpha = 1
}

function drawNode(ctx, id) {
  const state = (DATA.nodes[id] && DATA.nodes[id].state) || 1
  const c = stateColor(state)
  const pos = gp(id)
  const R = 34

  // 光环
  const gr = ctx.createRadialGradient(pos.x, pos.y, R * 0.4, pos.x, pos.y, R * 2.4)
  gr.addColorStop(0, c.stroke + '24')
  gr.addColorStop(0.55, c.stroke + '0a')
  gr.addColorStop(1, 'transparent')
  ctx.beginPath()
  ctx.arc(pos.x, pos.y, R * 2.4, 0, Math.PI * 2)
  ctx.fillStyle = gr
  ctx.fill()

  // 脉冲环
  const pulse = 0.5 + 0.5 * Math.sin(af * 0.045 + (pos.x * 0.01))
  ctx.beginPath()
  ctx.arc(pos.x, pos.y, R + 5 + pulse * 5, 0, Math.PI * 2)
  ctx.strokeStyle = c.stroke + '2a'
  ctx.lineWidth = 1.5
  ctx.stroke()

  ctx.beginPath()
  ctx.arc(pos.x, pos.y, R + 1.5, 0, Math.PI * 2)
  ctx.strokeStyle = c.stroke + '44'
  ctx.lineWidth = 1
  ctx.stroke()

  // 填充
  const fg = ctx.createRadialGradient(pos.x - R * 0.22, pos.y - R * 0.22, 0, pos.x, pos.y, R)
  fg.addColorStop(0, c.fill)
  fg.addColorStop(1, c.fill.replace('0.95', '0.98'))
  ctx.beginPath()
  ctx.arc(pos.x, pos.y, R, 0, Math.PI * 2)
  ctx.fillStyle = fg
  ctx.fill()
  ctx.strokeStyle = c.stroke
  ctx.lineWidth = 2
  ctx.stroke()

  // 文字
  ctx.fillStyle = c.stroke
  ctx.font = "bold 12px 'Orbitron', monospace"
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.shadowColor = c.stroke
  ctx.shadowBlur = 14
  ctx.fillText(id, pos.x, pos.y)
  ctx.shadowBlur = 0
}

function getNodeAtPosition(x, y) {
  for (const id of Object.keys(nodePositions)) {
    const pos = gp(id)
    if (Math.hypot(x - pos.x, y - pos.y) < 34) return id
  }
  return null
}

function seedPackets() {
  packets.length = 0
  DATA.links.forEach(([from, to, state]) => {
    const isBidi = state === 1 || state === 2 || state === 3
    if (!isBidi) return
    const count = 1 + Math.floor(Math.random() * 2)
    for (let i = 0; i < count; i++) {
      packets.push({ from, to, state, t: Math.random(), speed: 0.002 + Math.random() * 0.002 })
    }
  })
}

let animFrameId = null
function frame(ctx) {
  if (W === 0 || H === 0) { animFrameId = requestAnimationFrame(() => frame(ctx)); return }
  ctx.clearRect(0, 0, W, H)

  // 绘制连线
  DATA.links.forEach(([from, to, state]) => {
    if (!nodePositions[from] || !nodePositions[to]) return
    const fp = gp(from), tp = gp(to)
    drawLink(ctx, fp.x, fp.y, tp.x, tp.y, linkColor(state))
  })

  // 流动包
  packets = packets.filter(p => p.t < 1)
  packets.forEach(p => {
    p.t += p.speed
    const fp = gp(p.from), tp = gp(p.to)
    const dx = tp.x - fp.x, dy = tp.y - fp.y, len = Math.hypot(dx, dy)
    if (len < 1) return
    const ux = dx / len, uy = dy / len, rn = 42
    const sx = fp.x + ux * rn, sy = fp.y + uy * rn, ex = tp.x - ux * rn, ey = tp.y - uy * rn
    const px = lerp(sx, ex, p.t), py = lerp(sy, ey, p.t)
    const col = linkColor(p.state)
    const gr = ctx.createRadialGradient(px, py, 0, px, py, 5)
    gr.addColorStop(0, col + 'ff')
    gr.addColorStop(0.4, col + '88')
    gr.addColorStop(1, 'transparent')
    ctx.beginPath()
    ctx.arc(px, py, 5, 0, Math.PI * 2)
    ctx.fillStyle = gr
    ctx.fill()
    ctx.beginPath()
    ctx.arc(px, py, 2, 0, Math.PI * 2)
    ctx.fillStyle = col
    ctx.fill()
  })

  // 绘制节点
  Object.keys(nodePositions).forEach(id => drawNode(ctx, id))
  af++
  animFrameId = requestAnimationFrame(() => frame(ctx))
}

function resize(canvas, wrap) {
  W = canvas.width = wrap.clientWidth
  H = canvas.height = wrap.clientHeight
}

let topoTimer = null

onMounted(async () => {
  // 深蓝色背景
  document.body.style.background = '#050d1a'
  const appEl = document.getElementById('app')
  const mainEl = document.querySelector('.main-content')
  if (appEl) appEl.style.background = '#050d1a'
  if (mainEl) mainEl.style.background = '#050d1a'

  await loadData()

  const canvas = document.getElementById('tc')
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  const wrap = canvas.parentElement

  resize(canvas, wrap)
  window.addEventListener('resize', () => resize(canvas, wrap))
  animFrameId = requestAnimationFrame(() => frame(ctx))

  // 定时补充流动包
  topoTimer = setInterval(() => {
    DATA.links.forEach(([from, to, state]) => {
      if (Math.random() < 0.25) packets.push({ from, to, state, t: 0, speed: 0.002 + Math.random() * 0.002 })
    })
    if (packets.length > 80) packets = packets.slice(-50)
  }, 1000)

  // 节点拖拽
  canvas.addEventListener('pointerdown', (e) => {
    draggedNode = getNodeAtPosition(e.offsetX, e.offsetY)
  })
  canvas.addEventListener('pointermove', (e) => {
    if (!draggedNode) return
    const rect = canvas.getBoundingClientRect()
    const x = Math.max(24, Math.min(rect.width - 24, e.offsetX))
    const y = Math.max(24, Math.min(rect.height - 24, e.offsetY))
    nodePositions[draggedNode] = { nx: x / rect.width, ny: y / rect.height }
  })
  canvas.addEventListener('pointerup', () => { draggedNode = null })
  canvas.addEventListener('pointercancel', () => { draggedNode = null })
})

onBeforeUnmount(() => {
  if (animFrameId) cancelAnimationFrame(animFrameId)
  if (topoTimer) clearInterval(topoTimer)
  // 恢复背景
  document.body.style.background = ''
  const appEl = document.getElementById('app')
  const mainEl = document.querySelector('.main-content')
  if (appEl) appEl.style.background = ''
  if (mainEl) mainEl.style.background = ''
})
</script>

<style>
.po-root {
  width: 100vw;
  height: 100vh;
  background: transparent;
}

#tc {
  width: 100%;
  height: 100%;
  display: block;
}

/* 公开页面覆盖 body 背景为透明 */
body:has(#tc) {
  background: transparent !important;
}

/* 隐藏主内容区的背景 */
body:has(#tc) .main-content {
  background: transparent !important;
}
</style>
