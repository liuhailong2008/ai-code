<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

// ── 粒子系统 ──
const canvas = ref(null)
let ctx = null
let animId = null
let particles = []
const mouse = { x: -1000, y: -1000 }

const PARTICLE_COUNT = 80
const CONNECTION_DIST = 140
const CONNECTION_DIST_SQ = CONNECTION_DIST * CONNECTION_DIST
const ATTRACT_DIST = 250
const REPEL_DIST = 60

function initParticles(w, h) {
  particles = []
  for (let i = 0; i < PARTICLE_COUNT; i++) {
    particles.push({
      x: Math.random() * w,
      y: Math.random() * h,
      vx: (Math.random() - 0.5) * 0.6,
      vy: (Math.random() - 0.5) * 0.6,
      radius: Math.random() * 1.6 + 0.6,
      pulse: Math.random() * Math.PI * 2,
      density: Math.random() * 20 + 10,
    })
  }
}

function draw() {
  const c = canvas.value
  if (!c || !ctx) return
  const w = c.width = c.offsetWidth
  const h = c.height = c.offsetHeight
  if (particles.length === 0) initParticles(w, h)

  ctx.clearRect(0, 0, w, h)

  // 绘制粒子
  for (let i = 0; i < particles.length; i++) {
    const p = particles[i]
    p.pulse += 0.015

    // 自然漂浮
    p.x += p.vx
    p.y += p.vy

    // 鼠标吸引（远距离吸引）
    const dx = mouse.x - p.x
    const dy = mouse.y - p.y
    const dist = Math.sqrt(dx * dx + dy * dy)

    if (dist < ATTRACT_DIST && dist > REPEL_DIST && mouse.x > -500) {
      const force = ((ATTRACT_DIST - dist) / ATTRACT_DIST) * 0.05
      p.vx += (dx / dist) * force
      p.vy += (dy / dist) * force
    }

    // 鼠标排斥（近距离推开）
    if (dist < REPEL_DIST && dist > 0 && mouse.x > -500) {
      const force = ((REPEL_DIST - dist) / REPEL_DIST) * p.density * 0.03
      p.x -= (dx / dist) * force
      p.y -= (dy / dist) * force
    }

    // 速度衰减
    p.vx *= 0.99
    p.vy *= 0.99

    // 边界穿越
    if (p.x < 0) p.x = w
    if (p.x > w) p.x = 0
    if (p.y < 0) p.y = h
    if (p.y > h) p.y = 0

    // 绘制粒子
    const alpha = 0.25 + Math.sin(p.pulse) * 0.15
    ctx.beginPath()
    ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2)
    ctx.fillStyle = `rgba(0,245,196,${alpha})`
    ctx.fill()
  }

  // 粒子连线
  for (let i = 0; i < particles.length; i++) {
    for (let j = i + 1; j < particles.length; j++) {
      const dx = particles[i].x - particles[j].x
      const dy = particles[i].y - particles[j].y
      const distSq = dx * dx + dy * dy
      if (distSq < CONNECTION_DIST_SQ) {
        const alpha = (1 - Math.sqrt(distSq) / CONNECTION_DIST) * 0.1
        ctx.beginPath()
        ctx.moveTo(particles[i].x, particles[i].y)
        ctx.lineTo(particles[j].x, particles[j].y)
        ctx.strokeStyle = `rgba(0,245,196,${alpha})`
        ctx.lineWidth = 0.6
        ctx.stroke()
      }
    }
  }

  // 鼠标光晕
  if (mouse.x > -500) {
    const gradient = ctx.createRadialGradient(mouse.x, mouse.y, 0, mouse.x, mouse.y, 200)
    gradient.addColorStop(0, 'rgba(0,245,196,0.06)')
    gradient.addColorStop(1, 'rgba(0,0,0,0)')
    ctx.fillStyle = gradient
    ctx.fillRect(0, 0, w, h)
  }

  animId = requestAnimationFrame(draw)
}

// 鼠标追踪
function onMouseMove(e) {
  mouse.x = e.clientX
  mouse.y = e.clientY
}
function onMouseLeave() {
  mouse.x = -1000
  mouse.y = -1000
}

// 点击散开
function onClickCanvas(e) {
  const cx = e.clientX
  const cy = e.clientY
  for (let i = 0; i < particles.length; i++) {
    const p = particles[i]
    const dx = p.x - cx
    const dy = p.y - cy
    const dist = Math.sqrt(dx * dx + dy * dy)
    if (dist < 200) {
      const angle = Math.atan2(dy, dx)
      const force = (200 - dist) / 200 * 8
      p.vx += Math.cos(angle) * force
      p.vy += Math.sin(angle) * force
    }
  }
}

onMounted(async () => {
  // CAS 启用时自动跳转 CAS 登录页
  try {
    const res = await fetch('/api/auth/config')
    const data = await res.json()
    if (data.cas_enabled) {
      window.location.href = `${data.cas_login_url}?service=${encodeURIComponent(window.location.origin + '/api/auth/cas/callback')}`
      return
    }
  } catch {
    // 获取配置失败，使用本地登录模式
  }

  if (canvas.value) {
    ctx = canvas.value.getContext('2d')
    initParticles(window.innerWidth, window.innerHeight)
  }
  draw()
  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseleave', onMouseLeave)
  canvas.value?.addEventListener('click', onClickCanvas)
})

onUnmounted(() => {
  cancelAnimationFrame(animId)
  window.removeEventListener('mousemove', onMouseMove)
  window.removeEventListener('mouseleave', onMouseLeave)
  canvas.value?.removeEventListener('click', onClickCanvas)
})

async function handleLogin() {
  error.value = ''
  if (!username.value || !password.value) {
    error.value = '请输入用户名和密码'
    return
  }
  loading.value = true
  try {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username.value, password: password.value }),
    })
    const data = await res.json()
    if (data.success) {
      localStorage.setItem('pingmesh_logged_in', '1')
      localStorage.setItem('pingmesh_username', username.value)
      if (data.display_name) {
        localStorage.setItem('pingmesh_display_name', data.display_name)
      }
      router.push('/monitor')
    } else {
      error.value = data.message || '登录失败'
    }
  } catch (e) {
    error.value = '服务连接失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <canvas ref="canvas" class="particle-canvas"></canvas>

    <div class="login-container">
      <!-- Brand -->
      <div class="brand" style="animation-delay: 0s">
        <svg class="brand-icon" width="48" height="48" viewBox="0 0 48 48" fill="none">
          <!-- Outer halo -->
          <circle cx="24" cy="24" r="22" stroke="currentColor" stroke-width="0.5" opacity="0.08" stroke-dasharray="1.5 4"/>

          <!-- Hexagon outline -->
          <polygon points="24,5 40,14.5 40,33.5 24,43 8,33.5 8,14.5" stroke="currentColor" stroke-width="1.2" fill="none" opacity="0.5"/>

          <!-- Inner triangle 1 (top) -->
          <polygon points="24,5 40,33.5 8,33.5" stroke="currentColor" stroke-width="0.7" fill="none" opacity="0.2"/>
          <!-- Inner triangle 2 (bottom) -->
          <polygon points="24,43 40,14.5 8,14.5" stroke="currentColor" stroke-width="0.7" fill="none" opacity="0.2"/>

          <!-- Cross diagonals through center -->
          <line x1="24" y1="5" x2="24" y2="43" stroke="currentColor" stroke-width="0.8" opacity="0.3"/>
          <line x1="8" y1="14.5" x2="40" y2="33.5" stroke="currentColor" stroke-width="0.8" opacity="0.3"/>
          <line x1="40" y1="14.5" x2="8" y2="33.5" stroke="currentColor" stroke-width="0.8" opacity="0.3"/>

          <!-- Mid-edge connections (inner hexagon) -->
          <line x1="32" y1="9.75" x2="32" y2="38.25" stroke="currentColor" stroke-width="0.5" opacity="0.15"/>
          <line x1="16" y1="9.75" x2="16" y2="38.25" stroke="currentColor" stroke-width="0.5" opacity="0.15"/>
          <line x1="8" y1="24" x2="40" y2="24" stroke="currentColor" stroke-width="0.5" opacity="0.15"/>

          <!-- Vertex nodes -->
          <circle cx="24" cy="5" r="2.5" fill="currentColor"/>
          <circle cx="40" cy="14.5" r="2.5" fill="currentColor"/>
          <circle cx="40" cy="33.5" r="2.5" fill="currentColor"/>
          <circle cx="24" cy="43" r="2.5" fill="currentColor"/>
          <circle cx="8" cy="33.5" r="2.5" fill="currentColor"/>
          <circle cx="8" cy="14.5" r="2.5" fill="currentColor"/>

          <!-- Mid-edge nodes -->
          <circle cx="32" cy="9.75" r="1.5" fill="currentColor" opacity="0.6"/>
          <circle cx="40" cy="24" r="1.5" fill="currentColor" opacity="0.6"/>
          <circle cx="32" cy="38.25" r="1.5" fill="currentColor" opacity="0.6"/>
          <circle cx="16" cy="38.25" r="1.5" fill="currentColor" opacity="0.6"/>
          <circle cx="8" cy="24" r="1.5" fill="currentColor" opacity="0.6"/>
          <circle cx="16" cy="9.75" r="1.5" fill="currentColor" opacity="0.6"/>

          <!-- Center core -->
          <circle cx="24" cy="24" r="3.5" fill="currentColor" opacity="0.9"/>
          <circle cx="24" cy="24" r="6" stroke="currentColor" stroke-width="0.8" fill="none" opacity="0.25"/>
        </svg>
        <div class="brand-text">
          <span class="brand-main">PINGMESH</span>
          <span class="brand-sub">NETWORK MONITORING PLATFORM</span>
        </div>
      </div>

      <!-- Login Card -->
      <div class="login-card" style="animation-delay: 0.1s">
        <div class="card-header">
          <span class="card-blink"></span>
          <span class="card-title">系统认证</span>
          <span class="card-blink"></span>
        </div>

        <form class="login-form" @submit.prevent="handleLogin">
          <div class="field-group" style="animation-delay: 0.2s">
            <label class="field-label">
              <span class="label-icon">&#9654;</span>
              用户名
            </label>
            <div class="input-wrapper">
              <input
                v-model="username"
                type="text"
                placeholder="请输入用户名"
                class="tech-input"
                autocomplete="off"
                spellcheck="false"
              />
              <span class="input-border"></span>
            </div>
          </div>

          <div class="field-group" style="animation-delay: 0.3s">
            <label class="field-label">
              <span class="label-icon">&#9654;</span>
              密码
            </label>
            <div class="input-wrapper">
              <input
                v-model="password"
                type="password"
                placeholder="请输入密码"
                class="tech-input"
                autocomplete="off"
              />
              <span class="input-border"></span>
            </div>
          </div>

          <div v-if="error" class="error-msg">{{ error }}</div>

          <div class="form-actions" style="animation-delay: 0.4s">
            <button type="submit" class="login-btn" :class="{ loading }" :disabled="loading">
              <span v-if="!loading" class="btn-content">
                <span class="btn-bracket">[</span>
                认证登录
                <span class="btn-bracket">]</span>
              </span>
              <span v-else class="btn-loading">
                <span class="loading-dot"></span>
                <span class="loading-dot"></span>
                <span class="loading-dot"></span>
              </span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: radial-gradient(ellipse at 50% 0%, #0a1a2a 0%, #050c12 70%);
  overflow: hidden;
}

.particle-canvas {
  position: fixed;
  inset: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
}

/* ── Container ── */
.login-container {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 32px;
}

/* ── Brand ── */
.brand {
  display: flex;
  align-items: center;
  gap: 16px;
  opacity: 0;
  animation: fadeSlideIn 0.7s ease forwards;
}
.brand-icon {
  color: var(--cyan);
  filter: drop-shadow(0 0 12px rgba(0,245,196,0.3));
}
.brand-text {
  display: flex;
  flex-direction: column;
}
.brand-main {
  font-family: var(--font-head);
  font-size: 28px;
  font-weight: 900;
  letter-spacing: 8px;
  color: var(--cyan);
  text-shadow: 0 0 30px rgba(0,245,196,0.25);
}
.brand-sub {
  font-family: var(--font-mono);
  font-size: 9px;
  letter-spacing: 4px;
  color: var(--text-dim);
  margin-top: 2px;
}

/* ── Login Card ── */
.login-card {
  width: 380px;
  background: linear-gradient(180deg, rgba(12,22,36,0.95) 0%, rgba(8,15,24,0.98) 100%);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  padding: 32px 36px;
  position: relative;
  opacity: 0;
  animation: fadeSlideIn 0.7s ease forwards;
  backdrop-filter: blur(10px);
  box-shadow:
    0 0 60px rgba(0,245,196,0.04),
    inset 0 1px 0 rgba(0,245,196,0.04);
}
.login-card::before {
  content: '';
  position: absolute;
  inset: -1px;
  border-radius: var(--radius);
  padding: 1px;
  background: linear-gradient(135deg, rgba(0,245,196,0.15), transparent 40%, transparent 60%, rgba(0,245,196,0.08));
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  pointer-events: none;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 28px;
}
.card-blink {
  width: 6px;
  height: 6px;
  background: var(--cyan);
  border-radius: 50%;
  animation: pulse-dot 2s ease-in-out infinite;
}
.card-blink:last-child {
  animation-delay: 1s;
}
.card-title {
  font-family: var(--font-head);
  font-size: 13px;
  letter-spacing: 5px;
  color: var(--text-dim);
}

/* ── Form ── */
.login-form {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.field-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
  opacity: 0;
  animation: fadeSlideIn 0.6s ease forwards;
}

.field-label {
  font-family: var(--font-mono);
  font-size: 10px;
  letter-spacing: 2px;
  color: var(--muted);
  display: flex;
  align-items: center;
  gap: 6px;
  user-select: none;
}
.label-icon {
  font-size: 7px;
  color: var(--cyan);
}

.input-wrapper {
  position: relative;
}
.tech-input {
  width: 100%;
  background: rgba(0,245,196,0.03);
  border: 1px solid rgba(0,245,196,0.12);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  font-family: var(--font-mono);
  font-size: 13px;
  color: var(--text);
  outline: none;
  transition: all 0.25s;
  caret-color: var(--cyan);
}
.tech-input::placeholder {
  color: var(--muted);
  font-size: 11px;
  letter-spacing: 0.5px;
}
.tech-input:focus {
  border-color: var(--cyan-dim);
  background: rgba(0,245,196,0.05);
  box-shadow: 0 0 16px rgba(0,245,196,0.06), inset 0 0 16px rgba(0,245,196,0.02);
}

.input-border {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 1px;
  background: var(--cyan);
  transition: width 0.35s ease;
}
.tech-input:focus ~ .input-border {
  width: 60%;
}

/* ── Error ── */
.error-msg {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--err);
  background: rgba(255,69,96,0.06);
  border: 1px solid rgba(255,69,96,0.15);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  text-align: center;
  letter-spacing: 0.5px;
}

/* ── Button ── */
.form-actions {
  margin-top: 4px;
  opacity: 0;
  animation: fadeSlideIn 0.6s ease forwards;
}

.login-btn {
  width: 100%;
  padding: 13px;
  background: transparent;
  border: 1px solid var(--cyan-dim);
  border-radius: var(--radius-sm);
  color: var(--cyan);
  font-family: var(--font-head);
  font-size: 12px;
  letter-spacing: 4px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
  user-select: none;
}
.login-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(0,245,196,0.08), transparent);
  transform: translateX(-100%);
  transition: transform 0.5s;
}
.login-btn:hover:not(:disabled)::before {
  transform: translateX(100%);
}
.login-btn:hover:not(:disabled) {
  background: rgba(0,245,196,0.08);
  border-color: var(--cyan);
  box-shadow:
    0 0 24px rgba(0,245,196,0.12),
    inset 0 0 20px rgba(0,245,196,0.03);
  text-shadow: 0 0 12px rgba(0,245,196,0.4);
}
.login-btn:active:not(:disabled) {
  transform: scale(0.98);
}
.login-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}
.btn-bracket {
  color: var(--cyan-dim);
  font-weight: 400;
}

.btn-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}
.loading-dot {
  width: 6px;
  height: 6px;
  background: var(--cyan);
  border-radius: 50%;
  animation: pulse-dot 0.8s ease-in-out infinite;
}
.loading-dot:nth-child(2) { animation-delay: 0.15s; }
.loading-dot:nth-child(3) { animation-delay: 0.3s; }

/* ── Animations ── */
@keyframes fadeSlideIn {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse-dot {
  0%, 100% { opacity: 0.3; transform: scale(1); }
  50%      { opacity: 1;   transform: scale(1.4); }
}
</style>
