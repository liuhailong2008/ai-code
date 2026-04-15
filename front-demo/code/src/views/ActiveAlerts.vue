<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">未恢复告警</div>
        <div class="page-subtitle">当前未恢复的告警列表 · 共 {{ alerts.length }} 条</div>
      </div>
      <div class="page-actions">
        <div class="search-box" style="width: 200px;">
          <span>🔍</span>
          <input type="text" v-model="searchKeyword" placeholder="搜索告警...">
        </div>
        <select v-model="severityFilter" class="btn btn-outline" style="padding: 9px 14px;">
          <option value="all">全部严重性</option>
          <option value="critical">严重</option>
          <option value="warning">警告</option>
          <option value="info">通知</option>
        </select>
      </div>
    </div>

    <div class="card animate-in delay-1">
      <div class="card-body" style="padding:0;">
        <table class="data-table">
          <thead>
            <tr>
              <th>严重性</th>
              <th>告警标题</th>
              <th>源 → 目标</th>
              <th>延迟</th>
              <th>丢包率</th>
              <th>触发时间</th>
              <th>持续时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="alert in filteredAlerts" :key="alert.id" class="alert-row">
              <td>
                <span class="alert-severity" :class="alert.severity"></span>
                <span :style="{ color: getSeverityColor(alert.severity), fontWeight: 600 }">
                  {{ getSeverityLabel(alert.severity) }}
                </span>
              </td>
              <td class="node-name">{{ alert.title }}</td>
              <td>{{ alert.source }} → {{ alert.target }}</td>
              <td>
                <span :style="{ color: alert.latency > 250 ? '#ef4444' : alert.latency > 100 ? '#f59e0b' : '#10b981' }">
                  {{ alert.latency.toFixed(1) }} ms
                </span>
              </td>
              <td>
                <span class="loss-tag" :class="getLossClass(alert.loss)">
                  {{ alert.loss.toFixed(2) }}%
                </span>
              </td>
              <td style="color:var(--text-muted)">{{ alert.triggerTime }}</td>
              <td style="color:var(--accent-yellow)">{{ alert.duration }}</td>
              <td>
                <button class="btn btn-outline" style="padding:4px 10px;font-size:11px;" @click="handleAlert(alert)">
                  处理
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="stats-row animate-in delay-2" style="margin-top: 20px;">
      <div class="stat-card red">
        <div class="stat-header">
          <span class="stat-label">严重告警</span>
          <div class="stat-icon">⚠️</div>
        </div>
        <div class="stat-value">{{ criticalCount }}</div>
      </div>
      <div class="stat-card yellow">
        <div class="stat-header">
          <span class="stat-label">警告告警</span>
          <div class="stat-icon">⚡</div>
        </div>
        <div class="stat-value">{{ warningCount }}</div>
      </div>
      <div class="stat-card blue">
        <div class="stat-header">
          <span class="stat-label">通知</span>
          <div class="stat-icon">ℹ️</div>
        </div>
        <div class="stat-value">{{ infoCount }}</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const searchKeyword = ref('')
const severityFilter = ref('all')

const alerts = ref([
  { id: 1, severity: 'critical', title: '北京→上海链路延迟超标', source: 'IDC1-北京', target: 'IDC2-上海', latency: 285.5, loss: 1.25, triggerTime: '10:32', duration: '8分钟' },
  { id: 2, severity: 'critical', title: '深圳 dc-sz-sw-03 心跳超时', source: 'IDC3-深圳', target: 'Leaf-03', latency: 0, loss: 100, triggerTime: '10:28', duration: '12分钟' },
  { id: 3, severity: 'warning', title: '杭州→广州抖动异常', source: 'IDC4-杭州', target: 'IDC5-广州', latency: 185.2, loss: 0.45, triggerTime: '10:15', duration: '25分钟' },
  { id: 4, severity: 'warning', title: '成都节点丢包率升高', source: 'IDC6-成都', target: 'IDC1-北京', latency: 198.5, loss: 0.32, triggerTime: '10:05', duration: '35分钟' },
  { id: 5, severity: 'critical', title: '西安链路断开', source: 'IDC7-西安', target: 'IDC2-上海', latency: 0, loss: 100, triggerTime: '09:58', duration: '42分钟' },
  { id: 6, severity: 'info', title: '新增探针 dc-hz-srv-01 上线', source: 'IDC4-杭州', target: 'Leaf-12', latency: 2.5, loss: 0.01, triggerTime: '09:45', duration: '55分钟' },
  { id: 7, severity: 'warning', title: '上海 Leaf-08 延迟升高', source: 'IDC2-上海', target: 'Leaf-08', latency: 152.3, loss: 0.28, triggerTime: '09:30', duration: '1小时10分钟' }
])

const filteredAlerts = computed(() => {
  return alerts.value.filter(a => {
    const matchKeyword = a.title.includes(searchKeyword.value) || a.source.includes(searchKeyword.value) || a.target.includes(searchKeyword.value)
    const matchSeverity = severityFilter.value === 'all' || a.severity === severityFilter.value
    return matchKeyword && matchSeverity
  })
})

const criticalCount = computed(() => alerts.value.filter(a => a.severity === 'critical').length)
const warningCount = computed(() => alerts.value.filter(a => a.severity === 'warning').length)
const infoCount = computed(() => alerts.value.filter(a => a.severity === 'info').length)

function getSeverityColor(severity) {
  return severity === 'critical' ? '#ef4444' : severity === 'warning' ? '#f59e0b' : '#3b82f6'
}

function getSeverityLabel(severity) {
  return severity === 'critical' ? '严重' : severity === 'warning' ? '警告' : '通知'
}

function getLossClass(loss) {
  if (loss < 0.1) return 'loss-low'
  if (loss < 1) return 'loss-med'
  return 'loss-high'
}

function handleAlert(alert) {
  alert.severity = 'info'
  alert.duration = '已处理'
}
</script>

<style scoped>
.alert-row:hover {
  background: rgba(239,68,68,.04);
}
.stat-card.yellow .stat-icon {
  background: rgba(245,158,11,.15);
  color: var(--accent-yellow);
}
</style>
