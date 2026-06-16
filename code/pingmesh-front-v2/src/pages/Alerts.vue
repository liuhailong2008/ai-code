<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { apiFetch } from '../utils/api.js'

const rawAlerts = ref([])
const total = ref(0)
const pageSize = 5
const currentPage = ref(1)
const loading = ref(false)
const selectedIndex = ref(-1)
const selectedAlert = computed(() => selectedIndex.value >= 0 ? rawAlerts.value[selectedIndex.value] : null)
const filterStatus = ref('') // '' = 全部

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

// 分页数字按钮列表
const pageNumbers = computed(() => {
  const pages = []
  const tp = totalPages.value
  const cp = currentPage.value
  if (tp <= 7) {
    for (let i = 1; i <= tp; i++) pages.push(i)
  } else {
    pages.push(1)
    if (cp > 3) pages.push('...')
    const start = Math.max(2, cp - 1)
    const end = Math.min(tp - 1, cp + 1)
    for (let i = start; i <= end; i++) pages.push(i)
    if (cp < tp - 2) pages.push('...')
    pages.push(tp)
  }
  return pages
})

// 状态统计（从接口获取）
const statusStats = ref([])
const statusCountMap = computed(() => {
  const map = {}
  statusStats.value.forEach(s => { map[s.alert_status] = s.count })
  return map
})

function selectAlert(index) {
  selectedIndex.value = index
}

function goPage(page) {
  if (page < 1 || page > totalPages.value || page === currentPage.value) return
  currentPage.value = page
  selectedIndex.value = -1
  fetchAlerts()
}

const filterStatuses = ref(['告警中', '处置中']) // 默认选中

function toggleFilter(status) {
  if (status === '') {
    // 点击"全部"：清空筛选
    filterStatuses.value = []
    currentPage.value = 1
    selectedIndex.value = -1
    fetchAlerts()
    return
  }
  const idx = filterStatuses.value.indexOf(status)
  if (idx >= 0) {
    filterStatuses.value.splice(idx, 1)
  } else {
    filterStatuses.value.push(status)
  }
  // 如果所有状态都被取消，视为全选
  currentPage.value = 1
  selectedIndex.value = -1
  fetchAlerts()
}

function isActive(status) {
  if (status === '') return filterStatuses.value.length === 0
  return filterStatuses.value.includes(status)
}

function formatTimeStr(timeStr) {
  if (!timeStr) return ''
  const d = new Date(timeStr)
  return d.toLocaleDateString('zh-CN', { month: '2-digit', day: '2-digit' }) + ' ' +
    d.toLocaleTimeString('zh-CN', { hour12: false })
}

function formatFullTime(timeStr) {
  if (!timeStr) return '-'
  const d = new Date(timeStr)
  return d.toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }) + ' ' +
    d.toLocaleTimeString('zh-CN', { hour12: false })
}

function calcDuration(timeStr) {
  if (!timeStr) return ''
  const diff = Date.now() - new Date(timeStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 60) return mins + 'm'
  const hours = Math.floor(mins / 60)
  const remain = mins % 60
  return hours + 'h' + (remain > 0 ? remain + 'm' : '')
}

async function fetchAlerts() {
  loading.value = true
  try {
    const params = new URLSearchParams({ page: currentPage.value, pageSize })
    if (filterStatuses.value.length > 0) params.set('statuses', filterStatuses.value.join(','))
    const res = await apiFetch(`/api/alerts?${params}`)
    const data = await res.json()
    total.value = data.total || 0
    rawAlerts.value = data.list || []
  } catch (e) {
    console.error('获取告警数据失败:', e)
  } finally {
    loading.value = false
  }
}

async function fetchStats() {
  try {
    const res = await apiFetch('/api/alerts/stats')
    statusStats.value = await res.json()
  } catch {}
}

async function updateStatus(id, newStatus) {
  if (newStatus === '已处置') {
    if (!confirm('确认完成处置？')) return
  }
  try {
    await apiFetch('/api/alerts/update-status', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id, status: newStatus })
    })
    await fetchAlerts()
    await fetchStats()
  } catch (e) {
    console.error('更新状态失败:', e)
  }
}

const totalAll = computed(() => statusStats.value.reduce((s, i) => s + i.count, 0))

onMounted(() => {
  fetchAlerts()
  fetchStats()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">告警管理</h2>
      <div class="page-controls">
        <button
          class="ctrl-btn" :class="{ active: isActive('') }"
          @click="toggleFilter('')"
        >全部 ({{ totalAll }})</button>
        <button
          v-for="s in statusStats" :key="s.alert_status"
          class="ctrl-btn" :class="{ active: isActive(s.alert_status) }"
          @click="toggleFilter(s.alert_status)"
        >{{ s.alert_status }} ({{ s.count }})</button>
      </div>
    </div>
    <div class="alerts-layout">
      <div class="alerts-list">
        <div v-if="loading" class="detail-placeholder" style="padding:40px;text-align:center;">加载中...</div>
        <div v-else-if="rawAlerts.length === 0" class="detail-placeholder" style="padding:40px;text-align:center;">暂无告警数据</div>
        <div
          v-for="(a, i) in rawAlerts" :key="a.id"
          class="alert-card" :class="{ selected: selectedIndex === i }"
          @click="selectAlert(i)"
        >
          <div class="ac-level">
            <div class="ac-status-dot" :class="a.alert_status === '告警中' ? 'err' : a.alert_status === '处置中' ? 'warn' : 'ok'"></div>
          </div>
          <div class="ac-body">
            <div class="ac-header">
              <span class="ac-badge" :class="a.alert_status === '告警中' ? 'err' : a.alert_status === '处置中' ? 'warn' : 'ok'">{{ a.alert_status }}</span>
              <span class="ac-idc-tag">{{ a.alert_idc }}</span>
              <span class="ac-type-tag">{{ a.alert_type }}</span>
            </div>
            <div class="ac-desc">
              <span class="ac-device">{{ a.alert_leaf || a.alert_spine || '-' }}</span>
              <span v-if="a.alert_rack" class="ac-rack">{{ a.alert_rack }}</span>
            </div>
            <div class="ac-meta">
              <span class="ac-time">{{ formatTimeStr(a.create_time) }}</span>
              <span class="ac-duration">持续 {{ calcDuration(a.create_time) }}</span>
              <span v-if="a.create_by" class="ac-by">{{ a.create_by }}</span>
            </div>
          </div>
          <div class="ac-actions">
            <button v-if="a.alert_status === '告警中'" class="ac-btn" @click.stop="updateStatus(a.id, '处置中')">开始处理</button>
            <button v-if="a.alert_status === '处置中'" class="ac-btn" @click.stop="updateStatus(a.id, '已处置')">完成处置</button>
          </div>
        </div>
        <div v-if="!loading && rawAlerts.length > 0" class="pagination">
          <button class="ac-btn" :disabled="currentPage <= 1" @click="goPage(currentPage - 1)">上一页</button>
          <template v-for="p in pageNumbers" :key="p">
            <span v-if="p === '...'" class="page-info" style="padding:0 2px;">...</span>
            <button v-else class="page-num-btn" :class="{ cur: p === currentPage }" @click="goPage(p)">{{ p }}</button>
          </template>
          <button class="ac-btn" :disabled="currentPage >= totalPages" @click="goPage(currentPage + 1)">下一页</button>
        </div>
      </div>

      <div class="alerts-detail">
        <template v-if="selectedAlert">
          <div class="detail-header">
            <div class="dh-status" :class="selectedAlert.alert_status === '告警中' ? 'err' : selectedAlert.alert_status === '处置中' ? 'warn' : 'ok'">
              <span class="dh-dot"></span>{{ selectedAlert.alert_status }}
            </div>
            <div class="dh-id">#{{ selectedAlert.id }}</div>
          </div>

          <div class="detail-section">
            <div class="ds-title">告警信息</div>
            <div class="ds-row"><span class="ds-label">告警类型</span><span class="ds-val ds-highlight">{{ selectedAlert.alert_type }}</span></div>
            <div class="ds-row"><span class="ds-label">故障机房</span><span class="ds-val">{{ selectedAlert.alert_idc }}</span></div>
            <div class="ds-row"><span class="ds-label">影响设备</span><span class="ds-val">{{ selectedAlert.alert_leaf || selectedAlert.alert_spine || '-' }}</span></div>
            <div class="ds-row"><span class="ds-label">机柜位置</span><span class="ds-val">{{ selectedAlert.alert_rack || '-' }}</span></div>
          </div>

          <div class="detail-section">
            <div class="ds-title">设备拓扑</div>
            <div class="ds-topo-flow">
              <span class="ds-topo-item">{{ selectedAlert.alert_spine || 'Spine' }}</span>
              <span class="ds-topo-arrow">→</span>
              <span class="ds-topo-item">{{ selectedAlert.alert_leaf || 'Leaf' }}</span>
              <span v-if="selectedAlert.alert_rack" class="ds-topo-arrow">→</span>
              <span v-if="selectedAlert.alert_rack" class="ds-topo-item">{{ selectedAlert.alert_rack }}</span>
            </div>
          </div>

          <div class="detail-section">
            <div class="ds-title">处理记录</div>
            <div class="ds-row"><span class="ds-label">创建时间</span><span class="ds-val">{{ formatFullTime(selectedAlert.create_time) }}</span></div>
            <div class="ds-row"><span class="ds-label">创建人</span><span class="ds-val">{{ selectedAlert.create_by || '-' }}</span></div>
            <div class="ds-row"><span class="ds-label">更新时间</span><span class="ds-val">{{ formatFullTime(selectedAlert.update_time) }}</span></div>
            <div class="ds-row"><span class="ds-label">更新人</span><span class="ds-val">{{ selectedAlert.update_by || '-' }}</span></div>
            <div class="ds-row"><span class="ds-label">持续时长</span><span class="ds-val warn">{{ calcDuration(selectedAlert.create_time) }}</span></div>
          </div>

          <div class="detail-section">
            <div class="ds-title">操作</div>
            <div class="ds-actions">
              <button v-if="selectedAlert.alert_status === '告警中'" class="ac-btn ds-action-btn err" @click="updateStatus(selectedAlert.id, '处置中')">开始处理</button>
              <button v-if="selectedAlert.alert_status === '处置中'" class="ac-btn ds-action-btn warn" @click="updateStatus(selectedAlert.id, '已处置')">完成处置</button>
            </div>
          </div>
        </template>
        <div v-else class="detail-placeholder">&larr; 点击告警查看详情</div>
      </div>
    </div>
  </div>
</template>
