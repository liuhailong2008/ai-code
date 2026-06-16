<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { apiFetch } from '../utils/api.js'

const alerts = ref([])
let timer = null

async function fetchAlerts() {
  try {
    const res = await apiFetch('/api/dashboard-monitor/unresolved-alerts')
    const data = await res.json()
    alerts.value = data.names || []
  } catch {
    alerts.value = []
  }
}

onMounted(() => {
  fetchAlerts()
  timer = setInterval(fetchAlerts, 15000)
})

onBeforeUnmount(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="ticker-bar">
    <span class="ticker-label">实时告警</span>
    <div class="ticker-track">
      <div class="ticker-content" v-if="alerts.length">
        <template v-for="(a, i) in alerts" :key="i">
          {{ a }} <template v-if="i < alerts.length - 1">&nbsp;|&nbsp;</template>
        </template>
      </div>
      <div class="ticker-content" v-else>
        暂无告警
      </div>
    </div>
  </div>
</template>
