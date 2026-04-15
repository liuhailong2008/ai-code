<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">历史告警</div>
        <div class="page-subtitle">告警历史记录 · 共 {{ alerts.length }} 条</div>
      </div>
      <div class="page-actions">
        <div class="search-box" style="width: 200px;">
          <span>🔍</span>
          <input type="text" v-model="searchKeyword" placeholder="搜索告警...">
        </div>
        <button class="btn btn-outline" @click="exportData">📥 导出</button>
      </div>
    </div>

    <div class="card animate-in delay-1">
      <div class="card-body" style="padding:0;">
        <table class="data-table">
          <thead>
            <tr>
              <th>状态</th>
              <th>告警标题</th>
              <th>源 → 目标</th>
              <th>触发时间</th>
              <th>恢复时间</th>
              <th>持续时间</th>
              <th>峰值延迟</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="alert in filteredAlerts" :key="alert.id">
              <td>
                <span class="status-dot" :class="alert.status">
                  <span class="dot"></span>
                  {{ alert.status === 'resolved' ? '已恢复' : '处理中' }}
                </span>
              </td>
              <td class="node-name">{{ alert.title }}</td>
              <td>{{ alert.source }} → {{ alert.target }}</td>
              <td style="color:var(--text-muted)">{{ alert.triggerTime }}</td>
              <td style="color:var(--accent-green)">{{ alert.resolveTime }}</td>
              <td>{{ alert.duration }}</td>
              <td>{{ alert.peakLatency.toFixed(1) }} ms</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="pagination" style="margin-top: 20px; display: flex; justify-content: center; gap: 8px;">
      <button class="btn btn-outline" style="padding: 6px 12px;">上一页</button>
      <button class="btn btn-outline" style="padding: 6px 12px;">1 / 5</button>
      <button class="btn btn-outline" style="padding: 6px 12px;">下一页</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const searchKeyword = ref('')

const alerts = ref([
  { id: 1, status: 'resolved', title: '北京→上海链路抖动', source: 'IDC1-北京', target: 'IDC2-上海', triggerTime: '08:15', resolveTime: '09:42', duration: '1小时27分钟', peakLatency: 245.5 },
  { id: 2, status: 'resolved', title: '深圳节点离线', source: 'IDC3-深圳', target: 'Leaf-15', triggerTime: '07:30', resolveTime: '08:15', duration: '45分钟', peakLatency: 0 },
  { id: 3, status: 'resolved', title: '杭州丢包率异常', source: 'IDC4-杭州', target: 'IDC1-北京', triggerTime: '06:20', resolveTime: '07:05', duration: '45分钟', peakLatency: 185.2 },
  { id: 4, status: 'resolved', title: '广州 Leaf-05 延迟升高', source: 'IDC5-广州', target: 'Leaf-05', triggerTime: '05:10', resolveTime: '05:45', duration: '35分钟', peakLatency: 156.8 },
  { id: 5, status: 'processing', title: '成都网络拥塞', source: 'IDC6-成都', target: 'IDC2-上海', triggerTime: '04:25', resolveTime: '--', duration: '进行中', peakLatency: 312.5 }
])

const filteredAlerts = computed(() => {
  return alerts.value.filter(a => 
    a.title.includes(searchKeyword.value) || 
    a.source.includes(searchKeyword.value) || 
    a.target.includes(searchKeyword.value)
  )
})

function exportData() {
  alert('导出功能开发中...')
}
</script>
