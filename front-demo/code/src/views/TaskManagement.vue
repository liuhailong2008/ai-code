<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">任务管理</div>
        <div class="page-subtitle">管理全部探测任务 · 共 {{ tasks.length }} 个任务</div>
      </div>
      <div class="page-actions">
        <select v-model="statusFilter" class="btn btn-outline" style="padding: 9px 14px;">
          <option value="all">全部状态</option>
          <option value="pending">待下发</option>
          <option value="dispatched">已下发</option>
          <option value="running">执行中</option>
          <option value="completed">已完成</option>
          <option value="cleanup">待清理</option>
        </select>
        <button class="btn btn-outline">↻ 刷新</button>
      </div>
    </div>

    <div class="card animate-in delay-1">
      <div class="card-body" style="padding:0;overflow-x:auto;">
        <table class="data-table">
          <thead>
            <tr>
              <th>任务ID</th>
              <th>关联策略</th>
              <th>源机房</th>
              <th>目标机房</th>
              <th>节点数</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>开始时间</th>
              <th>完成时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in filteredTasks" :key="task.id">
              <td class="node-name">{{ task.id }}</td>
              <td>{{ task.strategy }}</td>
              <td><span class="region-tag">{{ task.sourceIDC }}</span></td>
              <td><span class="region-tag">{{ task.targetIDC }}</span></td>
              <td>{{ task.nodeCount }}</td>
              <td>
                <span class="status-dot" :class="getStatusClass(task.status)">
                  <span class="dot"></span>
                  {{ getStatusLabel(task.status) }}
                </span>
              </td>
              <td style="color:var(--text-muted)">{{ task.createdAt }}</td>
              <td style="color:var(--text-muted)">{{ task.startedAt || '--' }}</td>
              <td style="color:var(--text-muted)">{{ task.completedAt || '--' }}</td>
              <td>
                <button v-if="task.status === 'pending'" class="btn btn-outline" style="padding:4px 8px;font-size:11px;" @click="dispatch(task)">
                  下发
                </button>
                <button v-else-if="task.status === 'running'" class="btn btn-outline" style="padding:4px 8px;font-size:11px;" @click="viewProgress(task)">
                  查看进度
                </button>
                <button v-else class="btn btn-outline" style="padding:4px 8px;font-size:11px;" @click="viewDetail(task)">
                  详情
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="pagination" style="margin-top: 20px; display: flex; justify-content: flex-end; gap: 8px;">
      <button class="btn btn-outline" style="padding: 6px 12px;">上一页</button>
      <button class="btn btn-outline" style="padding: 6px 12px;">1 / 8</button>
      <button class="btn btn-outline" style="padding: 6px 12px;">下一页</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const statusFilter = ref('all')

const tasks = ref([
  { id: 'TASK-20240414-001', strategy: '北京-上海探测策略', sourceIDC: 'IDC1-北京', targetIDC: 'IDC2-上海', nodeCount: 15, status: 'running', createdAt: '10:30:00', startedAt: '10:30:05', completedAt: null },
  { id: 'TASK-20240414-002', strategy: '华东区域探测', sourceIDC: 'IDC2-上海', targetIDC: 'IDC4-杭州', nodeCount: 8, status: 'completed', createdAt: '09:30:00', startedAt: '09:30:02', completedAt: '09:35:28' },
  { id: 'TASK-20240414-003', strategy: '华南区域探测', sourceIDC: 'IDC3-深圳', targetIDC: 'IDC5-广州', nodeCount: 12, status: 'dispatched', createdAt: '09:00:00', startedAt: '09:00:05', completedAt: null },
  { id: 'TASK-20240413-018', strategy: '全量探测策略', sourceIDC: 'IDC1-北京', targetIDC: 'IDC6-成都', nodeCount: 35, status: 'completed', createdAt: '08:00:00', startedAt: '08:00:03', completedAt: '08:25:45' },
  { id: 'TASK-20240413-017', strategy: '西部探测', sourceIDC: 'IDC6-成都', targetIDC: 'IDC7-西安', nodeCount: 5, status: 'cleanup', createdAt: '22:00:00', startedAt: '22:00:02', completedAt: '22:08:15' },
  { id: 'TASK-20240414-004', strategy: '北京-上海探测策略', sourceIDC: 'IDC1-北京', targetIDC: 'IDC2-上海', nodeCount: 15, status: 'pending', createdAt: '10:35:00', startedAt: null, completedAt: null }
])

const filteredTasks = computed(() => {
  if (statusFilter.value === 'all') return tasks.value
  return tasks.value.filter(t => t.status === statusFilter.value)
})

function getStatusClass(status) {
  return {
    'pending': 'offline',
    'dispatched': 'warning',
    'running': 'online',
    'completed': 'online',
    'cleanup': 'warning'
  }[status] || 'offline'
}

function getStatusLabel(status) {
  return {
    'pending': '待下发',
    'dispatched': '已下发',
    'running': '执行中',
    'completed': '已完成',
    'cleanup': '待清理'
  }[status] || status
}

function dispatch(task) {
  alert(`下发任务 ${task.id}`)
}

function viewProgress(task) {
  alert(`查看任务 ${task.id} 进度`)
}

function viewDetail(task) {
  alert(`查看任务 ${task.id} 详情`)
}
</script>
