<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">策略管理</div>
        <div class="page-subtitle">管理探测策略 · 共 {{ strategies.length }} 条策略</div>
      </div>
      <div class="page-actions">
        <div class="search-box" style="width: 200px;">
          <span>🔍</span>
          <input type="text" v-model="searchKeyword" placeholder="搜索策略...">
        </div>
        <button class="btn btn-primary" @click="showAddDialog = true">+ 新建策略</button>
      </div>
    </div>

    <div class="card animate-in delay-1">
      <div class="card-body" style="padding:0;overflow-x:auto;">
        <table class="data-table">
          <thead>
            <tr>
              <th>策略名称</th>
              <th>源机房</th>
              <th>目标机房</th>
              <th>代理数量/机房</th>
              <th>目标数量/机房</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>最后执行</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="strategy in filteredStrategies" :key="strategy.id">
              <td class="node-name">{{ strategy.name }}</td>
              <td><span class="region-tag">{{ strategy.sourceIDC }}</span></td>
              <td><span class="region-tag">{{ strategy.targetIDC }}</span></td>
              <td>{{ strategy.proxyCount }}</td>
              <td>{{ strategy.targetCount }}</td>
              <td>
                <span class="status-dot" :class="getStatusClass(strategy.status)">
                  <span class="dot"></span>
                  {{ getStatusLabel(strategy.status) }}
                </span>
              </td>
              <td style="color:var(--text-muted)">{{ strategy.createdAt }}</td>
              <td style="color:var(--text-muted)">{{ strategy.lastRun || '从未' }}</td>
              <td>
                <button class="btn btn-outline" style="padding:4px 8px;font-size:11px;margin-right:4px;" @click="generateTask(strategy)">
                  生成任务
                </button>
                <button class="btn btn-outline" style="padding:4px 8px;font-size:11px;" @click="dispatchTask(strategy)">
                  下发
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

const showAddDialog = ref(false)
const searchKeyword = ref('')

const strategies = ref([
  { id: 1, name: '北京-上海探测策略', sourceIDC: 'IDC1-北京', targetIDC: 'IDC2-上海', proxyCount: 5, targetCount: 10, status: 'active', createdAt: '2024-04-01', lastRun: '2024-04-14 10:00' },
  { id: 2, name: '华东区域探测', sourceIDC: 'IDC2-上海', targetIDC: 'IDC4-杭州', proxyCount: 3, targetCount: 8, status: 'active', createdAt: '2024-04-02', lastRun: '2024-04-14 09:30' },
  { id: 3, name: '华南区域探测', sourceIDC: 'IDC3-深圳', targetIDC: 'IDC5-广州', proxyCount: 4, targetCount: 12, status: 'inactive', createdAt: '2024-04-03', lastRun: '2024-04-13 22:00' },
  { id: 4, name: '全量探测策略', sourceIDC: 'IDC1-北京', targetIDC: 'IDC6-成都', proxyCount: 10, targetCount: 25, status: 'active', createdAt: '2024-03-28', lastRun: '2024-04-14 08:00' },
  { id: 5, name: '西部探测', sourceIDC: 'IDC6-成都', targetIDC: 'IDC7-西安', proxyCount: 2, targetCount: 5, status: 'pending', createdAt: '2024-04-10', lastRun: null }
])

const filteredStrategies = computed(() => {
  return strategies.value.filter(s => s.name.includes(searchKeyword.value))
})

function getStatusClass(status) {
  return status === 'active' ? 'online' : status === 'inactive' ? 'offline' : 'warning'
}

function getStatusLabel(status) {
  return status === 'active' ? '启用' : status === 'inactive' ? '禁用' : '待下发'
}

function generateTask(strategy) {
  alert(`正在为策略「${strategy.name}」生成任务...`)
}

function dispatchTask(strategy) {
  alert(`正在下发策略「${strategy.name}」...`)
}
</script>
