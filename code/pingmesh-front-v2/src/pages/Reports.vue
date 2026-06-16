<script setup>
const reportCards = [
  { icon:'▤', iconClass:'cyan', title:'月度可用性报告', desc:'SLA 达标率分析 · 节点可用性 · 故障统计', meta:'上次生成 2025-03-31', featured: true },
  { icon:'▦', iconClass:'green', title:'流量容量规划', desc:'趋势预测 · 扩容建议 · 峰值分析', meta:'上次生成 2025-03-28', featured: false },
  { icon:'⚠', iconClass:'warn', title:'安全态势报告', desc:'威胁分析 · 攻击溯源 · 合规检测', meta:'上次生成 2025-04-01', featured: false },
  { icon:'◎', iconClass:'', title:'故障分析报告', desc:'MTTR · 根因分析 · 影响评估', meta:'上次生成 2025-03-25', featured: false },
  { icon:'⬡', iconClass:'cyan', title:'设备清单报告', desc:'硬件信息 · 版本合规 · 许可证状态', meta:'上次生成 2025-04-05', featured: false },
]
</script>

<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">报告中心</h2>
      <div class="page-controls">
        <button class="ctrl-btn accent-btn">+ 新建报告</button>
      </div>
    </div>
    <div class="reports-grid">
      <div
        v-for="(card, i) in reportCards" :key="i"
        class="report-card" :class="{featured:card.featured}"
      >
        <div class="rc-icon" :class="card.iconClass">{{ card.icon }}</div>
        <div class="rc-title">{{ card.title }}</div>
        <div class="rc-desc">{{ card.desc }}</div>
        <div class="rc-meta">{{ card.meta }}</div>
        <div class="rc-actions">
          <button class="rc-btn">生成</button>
          <button class="rc-btn">下载 PDF</button>
        </div>
      </div>
      <div class="report-card add-card">
        <div class="rc-icon" style="font-size:32px;opacity:0.3;">+</div>
        <div class="rc-title" style="opacity:0.4;">自定义报告</div>
        <div class="rc-desc" style="opacity:0.3;">配置指标 · 选择时间范围</div>
      </div>
    </div>
  </div>
</template>
