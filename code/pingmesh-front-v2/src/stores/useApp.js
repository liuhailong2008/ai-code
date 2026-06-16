import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiFetch } from '../utils/api.js'

export const useAppStore = defineStore('app', () => {
  const currentPage = ref('dashboard')

  // IDC 列表缓存：全局复用，避免重复请求
  const _idcCache = ref(null)

  async function fetchIdcs() {
    if (_idcCache.value) return _idcCache.value
    try {
      const res = await apiFetch('/api/idcs')
      _idcCache.value = await res.json()
      return _idcCache.value
    } catch (e) {
      console.error('fetchIdcs failed:', e)
      return []
    }
  }

  const pageLabels = {
    dashboard: '总览面板',
    monitor:   '监控大屏',
    'leaf-connectivity': '机房网络监控',
    latency:           '机房耗时监控',
    statistics:        '统计数据',
    topology:  '拓扑图',
    traffic:   '流量分析',
    alerts:    '告警管理',
    devices:   '设备管理',
    reports:   '报告中心',
    settings:  '系统设置',
    'settings-env': '环境管理',
    'metrics-manage': '指标管理',
    'metric-query': '指标查询',
  }

  const breadcrumb = computed(() => pageLabels[currentPage.value] || currentPage.value)

  function setPage(page) {
    currentPage.value = page
  }

  return { currentPage, pageLabels, breadcrumb, setPage, fetchIdcs }
})
