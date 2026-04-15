<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">告警配置</div>
        <div class="page-subtitle">配置告警规则和阈值</div>
      </div>
      <div class="page-actions">
        <button class="btn btn-primary" @click="showAddDialog = true">+ 添加规则</button>
      </div>
    </div>

    <div class="card animate-in delay-1">
      <div class="card-header">
        <span class="card-title">告警规则列表</span>
      </div>
      <div class="card-body" style="padding:0;">
        <table class="data-table">
          <thead>
            <tr>
              <th>规则名称</th>
              <th>告警类型</th>
              <th>阈值条件</th>
              <th>持续时间</th>
              <th>严重性</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="rule in rules" :key="rule.id">
              <td class="node-name">{{ rule.name }}</td>
              <td>{{ rule.type }}</td>
              <td>{{ rule.condition }}</td>
              <td>{{ rule.duration }}</td>
              <td>
                <span :style="{ color: getSeverityColor(rule.severity), fontWeight: 600 }">
                  {{ getSeverityLabel(rule.severity) }}
                </span>
              </td>
              <td>
                <span class="status-dot" :class="rule.enabled ? 'online' : 'offline'">
                  <span class="dot"></span>
                  {{ rule.enabled ? '启用' : '禁用' }}
                </span>
              </td>
              <td>
                <button class="btn btn-outline" style="padding:4px 8px;font-size:11px;margin-right:4px;" @click="toggleRule(rule)">
                  {{ rule.enabled ? '禁用' : '启用' }}
                </button>
                <button class="btn btn-outline" style="padding:4px 8px;font-size:11px;">编辑</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card animate-in delay-2" style="margin-top: 20px;">
      <div class="card-header">
        <span class="card-title">Prometheus 告警规则配置</span>
      </div>
      <div class="card-body">
        <textarea class="rule-editor" v-model="prometheusRules" placeholder="输入 Prometheus 告警规则..."></textarea>
      </div>
      <div class="card-footer" style="padding: 12px 22px; border-top: 1px solid var(--border-color);">
        <button class="btn btn-primary" @click="saveRules">保存配置</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const showAddDialog = ref(false)

const rules = ref([
  { id: 1, name: '链路延迟告警', type: 'latency', condition: '> 250ms', duration: '5分钟', severity: 'critical', enabled: true },
  { id: 2, name: '链路抖动告警', type: 'jitter', condition: '> 50ms', duration: '10分钟', severity: 'warning', enabled: true },
  { id: 3, name: '丢包率告警', type: 'loss', condition: '> 1%', duration: '5分钟', severity: 'critical', enabled: true },
  { id: 4, name: '节点离线告警', type: 'offline', condition: '心跳超时', duration: '3分钟', severity: 'critical', enabled: true },
  { id: 5, name: 'P99延迟告警', type: 'p99', condition: '> 200ms', duration: '15分钟', severity: 'warning', enabled: false }
])

const prometheusRules = ref(`groups:
- name: pingmesh_alerts
  rules:
  - alert: HighLatency
    expr: pingmesh_latency > 250
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High latency detected"
      description: "Link latency is {{ $value }}ms"

  - alert: HighPacketLoss
    expr: pingmesh_loss_rate > 1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High packet loss"
      description: "Packet loss rate is {{ $value }}%"`)

function getSeverityColor(severity) {
  return severity === 'critical' ? '#ef4444' : severity === 'warning' ? '#f59e0b' : '#3b82f6'
}

function getSeverityLabel(severity) {
  return severity === 'critical' ? '严重' : severity === 'warning' ? '警告' : '通知'
}

function toggleRule(rule) {
  rule.enabled = !rule.enabled
}

function saveRules() {
  alert('规则保存成功！')
}
</script>

<style scoped>
.rule-editor {
  width: 100%;
  height: 300px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 12px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 13px;
  color: var(--text-primary);
  resize: vertical;
}
</style>
