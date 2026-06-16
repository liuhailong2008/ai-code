<script setup>
import { ref, onMounted, computed, reactive } from 'vue'
import { apiFetch } from '../utils/api.js'

const metricMetaList = ref([])
const metricMetaLoading = ref(false)
const metricMetaError = ref('')
const metricMetaSaving = ref(false)
const metricMetaSaved = ref(false)

// ── 指标 → 页面/功能 映射 ──
const metricPageMap = {
  'ping_idc_lost':             { page: '监控大屏', feature: 'IDC 丢包率 → 拓扑图节点状态' },
  'ping_idc_link_status':      { page: '监控大屏', feature: '机房间链路状态' },
  'ping_idc_latency_p99':      { page: '监控大屏', feature: 'IDC P99 时延 → 面板状态判定' },
  'ping_idc_latency_avg':      { page: '监控大屏', feature: 'IDC 平均时延 → 底栏汇总' },
  'ping_latency_by_idc_max':   { page: '监控大屏', feature: '机房内最大时延 → 面板数据' },
  'ping_latency_by_idc_avg':   { page: '监控大屏', feature: '机房内平均时延 → 面板 + 树节点' },
  'ping_latency_by_idc_p99':   { page: '监控大屏', feature: '机房内 P99 时延 → 面板数据' },
  'ping_lost_by_idc_max':      { page: '监控大屏', feature: '机房内丢包率 → 面板数据' },
  'ping_latency_by_spine_avg': { page: '监控大屏', feature: 'Spine 节点时延 → 树节点状态' },
  'ping_latency_by_leaf_avg':  { page: '监控大屏 / 机房耗时', feature: 'Leaf 时延 → 树节点 / 散点图 X 轴' },
  'ping_latency_by_leaf_max':  { page: '机房耗时', feature: 'Leaf 最大时延 → 散点图 X 轴' },
  'ping_latency_by_leaf_p99':  { page: '机房耗时', feature: 'Leaf P99 时延 → 散点图 X 轴' },
  'ping_lost_by_leaf':         { page: '机房耗时', feature: 'Leaf 丢包率 → 散点图 Y 轴' },
  'ping_latency_overall_avg':  { page: '监控大屏', feature: '全局平均时延' },
  'ping_latency_overall_p99':  { page: '监控大屏', feature: '全局 P99 时延' },
  'ping_lost_overall':         { page: '监控大屏', feature: '全局丢包率' },
  'ping_leaf_avg':             { page: '机房网络监控', feature: 'Leaf 间平均时延 → 热力图/TOP 链路' },
  'ping_leaf_p99':             { page: '机房网络监控', feature: 'Leaf 间 P99 时延 → 热力图/TOP 链路' },
  'ping_leaf_max':             { page: '机房网络监控', feature: 'Leaf 间最大时延 → 热力图/TOP 链路' },
  'ping_leaf_max_all':         { page: '机房网络监控', feature: '所有机房 Leaf 间最大时延' },
}

function getPageInfo(metric) {
  return metricPageMap[metric] || { page: '-', feature: '-' }
}

// ── 测试状态 ──
const metricTestMap = reactive({})

function getTestState(metric) {
  return metricTestMap[metric] || { testing: false, count: null, error: null }
}

async function testMetric(metric) {
  const state = metricTestMap[metric] || { testing: false, count: null, error: null }
  state.testing = true
  state.count = null
  state.error = null
  metricTestMap[metric] = state
  try {
    const res = await apiFetch('/api/env-config/prometheus/metric-test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ metric })
    })
    const data = await res.json()
    if (data.status === 'ok') {
      state.count = data.count
    } else {
      state.error = data.error || '查询失败'
    }
  } catch (e) {
    state.error = '请求异常: ' + (e.message || e)
  } finally {
    state.testing = false
    metricTestMap[metric] = state
  }
}

async function testAllMetrics() {
  for (const item of filteredMetricList.value) {
    if (item.status === 'disable') continue
    await testMetric(item.metric)
  }
}

const typeOptions = [
  { value: 'latency', label: '时延' },
  { value: 'loss', label: '丢包' },
]
const scopeOptions = [
  { value: 'idc', label: 'Idc' },
  { value: 'spine', label: 'Spine' },
  { value: 'leaf', label: 'Leaf' },
]
const categoryOptions = [
  { value: 'raw', label: '原始值' },
  { value: 'max', label: '最大' },
  { value: 'avg', label: '平均' },
  { value: 'p99', label: 'P99' },
]

const filterType = ref('')
const filterScope = ref('')
const filterCategory = ref('')
const filterStatus = ref('')
const filterPage = ref('')
const searchText = ref('')

const typeLabels = { '': '全部', latency: '时延', loss: '丢包' }
const scopeLabels = { '': '全部', idc: 'Idc', spine: 'Spine', leaf: 'Leaf' }
const categoryLabels = { '': '全部', raw: '原始值', max: '最大', avg: '平均', p99: 'P99' }
const statusLabels = { '': '全部', 'new': '新发现', show: '展示', hide: '隐藏', disable: '停用' }

// 从 metricPageMap 提取所有不重复的页面值
const pageFilterKeys = [...new Set(Object.values(metricPageMap).map(v => v.page))]
const pageLabels = Object.fromEntries([['', '全部'], ...pageFilterKeys.map(k => [k, k])])

const statusOptions = [
  { value: 'new', label: '新发现' },
  { value: 'show', label: '展示' },
  { value: 'hide', label: '隐藏' },
  { value: 'disable', label: '停用' },
]

const filteredMetricList = computed(() => {
  let list = metricMetaList.value.filter(m => {
    if (searchText.value && !m.metric.toLowerCase().includes(searchText.value.toLowerCase())) return false
    if (filterType.value && m.type !== filterType.value) return false
    if (filterScope.value && m.scope !== filterScope.value) return false
    if (filterCategory.value && m.category !== filterCategory.value) return false
    if (filterStatus.value && m.status !== filterStatus.value) return false
    if (filterPage.value && getPageInfo(m.metric).page !== filterPage.value) return false
    return true
  })
  list.sort((a, b) => a.metric.localeCompare(b.metric))
  return list
})

async function loadMetricMeta() {
  metricMetaLoading.value = true
  metricMetaError.value = ''
  try {
    const res = await apiFetch('/api/statistics/metrics')
    const data = await res.json()
    if (Array.isArray(data)) {
      metricMetaList.value = data
      if (data.length === 0) {
        metricMetaError.value = '未获取到任何指标（请确认 Prometheus 已启动且包含 ping_ 前缀的指标）'
      }
    } else {
      metricMetaError.value = '接口返回格式异常：' + JSON.stringify(data)
    }
  } catch (e) {
    metricMetaError.value = '加载失败：' + (e.message || e)
  } finally {
    metricMetaLoading.value = false
  }
}

function onSaveClick() {
  if (!confirm('确认要保存指标配置到 config.yaml？')) return
  saveMetricMeta()
}

async function saveMetricMeta() {
  metricMetaSaving.value = true
  metricMetaSaved.value = false
  try {
    const res = await apiFetch('/api/statistics/metric-meta', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(metricMetaList.value)
    })
    const data = await res.json()
    if (data.status === 'ok') {
      metricMetaSaved.value = true
      setTimeout(() => { metricMetaSaved.value = false }, 3000)
    }
  } catch (e) {
    console.error('保存指标配置失败:', e)
  } finally {
    metricMetaSaving.value = false
  }
}

onMounted(() => {
  loadMetricMeta()
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">指标管理</h2>
    </div>
    <div class="settings-section">
      <div class="ts-group">
        <!-- 筛选区 + 操作按钮 -->
        <div class="stats-topbar">
          <div class="stats-filters">
            <div class="filter-row">
              <span class="filter-label">指标类型</span>
              <div class="filter-btns">
                <button v-for="key in ['', 'latency', 'loss']" :key="key" class="ctrl-btn" :class="{ active: filterType === key }" @click="filterType = key">{{ typeLabels[key] }}</button>
              </div>
            </div>
            <div class="filter-row">
              <span class="filter-label">数据范围</span>
              <div class="filter-btns">
                <button v-for="key in ['', 'idc', 'spine', 'leaf']" :key="key" class="ctrl-btn" :class="{ active: filterScope === key }" @click="filterScope = key">{{ scopeLabels[key] }}</button>
              </div>
            </div>
            <div class="filter-row">
              <span class="filter-label">统计类别</span>
              <div class="filter-btns">
                <button v-for="key in ['', 'raw', 'max', 'avg', 'p99']" :key="key" class="ctrl-btn" :class="{ active: filterCategory === key }" @click="filterCategory = key">{{ categoryLabels[key] }}</button>
              </div>
            </div>
            <div class="filter-row">
              <span class="filter-label">状态</span>
              <div class="filter-btns">
                <button v-for="key in ['', 'new', 'show', 'hide', 'disable']" :key="key" class="ctrl-btn" :class="{ active: filterStatus === key }" @click="filterStatus = key">{{ statusLabels[key] }}</button>
              </div>
            </div>
            <div class="filter-row">
              <span class="filter-label">相关页面</span>
              <div class="filter-btns">
                <button v-for="key in ['', ...pageFilterKeys]" :key="key" class="ctrl-btn" :class="{ active: filterPage === key }" @click="filterPage = key">{{ pageLabels[key] || key }}</button>
              </div>
            </div>
          </div>
          <div class="stats-actions">
            <button class="ctrl-btn" :disabled="metricMetaLoading" @click="loadMetricMeta">
              {{ metricMetaLoading ? '加载中...' : '刷新' }}
            </button>
            <button class="ctrl-btn accent-btn" :disabled="metricMetaSaving" @click="onSaveClick">
              {{ metricMetaSaving ? '保存中...' : '保存配置' }}
            </button>
            <button class="ctrl-btn" @click="testAllMetrics">
              指标测试
            </button>
            <span v-if="metricMetaSaved" class="ss-saved">保存成功</span>
          </div>
        </div>
        <!-- 搜索栏 -->
        <div class="stats-search">
          <span class="filter-label">关键字</span>
          <input class="ctrl-input" v-model="searchText" placeholder="按指标名称搜索..." style="width:280px;" />
          <span class="stats-count">共 {{ filteredMetricList.length }} 条</span>
        </div>
        <!-- 加载中 -->
        <div v-if="metricMetaLoading" class="detail-placeholder" style="padding:30px;">加载中...</div>
        <!-- 错误信息 -->
        <div v-else-if="metricMetaError" class="detail-placeholder" style="padding:30px;color:var(--err);">{{ metricMetaError }}</div>
        <!-- 指标表格 -->
        <div v-else class="idc-table">
          <div class="idc-thead">
            <span class="idc-th" style="width:100px;">测试</span>
            <span class="idc-th" style="width:55px;">状态</span>
            <span class="idc-th code">指标名称</span>
            <span class="idc-th page-col">相关页面</span>
            <span class="idc-th feature-col">页面功能</span>
            <span class="idc-th" style="width:70px;">指标类型</span>
            <span class="idc-th" style="width:70px;">数据范围</span>
            <span class="idc-th" style="width:70px;">统计类别</span>
            <span class="idc-th" style="width:55px;">策略</span>
            <span class="idc-th name">描述</span>
          </div>
          <div
            v-for="(item, idx) in filteredMetricList"
            :key="idx"
            class="idc-tr"
          >
            <span class="idc-td" style="width:100px;">
              <button class="ctrl-btn test-btn" :disabled="getTestState(item.metric).testing" @click="testMetric(item.metric)">
                {{ getTestState(item.metric).testing ? '...' : '测试' }}
              </button>
              <span v-if="getTestState(item.metric).count !== null" class="test-count" :class="getTestState(item.metric).count === 0 ? 'err' : 'ok'">{{ getTestState(item.metric).count }}条</span>
              <span v-if="getTestState(item.metric).error" class="test-count err">{{ getTestState(item.metric).error }}</span>
            </span>
            <span class="idc-td" style="width:55px;">
              <select class="ctrl-select status-select" :class="'status-' + (item.status || 'new')" v-model="item.status" style="width:52px;">
                <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </span>
            <span class="idc-td code" style="font-size:11px;">{{ item.metric }}</span>
            <span class="idc-td page-col" style="font-size:11px;">{{ getPageInfo(item.metric).page }}</span>
            <span class="idc-td feature-col" style="font-size:11px;white-space:normal;word-break:break-word;">{{ getPageInfo(item.metric).feature }}</span>
            <span class="idc-td" style="width:70px;">
              <select class="ctrl-select" v-model="item.type" style="width:60px;">
                <option value="">-</option>
                <option v-for="opt in typeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </span>
            <span class="idc-td" style="width:70px;">
              <select class="ctrl-select" v-model="item.scope" style="width:60px;">
                <option value="">-</option>
                <option v-for="opt in scopeOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </span>
            <span class="idc-td" style="width:70px;">
              <select class="ctrl-select" v-model="item.category" style="width:60px;">
                <option value="">-</option>
                <option v-for="opt in categoryOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </span>
            <span class="idc-td" style="width:55px;">
              <select class="ctrl-select" v-model="item.strategy" style="width:48px;">
                <option value="">-</option>
                <option value="00">00</option>
                <option value="01">01</option>
                <option value="23">23</option>
                <option value="02">02</option>
                <option value="03">03</option>
              </select>
            </span>
            <span class="idc-td name">
              <input class="idc-input" v-model="item.description" type="text" placeholder="输入描述..." />
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 指标筛选区 */
.stats-topbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding-bottom: 8px;
  gap: 12px;
}
.stats-search {
  display: flex;
  align-items: center;
  gap: 10px;
  padding-bottom: 12px;
}
.stats-count {
  font-size: 11px;
  color: var(--text-dim);
  font-family: var(--font-mono);
}

/* 状态下拉颜色 */
.status-select.status-new     { color: var(--text-dim); border-color: var(--bg-border); }
.status-select.status-show    { color: var(--cyan);     border-color: var(--cyan-dim); }
.status-select.status-hide    { color: var(--warn);     border-color: rgba(255,171,0,0.3); }
.status-select.status-disable { color: var(--err);      border-color: rgba(255,69,96,0.3); }
.stats-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  padding-top: 2px;
}
.stats-filters {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}
.filter-row {
  display: flex;
  align-items: center;
  gap: 10px;
}
.filter-label {
  font-size: 10px;
  color: var(--text-dim);
  letter-spacing: 1px;
  text-transform: uppercase;
  min-width: 56px;
  font-family: var(--font-head);
}
.filter-btns {
  display: flex;
  gap: 4px;
}
.filter-btns .ctrl-btn.active {
  border-color: var(--cyan);
  color: var(--cyan);
  background: var(--cyan-ghost);
}

/* 指标表格 */
.idc-table {
  width: 100%;
}
.idc-thead {
  display: flex;
  padding: 0 0 10px 0;
  border-bottom: 1px solid var(--bg-border);
  margin-bottom: 4px;
}
.idc-th {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-dim);
  text-transform: uppercase;
  letter-spacing: 1px;
}
.idc-th.code { width: 250px; }
.idc-th.page-col { width: 130px; }
.idc-th.feature-col { flex: 1; min-width: 140px; }
.idc-th.name { flex: 1; }

.idc-tr {
  display: flex;
  align-items: center;
  padding: 6px 0;
  border-bottom: 1px solid var(--bg-border2);
}
.idc-tr:last-child {
  border-bottom: none;
}
.idc-td {
  font-size: 12px;
  font-family: var(--font-mono);
  color: var(--text-dim);
}
.idc-td.code {
  width: 250px;
  color: var(--text);
  font-weight: 600;
}
.idc-td.page-col {
  width: 130px;
}
.idc-td.feature-col {
  flex: 1;
  min-width: 140px;
}
.idc-td.name {
  flex: 1;
}
.idc-input {
  background: var(--bg-deep);
  border: 1px solid var(--bg-border);
  color: var(--text);
  font-family: var(--font-mono);
  font-size: 12px;
  padding: 4px 8px;
  border-radius: 4px;
  width: 100%;
  max-width: 220px;
  outline: none;
  transition: border-color 0.15s;
}
.idc-input:focus {
  border-color: var(--cyan-dim);
}

.ss-saved {
  font-size: 12px;
  color: var(--green);
}

.ts-group {
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  padding: 14px 20px;
  margin-bottom: 12px;
}
.settings-section {
  width: 100%;
}

.test-btn {
  font-size: 10px;
  padding: 2px 8px;
  min-width: 40px;
}

.test-count {
  font-size: 10px;
  font-family: var(--font-mono);
  margin-left: 4px;
}

.test-count.ok {
  color: var(--green);
}

.test-count.err {
  color: var(--err);
}
</style>
