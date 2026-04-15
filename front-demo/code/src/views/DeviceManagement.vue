<template>
  <div>
    <div class="page-header animate-in">
      <div>
        <div class="page-title">设备管理</div>
        <div class="page-subtitle">管理全部主机设备 · CMDB 同步时间: {{ syncTime }}</div>
      </div>
      <div class="page-actions">
        <div class="search-box" style="width: 200px;">
          <span>🔍</span>
          <input type="text" v-model="searchKeyword" placeholder="搜索设备...">
        </div>
        <select v-model="roleFilter" class="btn btn-outline" style="padding: 9px 14px;">
          <option value="all">全部角色</option>
          <option value="agent">探测代理</option>
          <option value="target">探测目标</option>
          <option value="none">未配置</option>
        </select>
        <button class="btn btn-outline" @click="syncCMDB">↻ 同步CMDB</button>
        <button class="btn btn-primary" @click="batchSetRole">批量设置角色</button>
      </div>
    </div>

    <div class="card animate-in delay-1">
      <div class="card-body" style="padding:0;overflow-x:auto;">
        <table class="data-table">
          <thead>
            <tr>
              <th style="width: 40px;"><input type="checkbox" @change="toggleSelectAll"></th>
              <th>设备序列号</th>
              <th>IP地址</th>
              <th>所属机房</th>
              <th>网络区域</th>
              <th>Leaf</th>
              <th>机柜</th>
              <th>应用</th>
              <th>运维团队</th>
              <th>角色</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="device in filteredDevices" :key="device.serial">
              <td><input type="checkbox" v-model="device.selected"></td>
              <td class="node-name">{{ device.serial }}</td>
              <td>{{ device.ip }}</td>
              <td><span class="region-tag">{{ device.idc }}</span></td>
              <td>{{ device.networkZone }}</td>
              <td>{{ device.leaf }}</td>
              <td>{{ device.rack }}</td>
              <td>{{ device.app }}</td>
              <td>{{ device.opsTeam }}</td>
              <td>
                <select v-model="device.role" class="role-select" @change="updateRole(device)">
                  <option value="none">未配置</option>
                  <option value="agent">探测代理</option>
                  <option value="target">探测目标</option>
                </select>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="pagination" style="margin-top: 20px; display: flex; justify-content: space-between; align-items: center;">
      <span style="color: var(--text-muted); font-size: 13px;">共 {{ filteredDevices.length }} 台设备</span>
      <div style="display: flex; gap: 8px;">
        <button class="btn btn-outline" style="padding: 6px 12px;">上一页</button>
        <button class="btn btn-outline" style="padding: 6px 12px;">1 / 12</button>
        <button class="btn btn-outline" style="padding: 6px 12px;">下一页</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'

const searchKeyword = ref('')
const roleFilter = ref('all')
const syncTime = ref('2024-04-14 10:25:00')

const devices = ref([
  { serial: 'SN-BJ-001', ip: '10.1.1.10', idc: 'IDC1-北京', networkZone: '生产区', leaf: 'Leaf-01', rack: 'A-01', app: 'Web服务', opsTeam: '东部运维', role: 'agent', selected: false },
  { serial: 'SN-BJ-002', ip: '10.1.1.11', idc: 'IDC1-北京', networkZone: '生产区', leaf: 'Leaf-01', rack: 'A-01', app: 'API服务', opsTeam: '东部运维', role: 'target', selected: false },
  { serial: 'SN-SH-001', ip: '10.2.1.10', idc: 'IDC2-上海', networkZone: '生产区', leaf: 'Leaf-02', rack: 'B-03', app: '数据库', opsTeam: '华东运维', role: 'agent', selected: false },
  { serial: 'SN-SH-002', ip: '10.2.1.11', idc: 'IDC2-上海', networkZone: '测试区', leaf: 'Leaf-02', rack: 'B-03', app: '测试服务', opsTeam: '华东运维', role: 'none', selected: false },
  { serial: 'SN-SZ-001', ip: '10.3.1.10', idc: 'IDC3-深圳', networkZone: '生产区', leaf: 'Leaf-03', rack: 'C-05', app: '缓存服务', opsTeam: '华南运维', role: 'agent', selected: false },
  { serial: 'SN-HZ-001', ip: '10.4.1.10', idc: 'IDC4-杭州', networkZone: '生产区', leaf: 'Leaf-04', rack: 'D-02', app: '搜索服务', opsTeam: '阿里运维', role: 'target', selected: false },
  { serial: 'SN-GZ-001', ip: '10.5.1.10', idc: 'IDC5-广州', networkZone: '生产区', leaf: 'Leaf-05', rack: 'E-01', app: 'CDN节点', opsTeam: '华南运维', role: 'none', selected: false },
  { serial: 'SN-CD-001', ip: '10.6.1.10', idc: 'IDC6-成都', networkZone: '生产区', leaf: 'Leaf-06', rack: 'F-04', app: '视频服务', opsTeam: '西部运维', role: 'agent', selected: false }
])

const filteredDevices = computed(() => {
  return devices.value.filter(d => {
    const matchSearch = d.serial.includes(searchKeyword.value) || d.ip.includes(searchKeyword.value) || d.idc.includes(searchKeyword.value)
    const matchRole = roleFilter.value === 'all' || d.role === roleFilter.value
    return matchSearch && matchRole
  })
})

function toggleSelectAll(event) {
  filteredDevices.value.forEach(d => d.selected = event.target.checked)
}

function updateRole(device) {
  console.log('更新角色:', device)
}

function batchSetRole() {
  const selected = devices.value.filter(d => d.selected)
  if (selected.length === 0) {
    alert('请先选择设备')
    return
  }
  alert(`已选择 ${selected.length} 台设备，请选择要设置的角色`)
}

function syncCMDB() {
  syncTime.value = new Date().toLocaleString('zh-CN')
  alert('CMDB 同步完成')
}
</script>

<style scoped>
.role-select {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  padding: 4px 8px;
  font-size: 12px;
  color: var(--text-primary);
}
</style>
