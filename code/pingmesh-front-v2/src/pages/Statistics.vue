<script setup>
import { ref, computed, onMounted } from 'vue'
import { apiFetch } from '../utils/api.js'

const metrics = ref([])
const loading = ref(false)
const searchText = ref('')

// ── 浮出查询面板 ──
const showQueryModal = ref(false)
const queryMetricName = ref('')
const queryLoading = ref(false)
const queryError = ref('')
const queryResultType = ref('')
const queryResults = ref([])

async function doQuery(metricName) {
  queryMetricName.value = metricName
  queryLoading.value = true
  queryError.value = ''
  queryResults.value = []
  queryResultType.value = ''
  showQueryModal.value = true
  try {
    const res = await apiFetch('/api/env-config/prometheus/query', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ metric: metricName })
    })
    const data = await res.json()
    if (data.status === 'success') {
      queryResultType.value = data.data?.resultType || ''
      queryResults.value = data.data?.result || []
      if (queryResults.value.length === 0) {
        queryError.value = '查询成功，但无数据'
      }
    } else {
      queryError.value = data.error || '查询失败'
    }
  } catch (e) {
    queryError.value = '请求异常: ' + (e.message || e)
  } finally {
    queryLoading.value = false
  }
}

function closeQueryModal() {
  showQueryModal.value = false
}

function formatVal(v) {
  if (v == null) return '-'
  const n = Number(v)
  if (!isNaN(n) && v !== '') {
    if (n < 0.01) return n.toExponential(3)
    return n.toFixed(4)
  }
  return String(v)
}

function formatTs(ts) {
  if (!ts) return '-'
  return new Date(Number(ts) * 1000).toLocaleString('zh-CN')
}

// 四维筛选状态（单选，'' 表示全部）
const filterType = ref('')
const filterScope = ref('')
const filterCategory = ref('')
const filterRegion = ref('')

// 筛选选项
const typeOptions = ['', 'latency', 'loss']
const typeLabels = { '': '全部', latency: '时延', loss: '丢包' }
const scopeOptions = ['', 'idc', 'spine', 'leaf']
const scopeLabels = { '': '全部', idc: 'Idc', spine: 'Spine', leaf: 'Leaf' }
const categoryOptions = ['', 'raw', 'max', 'avg', 'p99']
const categoryLabels = { '': '全部', raw: '原始值', max: '最大', avg: '平均', p99: 'P99' }
const regionOptions = ['', 'global', 'inner', 'inter']
const regionLabels = { '': '全部', global: '全局', inner: '机房内', inter: '机房间' }

// 过滤后的指标列表（只展示 status=show 的指标）
const filteredMetrics = computed(() => {
  return metrics.value.filter(m => {
    if (m.status !== 'show') return false
    if (searchText.value && !m.metric.toLowerCase().includes(searchText.value.toLowerCase())) return false
    if (filterType.value && m.type !== filterType.value) return false
    if (filterScope.value && m.scope !== filterScope.value) return false
    if (filterCategory.value && m.category !== filterCategory.value) return false
    if (filterRegion.value) {
      const s = m.strategy
      if (filterRegion.value === 'global' && s !== '00') return false
      if (filterRegion.value === 'inner' && s !== '01') return false
      if (filterRegion.value === 'inter' && !['23','02','03'].includes(s)) return false
    }
    return true
  })
})

// 类型标签样式
function getTypeClass(type) {
  return type === 'latency' ? 'type-latency' : type === 'loss' ? 'type-loss' : ''
}

function getScopeClass(scope) {
  return scope === 'idc' ? 'scope-idc' : scope === 'spine' ? 'scope-spine' : scope === 'leaf' ? 'scope-leaf' : ''
}

function getCategoryClass(cat) {
  return cat === 'raw' ? 'cat-raw' : cat === 'max' ? 'cat-max' : cat === 'avg' ? 'cat-avg' : cat === 'p99' ? 'cat-p99' : ''
}

function getTypeLabel(type) {
  return typeLabels[type] || '-'
}

function getScopeLabel(scope) {
  return scopeLabels[scope] || '-'
}

function getCategoryLabel(cat) {
  return categoryLabels[cat] || '-'
}

// 加载指标列表
async function fetchMetrics() {
  loading.value = true
  try {
    const res = await apiFetch('/api/statistics/metrics')
    metrics.value = await res.json()
  } catch (e) {
    console.error('加载指标列表失败:', e)
  } finally {
    loading.value = false
  }
}

// 点击指标名称，浮出查询面板
function openMetric(metricName) {
  doQuery(metricName)
}

onMounted(() => {
  fetchMetrics()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">统计数据</h2>
    </div>

    <!-- 四维筛选区 -->
    <div class="stats-filters">
      <div class="filter-row">
        <span class="filter-label">区域</span>
        <div class="filter-btns">
          <button
            v-for="opt in regionOptions" :key="opt"
            class="ctrl-btn" :class="{ active: filterRegion === opt }"
            @click="filterRegion = opt"
          >{{ regionLabels[opt] }}</button>
        </div>
      </div>
      <div class="filter-row">
        <span class="filter-label">指标类型</span>
        <div class="filter-btns">
          <button
            v-for="opt in typeOptions" :key="opt"
            class="ctrl-btn" :class="{ active: filterType === opt }"
            @click="filterType = opt"
          >{{ typeLabels[opt] }}</button>
        </div>
      </div>
      <div class="filter-row">
        <span class="filter-label">数据范围</span>
        <div class="filter-btns">
          <button
            v-for="opt in scopeOptions" :key="opt"
            class="ctrl-btn" :class="{ active: filterScope === opt }"
            @click="filterScope = opt"
          >{{ scopeLabels[opt] }}</button>
        </div>
      </div>
      <div class="filter-row">
        <span class="filter-label">统计类别</span>
        <div class="filter-btns">
          <button
            v-for="opt in categoryOptions" :key="opt"
            class="ctrl-btn" :class="{ active: filterCategory === opt }"
            @click="filterCategory = opt"
          >{{ categoryLabels[opt] }}</button>
        </div>
      </div>
      <div class="filter-row">
        <span class="filter-label">关键字</span>
        <input
          class="ctrl-input"
          v-model="searchText"
          placeholder="搜索指标名称..."
        />
        <span class="stats-count">共 {{ filteredMetrics.length }} 条</span>
      </div>
    </div>

    <!-- 指标表格 -->
    <div class="devices-table-wrap">
      <div v-if="loading" class="detail-placeholder" style="padding:40px;text-align:center;">加载中...</div>
      <div v-else-if="filteredMetrics.length === 0" class="detail-placeholder" style="padding:40px;text-align:center;">暂无指标数据</div>
      <table v-else class="devices-table">
        <thead>
          <tr>
            <th>指标名称</th>
            <th>指标类型</th>
            <th>数据范围</th>
            <th>统计类别</th>
            <th>描述</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in filteredMetrics" :key="item.metric"
            @click="openMetric(item.metric)"
          >
            <td><span class="metric-name">{{ item.metric }}</span></td>
            <td>
              <span v-if="item.type" class="status-badge" :class="getTypeClass(item.type)">
                {{ getTypeLabel(item.type) }}
              </span>
              <span v-else class="text-dim">-</span>
            </td>
            <td>
              <span v-if="item.scope" class="status-badge" :class="getScopeClass(item.scope)">
                {{ getScopeLabel(item.scope) }}
              </span>
              <span v-else class="text-dim">-</span>
            </td>
            <td>
              <span v-if="item.category" class="status-badge" :class="getCategoryClass(item.category)">
                {{ getCategoryLabel(item.category) }}
              </span>
              <span v-else class="text-dim">-</span>
            </td>
            <td class="text-dim">{{ item.description || '-' }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>

  <!-- 浮出查询面板 -->
  <Teleport to="body">
    <div v-if="showQueryModal" class="modal-overlay" @click.self="closeQueryModal">
      <div class="modal-panel">
        <div class="modal-header">
          <h3>指标查询 - {{ queryMetricName }}</h3>
          <button class="ctrl-btn" @click="closeQueryModal">✕</button>
        </div>
        <div class="modal-body">
          <div v-if="queryLoading" class="detail-placeholder" style="padding:30px;">查询中...</div>
          <div v-else-if="queryError" class="query-error">{{ queryError }}</div>
          <div v-else-if="queryResults.length > 0">
            <div class="results-info">
              <span>类型: <strong>{{ queryResultType }}</strong></span>
              <span>数据条数: <strong>{{ queryResults.length }}</strong></span>
            </div>
            <div class="results-table-wrap">
              <table class="results-table">
                <thead>
                  <tr>
                    <th v-for="label in Object.keys(queryResults[0].metric).filter(k => k !== '__name__')" :key="label">{{ label }}</th>
                    <th>值</th>
                    <th>时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(row, idx) in queryResults" :key="idx">
                    <td v-for="label in Object.keys(queryResults[0].metric).filter(k => k !== '__name__')" :key="label">
                      {{ row.metric[label] }}
                    </td>
                    <td class="val-cell">{{ formatVal(row.value?.[1]) }}</td>
                    <td class="time-cell">{{ formatTs(row.value?.[0]) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
/* 三维筛选区 */
.stats-filters {
  padding: 12px 20px;
  border-bottom: 1px solid var(--bg-border2);
  display: flex;
  flex-direction: column;
  gap: 8px;
  flex-shrink: 0;
}

.filter-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.filter-label {
  font-size: 10px;
  color: var(--text-dim);
  letter-spacing: 1px;
  text-transform: uppercase;
  min-width: 60px;
  font-family: var(--font-head);
}

.filter-btns {
  display: flex;
  gap: 4px;
}

.stats-count {
  font-size: 12px;
  color: var(--text-dim);
  margin-left: 8px;
}

/* 指标名称 */
.metric-name {
  font-family: var(--font-mono);
  color: var(--cyan);
  cursor: pointer;
  transition: opacity 0.15s;
}
.metric-name:hover {
  opacity: 0.8;
}

.text-dim {
  color: var(--text-dim);
}

/* 表格行 hover */
.devices-table tbody tr {
  cursor: pointer;
  transition: background 0.15s;
}
.devices-table tbody tr:hover {
  background: var(--bg-hover);
}

/* 状态标签颜色 */
.status-badge.type-latency { background: rgba(57,255,126,0.1); color: var(--green); }
.status-badge.type-loss    { background: rgba(255,69,96,0.15); color: var(--err); }

.status-badge.scope-idc   { background: rgba(0,245,196,0.08); color: var(--cyan); }
.status-badge.scope-spine { background: rgba(0,245,196,0.08); color: var(--cyan); }
.status-badge.scope-leaf  { background: rgba(0,245,196,0.08); color: var(--cyan); }

.status-badge.cat-raw { background: rgba(255,194,51,0.15); color: var(--warn); }
.status-badge.cat-max { background: rgba(255,194,51,0.15); color: var(--warn); }
.status-badge.cat-avg { background: rgba(0,245,196,0.08); color: var(--cyan); }
.status-badge.cat-p99 { background: rgba(206,147,216,0.15); color: #ce93d8; }

/* ── 浮出查询面板 ── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-panel {
  background: var(--bg-deep);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  width: 90vw;
  max-width: 900px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 32px rgba(0,0,0,0.4);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 1px solid var(--bg-border);
}

.modal-header h3 {
  font-size: 16px;
  color: var(--cyan);
  font-family: var(--font-mono);
  margin: 0;
}

.modal-body {
  padding: 16px 20px;
  overflow-y: auto;
  flex: 1;
}

.modal-body .query-error {
  color: var(--err);
  font-size: 13px;
  padding: 12px;
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
}

.modal-body .results-info {
  display: flex;
  gap: 24px;
  font-size: 12px;
  color: var(--text-dim);
  margin-bottom: 14px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bg-border);
}

.modal-body .results-info strong {
  color: var(--cyan);
  font-family: var(--font-mono);
}

.modal-body .results-table-wrap {
  overflow-x: auto;
}

.modal-body .results-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.modal-body .results-table th {
  text-align: left;
  padding: 6px 12px;
  font-size: 10px;
  font-weight: 600;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 1px;
  border-bottom: 1px solid var(--bg-border);
  white-space: nowrap;
}

.modal-body .results-table td {
  padding: 6px 12px;
  color: var(--text);
  font-family: var(--font-mono);
  border-bottom: 1px solid var(--bg-border2);
  white-space: nowrap;
}

.modal-body .results-table tr:last-child td {
  border-bottom: none;
}

.modal-body .val-cell {
  color: var(--cyan) !important;
  font-weight: 600;
}

.modal-body .time-cell {
  color: var(--text-dim) !important;
  font-size: 11px;
}
</style>
