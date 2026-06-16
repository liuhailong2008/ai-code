<script setup>
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { protos, topSrc } from '@/composables/useData'
import Chart from 'chart.js/auto'

const range = ref('1h')
const ranges = [
  { key: '1h',  label: '1小时' },
  { key: '6h',  label: '6小时' },
  { key: '24h', label: '24小时' },
  { key: '7d',  label: '7天' },
]

function setRange(key) { range.value = key }

// ── Trend Chart ──
const trendCanvas = ref(null)
let trendChart = null

function initTrendChart() {
  if (!trendCanvas.value || trendChart) return
  const labels = Array.from({ length: 24 }, (_, i) => i + ':00')
  const d1 = Array.from({ length: 24 }, () => +(15 + Math.random() * 40).toFixed(1))
  const d2 = d1.map(v => +(v * 0.65).toFixed(1))
  trendChart = new Chart(trendCanvas.value, {
    type: 'line',
    data: {
      labels,
      datasets: [
        {
          label: '入站', data: d1,
          borderColor: '#00f5c4', borderWidth: 2, pointRadius: 3,
          pointBackgroundColor: '#00f5c4', fill: true,
          backgroundColor: 'rgba(0,245,196,0.06)', tension: 0.4,
        },
        {
          label: '出站', data: d2,
          borderColor: '#39ff7e', borderWidth: 2, pointRadius: 3,
          pointBackgroundColor: '#39ff7e', fill: true,
          backgroundColor: 'rgba(57,255,126,0.04)', tension: 0.4,
        },
      ],
    },
    options: {
      responsive: true, maintainAspectRatio: false,
      plugins: { legend: { display: false } },
      scales: {
        x: { ticks: { color: '#3a6070', font: { size: 9 } }, grid: { color: 'rgba(0,245,196,0.04)' }, border: { display: false } },
        y: { ticks: { color: '#3a6070', font: { size: 9 }, callback: v => v + 'G' }, grid: { color: 'rgba(0,245,196,0.04)' }, border: { display: false } },
      },
    },
  })
}

// ── Proto Bar Chart ──
const protoCanvas = ref(null)
let protoChart = null

function initProtoBarChart() {
  if (!protoCanvas.value || protoChart) return
  protoChart = new Chart(protoCanvas.value, {
    type: 'bar',
    data: {
      labels: protos.map(p => p.name),
      datasets: [{ data: protos.map(p => p.pct), backgroundColor: protos.map(p => p.color + 'aa'), borderColor: protos.map(p => p.color), borderWidth: 1 }],
    },
    options: {
      responsive: true, maintainAspectRatio: false,
      plugins: { legend: { display: false } },
      scales: {
        x: { ticks: { color: '#3a6070', font: { size: 9 } }, grid: { display: false }, border: { display: false } },
        y: { ticks: { color: '#3a6070', font: { size: 9 }, callback: v => v + '%' }, grid: { color: 'rgba(0,245,196,0.04)' }, border: { display: false } },
      },
    },
  })
}

onMounted(async () => {
  await nextTick()
  initTrendChart()
  initProtoBarChart()
})

onBeforeUnmount(() => {
  if (trendChart) { trendChart.destroy(); trendChart = null }
  if (protoChart) { protoChart.destroy(); protoChart = null }
})
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">流量分析</h2>
      <div class="page-controls">
        <button
          v-for="r in ranges" :key="r.key"
          class="ctrl-btn" :class="{active:range===r.key}"
          @click="setRange(r.key)"
        >{{ r.label }}</button>
      </div>
    </div>
    <div class="traffic-layout">
      <!-- KPI Row -->
      <div class="traffic-kpi-row">
        <div class="t-kpi"><div class="tkl">峰值带宽</div><div class="tkv cyan">52.3 <span>Gbps</span></div><div class="tks">今日 14:18</div></div>
        <div class="t-kpi"><div class="tkl">总流量</div><div class="tkv green">1.84 <span>TB</span></div><div class="tks">今日累计</div></div>
        <div class="t-kpi"><div class="tkl">平均利用率</div><div class="tkv">61.4 <span>%</span></div><div class="tks">基于容量</div></div>
        <div class="t-kpi warn"><div class="tkl">异常流量</div><div class="tkv warn">3.2 <span>%</span></div><div class="tks">疑似攻击</div></div>
      </div>

      <!-- Trend Chart -->
      <div class="dash-card">
        <div class="card-header">
          <span class="card-title">带宽趋势</span>
          <div class="legend-row">
            <span class="leg-item cyan">■ 入站</span>
            <span class="leg-item green">■ 出站</span>
          </div>
        </div>
        <canvas ref="trendCanvas" style="width:100%;height:220px;"></canvas>
      </div>

      <!-- Bottom Row -->
      <div class="traffic-bottom-row">
        <div class="dash-card">
          <div class="card-header"><span class="card-title">TOP 10 流量来源</span></div>
          <div class="top-table">
            <div v-for="(s, i) in topSrc" :key="s.ip" class="tt-row-item">
              <div class="tt-rank">{{ i + 1 }}</div>
              <div class="tt-ip">{{ s.ip }}</div>
              <div class="tt-bw">{{ s.bw }}</div>
              <div class="tt-bar-wrap"><div class="tt-bar-bg"><div class="tt-bar-fill" :style="{width:s.pct+'%'}"></div></div></div>
            </div>
          </div>
        </div>
        <div class="dash-card">
          <div class="card-header"><span class="card-title">协议流量占比</span></div>
          <canvas ref="protoCanvas" style="width:100%;height:180px;"></canvas>
        </div>
      </div>
    </div>
  </div>
</template>
