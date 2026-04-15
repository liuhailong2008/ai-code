<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">探测代理管理</div>
        <div class="page-subtitle">管理探测代理 · 在线 {{ onlineCount }} / 总计 {{ agents.length }}</div>
      </div>
      <div class="page-actions">
        <select v-model="statusFilter" class="btn btn-outline" style="padding: 9px 14px;">
          <option value="all">全部状态</option>
          <option value="online">在线</option>
          <option value="offline">离线</option>
          <option value="error">异常</option>
        </select>
        <button class="btn btn-outline" @click="batchStart">批量启动</button>
        <button class="btn btn-outline" @click="batchStop">批量停止</button>
        <button class="btn btn-primary" @click="batchUpgrade">批量升级</button>
      </div>
    </div>

    <div class="stats-row animate-in delay-1">
      <div class="stat-card green">
        <div class="stat-header">
          <span class="stat-label">在线代理</span>
          <div class="stat-icon">✓</div>
        </div>
        <div class="stat-value">{{ onlineCount }}</div>
        <div class="stat-footer trend-up">占比 {{ ((onlineCount / agents.length) * 100).toFixed(1) }}%</div>
      </div>
      <div class="stat-card red">
        <div class="stat-header">
          <span class="stat-label">离线代理</span>
          <div class="stat-icon">✗</div>
        </div>
        <div class="stat-value">{{ offlineCount }}</div>
        <div class="stat-footer trend-down">需检查</div>
      </div>
      <div class="stat-card blue">
        <div class="stat-header">
          <span class="stat-label">需升级</span>
          <div class="stat-icon">⬆</div>
        </div>
        <div class="stat-value">{{ upgradeCount }}</div>
        <div class="stat-footer trend-neutral">有新版本</div>
      </div>
      <div class="stat-card cyan">
        <div class="stat-header">
          <span class="stat-label">平均CPU</span>
          <div class="stat-icon">📊</div>
        </div>
        <div class="stat-value">{{ avgCPU }}%</div>
        <div class="stat-footer trend-neutral">运行正常</div>
      </div>
    </div>

    <div class="card animate-in delay-2">
      <div class="card-body" style="padding:0;overflow-x:auto;">
        <table class="data-table">
          <thead>
            <tr>
              <th style="width: 40px;"><input type="checkbox" @change="toggleSelectAll"></th>
              <th>代理名称</th>
              <th>IP地址</th>
              <th>所属机房</th>
              <th>版本</th>
              <th>最新版本</th>
              <th>状态</th>
              <th>运行时间</th>
              <th>CPU</th>
              <th>内存</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="agent in filteredAgents" :key="agent.id">
              <td><input type="checkbox" v-model="agent.selected"></td>
              <td class="node-name">{{ agent.name }}</td>
              <td>{{ agent.ip }}</td>
              <td><span class="region-tag">{{ agent.idc }}</span></td>
              <td>v{{ agent.version }}</td>
              <td>
                <span v-if="agent.version < agent.latestVersion" style="color: var(--accent-yellow);">
                  v{{ agent.latestVersion }}
                </span>
                <span v-else style="color: var(--accent-green);">最新</span>
              </td>
              <td>
                <span class="status-dot" :class="agent.status">
                  <span class="dot"></span>
                  {{ getStatusLabel(agent.status) }}
                </span>
              </td>
              <td style="color:var(--text-muted)">{{ agent.uptime }}</td>
              <td>
                <div class="usage-bar">
                  <div class="usage-fill" :style="{ width: agent.cpu + '%', background: getCPUColor(agent.cpu) }"></div>
                </div>
                <span style="font-size: 11px; color: var(--text-muted);">{{ agent.cpu }}%</span>
              </td>
              <td>
                <div class="usage-bar">
                  <div class="usage-fill" :style="{ width: agent.memory + '%', background: '#3b82f6' }"></div>
                </div>
                <span style="font-size: 11px; color: var(--text-muted);">{{ agent.memory }}%</span>
              </td>
              <td>
                <button v-if="agent.status === 'online'" class="btn btn-outline" style="padding:4px 8px;font-size:11px;margin-right:2px;" @click="stopAgent(agent)">
                  停止
                </button>
                <button v-else class="btn btn-outline" style="padding:4px 8px;font-size:11px;margin-right:2px;" @click="startAgent(agent)">
                  启动
                </button>
                <button v-if="agent.version < agent.latestVersion" class="btn btn-outline" style="padding:4px 8px;font-size:11px;" @click="upgradeAgent(agent)">
                  升级
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const statusFilter = ref('all')

const agents = ref([
  { id: 1, name: 'agent-bj-01', ip: '10.1.1.100', idc: 'IDC1-北京', version: '2.1.0', latestVersion: '2.2.0', status: 'online', uptime: '15天6小时', cpu: 35, memory: 48, selected: false },
  { id: 2, name: 'agent-bj-02', ip: '10.1.1.101', idc: 'IDC1-北京', version: '2.2.0', latestVersion: '2.2.0', status: 'online', uptime: '22天3小时', cpu: 28, memory: 42, selected: false },
  { id: 3, name: 'agent-sh-01', ip: '10.2.1.100', idc: 'IDC2-上海', version: '2.1.5', latestVersion: '2.2.0', status: 'online', uptime: '8天12小时', cpu: 52, memory: 65, selected: false },
  { id: 4, name: 'agent-sh-02', ip: '10.2.1.101', idc: 'IDC2-上海', version: '2.2.0', latestVersion: '2.2.0', status: 'offline', uptime: '--', cpu: 0, memory: 0, selected: false },
  { id: 5, name: 'agent-sz-01', ip: '10.3.1.100', idc: 'IDC3-深圳', version: '2.2.0', latestVersion: '2.2.0', status: 'online', uptime: '30天5小时', cpu: 42, memory: 55, selected: false },
  { id: 6, name: 'agent-hz-01', ip: '10.4.1.100', idc: 'IDC4-杭州', version: '2.1.0', latestVersion: '2.2.0', status: 'error', uptime: '2天1小时', cpu: 95, memory: 88, selected: false },
  { id: 7, name: 'agent-gz-01', ip: '10.5.1.100', idc: 'IDC5-广州', version: '2.2.0', latestVersion: '2.2.0', status: 'online', uptime: '18天8小时', cpu: 38, memory: 52, selected: false },
  { id: 8, name: 'agent-cd-01', ip: '10.6.1.100', idc: 'IDC6-成都', version: '2.2.0', latestVersion: '2.2.0', status: 'online', uptime: '12天14小时', cpu: 45, memory: 58, selected: false }
])

const filteredAgents = computed(() => {
  if (statusFilter.value === 'all') return agents.value
  return agents.value.filter(a => a.status === statusFilter.value)
})

const onlineCount = computed(() => agents.value.filter(a => a.status === 'online').length)
const offlineCount = computed(() => agents.value.filter(a => a.status === 'offline').length)
const upgradeCount = computed(() => agents.value.filter(a => a.version < a.latestVersion).length)
const avgCPU = computed(() => {
  const onlineAgents = agents.value.filter(a => a.status === 'online')
  return Math.round(onlineAgents.reduce((sum, a) => sum + a.cpu, 0) / onlineAgents.length)
})

function getStatusLabel(status) {
  return { online: '在线', offline: '离线', error: '异常' }[status] || status
}

function getCPUColor(cpu) {
  if (cpu < 50) return '#10b981'
  if (cpu < 80) return '#f59e0b'
  return '#ef4444'
}

function toggleSelectAll(event) {
  filteredAgents.value.forEach(a => a.selected = event.target.checked)
}

function startAgent(agent) {
  alert(`启动代理 ${agent.name}`)
  agent.status = 'online'
  agent.uptime = '0分钟'
}

function stopAgent(agent) {
  alert(`停止代理 ${agent.name}`)
  agent.status = 'offline'
  agent.uptime = '--'
}

function upgradeAgent(agent) {
  alert(`升级代理 ${agent.name} 到 v${agent.latestVersion}`)
  agent.version = agent.latestVersion
}

function batchStart() {
  const selected = agents.value.filter(a => a.selected && a.status !== 'online')
  if (selected.length === 0) {
    alert('请先选择要启动的代理')
    return
  }
  alert(`启动 ${selected.length} 个代理`)
}

function batchStop() {
  const selected = agents.value.filter(a => a.selected && a.status === 'online')
  if (selected.length === 0) {
    alert('请先选择要停止的代理')
    return
  }
  alert(`停止 ${selected.length} 个代理`)
}

function batchUpgrade() {
  const selected = agents.value.filter(a => a.selected && a.version < a.latestVersion)
  if (selected.length === 0) {
    alert('没有需要升级的代理')
    return
  }
  alert(`升级 ${selected.length} 个代理`)
}
</script>

<style scoped>
.usage-bar {
  width: 50px;
  height: 4px;
  background: var(--bg-secondary);
  border-radius: 2px;
  overflow: hidden;
  display: inline-block;
  vertical-align: middle;
  margin-right: 4px;
}

.usage-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s;
}
</style>
