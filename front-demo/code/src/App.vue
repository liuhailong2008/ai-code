<template>
  <div class="app">
    <aside class="sidebar" id="sidebar">
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <div class="logo-icon">P</div>
          <div class="logo-text">PingMesh</div>
        </div>
        <button class="sidebar-toggle" @click="toggleSidebar">◀</button>
      </div>
      <nav class="sidebar-nav">
        <div class="nav-section-title">监控中心</div>
        <router-link to="/" class="nav-item" :class="{ active: $route.path === '/' }">
          <span class="nav-icon">🔥</span> 整体网络监控
        </router-link>
        <router-link to="/leaf-connectivity" class="nav-item" :class="{ active: $route.path === '/leaf-connectivity' }">
          <span class="nav-icon">🌐</span> 机房网络监控
        </router-link>
        <router-link to="/latency" class="nav-item" :class="{ active: $route.path === '/latency' }">
          <span class="nav-icon">⏱️</span> 耗时监控
        </router-link>

        <div class="nav-section-title">告警管理</div>
        <router-link to="/alerts/active" class="nav-item" :class="{ active: $route.path === '/alerts/active' }">
          <span class="nav-icon">⚠️</span> 活跃告警<span class="nav-badge">3</span>
        </router-link>
        <router-link to="/alerts/history" class="nav-item" :class="{ active: $route.path === '/alerts/history' }">
          <span class="nav-icon">📋</span> 历史告警
        </router-link>
        <router-link to="/alerts/config" class="nav-item" :class="{ active: $route.path === '/alerts/config' }">
          <span class="nav-icon">⚙️</span> 告警配置
        </router-link>

        <div class="nav-section-title">任务中心</div>
        <router-link to="/devices" class="nav-item" :class="{ active: $route.path === '/devices' }">
          <span class="nav-icon">🖥️</span> 设备管理
        </router-link>
        <router-link to="/strategies" class="nav-item" :class="{ active: $route.path === '/strategies' }">
          <span class="nav-icon">📑</span> 策略管理
        </router-link>
        <router-link to="/tasks" class="nav-item" :class="{ active: $route.path === '/tasks' }">
          <span class="nav-icon">📝</span> 任务管理
        </router-link>

        <div class="nav-section-title">配置中心</div>
        <router-link to="/agents" class="nav-item" :class="{ active: $route.path === '/agents' }">
          <span class="nav-icon">🤖</span> 探测代理管理
        </router-link>
      </nav>
    </aside>

    <div class="main">
      <header class="header">
        <div class="header-left">
          <button class="menu-toggle" @click="toggleSidebar">☰</button>
          <div class="breadcrumb">监控面板 <span>/ Dashboard</span></div>
        </div>
        <div class="header-right">
          <div class="search-box">
            <span>🔍</span>
            <input type="text" placeholder="搜索节点、告警、IP...">
          </div>
          <button class="header-btn">🔔<span class="dot"></span></button>
          <button class="header-btn">⚙️</button>
          <div class="user-avatar">龙</div>
          <div class="theme-switcher">
            <span style="font-size:12px;color:var(--text-muted);">主题</span>
            <button
              v-for="theme in themes"
              :key="theme.id"
              class="theme-btn"
              :class="{ active: currentTheme === theme.id }"
              @click="switchTheme(theme.id)"
            >
              {{ theme.label }}
            </button>
          </div>
        </div>
      </header>

      <main class="content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const themes = [
  { id: 'style1', label: '深色' },
  { id: 'style2', label: '企业' },
  { id: 'style3', label: '专业' }
]

const currentTheme = ref('style1')

function switchTheme(theme) {
  currentTheme.value = theme
  document.documentElement.setAttribute('data-theme', theme)
  localStorage.setItem('pingmesh-theme', theme)
}

function toggleSidebar() {
  document.getElementById('sidebar').classList.toggle('open')
}

onMounted(() => {
  const saved = localStorage.getItem('pingmesh-theme') || 'style1'
  currentTheme.value = saved
  document.documentElement.setAttribute('data-theme', saved)
})
</script>
