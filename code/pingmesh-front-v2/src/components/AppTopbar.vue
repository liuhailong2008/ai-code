<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/useApp'
import { apiFetch } from '../utils/api.js'

const store = useAppStore()
const route = useRoute()
const overallAvgLatency = ref('--')
const overallP99Latency = ref('--')
const overallLoss = ref('--')

// 用户显示信息
const userName = ref(localStorage.getItem('pingmesh_display_name') || localStorage.getItem('pingmesh_username') || '未登录')
const userAvatar = ref(userName.value.slice(-1))

async function fetchUserInfo() {
  const token = localStorage.getItem('pingmesh_token')
  if (token) {
    try {
      const res = await fetch('/api/auth/cas/user', {
        headers: { 'Authorization': `Bearer ${token}` }
      })
      const data = await res.json()
      if (data.success && data.user) {
        localStorage.setItem('pingmesh_username', data.user.username || '')
        localStorage.setItem('pingmesh_display_name', data.user.display_name || data.user.username || '')
        userName.value = data.user.display_name || data.user.username || ''
        userAvatar.value = userName.value.slice(-1)
      }
    } catch {
      // ignore
    }
    return
  }
  // 本地登录：从 localStorage 读取
  const displayName = localStorage.getItem('pingmesh_display_name')
  const username = localStorage.getItem('pingmesh_username')
  userName.value = displayName || username || '未登录'
  userAvatar.value = userName.value.slice(-1)
}

async function fetchOverallStats() {
  try {
    const res = await apiFetch('/api/dashboard-monitor/overall-stats')
    const data = await res.json()
    overallAvgLatency.value = data.overall_avg_latency || '--'
    overallP99Latency.value = data.overall_p99_latency || '--'
    overallLoss.value = data.overall_loss || '--'
  } catch {
    // ignore
  }
}

let statsTimer = null
onMounted(() => {
  fetchUserInfo()
  fetchOverallStats()
  statsTimer = setInterval(fetchOverallStats, 5000)
})
onBeforeUnmount(() => {
  if (statsTimer) clearInterval(statsTimer)
})

// 路由变化时重新读取用户信息（v-show 切换不会触发 onMounted）
watch(() => route.name, (name) => {
  if (name && name !== 'login') {
    fetchUserInfo()
  }
})
</script>

<template>
  <header class="topbar">
    <div class="topbar-left">
      <div class="breadcrumb">{{ store.breadcrumb }}</div>
      <div class="live-indicator">
        <div class="live-dot"></div>
        <span>实时监控</span>
      </div>
    </div>
    <div class="topbar-right">
      <div class="topbar-stat">
        <span class="ts-label">整体时延(平均)</span>
        <span class="ts-val green">{{ overallAvgLatency }} <span class="ts-unit">ms</span></span>
      </div>
      <div class="topbar-stat">
        <span class="ts-label">整体时延(P99)</span>
        <span class="ts-val dim">{{ overallP99Latency }} <span class="ts-unit">ms</span></span>
      </div>
      <div class="topbar-stat">
        <span class="ts-label">丢包率</span>
        <span class="ts-val dim">{{ overallLoss }} <span class="ts-unit">%</span></span>
      </div>
      <div class="topbar-sep"></div>
      <div class="topbar-user">
        <div class="user-avatar">{{ userAvatar }}</div>
        <span>{{ userName }}</span>
      </div>
    </div>
  </header>
</template>
