<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAppStore } from '@/stores/useApp'
import { formatTime } from '@/composables/useData'
import { apiFetch } from '../utils/api.js'

const router = useRouter()
const route = useRoute()
const store = useAppStore()
const sidebarTime = ref(formatTime())

let timer = null
let badgeTimer = null
const showHidden = ref(false)
const alertBadge = ref(null)

async function fetchBadge() {
  try {
    const res = await apiFetch('/api/alerts/unresolved-count')
    const data = await res.json()
    const count = data.count || 0
    alertBadge.value = count > 0 ? String(count) : null
  } catch {
    alertBadge.value = null
  }
}

function onKeydown(e) {
  if (e.ctrlKey && e.key.toLowerCase() === 'd') {
    e.preventDefault()
    showHidden.value = !showHidden.value
  }
}

onMounted(() => {
  timer = setInterval(() => { sidebarTime.value = formatTime() }, 1000)
  document.addEventListener('keydown', onKeydown)
  fetchBadge()
  badgeTimer = setInterval(fetchBadge, 30000)
})
onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
  if (badgeTimer) clearInterval(badgeTimer)
  document.removeEventListener('keydown', onKeydown)
})

const navItems = [
  { section: '监控', items: [
    { page: 'dashboard', icon:'grid',    label:'总览面板', badge: null,  dot: true,  hidden: true },
    { page: 'monitor',           icon:'monitor', label:'监控大屏',     badge: null,  dot: false },
    { page: 'leaf-connectivity', icon:'leaf',    label:'机房网络监控',  badge: null,  dot: false },
    { page: 'latency', icon:'latency', label:'机房耗时监控',  badge: null,  dot: false },
    { page: 'topology',  icon:'topo',    label:'拓扑图',   badge: null,  dot: false, hidden: true },
    { page: 'traffic',   icon:'chart',   label:'流量分析', badge: null,  dot: false, hidden: true },
  ]},
  { section: '管理', items: [
    { page: 'alerts',    icon:'alert',   label:'告警管理', badge: alertBadge,   dot: false },
    { page: 'statistics', icon:'chart', label:'统计数据', badge: null,  dot: false },
    { page: 'devices',   icon:'device',  label:'设备管理', badge: null,  dot: false, hidden: true },
    { page: 'reports',   icon:'report',  label:'报告中心', badge: null,  dot: false, hidden: true },
  ]},
  { section: '系统', items: [
    { page: 'settings-env', icon:'settings', label:'环境管理', badge: null,  dot: false },
    { page: 'metrics-manage', icon:'chart', label:'指标管理', badge: null,  dot: false },
    { page: 'settings',  icon:'settings',label:'系统设置', badge: null,  dot: false },
  ]},
]
</script>

<template>
  <nav class="sidebar">
    <div class="sidebar-logo">
      <div class="logo-icon">
        <svg width="28" height="28" viewBox="0 0 48 48" fill="none">
          <circle cx="24" cy="24" r="22" stroke="#00f5c4" stroke-width="0.5" opacity="0.08" stroke-dasharray="1.5 4"/>
          <polygon points="24,5 40,14.5 40,33.5 24,43 8,33.5 8,14.5" stroke="#00f5c4" stroke-width="1.2" fill="none" opacity="0.5"/>
          <polygon points="24,5 40,33.5 8,33.5" stroke="#00f5c4" stroke-width="0.7" fill="none" opacity="0.2"/>
          <polygon points="24,43 40,14.5 8,14.5" stroke="#00f5c4" stroke-width="0.7" fill="none" opacity="0.2"/>
          <line x1="24" y1="5" x2="24" y2="43" stroke="#00f5c4" stroke-width="0.8" opacity="0.3"/>
          <line x1="8" y1="14.5" x2="40" y2="33.5" stroke="#00f5c4" stroke-width="0.8" opacity="0.3"/>
          <line x1="40" y1="14.5" x2="8" y2="33.5" stroke="#00f5c4" stroke-width="0.8" opacity="0.3"/>
          <line x1="32" y1="9.75" x2="32" y2="38.25" stroke="#00f5c4" stroke-width="0.5" opacity="0.15"/>
          <line x1="16" y1="9.75" x2="16" y2="38.25" stroke="#00f5c4" stroke-width="0.5" opacity="0.15"/>
          <line x1="8" y1="24" x2="40" y2="24" stroke="#00f5c4" stroke-width="0.5" opacity="0.15"/>
          <circle cx="24" cy="5" r="2.5" fill="#00f5c4"/>
          <circle cx="40" cy="14.5" r="2.5" fill="#00f5c4"/>
          <circle cx="40" cy="33.5" r="2.5" fill="#00f5c4"/>
          <circle cx="24" cy="43" r="2.5" fill="#00f5c4"/>
          <circle cx="8" cy="33.5" r="2.5" fill="#00f5c4"/>
          <circle cx="8" cy="14.5" r="2.5" fill="#00f5c4"/>
          <circle cx="32" cy="9.75" r="1.5" fill="#00f5c4" opacity="0.6"/>
          <circle cx="40" cy="24" r="1.5" fill="#00f5c4" opacity="0.6"/>
          <circle cx="32" cy="38.25" r="1.5" fill="#00f5c4" opacity="0.6"/>
          <circle cx="16" cy="38.25" r="1.5" fill="#00f5c4" opacity="0.6"/>
          <circle cx="8" cy="24" r="1.5" fill="#00f5c4" opacity="0.6"/>
          <circle cx="16" cy="9.75" r="1.5" fill="#00f5c4" opacity="0.6"/>
          <circle cx="24" cy="24" r="3.5" fill="#00f5c4" opacity="0.9"/>
          <circle cx="24" cy="24" r="6" stroke="#00f5c4" stroke-width="0.8" fill="none" opacity="0.25"/>
        </svg>
      </div>
      <div class="logo-text">
        <span class="logo-main">Pingmesh</span>
        <span class="logo-sub">v1.0.0</span>
      </div>
    </div>

    <template v-for="group in navItems" :key="group.section">
      <div
        class="nav-section-label"
        v-if="group.items.some(i => !i.hidden || showHidden)"
      >{{ group.section }}</div>
      <router-link
        v-for="item in group.items"
        :key="item.page"
        v-show="!item.hidden || showHidden"
        :to="{ name: item.page }"
        class="nav-item"
      >
        <!-- Grid icon -->
        <svg v-if="item.icon==='grid'" width="16" height="16" viewBox="0 0 16 16">
          <rect x="1" y="1" width="6" height="6" rx="1" fill="currentColor" opacity="0.8"/><rect x="9" y="1" width="6" height="6" rx="1" fill="currentColor" opacity="0.8"/><rect x="1" y="9" width="6" height="6" rx="1" fill="currentColor" opacity="0.8"/><rect x="9" y="9" width="6" height="6" rx="1" fill="currentColor" opacity="0.8"/>
        </svg>
        <!-- Monitor icon -->
        <svg v-else-if="item.icon==='monitor'" width="16" height="16" viewBox="0 0 16 16">
          <rect x="1" y="2" width="14" height="10" rx="1" fill="none" stroke="currentColor" stroke-width="1.4"/><line x1="5" y1="12" x2="11" y2="12" stroke="currentColor" stroke-width="1.2"/><polyline points="3,14 8,12 13,14" fill="none" stroke="currentColor" stroke-width="1.2"/>
        </svg>
        <!-- Leaf icon -->
        <svg v-else-if="item.icon==='leaf'" width="16" height="16" viewBox="0 0 16 16">
          <circle cx="5" cy="5" r="2" fill="currentColor" opacity="0.8"/><circle cx="11" cy="5" r="2" fill="currentColor" opacity="0.8"/><circle cx="5" cy="11" r="2" fill="currentColor" opacity="0.8"/><circle cx="11" cy="11" r="2" fill="currentColor" opacity="0.8"/><line x1="7" y1="5" x2="9" y2="5" stroke="currentColor" stroke-width="1.2" opacity="0.6"/><line x1="5" y1="7" x2="5" y2="9" stroke="currentColor" stroke-width="1.2" opacity="0.6"/><line x1="11" y1="7" x2="11" y2="9" stroke="currentColor" stroke-width="1.2" opacity="0.6"/><line x1="7" y1="11" x2="9" y2="11" stroke="currentColor" stroke-width="1.2" opacity="0.6"/>
        </svg>
        <!-- Latency icon -->
        <svg v-else-if="item.icon==='latency'" width="16" height="16" viewBox="0 0 16 16">
          <circle cx="8" cy="8" r="6" fill="none" stroke="currentColor" stroke-width="1.4"/><polyline points="8,3 8,8 12,10" fill="none" stroke="currentColor" stroke-width="1.4"/><circle cx="8" cy="8" r="1" fill="currentColor"/>
        </svg>
        <!-- Topo icon -->
        <svg v-else-if="item.icon==='topo'" width="16" height="16" viewBox="0 0 16 16">
          <circle cx="8" cy="3" r="2" fill="currentColor" opacity="0.8"/><circle cx="2" cy="13" r="2" fill="currentColor" opacity="0.8"/><circle cx="14" cy="13" r="2" fill="currentColor" opacity="0.8"/><line x1="8" y1="5" x2="2" y2="11" stroke="currentColor" stroke-width="1.2" opacity="0.6"/><line x1="8" y1="5" x2="14" y2="11" stroke="currentColor" stroke-width="1.2" opacity="0.6"/><line x1="4" y1="13" x2="12" y2="13" stroke="currentColor" stroke-width="1.2" opacity="0.6"/>
        </svg>
        <!-- Chart icon -->
        <svg v-else-if="item.icon==='chart'" width="16" height="16" viewBox="0 0 16 16">
          <polyline points="1,12 4,7 7,9 10,4 13,8 15,5" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.8"/>
        </svg>
        <!-- Alert icon -->
        <svg v-else-if="item.icon==='alert'" width="16" height="16" viewBox="0 0 16 16">
          <path d="M8 1L1 13h14L8 1z" fill="none" stroke="currentColor" stroke-width="1.4" opacity="0.8"/><line x1="8" y1="6" x2="8" y2="9" stroke="currentColor" stroke-width="1.5"/><circle cx="8" cy="11" r="0.8" fill="currentColor"/>
        </svg>
        <!-- Device icon -->
        <svg v-else-if="item.icon==='device'" width="16" height="16" viewBox="0 0 16 16">
          <rect x="1" y="3" width="14" height="10" rx="1" fill="none" stroke="currentColor" stroke-width="1.4"/><line x1="5" y1="13" x2="5" y2="15" stroke="currentColor" stroke-width="1.4"/><line x1="11" y1="13" x2="11" y2="15" stroke="currentColor" stroke-width="1.4"/><line x1="3" y1="15" x2="13" y2="15" stroke="currentColor" stroke-width="1.4"/>
        </svg>
        <!-- Report icon -->
        <svg v-else-if="item.icon==='report'" width="16" height="16" viewBox="0 0 16 16">
          <rect x="2" y="1" width="12" height="14" rx="1" fill="none" stroke="currentColor" stroke-width="1.4"/><line x1="5" y1="5" x2="11" y2="5" stroke="currentColor" stroke-width="1.2"/><line x1="5" y1="8" x2="11" y2="8" stroke="currentColor" stroke-width="1.2"/><line x1="5" y1="11" x2="9" y2="11" stroke="currentColor" stroke-width="1.2"/>
        </svg>
        <!-- Settings icon -->
        <svg v-else-if="item.icon==='settings'" width="16" height="16" viewBox="0 0 16 16">
          <circle cx="8" cy="8" r="2.5" fill="none" stroke="currentColor" stroke-width="1.4"/><path d="M8 1v2M8 13v2M1 8h2M13 8h2M3.1 3.1l1.4 1.4M11.5 11.5l1.4 1.4M3.1 12.9l1.4-1.4M11.5 4.5l1.4-1.4" stroke="currentColor" stroke-width="1.3"/>
        </svg>
        <span>{{ item.label }}</span>
        <div v-if="item.dot" class="nav-dot"></div>
        <div v-if="item.badge" class="nav-badge">{{ item.badge }}</div>
      </router-link>
    </template>

    <div class="sidebar-footer">
      <div class="system-health">
        <div class="health-dot"></div>
        <span>系统运行正常</span>
      </div>
      <div class="footer-time">{{ sidebarTime }}</div>
    </div>
  </nav>
</template>
