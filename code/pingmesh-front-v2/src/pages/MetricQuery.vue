<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch } from '../utils/api.js'

const route = useRoute()
const router = useRouter()

const metricName = ref(route.query.metric || '')
const loading = ref(false)
const error = ref('')
const resultType = ref('')
const results = ref([])

async function queryMetric() {
  if (!metricName.value.trim()) {
    error.value = '请输入指标名称'
    return
  }
  loading.value = true
  error.value = ''
  results.value = []
  resultType.value = ''
  try {
    const res = await apiFetch('/api/env-config/prometheus/query', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ metric: metricName.value.trim() })
    })
    const data = await res.json()
    if (data.status === 'success') {
      resultType.value = data.data?.resultType || ''
      results.value = data.data?.result || []
      if (results.value.length === 0) {
        error.value = '查询成功，但无数据'
      }
    } else if (data.status === 'error') {
      error.value = data.error || 'Prometheus 查询失败'
    } else {
      // 可能是我们的错误格式
      error.value = data.error || JSON.stringify(data)
    }
  } catch (e) {
    error.value = '请求异常: ' + (e.message || e)
  } finally {
    loading.value = false
  }
}

function goBack() {
  router.back()
}

onMounted(() => {
  if (metricName.value) {
    queryMetric()
  }
})

// 格式化数值
function formatValue(v) {
  if (v == null) return '-'
  const n = Number(v)
  if (!isNaN(n) && v !== '') {
    if (n < 0.01) return n.toExponential(3)
    return n.toFixed(4)
  }
  return String(v)
}

// 格式化时间戳
function formatTimestamp(ts) {
  if (!ts) return '-'
  return new Date(Number(ts) * 1000).toLocaleString('zh-CN')
}
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <button class="ctrl-btn back-btn" @click="goBack">&larr; 返回</button>
      <h2 class="page-title">指标查询</h2>
    </div>

    <div class="query-section">
      <div class="query-bar">
        <input
          class="ctrl-input query-input"
          v-model="metricName"
          placeholder="输入指标名称，如 ping_idc_latency_avg"
          @keyup.enter="queryMetric"
        />
        <button class="ctrl-btn accent-btn" :disabled="loading" @click="queryMetric">
          {{ loading ? '查询中...' : '查询' }}
        </button>
      </div>

      <!-- 错误 -->
      <div v-if="error" class="query-error">{{ error }}</div>

      <!-- 结果区域 -->
      <div v-if="results.length > 0" class="query-results">
        <div class="results-info">
          <span>指标: <strong>{{ metricName }}</strong></span>
          <span>类型: <strong>{{ resultType }}</strong></span>
          <span>数据条数: <strong>{{ results.length }}</strong></span>
        </div>

        <div class="results-table-wrap">
          <table class="results-table">
            <thead>
              <tr>
                <th v-for="label in Object.keys(results[0].metric)" :key="label">{{ label }}</th>
                <th>值</th>
                <th>时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, idx) in results" :key="idx">
                <td v-for="label in Object.keys(results[0].metric)" :key="label">
                  {{ row.metric[label] }}
                </td>
                <td class="val-cell">{{ formatValue(row.value?.[1]) }}</td>
                <td class="time-cell">{{ formatTimestamp(row.value?.[0]) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 20px;
}

.back-btn {
  font-size: 12px;
  padding: 4px 12px;
}

.page-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--text);
  margin: 0;
}

.query-section {
  width: 100%;
}

.query-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 16px;
}

.query-input {
  flex: 1;
  max-width: 500px;
  font-family: var(--font-mono);
  font-size: 13px;
}

.query-error {
  color: var(--err);
  font-size: 13px;
  padding: 12px;
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
}

.query-results {
  background: var(--bg-card);
  border: 1px solid var(--bg-border);
  border-radius: var(--radius);
  padding: 14px 20px;
}

.results-info {
  display: flex;
  gap: 24px;
  font-size: 12px;
  color: var(--text-dim);
  margin-bottom: 14px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--bg-border);
}

.results-info strong {
  color: var(--cyan);
  font-family: var(--font-mono);
}

.results-table-wrap {
  overflow-x: auto;
}

.results-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.results-table th {
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

.results-table td {
  padding: 6px 12px;
  color: var(--text);
  font-family: var(--font-mono);
  border-bottom: 1px solid var(--bg-border2);
  white-space: nowrap;
}

.results-table tr:last-child td {
  border-bottom: none;
}

.val-cell {
  color: var(--cyan) !important;
  font-weight: 600;
}

.time-cell {
  color: var(--text-dim) !important;
  font-size: 11px;
}
</style>
