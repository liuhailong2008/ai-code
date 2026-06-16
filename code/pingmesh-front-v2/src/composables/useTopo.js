import { ref, onMounted, onBeforeUnmount } from 'vue'
import { topoNodes, topoLinks, nodes } from './useData'

/**
 * 使用 Canvas 绘制网络拓扑图
 * 支持迷你拓扑和全屏拓扑两种模式
 */
export function useTopo(canvasRef, options = {}) {
  const {
    fullMode = false,
    headerH = 0,
    onTooltip = null,
  } = options

  const tooltipData = ref(null)
  const tooltipStyle = ref({ display: 'none' })
  const scale = ref(1)
  const offset = ref({ x: 0, y: 0 })

  let rafId = null
  let frameCount = 0
  let canvas = null
  let ctx = null
  let W = 0, H = 0
  let nodePositions = {}

  function calcNodePositions() {
    const map = {}
    topoNodes.forEach(n => {
      map[n.id] = { ...n, px: n.x * W, py: n.y * H }
    })
    return map
  }

  function findNode(mx, my) {
    for (const n of topoNodes) {
      const nd = nodePositions[n.id]
      if (!nd) continue
      const dist = Math.hypot(mx - nd.px, my - nd.py)
      if (dist < nd.r + 8) return { topoNode: n, pos: nd }
    }
    return null
  }

  function findRealNode(topoId) {
    const idMap = {
      internet: null, fw: 'FW-EDGE-01',
      cr1: 'CORE-RTR-01', cr2: 'CORE-RTR-02',
      sw1: 'CORE-SW-01',  sw2: 'CORE-SW-02', sw3: null,
      e1: 'EDGE-BJ-01', e2: 'EDGE-SH-01', e3: 'EDGE-GZ-01',
      e4: 'EDGE-CD-01', e5: 'EDGE-WH-01',
    }
    const realId = idMap[topoId]
    if (!realId) return null
    return nodes.find(d => d.id === realId)
  }

  function draw() {
    if (!canvas || !ctx) return

    if (fullMode) {
      ctx.fillStyle = '#080f16'
      ctx.fillRect(0, 0, W, H)

      // Grid
      ctx.strokeStyle = 'rgba(0,245,196,0.035)'
      ctx.lineWidth = 1
      for (let x = 0; x < W; x += 50) { ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, H); ctx.stroke() }
      for (let y = 0; y < H; y += 50) { ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(W, y); ctx.stroke() }
    } else {
      ctx.clearRect(0, 0, W, H)
    }

    // Links
    topoLinks.forEach(([a, b, ok]) => {
      const na = nodePositions[a], nb = nodePositions[b]
      if (!na || !nb) return

      const lw = fullMode ? 2 : 1.2
      const col = ok ? '#00f5c488' : '#ff456088'
      const colWeak = ok ? '#00f5c422' : '#ff456022'

      const g = ctx.createLinearGradient(na.px, na.py, nb.px, nb.py)
      g.addColorStop(0, colWeak)
      g.addColorStop(0.5, col)
      g.addColorStop(1, colWeak)

      ctx.beginPath()
      ctx.moveTo(na.px, na.py)
      ctx.lineTo(nb.px, nb.py)
      ctx.strokeStyle = g
      ctx.lineWidth = lw
      ctx.setLineDash(ok ? [] : [fullMode ? 6 : 4, fullMode ? 6 : 5])
      ctx.stroke()
      ctx.setLineDash([])

      // Animated packets
      if (ok) {
        const speed = fullMode ? 0.006 : 0.003
        const t = ((frameCount * speed) + (a.charCodeAt(0) * 0.17)) % 1
        const px = na.px + (nb.px - na.px) * t
        const py = na.py + (nb.py - na.py) * t
        ctx.beginPath()
        ctx.arc(px, py, fullMode ? 3 : 2, 0, Math.PI * 2)
        ctx.fillStyle = '#00f5c4'
        if (fullMode) {
          ctx.shadowColor = '#00f5c4'
          ctx.shadowBlur = 8
        }
        ctx.fill()
        ctx.shadowBlur = 0
      }
    })

    // Nodes
    topoNodes.forEach(n => {
      const nd = nodePositions[n.id]

      if (fullMode) {
        // Glow rings
        [nd.r * 3, nd.r * 2].forEach(r => {
          const gr = ctx.createRadialGradient(nd.px, nd.py, 0, nd.px, nd.py, r)
          gr.addColorStop(0, nd.color + '18')
          gr.addColorStop(1, 'transparent')
          ctx.beginPath(); ctx.arc(nd.px, nd.py, r, 0, Math.PI * 2)
          ctx.fillStyle = gr; ctx.fill()
        })

        // Outer ring
        ctx.beginPath(); ctx.arc(nd.px, nd.py, nd.r * 1.4, 0, Math.PI * 2)
        ctx.fillStyle = '#0c1624'; ctx.fill()
        ctx.strokeStyle = nd.color + '44'; ctx.lineWidth = 1; ctx.stroke()

        // Inner circle
        ctx.beginPath(); ctx.arc(nd.px, nd.py, nd.r, 0, Math.PI * 2)
        ctx.fillStyle = '#0c1624'; ctx.fill()
        ctx.strokeStyle = nd.color; ctx.lineWidth = 2; ctx.stroke()

        // Label
        ctx.fillStyle = nd.color
        ctx.font = 'bold 11px JetBrains Mono, monospace'
        ctx.textAlign = 'center'; ctx.textBaseline = 'middle'
        ctx.fillText(nd.label, nd.px, nd.py)

        // Sub label
        const subLabels = {
          internet: 'INTERNET', fw: 'FW-EDGE', cr1: 'RTR-01', cr2: 'RTR-02',
          sw1: 'SW-01', sw2: 'SW-02', sw3: 'SW-03',
          e1: 'EDGE-BJ', e2: 'EDGE-SH', e3: 'EDGE-GZ', e4: 'EDGE-CD', e5: 'EDGE-WH',
        }
        ctx.fillStyle = '#5a8898'
        ctx.font = '9px JetBrains Mono, monospace'
        ctx.fillText(subLabels[n.id] || '', nd.px, nd.py + nd.r + 12)
      } else {
        // Mini mode
        const gr = ctx.createRadialGradient(nd.px, nd.py, 0, nd.px, nd.py, nd.r * 2.5)
        gr.addColorStop(0, nd.color + '20')
        gr.addColorStop(1, 'transparent')
        ctx.beginPath(); ctx.arc(nd.px, nd.py, nd.r * 2.5, 0, Math.PI * 2)
        ctx.fillStyle = gr; ctx.fill()

        ctx.beginPath(); ctx.arc(nd.px, nd.py, nd.r, 0, Math.PI * 2)
        ctx.fillStyle = '#0c1624'; ctx.fill()
        ctx.strokeStyle = nd.color; ctx.lineWidth = 1.5; ctx.stroke()

        ctx.fillStyle = nd.color
        ctx.font = '500 9px JetBrains Mono, monospace'
        ctx.textAlign = 'center'; ctx.textBaseline = 'middle'
        ctx.fillText(nd.label.slice(0, 6), nd.px, nd.py)
      }
    })

    frameCount++
    rafId = requestAnimationFrame(draw)
  }

  function init() {
    canvas = canvasRef.value
    if (!canvas) return

    const parent = canvas.parentElement
    W = parent.clientWidth
    H = fullMode ? parent.clientHeight : parent.clientHeight - headerH
    canvas.width = W
    canvas.height = H
    ctx = canvas.getContext('2d')
    nodePositions = calcNodePositions()

    if (rafId) cancelAnimationFrame(rafId)
    draw()
  }

  function handleMouseMove(e) {
    if (!canvas || !fullMode) return
    const rect = canvas.getBoundingClientRect()
    const mx = e.clientX - rect.left
    const my = e.clientY - rect.top
    const hit = findNode(mx, my)

    if (hit) {
      const realNode = findRealNode(hit.topoNode.id)
      tooltipData.value = { topoNode: hit.topoNode, realNode }
      tooltipStyle.value = {
        display: 'block',
        left: (mx + 14) + 'px',
        top: (my - 10) + 'px',
      }
    } else {
      tooltipStyle.value = { display: 'none' }
    }
  }

  function handleMouseLeave() {
    tooltipStyle.value = { display: 'none' }
  }

  function resize() {
    init()
  }

  function zoomIn() { scale.value = Math.min(3, scale.value * 1.2) }
  function zoomOut() { scale.value = Math.max(0.5, scale.value * 0.8) }
  function reset() { scale.value = 1; offset.value = { x: 0, y: 0 } }

  let resizeObserver = null
  onMounted(() => {
    init()
    if (fullMode) {
      resizeObserver = new ResizeObserver(() => init())
      const parent = canvas?.parentElement
      if (parent) resizeObserver.observe(parent)
    }
  })

  onBeforeUnmount(() => {
    if (rafId) cancelAnimationFrame(rafId)
    if (resizeObserver) resizeObserver.disconnect()
  })

  return {
    tooltipData,
    tooltipStyle,
    scale,
    offset,
    init,
    resize,
    zoomIn,
    zoomOut,
    reset,
    handleMouseMove,
    handleMouseLeave,
  }
}
