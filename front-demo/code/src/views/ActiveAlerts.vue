<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">活跃告警</div>
        <div class="page-subtitle">当前活跃告警列表 · 共 {{ alerts.length }} 条</div>
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

    <div class="stats-row animate-in delay-1">
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

    <div class="card animate-in delay-2">
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
                <button class="btn btn-outline" style="padding:4px 10px;font-size:11px;" @click="openProcessDialog(alert)">
                  处理
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 处理对话框 -->
    <div v-if="showDialog" class="dialog-overlay" @click="closeDialog">
      <div class="dialog-content dialog-wide" @click.stop>
        <div class="dialog-header">
          <span class="dialog-title">处理告警</span>
          <button class="dialog-close" @click="closeDialog">×</button>
        </div>
        <div class="dialog-body dialog-body-wide">
          <!-- 左侧：告警详情 -->
          <div class="alert-detail">
            <div class="detail-row">
              <span class="detail-label">严重性：</span>
              <span class="detail-value" :style="{ color: getSeverityColor(currentAlert.severity) }">{{ getSeverityLabel(currentAlert.severity) }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">告警标题：</span>
              <span class="detail-value">{{ currentAlert.title }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">源 → 目标：</span>
              <span class="detail-value">{{ currentAlert.source }} → {{ currentAlert.target }}</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">延迟：</span>
              <span class="detail-value" :style="{ color: currentAlert.latency > 250 ? '#ef4444' : currentAlert.latency > 100 ? '#f59e0b' : '#10b981' }">{{ currentAlert.latency.toFixed(1) }} ms</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">丢包率：</span>
              <span class="detail-value" :style="{ color: currentAlert.loss >= 100 ? '#ef4444' : 'inherit' }">{{ currentAlert.loss.toFixed(2) }}%</span>
            </div>
            <div class="detail-row">
              <span class="detail-label">触发时间：</span>
              <span class="detail-value">{{ currentAlert.triggerTime }}</span>
            </div>
<div class="detail-row">
              <span class="detail-label">持续时间：</span>
              <span class="detail-value">{{ currentAlert.duration }}</span>
            </div>
            <!-- Leaf设备详细信息 -->
            <div class="device-info-box">
              <div class="detail-row">
                <span class="detail-label">{{ currentAlert.source }}带外IP：</span>
                <span class="detail-value">{{ getDeviceInfo(currentAlert.source).ip }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">{{ currentAlert.source }}负责人：</span>
                <span class="detail-value">{{ getDeviceInfo(currentAlert.source).owner }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">{{ currentAlert.target }}带外IP：</span>
                <span class="detail-value">{{ getDeviceInfo(currentAlert.target).ip }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">{{ currentAlert.target }}负责人：</span>
                <span class="detail-value">{{ getDeviceInfo(currentAlert.target).owner }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">今日值班：</span>
                <span class="detail-value">{{ getDeviceInfo(currentAlert.source).dutyContact }}</span>
              </div>
            </div>
          </div>
          <!-- 右侧：网络拓扑图和图表 -->
          <div class="topology-section">
            <div class="section-title">当前机房网络设备拓扑</div>
            <div ref="topologyRef" class="topology-chart" style="height: 180px;"></div>
            
            <div class="charts-row">
              <div class="chart-half">
                <div class="section-title">Leaf 连通性热力图</div>
                <div ref="heatmapRef" class="chart-container" style="height: 200px;"></div>
              </div>
              <div class="chart-half">
                <div class="section-title">耗时-丢包率散点图</div>
                <div ref="scatterRef" class="chart-container" style="height: 200px;"></div>
              </div>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-outline" @click="closeDialog">取消</button>
          <button class="btn btn-primary" @click="confirmProcess">确认处理</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick } from 'vue'
import * as echarts from 'echarts'

const searchKeyword = ref('')
const severityFilter = ref('all')
const showDialog = ref(false)
const currentAlert = ref({})
const topologyRef = ref(null)
const heatmapRef = ref(null)
const scatterRef = ref(null)
let topologyChart = null
let heatmapChart = null
let scatterChart = null

const alerts = ref([
  { id: 1, severity: 'critical', title: '北京IDC-1 Leaf-03 与 Leaf-10 链路超时', source: 'Leaf-03', target: 'Leaf-10', latency: 300, loss: 100, triggerTime: '10:32', duration: '8分钟' },
  { id: 2, severity: 'warning', title: '杭州IDC-1 Leaf-05 与 Leaf-08 丢包率升高', source: 'Leaf-05', target: 'Leaf-08', latency: 185.2, loss: 2.5, triggerTime: '10:15', duration: '25分钟' },
  { id: 3, severity: 'critical', title: '广州IDC-1 Leaf-12 与 Leaf-14 链路超时', source: 'Leaf-12', target: 'Leaf-14', latency: 300, loss: 100, triggerTime: '09:58', duration: '42分钟' }
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

function getDeviceInfo(leafName) {
  const deviceMap = {
    'Leaf-03': { ip: '10.78.0.32', owner: '张瞻（17678190026）', duty: '今日值班', dutyContact: '陈灿（18600107299）' },
    'Leaf-10': { ip: '10.78.0.75', owner: '陈鹏（13572109908）', duty: '今日值班', dutyContact: '陈灿（18600107299）' },
    'Leaf-05': { ip: '10.78.0.45', owner: '李明（13800138000）', duty: '今日值班', dutyContact: '陈灿（18600107299）' },
    'Leaf-08': { ip: '10.78.0.58', owner: '王强（13900139000）', duty: '今日值班', dutyContact: '陈灿（18600107299）' },
    'Leaf-12': { ip: '10.78.0.88', owner: '赵磊（13700137000）', duty: '今日值班', dutyContact: '陈灿（18600107299）' },
    'Leaf-14': { ip: '10.78.0.92', owner: '刘洋（13600136000）', duty: '今日值班', dutyContact: '陈灿（18600107299）' }
  }
  return deviceMap[leafName] || { ip: '-', owner: '-', duty: '今日值班', dutyContact: '陈灿（18600107299）' }
}

function openProcessDialog(alert) {
  currentAlert.value = alert
  showDialog.value = true
  nextTick(() => {
    initTopology()
    initHeatmap()
    initScatter()
  })
}

function closeDialog() {
  showDialog.value = false
  if (topologyChart) {
    topologyChart.dispose()
    topologyChart = null
  }
  if (heatmapChart) {
    heatmapChart.dispose()
    heatmapChart = null
  }
  if (scatterChart) {
    scatterChart.dispose()
    scatterChart = null
  }
}

function confirmProcess() {
  const alert = alerts.value.find(a => a.id === currentAlert.value.id)
  if (alert) {
    alert.severity = 'info'
    alert.duration = '已处理'
  }
  closeDialog()
}

function initTopology() {
  if (!topologyRef.value) return
  topologyChart = echarts.init(topologyRef.value)
  updateTopology()
}

function updateTopology() {
  if (!topologyChart) return
  
  // 根据当前告警获取异常的两台接入交换机
  const alertSource = currentAlert.value.source || ''
  const alertTarget = currentAlert.value.target || ''
  
  // 判断哪些接入交换机有异常
  const abnormalSwitches = [alertSource, alertTarget].filter(s => s.includes('Leaf'))
  
  // 使用树图实现四层网络拓扑，显示主机数量
  // 每个机柜3台主机，正常显示3/3，有超时显示如1/3
  // 第三层汇总第四层，如9/9
  
  // 计算第四层每个机柜的状态
  const getCabinetStatus = (switchName) => {
    if (abnormalSwitches.includes(switchName)) {
      // 随机一台主机超时
      return { total: 3, online: 2, name: '2/3' }
    }
    return { total: 3, online: 3, name: '3/3' }
  }
  
  // 计算第三层汇总
  const getAccessStatus = (switchName, cabinets) => {
    let online = 0
    let total = 0
    cabinets.forEach(cab => {
      const status = getCabinetStatus(switchName)
      online += status.online
      total += status.total
    })
    return { online, total, name: `${online}/${total}` }
  }
  
  // APP-接入交换机
  const app1Status = getCabinetStatus('Leaf-03')
  const app2Status = getCabinetStatus('Leaf-10')
  const app3Status = getCabinetStatus('Leaf-12')
  const app1Total = getAccessStatus('Leaf-03', [app1Status])
  const app2Total = getAccessStatus('Leaf-10', [app2Status])
  const app3Total = getAccessStatus('Leaf-12', [app3Status])
  
  // OM-接入交换机
  const om1Status = getCabinetStatus('Leaf-05')
  const om2Status = getCabinetStatus('Leaf-08')
  const om1Total = getAccessStatus('Leaf-05', [om1Status])
  const om2Total = getAccessStatus('Leaf-08', [om2Status])
  
  // APP汇聚汇总
  const appTotalOnline = app1Status.online + app2Status.online + app3Status.online
  const appTotalAll = app1Status.total + app2Status.total + app3Status.total
  
  // OM汇聚汇总
  const omTotalOnline = om1Status.online + om2Status.online
  const omTotalAll = om1Status.total + om2Status.total
  
  const data = {
    name: `核心交换机\n9/9`,
    children: [
      {
        name: `APP-汇聚\n${appTotalOnline}/${appTotalAll}`,
        children: [
          {
            name: `Leaf-03\n${app1Status.name}`,
            children: [
              { name: `机柜-1-1\n3/3` },
              { name: `机柜-1-2\n3/3` },
              { name: `机柜-1-3\n${app1Status.name}` }
            ]
          },
          {
            name: `Leaf-10\n${app2Status.name}`,
            children: [
              { name: `机柜-2-1\n3/3` },
              { name: `机柜-2-2\n3/3` },
              { name: `机柜-2-3\n${app2Status.name}` }
            ]
          },
          {
            name: `Leaf-12\n${app3Status.name}`,
            children: [
              { name: `机柜-3-1\n3/3` },
              { name: `机柜-3-2\n3/3` },
              { name: `机柜-3-3\n3/3` }
            ]
          }
        ]
      },
      {
        name: `OM-汇聚\n${omTotalOnline}/${omTotalAll}`,
        children: [
          {
            name: `Leaf-05\n${om1Status.name}`,
            children: [
              { name: `机柜-4-1\n3/3` },
              { name: `机柜-4-2\n3/3` },
              { name: `机柜-4-3\n3/3` }
            ]
          },
          {
            name: `Leaf-08\n${om2Status.name}`,
            children: [
              { name: `机柜-5-1\n3/3` },
              { name: `机柜-5-2\n3/3` },
              { name: `机柜-5-3\n3/3` }
            ]
          }
        ]
      }
    ]
  }
  
  const option = {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'item',
      triggerOn: 'mousemove',
      formatter: '{b}'
    },
    series: [
      {
        type: 'tree',
        data: [data],
        top: '5%',
        left: '10%',
        bottom: '5%',
        right: '20%',
        symbol: 'circle',
        symbolSize: 18,
        itemStyle: {
          color: (params) => {
            const name = params.data.name || ''
            // 判断是否为异常的Leaf交换机（Leaf-03或Leaf-10）
            const isAbnormalLeaf = name.includes('Leaf-03') || name.includes('Leaf-10')
            if (isAbnormalLeaf) {
              return '#ef4444' // 红色
            }
            // 判断是否为满配：3/3 或 9/9 等
            const isFull = name.includes('/3') || name.includes('/9') || name.includes('/15')
            // 如果不是满配（异常），显示橙黄色
            if (!isFull && name.includes('/')) {
              return '#f59e0b'
            }
            return '#10b981'
          },
          borderColor: (params) => {
            const name = params.data.name || ''
            const isAbnormalLeaf = name.includes('Leaf-03') || name.includes('Leaf-10')
            if (isAbnormalLeaf) {
              return '#ef4444'
            }
            const isFull = name.includes('/3') || name.includes('/9') || name.includes('/15')
            if (!isFull && name.includes('/')) {
              return '#f59e0b'
            }
            return '#10b981'
          }
        },
        lineStyle: {
          color: '#10b981',
          width: 1.5
        },
        label: {
          show: true,
          position: 'right',
          distance: 8,
          fontSize: 9,
          color: '#e8ecf4',
          lineHeight: 14
        },
        leaves: {
          label: {
            position: 'right',
            distance: 5,
            fontSize: 8,
            lineHeight: 12
          }
        },
        emphasis: {
          focus: 'descendant'
        },
        expandAndCollapse: false,
        initialTreeDepth: -1
      }
    ]
  }
  
  topologyChart.setOption(option)
}

function getThemeColors() {
  const theme = document.documentElement.getAttribute('data-theme') || 'style1'
  if (theme === 'style2') {
    return { bg: '#FFFFFF', text: '#111827', textMuted: '#9CA3AF', grid: '#E5E7EB' }
  }
  if (theme === 'style3') {
    return { bg: '#FFFFFF', text: '#1F2937', textMuted: '#6B7280', grid: '#E5E7EB' }
  }
  return { bg: '#1a2235', text: '#e8ecf4', textMuted: '#5a6a85', grid: 'rgba(30,42,63,.5)' }
}

function initHeatmap() {
  if (!heatmapRef.value) return
  heatmapChart = echarts.init(heatmapRef.value)
  updateHeatmap()
}

function updateHeatmap() {
  const colors = getThemeColors()
  const leafs = Array.from({ length: 15 }, (_, i) => `Leaf-${String(i + 1).padStart(2, '0')}`)
  
  // 获取当前告警对应的异常Leaf
  const alertSource = currentAlert.value.source || ''
  const alertTarget = currentAlert.value.target || ''
  const abnormalPairs = [[alertSource, alertTarget], [alertTarget, alertSource]]
  
  // 生成热力图数据
  const data = []
  for (let i = 0; i < leafs.length; i++) {
    for (let j = 0; j < leafs.length; j++) {
      let latency = i === j ? 0 : Math.random() * 8.5
      // 如果是异常Leaf对，设置latency为300
      if (abnormalPairs.some(pair => pair[0] === leafs[j] && pair[1] === leafs[i])) {
        latency = 300
      }
      data.push([j, i, parseFloat(latency.toFixed(2))])
    }
  }
  
  const option = {
    backgroundColor: colors.bg,
    tooltip: { position: 'top', formatter: (p) => `${leafs[p.value[1]]} → ${leafs[p.value[0]]}<br/>延迟: ${p.value[2].toFixed(2)} ms` },
    grid: { top: 30, left: 70, right: 20, bottom: 40 },
    xAxis: { type: 'category', data: leafs, splitArea: { show: true }, axisLabel: { color: colors.textMuted, fontSize: 9, interval: 4, rotate: 45 }, axisLine: { lineStyle: { color: colors.grid } } },
    yAxis: { type: 'category', data: leafs, splitArea: { show: true }, axisLabel: { color: colors.textMuted, fontSize: 9, interval: 4 }, axisLine: { lineStyle: { color: colors.grid } } },
    visualMap: { min: 0, max: 300, calculable: true, orient: 'horizontal', left: 'center', bottom: 0, inRange: { color: ['#10b981', '#3b82f6', '#f59e0b', '#ef4444'] }, textStyle: { color: colors.textMuted } },
    series: [{ type: 'heatmap', data: data, label: { show: true, color: colors.text, fontSize: 7, formatter: (p) => p.value[2] === 0 ? '-' : p.value[2].toFixed(1) }, itemStyle: { borderColor: colors.grid, borderWidth: 0.5 } }]
  }
  
  heatmapChart.setOption(option)
}

function initScatter() {
  if (!scatterRef.value) return
  scatterChart = echarts.init(scatterRef.value)
  updateScatter()
}

function updateScatter() {
  const colors = getThemeColors()
  
  // 获取当前告警对应的异常Leaf
  const alertSource = currentAlert.value.source || ''
  const alertTarget = currentAlert.value.target || ''
  const abnormalLeaves = [alertSource, alertTarget].filter(s => s.includes('Leaf'))
  const abnormalLeafNames = abnormalLeaves.map(s => s.replace('Leaf-', '').padStart(2, '0'))
  
  const data = []
  
  for (let i = 0; i < 25; i++) {
    const leafNum = String(i + 1).padStart(2, '0')
    let latency = Math.random() * 200
    let loss = Math.random() * 3
    let status = 'normal'
    let color = '#10b981'
    
    // 判断是否为异常的Leaf
    if (abnormalLeafNames.includes(leafNum)) {
      latency = 300
      loss = 100
      status = 'critical'
      color = '#ef4444'
    }
    
    data.push({
      value: [parseFloat(latency.toFixed(2)), parseFloat(loss.toFixed(3)), `Leaf-${leafNum}`, status],
      itemStyle: { color }
    })
  }
  
  const option = {
    backgroundColor: colors.bg,
    tooltip: { trigger: 'item', formatter: (p) => `${p.value[2]}<br/>延迟: ${p.value[0]} ms<br/>丢包率: ${p.value[1]}%` },
    grid: { top: 30, left: 60, right: 30, bottom: 50 },
    xAxis: { type: 'value', name: '延迟 (ms)', nameLocation: 'middle', nameGap: 25, nameTextStyle: { color: colors.textMuted }, axisLabel: { color: colors.textMuted }, axisLine: { lineStyle: { color: colors.grid } }, splitLine: { lineStyle: { color: colors.grid, type: 'dashed' } }, min: 0, max: 350 },
    yAxis: { type: 'value', name: '丢包率 (%)', nameLocation: 'middle', nameGap: 40, nameTextStyle: { color: colors.textMuted }, axisLabel: { color: colors.textMuted, formatter: v => v.toFixed(1) }, axisLine: { lineStyle: { color: colors.grid } }, splitLine: { lineStyle: { color: colors.grid, type: 'dashed' } }, min: 0, max: 5 },
    series: [{ type: 'scatter', data, symbolSize: 10, emphasis: { scale: 1.5 } }]
  }
  
  scatterChart.setOption(option)
}

function handleResize() {
  topologyChart?.resize()
  heatmapChart?.resize()
  scatterChart?.resize()
}

window.addEventListener('resize', handleResize)
</script>

<style scoped>
.alert-row:hover {
  background: rgba(239,68,68,.04);
}
.stat-card.yellow .stat-icon {
  background: rgba(245,158,11,.15);
  color: var(--accent-yellow);
}

/* 对话框样式 */
.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog-content {
  background: var(--bg-card);
  border-radius: var(--radius);
  width: 600px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.dialog-content.dialog-wide {
  width: 1200px;
}

.dialog-body-wide {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

.dialog-body-wide .alert-detail {
  width: 280px;
  flex-shrink: 0;
}

.dialog-body-wide .topology-section {
  flex: 1;
  min-width: 0;
}

/* 图表横向排列 */
.charts-row {
  display: flex;
  gap: 16px;
}

.chart-half {
  flex: 1;
  min-width: 0;
}

/* 设备信息卡片 */
.device-info-section {
  display: flex;
  gap: 16px;
  margin-top: 16px;
  padding: 16px;
  background: var(--bg-page);
  border-radius: var(--radius-sm);
}

.device-card {
  flex: 1;
  min-width: 0;
}

.device-card .device-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

/* 设备信息横向排列 */
.device-info-inline {
  display: flex;
  gap: 24px;
  margin-top: 16px;
  padding: 12px 16px;
  background: var(--bg-page);
  border-radius: var(--radius-sm);
}

.device-info-item {
  flex: 1;
}

.device-info-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.device-info-text {
  font-size: 12px;
  color: var(--text-muted);
  line-height: 1.6;
}

/* 设备信息盒子 - 与告警概览格式统一 */
.device-info-box {
  margin-top: 12px;
  padding: 12px 16px;
  background: var(--bg-page);
  border-radius: var(--radius-sm);
}

.device-info-box .detail-row {
  margin-bottom: 4px;
}

.device-info-box .detail-row:last-child {
  margin-bottom: 0;
}

.device-info-box .detail-label {
  width: 100px;
  white-space: nowrap;
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.dialog-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.dialog-close {
  background: none;
  border: none;
  font-size: 24px;
  color: var(--text-muted);
  cursor: pointer;
  padding: 0;
  line-height: 1;
}

.dialog-close:hover {
  color: var(--text-primary);
}

.dialog-body {
  padding: 20px;
  overflow-y: auto;
  flex: 1;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--border-color);
}

.alert-detail {
  margin-bottom: 20px;
  padding: 16px;
  background: var(--bg-page);
  border-radius: var(--radius-sm);
}

.detail-row {
  display: flex;
  margin-bottom: 8px;
}

.detail-row:last-child {
  margin-bottom: 0;
}

.detail-label {
  width: 100px;
  color: var(--text-muted);
  font-size: 13px;
}

.detail-value {
  flex: 1;
  color: var(--text-primary);
  font-size: 13px;
}

.topology-section {
  margin-top: 20px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 12px;
}

.topology-chart {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
}

.btn-primary {
  background: var(--accent-blue);
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
}

.btn-primary:hover {
  background: var(--accent-cyan);
}
</style>
