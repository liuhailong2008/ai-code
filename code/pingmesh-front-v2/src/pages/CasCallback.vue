<script setup>
import { onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()
const error = ref('')
const loading = ref(true)

onMounted(async () => {
  const token = route.query.token
  if (token) {
    localStorage.setItem('pingmesh_token', token)

    // 获取 CAS 用户信息并存入 localStorage
    try {
      const res = await fetch('/api/auth/cas/user', {
        headers: { 'Authorization': `Bearer ${token}` }
      })
      const data = await res.json()
      if (data.success && data.user) {
        localStorage.setItem('pingmesh_username', data.user.username || '')
        localStorage.setItem('pingmesh_display_name', data.user.display_name || data.user.username || '')
      }
    } catch {
      // 获取用户信息失败不影响登录流程
    }

    router.push('/monitor')
    return
  }

  const ticket = route.query.ticket
  if (!ticket) {
    error.value = '缺少认证参数'
    loading.value = false
    return
  }

  // 直接带 ticket 访问（备用路径：重定向到后端 CAS 回调）
  window.location.href = `/api/auth/cas/callback?ticket=${encodeURIComponent(ticket)}`
})
</script>

<template>
  <div class="cas-callback-page">
    <div class="callback-box">
      <template v-if="loading">
        <div class="loading-spinner"></div>
        <p class="loading-text">正在验证 CAS 认证...</p>
      </template>
      <template v-else-if="error">
        <p class="error">{{ error }}</p>
        <a href="/#/login" class="back-link">返回登录</a>
      </template>
    </div>
  </div>
</template>

<style scoped>
.cas-callback-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: #0a0f1a;
}
.callback-box {
  text-align: center;
}
.loading-spinner {
  width: 32px;
  height: 32px;
  border: 2px solid rgba(0,245,196,0.2);
  border-top-color: #00f5c4;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin: 0 auto 16px;
}
.loading-text {
  color: rgba(0,245,196,0.7);
  font-family: 'JetBrains Mono', monospace;
  font-size: 13px;
  letter-spacing: 1px;
}
.error {
  color: #f56c6c;
  margin-bottom: 16px;
  font-size: 14px;
}
.back-link {
  color: #00f5c4;
  text-decoration: none;
  font-size: 13px;
}
.back-link:hover {
  text-decoration: underline;
}
@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
