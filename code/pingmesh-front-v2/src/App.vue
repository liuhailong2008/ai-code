<script setup>
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAppStore } from '@/stores/useApp'
import { watch } from 'vue'
import AppSidebar from '@/components/AppSidebar.vue'
import AppTopbar from '@/components/AppTopbar.vue'
import AppTicker from '@/components/AppTicker.vue'

const route = useRoute()
const router = useRouter()
const store = useAppStore()
const fullscreen = ref(false)
const routerReady = ref(false)

router.isReady().then(() => { routerReady.value = true })

const isLoginPage = computed(() => route.name === 'login')
const isPublicPage = computed(() => route.name === 'public-monitor')

function toggleFullscreen() {
  fullscreen.value = !fullscreen.value
}

function handleLogout() {
  // 清除所有本地登录状态
  localStorage.removeItem('pingmesh_token')
  localStorage.removeItem('pingmesh_username')
  localStorage.removeItem('pingmesh_display_name')
  localStorage.removeItem('pingmesh_logged_in')
  localStorage.removeItem('pingmesh_user')
  // 跳转到后端登出接口，后端会清理 token 并重定向到 CAS 登出页
  window.location.href = '/api/auth/logout'
}

watch(() => route.name, (name) => {
  if (name) store.setPage(name)
}, { immediate: true })
</script>

<template>
  <div v-if="!isPublicPage" class="scan-line"></div>
  <div v-if="!isPublicPage" class="noise-overlay"></div>

  <AppSidebar v-if="routerReady && !fullscreen && !isLoginPage && !isPublicPage" />
  <main class="main-content" :class="{ fullscreen: fullscreen || isPublicPage, login: isLoginPage }">
    <div v-show="routerReady && !isLoginPage && !isPublicPage" class="fs-toggle" @click="toggleFullscreen" :title="fullscreen ? '退出全屏' : '全屏展示'">
      <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
        <template v-if="!fullscreen">
          <rect x="2" y="2" width="5" height="5" stroke="currentColor" stroke-width="1.3" fill="none"/>
          <rect x="9" y="2" width="5" height="5" stroke="currentColor" stroke-width="1.3" fill="none"/>
          <rect x="2" y="9" width="5" height="5" stroke="currentColor" stroke-width="1.3" fill="none"/>
          <rect x="9" y="9" width="5" height="5" stroke="currentColor" stroke-width="1.3" fill="none"/>
        </template>
        <template v-else>
          <rect x="3" y="3" width="10" height="10" stroke="currentColor" stroke-width="1.3" fill="none"/>
          <line x1="1" y1="1" x2="4" y2="4" stroke="currentColor" stroke-width="1.3"/>
          <line x1="15" y1="1" x2="12" y2="4" stroke="currentColor" stroke-width="1.3"/>
          <line x1="1" y1="15" x2="4" y2="12" stroke="currentColor" stroke-width="1.3"/>
          <line x1="15" y1="15" x2="12" y2="12" stroke="currentColor" stroke-width="1.3"/>
        </template>
      </svg>
    </div>
    <div v-show="routerReady && !isLoginPage && !isPublicPage" class="logout-btn" @click="handleLogout" title="注销登录">
      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/>
        <polyline points="16 17 21 12 16 7"/>
        <line x1="21" y1="12" x2="9" y2="12"/>
      </svg>
    </div>
    <AppTopbar v-if="routerReady && !fullscreen && !isLoginPage && !isPublicPage" />
    <AppTicker v-if="routerReady && !fullscreen && !isLoginPage && !isPublicPage" />
    <router-view v-slot="{ Component }">
      <transition name="page-fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </main>
</template>

<style scoped>
.fs-toggle {
  position: fixed;
  top: 10px;
  right: 14px;
  z-index: 200;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
  color: #3a6070;
  transition: color 0.15s;
}
.fs-toggle:hover {
  color: #7eb8c8;
}
.logout-btn {
  position: fixed;
  top: 10px;
  right: 42px;
  z-index: 200;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
  color: #3a6070;
  transition: color 0.15s;
}
.logout-btn:hover {
  color: #ff4560;
}
.main-content.fullscreen {
  margin-left: 0;
}
.main-content.login {
  margin-left: 0;
}

.main-content :deep(.topbar-right) { padding-right: 72px; }
.main-content :deep(.ticker-bar) { padding-right: 36px; }

.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 0.15s ease;
}
.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
}
</style>
